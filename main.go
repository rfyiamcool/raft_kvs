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
	"flag"
	"strings"

	"github.com/rfyiamcool/raft_kvs/consensus/raft/raftpb"
)

var (
	clusterList = flag.String("cluster", "http://127.0.0.1:9021", "comma separated cluster peers")
	id          = flag.Int("id", 1, "node ID")
	kvport      = flag.Int("port", 9121, "key-value server port")
	join        = flag.Bool("join", false, "join an existing cluster")

	kvs *kvstore
)

func main() {
	flag.Parse()

	// 把logEntries传递给raft状态机
	proposeC := make(chan string)
	defer close(proposeC)

	// 告知raft节点的变更
	confChangeC := make(chan raftpb.ConfChange)
	defer close(confChangeC)

	// raft provides a commit stream for the proposals from the http api
	getSnapshot := func() ([]byte, error) { return kvs.getSnapshot() }

	// 创建raft节点, proposeC/confChangeC的接收端
	rc, commitC, errorC, snapshotterReady := newRaftNode(
		*id,
		strings.Split(*clusterList, ","),
		*join,
		getSnapshot,
		proposeC,
		confChangeC,
	)

	// proposeC 发送端 commitC接收端
	kvs = newKVStore(<-snapshotterReady, proposeC, commitC, errorC)
	kvs.start()

	// 启动动http api服务器,处理发送到的raft请求
	serveHttpKVServer(kvs, *kvport, rc, confChangeC, errorC)
}
