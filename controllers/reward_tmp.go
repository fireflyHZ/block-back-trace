package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"profit-allocation/models"
	"profit-allocation/tool/bit"
	"profit-allocation/tool/data"
	"profit-allocation/tool/log"
	"strconv"
	"time"
)

type RewardTmpController struct {
	beego.Controller
}

func (c *RewardTmpController) GetRewardAndPledge() {
	t := c.GetString("time")

	if t == "" {
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

	num, err := o.QueryTable("fly_reward_info_tmp").Filter("time", t).All(&rewardInfo)
	log.Logger.Debug("DEBUG: QueryRewardInfo() reward: %+v ", rewardInfo)
	if err != nil || num == 0 {
		resp := data.RewardRespTmp{
			Code:   "faile",
			Msg:    "haven't this miner reward info",
			Reward: 0.0,
			Pledge: 0.0,
			Power:  0.0,
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	} else {

		var reward float64
		var pledge float64
		var power float64
		var timeStamp int64
		for _, info := range rewardInfo {
			if timeStamp<info.UpdateTime{
				timeStamp=info.UpdateTime
			}
			r, _ := strconv.ParseFloat(info.Value, 64)
			reward += r
			pledge += info.Pledge
			power += info.Power
		}

		resp := data.RewardRespTmp{
			Code:   "ok",
			Msg:    "successful",
			Reward: reward,
			Pledge: pledge,
			Power:  power,
			Update: time.Unix(timeStamp,0).Format("2006-01-02 15:04:05"),
		}
		c.Data["json"] = &resp
		c.ServeJSON()
	}
	return
}

func  (c *RewardTmpController) GetMessagesGas()  {
	t := c.GetString("time")

	if t == "" {
		resp := data.MessageGasTmp{
			Code:   "faile",
			Msg:    "time  is nil",
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	}

	expendInfo := make([]models.ExpendInfo, 0)
	o := orm.NewOrm()

	num, err := o.QueryTable("fly_expend_info").Filter("time", t).All(&expendInfo)
	log.Logger.Debug("DEBUG: QueryRewardInfo() reward: %+v ", expendInfo)
	if err != nil || num == 0 {
		resp := data.MessageGasTmp{
			Code:   "faile",
			Msg:    "haven't this miner reward info",
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	} else {
		gas:="0.0"
		for _, info := range expendInfo {
			gas=bit.CalculateReward(gas,info.Gas)
			gas=bit.CalculateReward(gas,info.BaseBurnFee)
			gas=bit.CalculateReward(gas,info.OverEstimationBurn)
		}
		resp := data.MessageGasTmp{
			Code:   "ok",
			Msg:    "successful",
			Gas: gas,
		}
		c.Data["json"] = &resp
		c.ServeJSON()
	}
	return
}
