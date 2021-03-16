package controllers

import (
	"fmt"
	"github.com/beego/beego/v2/server/web"
	logging "github.com/ipfs/go-log/v2"
	"profit-allocation/lotus/block"
)

var blockLog = logging.Logger("block-ctr-log")

type BlockController struct {
	web.Controller
}

func (c *BlockController) GetMinerMineBlockPercentage() {
	start := c.GetString("start")
	end := c.GetString("end")
	miner := c.GetString("miner")
	fmt.Println(start, end, miner)
	percentage, err := block.GetMinerMineBlockPercentage(start, end, miner)
	if err != nil {
		c.Ctx.WriteString(fmt.Sprintf("%+v", err))
		return
	}
	c.Ctx.WriteString(fmt.Sprintf("%.2f", percentage))
	return
}
