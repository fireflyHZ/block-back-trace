package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"profit-allocation/models"
	"profit-allocation/tool/data"
	"profit-allocation/tool/log"
	"strconv"
)

type RewardTmpController struct {
	beego.Controller
}

func (c *RewardTmpController) GetRewardAndPledge() {
	time := c.GetString("time")

	if time == "" {
		resp := data.RewardRespTmp{
			Code:   "faile",
			Msg:    "time  is nil",
			Reward: 0.0,
			Pledge: 0.0,
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	}

	rewardInfo := make([]models.RewardInfoTmp, 0)
	o := orm.NewOrm()

	num, err := o.QueryTable("fly_reward_info_tmp").Filter("time", time).All(&rewardInfo)
	log.Logger.Debug("DEBUG: QueryRewardInfo() reward: %+v ", rewardInfo)
	if err != nil || num == 0 {
		resp := data.RewardRespTmp{
			Code:   "faile",
			Msg:    "haven't this miner reward info",
			Reward: 0.0,
			Pledge: 0.0,
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	} else {

		var reward float64
		var pledge float64
		for _, info := range rewardInfo {
			r, _ := strconv.ParseFloat(info.Value, 64)
			reward += r
			pledge += info.Pledge
		}
		resp := data.RewardRespTmp{
			Code:   "ok",
			Msg:    "successful",
			Reward: reward,
			Pledge: pledge,
		}
		c.Data["json"] = &resp
		c.ServeJSON()
	}
	return
}
