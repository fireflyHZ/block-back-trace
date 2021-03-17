package controllers

import (
	"fmt"
	"github.com/beego/beego/v2/server/web"
	logging "github.com/ipfs/go-log/v2"
	"profit-allocation/lotus/block"
	"profit-allocation/models"
)

var blockLog = logging.Logger("block-ctr-log")

type BlockController struct {
	web.Controller
}

func (c *BlockController) GetMinerMineBlockPercentage() {
	start := c.GetString("start")
	end := c.GetString("end")
	miner := c.GetString("miner")

	resp := new(models.GetBlockPercentageResp)
	percentage, missed, mined, err := block.GetMinerMineBlockPercentage(end, start, miner)
	if err != nil {
		resp.Code = "failed"
		resp.Msg = fmt.Sprintf("Get miner mine block percentage error : %+v", err)
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	}
	resp.Code = "success"
	resp.Msg = "Get miner mine block percentage success"
	resp.MinedPercentage = fmt.Sprintf("%.2f%%", percentage*100)
	resp.Mined = mined
	resp.Missed = missed
	c.Data["json"] = &resp
	c.ServeJSON()
	return
}
