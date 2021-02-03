package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/firefly/profit/total_reward_info", totalRewardInfo)
	http.HandleFunc("/firefly/profit/total_miner_info", totalMinerInfo)
	err := http.ListenAndServe("172.16.0.7:20011", nil)
	if err != nil {
		log.Printf("listen 172.16.0.7:20011 error:%+v\n", err)
	}
}
func totalRewardInfo(writer http.ResponseWriter, request *http.Request) {
	value := request.URL.Query()
	minePool := value.Get("mp")
	t := value.Get("time")

	switch minePool {
	case "f02420":
		url := fmt.Sprintf("http://172.16.0.7:40001/firefly/profit/total_reward_info?time=%s", t)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f044315":
		url := fmt.Sprintf("http://172.16.0.7:40002/firefly/profit/total_reward_info?time=%s", t)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f055446":
		url := fmt.Sprintf("http://172.16.0.7:40003/firefly/profit/total_reward_info?time=%s", t)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f099132":
		url := fmt.Sprintf("http://172.16.0.7:40005/firefly/profit/total_reward_info?time=%s", t)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f088290":
		url := fmt.Sprintf("http://172.16.0.7:40004/firefly/profit/total_reward_info?time=%s", t)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f0104398":
		url := fmt.Sprintf("http://172.16.0.7:40006/firefly/profit/total_reward_info?time=%s", t)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f0117450":
		url := fmt.Sprintf("http://172.16.0.7:40007/firefly/profit/total_reward_info?time=%s", t)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f0122533":
		url := fmt.Sprintf("http://172.16.0.7:40008/firefly/profit/total_reward_info?time=%s", t)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f0130686":
		url := fmt.Sprintf("http://172.16.0.7:40009/firefly/profit/total_reward_info?time=%s", t)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f0144528":
		url := fmt.Sprintf("http://172.16.0.7:50010/firefly/profit/total_reward_info?time=%s", t)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f0148452":
		url := fmt.Sprintf("http://172.16.0.7:50011/firefly/profit/total_reward_info?time=%s", t)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	default:
		writer.Write([]byte("mine pool not match"))
	}
}
func totalMinerInfo(writer http.ResponseWriter, request *http.Request) {

	value := request.URL.Query()
	minePool := value.Get("mp")
	t := value.Get("time")
	miner := value.Get("miner")

	switch minePool {
	case "f02420":
		url := fmt.Sprintf("http://172.16.0.7:40001/firefly/profit/total_miner_info?time=%s&miner=%s", t, miner)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f044315":
		url := fmt.Sprintf("http://172.16.0.7:40002/firefly/profit/total_miner_info?time=%s&miner=%s", t, miner)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f055446":
		url := fmt.Sprintf("http://172.16.0.7:40003/firefly/profit/total_miner_info?time=%s&miner=%s", t, miner)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f099132":
		url := fmt.Sprintf("http://172.16.0.7:40005/firefly/profit/total_miner_info?time=%s&miner=%s", t, miner)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f088290":
		url := fmt.Sprintf("http://172.16.0.7:40004/firefly/profit/total_miner_info?time=%s&miner=%s", t, miner)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f0104398":
		url := fmt.Sprintf("http://172.16.0.7:40006/firefly/profit/total_miner_info?time=%s&miner=%s", t, miner)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f0117450":
		url := fmt.Sprintf("http://172.16.0.7:40007/firefly/profit/total_miner_info?time=%s&miner=%s", t, miner)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f0122533":
		url := fmt.Sprintf("http://172.16.0.7:40008/firefly/profit/total_miner_info?time=%s&miner=%s", t, miner)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f0130686":
		url := fmt.Sprintf("http://172.16.0.7:40009/firefly/profit/total_miner_info?time=%s&miner=%s", t, miner)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f0144528":
		url := fmt.Sprintf("http://172.16.0.7:50010/firefly/profit/total_miner_info?time=%s&miner=%s", t, miner)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f0148452":
		url := fmt.Sprintf("http://172.16.0.7:50011/firefly/profit/total_miner_info?time=%s&miner=%s", t, miner)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	default:
		writer.Write([]byte("mine pool not match"))
	}
}
