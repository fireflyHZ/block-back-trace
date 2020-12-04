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
	var reward, pledge, power, totalPower float64
	var timeStamp int64
	gas := "0.0"

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

	_, err := o.QueryTable("fly_reward_info_tmp").Filter("time", t).All(&rewardInfo)
	log.Logger.Debug("DEBUG: QueryRewardInfo() reward: %+v ", rewardInfo)
	if err != nil  {
		resp := data.RewardRespTmp{
			Code:       "fail",
			Msg:        "get reward info fail",
			Reward:     reward,
			Pledge:     pledge,
			Power:      power,
			Gas:        gas,
			TotalPower: totalPower,
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	} else {

		for _, info := range rewardInfo {
			if timeStamp < info.UpdateTime {
				timeStamp = info.UpdateTime
			}
			r, _ := strconv.ParseFloat(info.Value, 64)
			reward += r
			pledge += info.Pledge
			power += info.Power
		}

	}

	expendInfo := make([]models.ExpendInfo, 0)

	_, err = o.QueryTable("fly_expend_info").Filter("time", t).All(&expendInfo)
	log.Logger.Debug("DEBUG: QueryRewardInfo() reward: %+v ", expendInfo)
	if err != nil   {
		resp := data.RewardRespTmp{
			Code:       "fail",
			Msg:        "get expend info fail",
			Reward:     reward,
			Pledge:     pledge,
			Power:      power,
			Gas:        gas,
			TotalPower: totalPower,
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	} else{
		for _, info := range expendInfo {
			gas = bit.CalculateReward(gas, info.Gas)
			gas = bit.CalculateReward(gas, info.BaseBurnFee)
			gas = bit.CalculateReward(gas, info.OverEstimationBurn)
		}
	}
	//todo totalpower
	minerPowerStatus:=make([]models.MinerPowerStatus,0)
	_,err=o.QueryTable("fly_miner_power_status").Filter("time",t).All(&minerPowerStatus)
	if err != nil  {
		resp := data.RewardRespTmp{
			Code:       "fail",
			Msg:        "get miner power info fail",
			Reward:     reward,
			Pledge:     pledge,
			Power:      power,
			Gas:        gas,
			TotalPower: totalPower,
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	}else {
		for _,minerPower:=range minerPowerStatus{
			totalPower+=minerPower.Power
		}

	}

	resp := data.RewardRespTmp{
		Code:       "ok",
		Msg:        "successful",
		Reward:     reward,
		Pledge:     pledge,
		Power:      power,
		Gas:        gas,
		TotalPower: totalPower,
		Update:     time.Unix(timeStamp, 0).Format("2006-01-02 15:04:05"),
	}
	c.Data["json"] = &resp
	c.ServeJSON()
	return
}

func (c *RewardTmpController) GetMessagesGas() {
	t := c.GetString("time")

	if t == "" {
		resp := data.MessageGasTmp{
			Code: "faile",
			Msg:  "time  is nil",
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
			Code: "faile",
			Msg:  "haven't this miner reward info",
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	} else {
		gas := "0.0"
		for _, info := range expendInfo {
			gas = bit.CalculateReward(gas, info.Gas)
			gas = bit.CalculateReward(gas, info.BaseBurnFee)
			gas = bit.CalculateReward(gas, info.OverEstimationBurn)
		}
		resp := data.MessageGasTmp{
			Code: "ok",
			Msg:  "successful",
			Gas:  gas,
		}
		c.Data["json"] = &resp
		c.ServeJSON()
	}
	return
}


func (c *RewardTmpController) GetMinerInfo() {
	var reward, pledge, power, totalPower float64
	var timeStamp int64
	gas := "0.0"

	miner := c.GetString("miner")
	t := c.GetString("time")
	if miner == "" {
		resp := data.RewardRespTmp{
			Code:   "faile",
			Msg:    "miner  is nil",
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	}
	if t == "" {
		resp := data.RewardRespTmp{
			Code:   "faile",
			Msg:    "time  is nil",
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	}

	rewardInfo :=new(models.RewardInfoTmp)
	o := orm.NewOrm()

	num, err := o.QueryTable("fly_reward_info_tmp").Filter("miner_id",miner).Filter("time", t).All(rewardInfo)
	//log.Logger.Debug("DEBUG: QueryRewardInfo() reward: %+v ", rewardInfo)
	if err != nil  {
		resp := data.RewardRespTmp{
			Code:       "fail",
			Msg:        "get reward info fail",
			Reward:     reward,
			Pledge:     pledge,
			Power:      power,
			Gas:        gas,
			TotalPower: totalPower,
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	} else if num!=0 {
			r, _ := strconv.ParseFloat(rewardInfo.Value, 64)
			reward = r
			pledge = rewardInfo.Pledge
			power = rewardInfo.Power
	}
	minerAndWalletRelations:=make([]models.MinerAndWalletRelation,0)
	num, err = o.QueryTable("fly_miner_and_wallet_relation").Filter("miner_id",miner).All(&minerAndWalletRelations)
	if err != nil || num==0{
		resp := data.RewardRespTmp{
			Code:       "fail",
			Msg:        "get wallet info fail",
			Reward:     reward,
			Pledge:     pledge,
			Power:      power,
			Gas:        gas,
			TotalPower: totalPower,
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	}else {
		for _,wallet:=range minerAndWalletRelations{
			expendInfo := new(models.ExpendInfo)
			num, err = o.QueryTable("fly_expend_info").Filter("wallet_id",wallet.WalletId).Filter("time", t).All(expendInfo)
			//log.Logger.Debug("DEBUG: QueryRewardInfo() reward: %+v ", expendInfo)
			if err != nil  {
				resp := data.RewardRespTmp{
					Code:       "fail",
					Msg:        "get expend info fail",
					Reward:     reward,
					Pledge:     pledge,
					Power:      power,
					Gas:        gas,
					TotalPower: totalPower,
				}
				c.Data["json"] = &resp
				c.ServeJSON()
				return
			} else if num!=0{
					gas = bit.CalculateReward(gas, expendInfo.Gas)
					gas = bit.CalculateReward(gas, expendInfo.BaseBurnFee)
					gas = bit.CalculateReward(gas, expendInfo.OverEstimationBurn)
				}
			}
		}
	//todo totalpower
	minerPowerStatus:=new(models.MinerPowerStatus)
	_,err=o.QueryTable("fly_miner_power_status").Filter("miner_id",miner).Filter("time",t).All(minerPowerStatus)
	if err != nil  {
		resp := data.RewardRespTmp{
			Code:       "fail",
			Msg:        "get miner power info fail",
			Reward:     reward,
			Pledge:     pledge,
			Power:      power,
			Gas:        gas,
			TotalPower: totalPower,
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	}else {
		totalPower=minerPowerStatus.Power
	}


	resp := data.RewardRespTmp{
		Code:       "ok",
		Msg:        "successful",
		Reward:     reward,
		Pledge:     pledge,
		Power:      power,
		Gas:        gas,
		TotalPower: totalPower,
		Update:     time.Unix(timeStamp, 0).Format("2006-01-02 15:04:05"),
	}
	c.Data["json"] = &resp
	c.ServeJSON()
	return
}