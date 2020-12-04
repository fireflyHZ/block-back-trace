package reward

import (
	"github.com/astaxie/beego/orm"
	"profit-allocation/models"
	"profit-allocation/tool/log"
	"time"
)

var TimeFlag int64 = 1602633600         //"2020-10-14"日时间戳
var AllocateTimeFlag int64 = 1603584000 //"2020-10-25"日时间戳

func CalculateUserFund(newTimeStamp int64) {
	//遍历user
	o := orm.NewOrm()
	usersInfo := make([]models.UserInfo, 0)
	_, err := o.QueryTable("fly_user_info").All(&usersInfo)
	if err != nil {
		log.Logger.Error("Error CalculateUserFund QueryTable user info err:%+v", err)
		return
	}
	//查看更新时间
	updateTime := usersInfo[0].UpdateTime
	if updateTime != "" {
		tflag, err := time.Parse("2006-01-02", updateTime)
		if err != nil {
			log.Logger.Error("Error CalculateUserFund user info update pares err:%+v", err)
			return
		}
		TimeFlag = tflag.Unix()
	}
	//要更新的日期
	updateTimeStr := time.Unix(TimeFlag, 0).AddDate(0, 0, 1).Format("2006-01-02")
	updateTimeStamp, err := time.Parse("2006-01-02", updateTimeStr)
	if err != nil {
		log.Logger.Error("ERROR: handleRequestInfo() calculateMineReward updateTimeStamp parse err=%+v", err)
		return
	}
	//当前日期
	//nowTimeStr := time.Now().Format("2006-01-02")
	//nowTimeStamp, err := time.Parse("2006-01-02", t)

	//if err != nil {
	//	log.Logger.Error("ERROR: handleRequestInfo() calculateMineReward nowTimeStamp parse err=%+v", err)
	//	return
	//}
	//log.Logger.Debug("Debug update :%+v  now:%+v ", updateTimeStamp.Format("2006-01-02"), nowTimeStamp.Format("2006-01-02"))

	for {
		if newTimeStamp > updateTimeStamp.Unix() {
			t := updateTimeStamp.Format("2006-01-02")
			//从10-25日开始，立即释放当天的25%
			if updateTimeStamp.Unix() >= AllocateTimeFlag {
				err := allocateUsersProfit(usersInfo, t, 0.25, 0.75)
				if err != nil {
					log.Logger.Error("Error allocateUsersProfit err :%+v ", err)
					return
				}
				updateTimeStamp = updateTimeStamp.AddDate(0, 0, 1)
			} else {
				err := allocateUsersProfit(usersInfo, t, 0, 1)
				if err != nil {
					log.Logger.Error("Error allocateUsersProfit err :%+v ", err)
					return
				}
				updateTimeStamp = updateTimeStamp.AddDate(0, 0, 1)
			}
		} else {
			break
		}

	}

}

