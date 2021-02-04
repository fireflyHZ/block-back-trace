package controllers

import (
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	logging "github.com/ipfs/go-log/v2"
	"profit-allocation/models"
	"profit-allocation/tool/bit"
	"strconv"
	"time"
)

var rewardLog = logging.Logger("reward-ctr-log")

type RewardController struct {
	web.Controller
}

func (c *RewardController) GetRewardAndPledge() {
	var blockNum int
	var winCount int64
	var reward, pledge, power, totalPower, totalAvailable, totalPreCommit, totalVesting, totalPleage float64
	var timeStamp time.Time
	gas := "0.0"

	t := c.GetString("time")

	if t == "" {
		resp := models.RewardResp{
			Code:   "faile",
			Msg:    "time  is nil",
			Reward: 0.0,
			Pledge: 0.0,
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	}
	rewardLog.Infof("new request time:%+v", t)
	rewardInfo := make([]models.RewardInfo, 0)
	o := orm.NewOrm()
	num, err := o.QueryTable("fly_reward_info").Filter("time", t).All(&rewardInfo)
	//rewardLog.Debug("DEBUG: QueryRewardInfo() reward: %+v ", rewardInfo)
	if err != nil || num == 0 {
		resp := models.RewardResp{
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
			if timeStamp.Before(info.UpdateTime) {
				timeStamp = info.UpdateTime
			}
			r, _ := strconv.ParseFloat(info.Value, 64)
			reward += r
			pledge += info.Pledge
			power += info.Power
			blockNum += info.BlockNum
			winCount += info.WinCounts
		}

	}

	expendInfo := make([]models.ExpendInfo, 0)

	num, err = o.QueryTable("fly_expend_info").Filter("time", t).All(&expendInfo)
	//rewardLog.Debug("DEBUG: QueryRewardInfo() reward: %+v ", expendInfo)
	if err != nil || num == 0 {
		resp := models.RewardResp{
			Code:        "fail",
			Msg:         "get expend info fail",
			Reward:      reward,
			Pledge:      pledge,
			Power:       power,
			Gas:         gas,
			BlockNumber: blockNum,
			WinCount:    winCount,
			TotalPower:  totalPower,
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	} else {
		for _, info := range expendInfo {
			gas = bit.CalculateReward(gas, info.Gas)
			gas = bit.CalculateReward(gas, info.BaseBurnFee)
			gas = bit.CalculateReward(gas, info.OverEstimationBurn)
		}
	}
	//todo totalpower
	minerPowerStatus := make([]models.MinerPowerStatus, 0)
	num, err = o.QueryTable("fly_miner_power_status").Filter("time", t).All(&minerPowerStatus)
	if err != nil || num == 0 {
		resp := models.RewardResp{
			Code:        "fail",
			Msg:         "get miner power info fail",
			Reward:      reward,
			Pledge:      pledge,
			Power:       power,
			Gas:         gas,
			BlockNumber: blockNum,
			WinCount:    winCount,
			TotalPower:  totalPower,
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	} else {
		for _, minerPower := range minerPowerStatus {
			totalPower += minerPower.Power
			totalAvailable += minerPower.Available
			totalPreCommit += minerPower.PreCommit
			totalPleage += minerPower.Pleage
			totalVesting += minerPower.Vesting
		}

	}

	resp := models.RewardResp{
		Code:           "ok",
		Msg:            "successful",
		Reward:         reward,
		Pledge:         pledge,
		Power:          power,
		Gas:            gas,
		BlockNumber:    blockNum,
		WinCount:       winCount,
		TotalPower:     totalPower,
		TotalAvailable: totalAvailable,
		TotalPreCommit: totalPreCommit,
		TotalPleage:    totalPleage,
		TotalVesting:   totalVesting,
		Update:         timeStamp,
	}
	c.Data["json"] = &resp
	c.ServeJSON()
	return
}

func (c *RewardController) GetMessagesGas() {
	t := c.GetString("time")

	if t == "" {
		resp := models.MessageGasTmp{
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
	rewardLog.Debug("DEBUG: QueryRewardInfo() reward: %+v ", expendInfo)
	if err != nil || num == 0 {
		resp := models.MessageGasTmp{
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
		resp := models.MessageGasTmp{
			Code: "ok",
			Msg:  "successful",
			Gas:  gas,
		}
		c.Data["json"] = &resp
		c.ServeJSON()
	}
	return
}

func (c *RewardController) GetMinerInfo() {
	var blockNum int
	var winCount int64
	var reward, pledge, power, totalPower, totalAvailable, totalPreCommit, totalVesting, totalPleage float64
	var timeStamp time.Time
	gas := "0.0"

	miner := c.GetString("miner")
	t := c.GetString("time")
	if miner == "" {
		resp := models.RewardResp{
			Code: "faile",
			Msg:  "miner is nil",
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	}
	if t == "" {
		resp := models.RewardResp{
			Code: "faile",
			Msg:  "time is nil",
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	}
	rewardLog.Infof("new request miner:%+v time:%+v", miner, t)
	rewardInfo := new(models.RewardInfo)
	o := orm.NewOrm()
	num, err := o.QueryTable("fly_reward_info").Filter("miner_id", miner).Filter("time", t).All(rewardInfo)
	//rewardLog.Debug("DEBUG: QueryRewardInfo() reward: %+v ", rewardInfo)
	if err != nil || num == 0 {
		resp := models.RewardResp{
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
		timeStamp = rewardInfo.UpdateTime
		r, _ := strconv.ParseFloat(rewardInfo.Value, 64)
		reward = r
		pledge = rewardInfo.Pledge
		power = rewardInfo.Power
		blockNum = rewardInfo.BlockNum
		winCount = rewardInfo.WinCounts
	}
	minerAndWalletRelations := make([]models.MinerAndWalletRelation, 0)
	num, err = o.QueryTable("fly_miner_and_wallet_relation").Filter("miner_id", miner).All(&minerAndWalletRelations)
	if err != nil || num == 0 {
		resp := models.RewardResp{
			Code:        "fail",
			Msg:         "get wallet info fail",
			Reward:      reward,
			Pledge:      pledge,
			Power:       power,
			Gas:         gas,
			BlockNumber: blockNum,
			WinCount:    winCount,
			TotalPower:  totalPower,
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	} else {
		for _, wallet := range minerAndWalletRelations {
			expendInfo := new(models.ExpendInfo)
			num, err = o.QueryTable("fly_expend_info").Filter("wallet_id", wallet.WalletId).Filter("time", t).All(expendInfo)
			//rewardLog.Debug("DEBUG: QueryRewardInfo() reward: %+v ", expendInfo)
			if err != nil {
				resp := models.RewardResp{
					Code:        "fail",
					Msg:         "get expend info fail",
					Reward:      reward,
					Pledge:      pledge,
					Power:       power,
					Gas:         gas,
					BlockNumber: blockNum,
					WinCount:    winCount,
					TotalPower:  totalPower,
				}
				c.Data["json"] = &resp
				c.ServeJSON()
				return
			} else if num != 0 {
				gas = bit.CalculateReward(gas, expendInfo.Gas)
				gas = bit.CalculateReward(gas, expendInfo.BaseBurnFee)
				gas = bit.CalculateReward(gas, expendInfo.OverEstimationBurn)
			}
		}
	}
	//todo totalpower
	minerPowerStatus := new(models.MinerPowerStatus)
	num, err = o.QueryTable("fly_miner_power_status").Filter("miner_id", miner).Filter("time", t).All(minerPowerStatus)
	if err != nil || num == 0 {
		resp := models.RewardResp{
			Code:        "fail",
			Msg:         "get miner power info fail",
			Reward:      reward,
			Pledge:      pledge,
			Power:       power,
			Gas:         gas,
			BlockNumber: blockNum,
			WinCount:    winCount,
			TotalPower:  totalPower,
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	} else {
		totalPower = minerPowerStatus.Power
		totalAvailable = minerPowerStatus.Available
		totalPreCommit = minerPowerStatus.PreCommit
		totalPleage = minerPowerStatus.Pleage
		totalVesting = minerPowerStatus.Vesting
	}

	resp := models.RewardResp{
		Code:           "ok",
		Msg:            "successful",
		Reward:         reward,
		Pledge:         pledge,
		Power:          power,
		Gas:            gas,
		BlockNumber:    blockNum,
		WinCount:       winCount,
		TotalPower:     totalPower,
		TotalAvailable: totalAvailable,
		TotalPreCommit: totalPreCommit,
		TotalPleage:    totalPleage,
		TotalVesting:   totalVesting,
		Update:         timeStamp,
	}
	c.Data["json"] = &resp
	c.ServeJSON()
	return
}
