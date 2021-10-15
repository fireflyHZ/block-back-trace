package power

import (
	"bytes"
	"encoding/json"
	"fmt"
	logging "github.com/ipfs/go-log/v2"
	"io/ioutil"
	"net/http"
	"profit-allocation/models"
	"profit-allocation/util/dingTalk"
	"strconv"
	"strings"
	"time"
)

var log = logging.Logger("reward-log")

func PartitionCheck() {
	for {
		//获取所有miner
		miners := ""
		for minerId, _ := range models.Miners {
			miners += fmt.Sprintf(",\"%s\"", minerId)
		}
		miners = strings.TrimLeft(miners, ",")

		checkMinersPartitions(miners)
		time.Sleep(time.Minute * 3)
	}
}

func checkMinersPartitions(miners string) {
	client := http.Client{}
	requestData := fmt.Sprintf(`{"jsonrpc": "2.0", "method": "Filecoin.StateMinerDeadlineReport", "params": [[%s],null], "id": 3}`, miners)

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

		if min*60+sec > 60*3*(1+info.AllPartitions/2) {
			if info.AllPartitions > info.ProvenPartitions {
				err = dingTalk.SendPowerDingtalkData(&info)
				if err != nil {
					log.Errorf("SendPowerDingtalkData call error:%+v", err)
				}
			}
		}
	}
	return
}
