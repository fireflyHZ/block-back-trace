package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"profit-allocation/models"
	"profit-allocation/tool/data"
	"profit-allocation/tool/log"
)

type RewardController struct {
	beego.Controller
}



func (c *RewardController) QueryRewardInfo() {
	time := c.GetString("time")
	minerId := c.GetString("miner_id")
	if time == "" {
		resp := data.RewardResp{
			Code:          "faile",
			Msg:           "time  is nil",
			Reward: "0.0",
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	}
	if minerId == "" {
		resp := data.RewardResp{
			Code:          "faile",
			Msg:           "miner_id  is nil",
			Reward: "0.0",
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	}
	log.Logger.Debug("DEBUG: QueryRewardInfo() time: %+v -- minerid: %+v", time,minerId)
	rewardInfo := make([]models.RewardInfo, 0)
	o := orm.NewOrm()

	num, err := o.QueryTable("fly_reward_info").Filter("time", time).Filter("miner_id",minerId).All(&rewardInfo)
	log.Logger.Debug("DEBUG: QueryRewardInfo() reward: %+v ", rewardInfo)
	if err != nil || num == 0 {
		resp := data.RewardResp{
			Code:          "faile",
			Msg: "haven't this miner reward info",
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	} else {
		resp := data.RewardResp{
			Code:          "ok",
			Msg:           "successful",
			Reward: rewardInfo[0].Reward,
		}
		c.Data["json"] = &resp
		c.ServeJSON()
	}
	return
}

func (c *RewardController) QueryOrderDailyReward()()  {
	time := c.GetString("time")

	if time == "" {
		resp := data.OrderDailyRewardResp{
			Code:          "faile",
			Msg:           "time  is nil",
			Data: nil,
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	}
	log.Logger.Debug("DEBUG: QueryOrderDailyReward() time: %+v ", time)
	orderInfo := make([]models.OrderDailyRewardInfo, 0)
	o := orm.NewOrm()

	num, err := o.QueryTable("fly_order_daily_reward_info").Filter("time", time).All(&orderInfo)
	//log.Logger.Debug("DEBUG: QueryRewardInfo() reward: %+v ", orderInfo)
	if err != nil || num == 0 {
		resp := data.OrderDailyRewardResp{
			Code:          "faile",
			Msg:           "time  is nil",
			Data: nil,
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	} else {
		resp := data.OrderDailyRewardResp{
			Code:          "ok",
			Msg:           "successful",
			Data: orderInfo,
		}
		c.Data["json"] = &resp
		c.ServeJSON()
	}
	return
}