package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	node1 = "http://127.0.0.1:11111"
	node2 = "http://127.0.0.1:22222"
	node3 = "http://127.0.0.1:33333"
)

var (
	nodes = map[string]string{
		"node1": node1,
		"node2": node2,
		"node3": node3,
	}

	cmd        = flag.String("cmd", "", "cmd")
	count      = flag.String("count", "10000", "cmd")
	concurrent = flag.String("concurrent", "1", "cmd")
)

func main() {
	flag.Parse()

	var (
		leaderNode = getLeaderNode()
		leaderApi  = nodes[leaderNode]
		value      = randStr(8)
	)

	fmt.Println("current leader node: ", leaderNode)

	switch *cmd {
	case "leader":

	case "batch":
		fmt.Println("post value: ", value)
		batchPut(leaderApi, value)

	case "check":
		checkDataOnce()

	case "all":
		fmt.Println("post value: ", value)
		batchPut(leaderApi, value)

		var (
			m     = map[string]string{}
			start = time.Now()
		)
		for {
			if !checkData(value, m) {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			end := time.Now()
			fmt.Println("take cost :", end.Sub(start))
			break
		}

	default:

	}
}

func getLeaderNode() string {
	var leader = ""
	for nodeName, nodeUri := range nodes {
		resp, err := httpGetRequest(nodeUri + "/leader")
		if err != nil {
			fmt.Println(err)
		}

		if resp == "StateLeader" {
			leader = nodeName
		}
	}

	if leader == "" {
		fmt.Println("not get leader node")
		os.Exit(99)
	}

	return leader
}

func batchPut(api, value string) {
	m := map[string]string{
		"count":      *count,
		"concurrent": *concurrent,
		"value":      value,
	}

	fmt.Println(httpPostRequest(fmt.Sprintf("%s/batch_put", api), m))
}

func checkDataOnce() {
	for nodeName, nodeUri := range nodes {
		resp, err := httpGetRequest(fmt.Sprintf("%s/get?key=%s", nodeUri, *count))
		fmt.Println(nodeName, resp, err)
	}
}

func checkData(value string, m map[string]string) bool {
	for nodeName, nodeUri := range nodes {
		// already done
		if _, ok := m[nodeName]; ok {
			continue
		}

		resp, err := httpGetRequest(fmt.Sprintf("%s/get?key=%s", nodeUri, *count))
		fmt.Println(nodeName, resp)
		if err != nil {
			panic(err.Error())
		}
		if resp != value {
			continue
		}

		m[nodeName] = ""
	}

	if len(m) == len(nodes) {
		return true
	}

	return false
}

func httpGetRequest(api string) (string, error) {
	req, err := http.NewRequest("GET", api, nil)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)
	return string(respBody), nil
}

func httpPostRequest(api string, params map[string]string) (string, error) {
	var (
		response *http.Response
		body     []byte
		err      error
	)

	args := url.Values{}
	for k, v := range params {
		args.Set(k, v)
	}

	client := &http.Client{}
	response, err = client.PostForm(api, args)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()
	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func randStr(strlen int) string {
	rand.Seed(time.Now().UnixNano())
	data := make([]byte, strlen)
	var num int
	for i := 0; i < strlen; i++ {
		num = rand.Intn(57) + 65
		for {
			if num > 90 && num < 97 {
				num = rand.Intn(57) + 65
			} else {
				break
			}
		}
		data[i] = byte(num)
	}
	return string(data)
}
