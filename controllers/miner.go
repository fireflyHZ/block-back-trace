package controllers

import (
	"fmt"
	"github.com/beego/beego/v2/server/web"
	logging "github.com/ipfs/go-log/v2"
	"profit-allocation/lotus/reward"
)

type MinerController struct {
	web.Controller
}

type MinerBalanceResp struct {
	Code         string
	Msg          string
	MinerBalacne map[string][]*reward.MinerBalance
}

var minerLog = logging.Logger("miner-ctr-log")

func (c *MinerController) GetMinerBalance() {
	minerId := c.GetString("miner")
	if minerId == "" {
		resp := MinerBalanceResp{
			Code: "faile",
			Msg:  "wallet id is nil",
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	}
	minerLog.Infof("Query miner balance :%+v", minerId)
	result, err := reward.QueryMinerAddressBalance(minerId)
	if err != nil {
		resp := MinerBalanceResp{
			Code: "faile",
			Msg:  fmt.Sprintf("Query miner address balance error:%v", err),
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	}
	resp := MinerBalanceResp{
		Code:         "ok",
		Msg:          "success",
		MinerBalacne: result,
	}
	c.Data["json"] = &resp
	c.ServeJSON()
	return
}
