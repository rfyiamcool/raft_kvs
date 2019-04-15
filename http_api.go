// Copyright 2015 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"go.etcd.io/etcd/raft/raftpb"
)

// Handler for a http based key-value store backed by raft
type httpKVAPI struct {
	store       *kvstore
	raftNode    *raftNode
	confChangeC chan<- raftpb.ConfChange
}

func makeServerId() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%d", rand.Int63())
}

func (h *httpKVAPI) handlePut(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	var (
		key   = r.PostForm.Get("key")
		value = r.PostForm.Get("value")
	)

	h.store.Propose(key, value)

	w.WriteHeader(http.StatusNoContent)
}

func (h *httpKVAPI) handleBatchPut(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	var (
		countStr      = r.PostForm.Get("count")
		count         = 1
		concurrent    = 1
		concurrentStr = r.PostForm.Get("concurrent")
		err           error
	)

	// if !h.raftNode.isLeader() {
	// 	http.Error(w, "the node is follower, not leader", http.StatusBadRequest)
	// 	return
	// }

	count, err = strconv.Atoi(countStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("count conv error"))
		return
	}

	concurrent, err = strconv.Atoi(concurrentStr)
	if err != nil {
		concurrent = 1
	}

	var (
		start = time.Now()
		wg    sync.WaitGroup
	)

	var incr int64 = 0
	for index := 0; index < concurrent; index++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for {
				gid := atomic.AddInt64(&incr, 1)
				if gid > int64(count) {
					break
				}

				idStr := fmt.Sprintf("%v", gid)
				h.store.Propose(idStr, start.String())
			}
		}()
	}

	wg.Wait()

	take := time.Now().Sub(start)
	resp := fmt.Sprintf("total count: %d, thread: %v, time cost: %v",
		count,
		concurrent,
		take.String(),
	)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(resp))
}

func (h *httpKVAPI) handleGet(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	// key := r.URL.Query()["key"]
	key := r.Form.Get("key")

	if v, ok := h.store.Lookup(key); ok {
		w.Write([]byte(v))
		return
	}

	w.Write([]byte("null"))
}

func (h *httpKVAPI) handleJoin(w http.ResponseWriter, r *http.Request) {
	var key = r.RequestURI
	defer r.Body.Close()

	url, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read on POST (%v)\n", err)
		http.Error(w, "Failed on POST", http.StatusBadRequest)
		return
	}

	nodeId, err := strconv.ParseUint(key[1:], 0, 64)
	if err != nil {
		log.Printf("Failed to convert ID for conf change (%v)\n", err)
		http.Error(w, "Failed on POST", http.StatusBadRequest)
		return
	}

	cc := raftpb.ConfChange{
		Type:    raftpb.ConfChangeAddNode,
		NodeID:  nodeId,
		Context: url,
	}
	h.confChangeC <- cc

	// As above, optimistic that raft will apply the conf change
	w.WriteHeader(http.StatusNoContent)
}

func (h *httpKVAPI) handleRaftInfo(w http.ResponseWriter, r *http.Request) {
	info := h.raftNode.getSelfState()
	js, err := json.Marshal(info)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Write(js)
	w.WriteHeader(http.StatusNoContent)
}

func (h *httpKVAPI) handleLeader(w http.ResponseWriter, r *http.Request) {
	info := h.raftNode.getSelfState()
	state, ok := info["self_state"]
	if !ok {
		http.Error(w, "not get self_state", http.StatusBadRequest)
		return
	}

	w.Write([]byte(
		state.(string),
	))
	w.WriteHeader(http.StatusNoContent)
}

func (h *httpKVAPI) handleLeave(w http.ResponseWriter, r *http.Request) {
	var key = r.RequestURI
	defer r.Body.Close()

	nodeId, err := strconv.ParseUint(key[1:], 0, 64)
	if err != nil {
		log.Printf("Failed to convert ID for conf change (%v)\n", err)
		http.Error(w, "Failed on DELETE", http.StatusBadRequest)
		return
	}

	cc := raftpb.ConfChange{
		Type:   raftpb.ConfChangeRemoveNode,
		NodeID: nodeId,
	}
	h.confChangeC <- cc

	// As above, optimistic that raft will apply the conf change
	w.WriteHeader(http.StatusNoContent)
}

// serveHttpKVAPI starts a key-value server with a GET/PUT API and listens.
func serveHttpKVServer(kv *kvstore, port int, rc *raftNode, confChangeC chan<- raftpb.ConfChange, errorC <-chan error) {
	var (
		srv = http.Server{
			Addr: ":" + strconv.Itoa(port),
		}

		callHandler = &httpKVAPI{
			store:       kv,
			confChangeC: confChangeC,
			raftNode:    rc,
		}
	)

	http.HandleFunc("/batch_put", callHandler.handleBatchPut)
	http.HandleFunc("/put", callHandler.handlePut)
	http.HandleFunc("/get", callHandler.handleGet)
	http.HandleFunc("/leader", callHandler.handleLeader)
	http.HandleFunc("/info", callHandler.handleRaftInfo)
	http.HandleFunc("/join", callHandler.handleJoin)
	http.HandleFunc("/leave", callHandler.handleLeave)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	// exit when raft goes down
	if err, ok := <-errorC; ok {
		log.Fatal(err)
	}
}
