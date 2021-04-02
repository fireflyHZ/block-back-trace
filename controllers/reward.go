package controllers

import (
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	logging "github.com/ipfs/go-log/v2"
	"profit-allocation/models"
	"time"
)

var rewardLog = logging.Logger("reward-ctr-log")

type RewardController struct {
	web.Controller
}

//获取矿池信息
func (c *RewardController) GetRewardAndPledge() {
	var blockNum int
	var winCount int64
	var reward, pledge, power, totalPower, totalAvailable, totalPreCommit, totalVesting, totalPleage float64
	var timeStamp time.Time
	var gas float64

	t := c.GetString("time")
	mp := c.GetString("mp")

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

	rewardLog.Infof("new request mp:%+v time:%+v", mp, t)
	rewardInfos := make([]models.MinerStatusAndDailyChange, 0)
	o := orm.NewOrm()
	var num int64
	//var err error
	queryTime, err := time.ParseInLocation("2006-01-02", t, time.Local)
	if err != nil {
		resp := models.RewardResp{
			Code:   "faile",
			Msg:    fmt.Sprintf("parse time err:", err),
			Reward: 0.0,
			Pledge: 0.0,
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	}
	if mp == "f02420" {
		num, err = o.QueryTable("fly_miner_status_and_daily_change").Filter("miner_id__in", "f02420", "f021695", "f021704").Filter("time", queryTime).All(&rewardInfos)
		//num, err = o.Raw("select * from fly_miner_status_and_daily_change where miner_id=? or miner_id=? or miner_id=? and update_time::date=to_date(?,'YYYY-MM-DD')", "f02420", "f021695", "f021704", t).QueryRows(&rewardInfos)
	} else {
		num, err = o.QueryTable("fly_miner_status_and_daily_change").Filter("miner_id", mp).Filter("time", queryTime).All(&rewardInfos)
		//num, err = o.Raw("select * from fly_miner_status_and_daily_change where miner_id=? and update_time::date=to_date(?,'YYYY-MM-DD')", mp, t).QueryRows(&rewardInfos)
	}

	if err != nil || num == 0 {
		rewardLog.Errorf("get miner status and daily change err:%+v,num:%+v", err, num)
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
		for _, info := range rewardInfos {
			if timeStamp.Before(info.UpdateTime) {
				timeStamp = info.UpdateTime
			}
			reward += info.Reward
			pledge += info.Pledge
			power += info.Power
			blockNum += info.BlockNum
			winCount += info.WinCounts
			//当天状态
			totalPower += info.TotalPower
			totalAvailable += info.TotalAvailable
			totalPreCommit += info.TotalPreCommit
			totalPleage += info.TotalPledge
			totalVesting += info.TotalVesting
		}
	}

	expendInfo := make([]models.ExpendInfo, 0)
	if mp == "f02420" {
		num, err = o.Raw("select * from fly_expend_info where miner_id=? or miner_id=? or miner_id=? and update_time::date=to_date(?,'YYYY-MM-DD')", "f02420", "f021695", "f021704", t).QueryRows(&expendInfo)
		//num, err = o.QueryTable("fly_expend_info").Filter("miner_id_in", "f02420", "f021695", "f021704").Filter("time", queryTime).All(&expendInfo)
	} else {
		//num, err = o.QueryTable("fly_expend_info").Filter("miner_id", mp).Filter("time", queryTime).All(&expendInfo)
		num, err = o.Raw("select * from fly_expend_info where miner_id=? and update_time::date=to_date(?,'YYYY-MM-DD')", mp, t).QueryRows(&expendInfo)
	}
	if err != nil || num == 0 {
		rewardLog.Errorf("get expend info err:%+v,num:%+v", err, num)
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
			gas += info.Gas
			gas += info.BaseBurnFee
			gas += info.OverEstimationBurn
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
		var gas float64
		for _, info := range expendInfo {
			gas += info.Gas
			gas += info.BaseBurnFee
			gas += info.OverEstimationBurn
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
	var gas float64

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
	queryTime, err := time.ParseInLocation("2006-01-02", t, time.Local)
	if err != nil {
		resp := models.RewardResp{
			Code: "faile",
			Msg:  "time is nil",
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	}
	rewardInfo := new(models.MinerStatusAndDailyChange)
	o := orm.NewOrm()
	num, err := o.QueryTable("fly_miner_status_and_daily_change").Filter("miner_id", miner).Filter("time", queryTime).All(rewardInfo)
	//	num, err := o.Raw("select * from fly_miner_status_and_daily_change where miner_id=? and update_time::date=to_date(?,'YYYY-MM-DD')", miner, t).QueryRows(&rewardInfos)
	if err != nil || num == 0 {
		rewardLog.Errorf("get miner status and daily change err:%+v,num:%+v", err, num)
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
		//rewardInfo := rewardInfos[0]
		timeStamp = rewardInfo.UpdateTime
		reward = rewardInfo.Reward
		pledge = rewardInfo.Pledge
		power = rewardInfo.Power
		blockNum = rewardInfo.BlockNum
		winCount = rewardInfo.WinCounts
		//
		totalPower = rewardInfo.TotalPower
		totalAvailable = rewardInfo.TotalAvailable
		totalPreCommit = rewardInfo.TotalPreCommit
		totalPleage = rewardInfo.TotalPledge
		totalVesting = rewardInfo.TotalVesting
	}

	expendInfos := make([]models.ExpendInfo, 0)
	//num, err = o.QueryTable("fly_expend_info").Filter("miner_id", miner).Filter("time", queryTime).All(&expendInfos)
	num, err = o.Raw("select * from fly_expend_info where miner_id=? and update_time::date=to_date(?,'YYYY-MM-DD')", miner, t).QueryRows(&expendInfos)
	if err != nil {
		rewardLog.Errorf("get expend info err:%+v,num:%+v", err, num)
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
		for _, expendInfo := range expendInfos {
			gas += expendInfo.Gas
			gas += expendInfo.BaseBurnFee
			gas += expendInfo.OverEstimationBurn
		}
	}
	expendMsgs := make([]models.ExpendMessages, 0)

	sectorNum, err := o.Raw("select * from fly_expend_messages where miner_id=? and method=7 and create_time::date=to_date(?,'YYYY-MM-DD')", miner, t).QueryRows(&expendMsgs)
	if err != nil {
		rewardLog.Errorf("get expend message info err:%+v,num:%+v", err, num)
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
	}
	resp := models.RewardResp{
		Code:           "ok",
		Msg:            "successful",
		Reward:         reward,
		Pledge:         pledge,
		Power:          power,
		Gas:            gas,
		BlockNumber:    blockNum,
		SectorsNumber:  sectorNum,
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
