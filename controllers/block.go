package controllers

import (
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	logging "github.com/ipfs/go-log/v2"
	"profit-allocation/lotus/block"
	"profit-allocation/models"
	"time"
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

func (c *BlockController) GetMinersLuck() {
	resp := new(models.GetMinersLuckResp)
	c.Data["json"] = &resp
	from, err := c.GetFloat("from")
	if err != nil {
		resp.Code = "failed"
		resp.Msg = fmt.Sprintf("Get from power error : %+v", err)
		c.ServeJSON()
		return
	}

	to, err := c.GetFloat("to")
	if err != nil {
		resp.Code = "failed"
		resp.Msg = fmt.Sprintf("Get to power error : %+v", err)
		c.ServeJSON()
		return
	}

	days, err := c.GetInt("days")
	if err != nil {
		resp.Code = "failed"
		resp.Msg = fmt.Sprintf("Get days error : %+v", err)
		c.ServeJSON()
		return
	}

	miners := make([]models.AllMinersMined, 0)
	o := orm.NewOrm()
	t := time.Now().Add(-time.Hour * 24 * time.Duration(days))
	num, err := o.QueryTable("fly_all_miners_mined").Filter("power__gte", from).Filter("power__lte", to).Filter("time__gte", t).OrderBy("power").All(&miners)
	if num == 0 {
		resp.Code = "failed"
		resp.Msg = fmt.Sprintf("Get miners mined blocks info error : %+v", err)
		c.ServeJSON()
		return
	}

	total, err := o.QueryTable("fly_all_miners_mined").Filter("time__gte", t).Count()
	if total == 0 {
		resp.Code = "failed"
		resp.Msg = fmt.Sprintf("Get miners total mined blocks number error : %+v", err)
		c.ServeJSON()
		return
	}
	totalBlockNum := float64(total)
	record := make(map[string][]models.AllMinersMined)
	//对miner进行分组
	for _, miner := range miners {
		if _, ok := record[miner.MinerId]; ok {
			record[miner.MinerId] = append(record[miner.MinerId], miner)
		} else {
			tmp := make([]models.AllMinersMined, 0)
			tmp = append(tmp, miner)
			record[miner.MinerId] = tmp
		}
	}
	//对miner数据进行处理
	minersLuck := make([]models.MinerLuck, 0)
	for _, infos := range record {
		power := (infos[0].Power + infos[len(infos)-1].Power) / 2
		totalPower := (infos[0].TotalPower + infos[len(infos)-1].TotalPower) / 2
		powerPercent := power / float64(totalPower)
		theoBlockNum := powerPercent * totalBlockNum
		actBlockNum := len(infos)
		luckyValue := float64(actBlockNum) / theoBlockNum
		totalReward := 0.0
		for _, info := range infos {
			totalReward += info.Reward
		}
		minerLuck := models.MinerLuck{
			Miner:       infos[0].MinerId,
			Luck:        fmt.Sprintf("%.2f%%", luckyValue*100),
			BlockNumber: actBlockNum,
			TotalValue:  totalReward,
		}
		minersLuck = append(minersLuck, minerLuck)
	}
	resp.Code = "success"
	resp.Msg = "ok"
	resp.MinersLuck = minersLuck
	c.ServeJSON()
}
