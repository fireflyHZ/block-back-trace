package main

import (
	"bytes"
	"fmt"
	logging "github.com/ipfs/go-log/v2"
	"io/ioutil"
	"log"
	"net/http"
)

var forwardLog = logging.Logger("forwardLog")

func main() {
	err := initlog()
	if err != nil {
		fmt.Println("init log", err)
	}
	http.HandleFunc("/firefly/profit/total_reward_info", totalRewardInfo)
	http.HandleFunc("/firefly/profit/total_miner_info", totalMinerInfo)
	err = http.ListenAndServe("172.16.0.7:20011", nil)
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
		forwardLog.Info("f02420 total reward")
		url := fmt.Sprintf("http://172.16.0.7:40001/firefly/profit/total_reward_info?time=%s", t)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f044315":
		forwardLog.Info("f044315 total reward")
		url := fmt.Sprintf("http://172.16.0.7:40002/firefly/profit/total_reward_info?time=%s", t)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f055446":
		forwardLog.Info("f055446 total reward")
		url := fmt.Sprintf("http://172.16.0.7:40003/firefly/profit/total_reward_info?time=%s", t)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f099132":
		forwardLog.Info("f099132 total reward")
		url := fmt.Sprintf("http://172.16.0.7:40005/firefly/profit/total_reward_info?time=%s", t)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f088290":
		forwardLog.Info("f088290 total reward")
		url := fmt.Sprintf("http://172.16.0.7:40004/firefly/profit/total_reward_info?time=%s", t)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f0104398":
		forwardLog.Info("f0104398 total reward")
		url := fmt.Sprintf("http://172.16.0.7:40006/firefly/profit/total_reward_info?time=%s", t)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f0117450":
		forwardLog.Info("f0117450 total reward")
		url := fmt.Sprintf("http://172.16.0.7:40007/firefly/profit/total_reward_info?time=%s", t)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f0122533":
		forwardLog.Info("f0122533 total reward")
		url := fmt.Sprintf("http://172.16.0.7:40008/firefly/profit/total_reward_info?time=%s", t)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f0130686":
		forwardLog.Info("f0130686 total reward")
		url := fmt.Sprintf("http://172.16.0.7:40009/firefly/profit/total_reward_info?time=%s", t)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f0144528":
		forwardLog.Info("f0144528 total reward")
		url := fmt.Sprintf("http://172.16.0.7:50010/firefly/profit/total_reward_info?time=%s", t)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f0144530":
		forwardLog.Info("f0144530 total reward")
		url := fmt.Sprintf("http://172.16.0.7:50011/firefly/profit/total_reward_info?time=%s", t)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f0148452":
		forwardLog.Info("f0148452 total reward")
		url := fmt.Sprintf("http://172.16.0.7:50012/firefly/profit/total_reward_info?time=%s", t)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f0161819":
		forwardLog.Info("f0161819 total reward")
		url := fmt.Sprintf("http://172.16.0.7:50013/firefly/profit/total_reward_info?time=%s", t)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	default:
		forwardLog.Info("default total reward")
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
		forwardLog.Info("f02420 miner reward")
		url := fmt.Sprintf("http://172.16.0.7:40001/firefly/profit/total_miner_info?time=%s&miner=%s", t, miner)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f044315":
		forwardLog.Info("f044315 miner reward")
		url := fmt.Sprintf("http://172.16.0.7:40002/firefly/profit/total_miner_info?time=%s&miner=%s", t, miner)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f055446":
		forwardLog.Info("f055446 miner reward")
		url := fmt.Sprintf("http://172.16.0.7:40003/firefly/profit/total_miner_info?time=%s&miner=%s", t, miner)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f099132":
		forwardLog.Info("f099132 miner reward")
		url := fmt.Sprintf("http://172.16.0.7:40005/firefly/profit/total_miner_info?time=%s&miner=%s", t, miner)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f088290":
		forwardLog.Info("f088290 miner reward")
		url := fmt.Sprintf("http://172.16.0.7:40004/firefly/profit/total_miner_info?time=%s&miner=%s", t, miner)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f0104398":
		forwardLog.Info("f0104398 miner reward")
		url := fmt.Sprintf("http://172.16.0.7:40006/firefly/profit/total_miner_info?time=%s&miner=%s", t, miner)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f0117450":
		forwardLog.Info("f0117450 miner reward")
		url := fmt.Sprintf("http://172.16.0.7:40007/firefly/profit/total_miner_info?time=%s&miner=%s", t, miner)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f0122533":
		forwardLog.Info("f0122533 miner reward")
		url := fmt.Sprintf("http://172.16.0.7:40008/firefly/profit/total_miner_info?time=%s&miner=%s", t, miner)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f0130686":
		forwardLog.Info("f0130686 miner reward")
		url := fmt.Sprintf("http://172.16.0.7:40009/firefly/profit/total_miner_info?time=%s&miner=%s", t, miner)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f0144528":
		forwardLog.Info("f0144528 miner reward")
		url := fmt.Sprintf("http://172.16.0.7:50010/firefly/profit/total_miner_info?time=%s&miner=%s", t, miner)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f0144530":
		forwardLog.Info("f0144530 miner reward")
		url := fmt.Sprintf("http://172.16.0.7:50011/firefly/profit/total_miner_info?time=%s&miner=%s", t, miner)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f0148452":
		forwardLog.Info("f0148452 miner reward")
		url := fmt.Sprintf("http://172.16.0.7:50012/firefly/profit/total_miner_info?time=%s&miner=%s", t, miner)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	case "f0161819":
		forwardLog.Info("f0161819 miner reward")
		url := fmt.Sprintf("http://172.16.0.7:50013/firefly/profit/total_miner_info?time=%s&miner=%s", t, miner)
		r, _ := http.NewRequest("GET", url, bytes.NewReader([]byte{}))
		rsp, _ := http.DefaultClient.Do(r)
		rspData, _ := ioutil.ReadAll(rsp.Body)
		writer.Write(rspData)
	default:
		forwardLog.Info("default miner reward")
		writer.Write([]byte("mine pool not match"))
	}
}