func allocateUsersProfit(usersInfo []models.UserInfo, t string, releaseProportion, vestingProportion float64) error {
	o := orm.NewOrm()
	for _, userInfo := range usersInfo {
		err := o.Begin()
		if err != nil {
			log.Logger.Error("Error  orm transation begin err:%+v", err)
			return err
		}
		ordersInfo := make([]models.OrderInfo, 0)
		//通过uid找到orderid
		_, err = o.QueryTable("fly_order_info").Filter("user_id", userInfo.UserId).All(&ordersInfo)
		if err != nil {
			log.Logger.Error("Error CalculateUserFund QueryTable orders err:%+v", err)
			err = o.Rollback()
			if err != nil {
				log.Logger.Error("Error  QueryTable orders rollback err:%+v", err)
			}
			return err
		}
		//由orderid找到当天的orderinfo
		ordersRewardInfo := make([]*models.OrderDailyRewardInfo, 0)
		var userDailyReward float64
		var userDailyPleage float64
		var userDailyPower float64
		var userDailyFee float64
		var userTotalShare int
		//var t string

		for _, orderInfo := range ordersInfo {
			orderRewardInfo := new(models.OrderDailyRewardInfo)
			n, err := o.QueryTable("fly_order_daily_reward_info").Filter("order_id", orderInfo.OrderId).Filter("time", t).All(orderRewardInfo)
			if err != nil {
				log.Logger.Error("Error CalculateUserFund QueryTable order inifo err:%+v t:%+v", err, t)
				err = o.Rollback()
				if err != nil {
					log.Logger.Error("Error  QueryTable orders daily reward info rollback err:%+v", err)
				}
				return err
			}

			if n == 0 {
				log.Logger.Error("Error CalculateUserFund QueryTable order inifo  n:%+v t:%+v", n, t)
				err = o.Rollback()
				if err != nil {
					log.Logger.Error("Error  QueryTable orders daily reward info rollback err:%+v", err)
				}
				return err
			}
			//计算userid对应的当日的总收益和支出
			ordersRewardInfo = append(ordersRewardInfo, orderRewardInfo)
			userDailyReward += orderRewardInfo.Reward
			userDailyPleage += orderRewardInfo.Pleage
			userDailyPower += orderRewardInfo.Power
			userDailyFee += orderRewardInfo.Fee
			userTotalShare += orderInfo.Share
			//t = orderRewardInfo.Time
		}
		userVestings := make([]models.VestingInfo, 0)
		n, err := o.QueryTable("fly_vesting_info").Filter("user_id", userInfo.UserId).Filter("times__lt", 180).All(&userVestings)
		if err != nil {
			log.Logger.Error("Error CalculateUserFund QueryTable vesting inifo err:%+v  n:%+v", err, n)
			err = o.Rollback()
			if err != nil {
				log.Logger.Error("Error  QueryTable vesting info rollback err:%+v", err)
			}
			return err
		}
		var totalRelease float64
		var totalVesting float64
		if n == 0 {
			totalRelease = 0
			totalVesting = 0

			userInfo.Reward += userDailyReward
			userInfo.Power += userDailyPower
			userInfo.TotalPleage += userDailyPleage

			userInfo.Share = userTotalShare
			userInfo.Release += userDailyReward * releaseProportion

			//vesting
			userInfo.Vesting = userDailyReward * vestingProportion
			userInfo.Fee += userDailyFee

			if userInfo.Available > 0 {
				userInfo.AdvancePleage += userDailyPleage - userDailyReward*releaseProportion - userInfo.Available
			} else {
				userInfo.AdvancePleage += userDailyPleage - userDailyReward*releaseProportion
			}

			if userInfo.AdvancePleage < 0 {
				userInfo.AdvancePleage = 0
			}

			//available 最新可用余额=已有的可用余额-今天需要额外质押的
			//1. 初始余额-当前余额+当前释放=总质押
			//2. 余额为负时，余额=垫付质押
			userInfo.Available = userInfo.Available - userDailyPleage + userDailyReward*releaseProportion - userDailyFee

			userInfo.UpdateTime = t
			_, err := o.Update(&userInfo)
			if err != nil {
				log.Logger.Error("Error CalculateUserFund UpdateTable users inifo err:%+v ", err)
				err = o.Rollback()
				if err != nil {
					log.Logger.Error("Error  update user info rollback err:%+v", err)
				}
				return err
			}

			//插入今天的数据
			vesting := models.VestingInfo{
				UserId:    userInfo.UserId,
				Vesting:   userDailyReward * vestingProportion,
				Release:   userDailyReward * vestingProportion / 180,
				Times:     0,
				StartTime: t,
			}
			_, err = o.Insert(&vesting)
			if err != nil {
				log.Logger.Error("Error CalculateUserFund Insret Table vesting info err:%+v ", err)
				err = o.Rollback()
				if err != nil {
					log.Logger.Error("Error  insert vesting info rollback err:%+v", err)
				}
				return err
			}
		} else {
			//计算历史释放

			for _, userVesting := range userVestings {
				totalRelease += userVesting.Release
				userVesting.Times++
				totalVesting += userVesting.Vesting - float64(userVesting.Times)*userVesting.Release
				_, err := o.Update(&userVesting)
				if err != nil {
					log.Logger.Error("Error CalculateUserFund UpdateTable vesting inifo err:%+v  ", err)
					err = o.Rollback()
					if err != nil {
						log.Logger.Error("Error  update Table vesting rollback err:%+v", err)
					}
					return err
				}
			}
			userInfo.Reward += userDailyReward
			userInfo.Power += userDailyPower
			userInfo.TotalPleage += userDailyPleage

			userInfo.Share = userTotalShare
			userInfo.Release += totalRelease + userDailyReward*releaseProportion

			//vesting
			userInfo.Vesting = userDailyReward*vestingProportion + totalVesting
			userInfo.Fee += userDailyFee

			if userInfo.Available > 0 {
				userInfo.AdvancePleage += userDailyPleage - totalRelease - userDailyReward*releaseProportion - userInfo.Available
			} else {
				userInfo.AdvancePleage += userDailyPleage - totalRelease - userDailyReward*releaseProportion
			}
			if userInfo.AdvancePleage < 0 {
				userInfo.AdvancePleage = 0
			}

			//available
			userInfo.Available = userInfo.Available - userDailyPleage + userDailyReward*releaseProportion + totalRelease - userDailyFee

			userInfo.UpdateTime = t

			_, err := o.Update(&userInfo)
			if err != nil {
				log.Logger.Error("Error CalculateUserFund UpdateTable users info err:%+v ", err)
				err = o.Rollback()
				if err != nil {
					log.Logger.Error("Error  update Table user info rollback err:%+v", err)
				}
				return err
			}
			//插入今天的数据
			vesting := models.VestingInfo{
				UserId:    userInfo.UserId,
				Vesting:   userDailyReward * vestingProportion,
				Release:   userDailyReward * vestingProportion / 180,
				Times:     0,
				StartTime: t,
			}
			_, err = o.Insert(&vesting)
			if err != nil {
				log.Logger.Error("Error CalculateUserFund Insret Table vesting inifo err:%+v ", err)
				err = o.Rollback()
				if err != nil {
					log.Logger.Error("Error  insert vesting info rollback err:%+v", err)
				}
				return err
			}
		}

		userDailyRewardInfo := models.UserDailyRewardInfo{
			UserId:           userInfo.UserId,
			Reward:           userDailyReward,
			Power:            userDailyPower,
			Pledge:           userDailyPleage,
			Fee:              userDailyFee,
			ImmediateRelease: userDailyReward * releaseProportion,
			LinearRelease:    totalRelease,
			Time:             t,
		}
		_, err = o.Insert(&userDailyRewardInfo)
		if err != nil {
			log.Logger.Error("Error CalculateUserFund Insret Table user daily reward inifo err:%+v ", err)
			err = o.Rollback()
			if err != nil {
				log.Logger.Error("Error  user daily reward info rollback err:%+v", err)
			}
			return err
		}
		err = o.Commit()
		if err != nil {
			log.Logger.Error("Error  orm transation commit err:%+v", err)
		}
	}
	return nil
}
