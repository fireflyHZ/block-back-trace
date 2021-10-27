package power

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/chain/types"
	logging "github.com/ipfs/go-log/v2"
	"io/ioutil"
	"net/http"
	client2 "profit-allocation/lotus/client"
	"profit-allocation/models"
	"profit-allocation/util/dingTalk"
	"strconv"
	"strings"
	"time"
)

var log = logging.Logger("reward-log")

func PartitionCheck() {
	//doOneceTest()
	for {
		//获取所有miner
		miners := ""
		for minerId, _ := range models.Miners {
			miners += fmt.Sprintf(",\"%s\"", minerId)
		}
		miners = strings.TrimLeft(miners, ",")
		checkMinersPartitions(miners)
		log.Info("checked")
		time.Sleep(time.Minute * 3)
	}
}

func checkMinersPartitions(miners string) {

	client := http.Client{}
	requestData := fmt.Sprintf(`{"jsonrpc": "2.0", "method": "Filecoin.StateMinerDeadlineReport", "params": [[%s],null], "id": 3}`, miners)
	//fmt.Println(requestData,models.LotusHost)
	request, err := http.NewRequest("POST", models.LotusHost, bytes.NewBuffer([]byte(requestData)))
	if err != nil {
		err = dingTalk.SendCheckPowerErrorDingtalkData(fmt.Sprintf("http new request error:%+v", err))
		if err != nil {
			log.Errorf("SendCheckPowerErrorDingtalkData call error:%+v", err)
		}
		return
	}

	request.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(request)
	if err != nil {
		err = dingTalk.SendCheckPowerErrorDingtalkData(fmt.Sprintf("http do request error:%+v", err))
		if err != nil {
			log.Errorf("SendCheckPowerErrorDingtalkData call error:%+v", err)
		}
		return
	}
	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = dingTalk.SendCheckPowerErrorDingtalkData(fmt.Sprintf("read check reaponse error:%+v", err))
		if err != nil {
			log.Errorf("SendCheckPowerErrorDingtalkData call error:%+v", err)
		}
		return
	}
	defer resp.Body.Close()
	//fmt.Println(string(responseData))
	cpr := new(models.CheckPartitionsResp)

	err = json.Unmarshal(responseData, cpr)
	if err != nil {
		err = dingTalk.SendCheckPowerErrorDingtalkData(fmt.Sprintf("unmarshal check reaponse error:%+v", err))
		if err != nil {
			log.Errorf("SendCheckPowerErrorDingtalkData call error:%+v", err)
		}
		return
	}

	for _, info := range cpr.Result {

		elapsed := strings.Split(info.DeadlineElapsedTime, ":")
		min, err := strconv.Atoi(elapsed[0])
		if err != nil {
			err = dingTalk.SendCheckPowerErrorDingtalkData(fmt.Sprintf("get elapsed minute error:%+v miner:%+v", err, info.Maddr))
			if err != nil {
				log.Errorf("SendCheckPowerErrorDingtalkData call error:%+v", err)
			}
			continue
		}
		sec, err := strconv.Atoi(elapsed[1])
		if err != nil {
			err = dingTalk.SendCheckPowerErrorDingtalkData(fmt.Sprintf("get elapsed second error:%+v miner:%+v", err, info.Maddr))
			if err != nil {
				log.Errorf("SendCheckPowerErrorDingtalkData call error:%+v", err)
			}
			continue
		}

		if min*60+sec > 210*info.AllPartitions {
			if info.AllPartitions > info.ProvenPartitions {
				log.Warnf("window post warning,miner:%+v,deadline:%+v", info.Maddr, info.DeadlineIndex)
				err = dingTalk.SendPowerDingtalkData(&info)
				if err != nil {
					log.Errorf("SendPowerDingtalkData call error:%+v", err)
				}
			}
		}
	}
	return
}

func doOneceTest() {
	ctx := context.Background()
	miners := ""
	for minerId, _ := range models.Miners {
		miners += fmt.Sprintf(",\"%s\"", minerId)
	}
	miners = strings.TrimLeft(miners, ",")
	for i := 1200240; i < 1200241; i++ {
		tip, err := client2.Client.ChainGetTipSetByHeight(ctx, abi.ChainEpoch(i), types.EmptyTSK)
		if err != nil {
			fmt.Println(err)
			return
		}
		k := tip.Key()
		data, err := json.Marshal(&k)
		if err != nil {
			fmt.Println("marshal error", err)
			return
		}
		client := http.Client{}
		requestData := fmt.Sprintf(`{"jsonrpc": "2.0", "method": "Filecoin.StateMinerDeadlineReport", "params": [[%s],%s], "id": 3}`, miners, string(data))
		fmt.Println(requestData)
		request, err := http.NewRequest("POST", models.LotusHost, bytes.NewBuffer([]byte(requestData)))
		if err != nil {
			return
		}

		request.Header.Add("Content-Type", "application/json")
		resp, err := client.Do(request)
		if err != nil {
			return
		}
		responseData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return
		}
		//fmt.Println(string(responseData))
		cpr := new(models.CheckPartitionsResp)

		err = json.Unmarshal(responseData, cpr)
		if err != nil {

			return
		}

		for _, info := range cpr.Result {

			elapsed := strings.Split(info.DeadlineElapsedTime, ":")
			min, err := strconv.Atoi(elapsed[0])
			if err != nil {
				continue
			}
			sec, err := strconv.Atoi(elapsed[1])
			if err != nil {
				continue
			}
			fmt.Println(min, sec)
			if min*60+sec > 60*3*(1+info.AllPartitions/2) {
				if info.Maddr == "f044315" {
					fmt.Println(min*60 + sec)
					fmt.Println(60 * 3 * (1 + info.AllPartitions/2))
					fmt.Println(info.AllPartitions)
					fmt.Println(info.ProvenPartitions)
				}
				if info.AllPartitions > info.ProvenPartitions {
					fmt.Println("=========baojing")
					err := dingTalk.SendTestPowerDingtalkData(&info)
					fmt.Println(err)
				}
			}
		}
	}
}
