package wallet

import (
	"context"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/filecoin-project/go-address"
	"net/http"
	"profit-allocation/models"
	"profit-allocation/tool/bit"
	"profit-allocation/tool/log"
	"profit-allocation/tool/sync"
	"strings"
	"time"

	lotusClient "github.com/filecoin-project/lotus/api/client"
)

func CollectWalletData() {
	defer sync.Wg.Done()
	//获取钱包地址
	walletsStr := beego.AppConfig.String("wallets")
	wallets := strings.Split(walletsStr, ",")
	//log.Logger.Debug("Debug collectWalletData wallets:%+v", wallets)
	//查询数据
	o := orm.NewOrm()
	//	walletsInfo:=make([]models.WalletBaseinfo,0)
	walletsHistoryInfo := make([]models.WalletHistoryData, 0)
	//轮询
	err := o.Begin()
	if err != nil {
		log.Logger.Debug("DEBUG: collectWalletData orm transation begin error: %+v", err)
		return
	}

	for _, wallet := range wallets {
		fil, attoFil := getWalletinfo(wallet)
		walletInfo := new(models.WalletBaseinfo)
		n, err := o.QueryTable("fly_wallet_baseinfo").Filter("wallet_id", wallet).All(walletInfo)
		if err != nil {
			log.Logger.Error("Error  QueryTable wallet:%+v err:%+v num:%+v", wallet, err, n)
			err := o.Rollback()
			if err != nil {
				log.Logger.Debug("DEBUG: collectWalletData orm transation rollback error: %+v", err)
			}
			return
		}
		if n == 0 {
			walletInfo.WalletId = wallet
			walletInfo.BalanceFil = fil
			walletInfo.BalanceAttofil = attoFil
			walletInfo.CreateTime = time.Now().Unix()
			walletInfo.UpdateTime = time.Now().Unix()
			walletInfo.Status = "0"
			_, err := o.Insert(walletInfo)
			if err != nil {
				log.Logger.Error("Error  Insert wallet:%+v err:%+v ", wallet, err)
				err := o.Rollback()
				if err != nil {
					log.Logger.Debug("DEBUG: collectWalletData orm transation rollback error: %+v", err)
				}
				return
			}
		} else {
			//更新walletinfo
			walletInfo.BalanceFil = fil
			walletInfo.BalanceAttofil = attoFil
			walletInfo.UpdateTime = time.Now().Unix()
			_, err := o.Update(walletInfo)
			if err != nil {
				log.Logger.Error("Error  Update wallet:%+v err:%+v num:%+v", wallet, err)
				err := o.Rollback()
				if err != nil {
					log.Logger.Debug("DEBUG: collectWalletData orm transation rollback error: %+v", err)
				}
				return
			}
		}
		//插入historyinfo
		walletHistoryInfo := models.WalletHistoryData{
			// Id:             0,
			WalletId:       wallet,
			BalanceFil:     fil,
			BalanceAttofil: attoFil,
			CreateTime:     time.Now().Unix(),
			Status:         "0",
		}
		walletsHistoryInfo = append(walletsHistoryInfo, walletHistoryInfo)
	}
	amount := len(walletsHistoryInfo)

	num, err := o.InsertMulti(amount, walletsHistoryInfo)
	if err != nil {
		log.Logger.Error("Error  InsertMulti wallet history info  err:%+v ", err)
		err := o.Rollback()
		if err != nil {
			log.Logger.Debug("DEBUG: collectWalletData orm transation rollback error: %+v", err)
		}
		return
	}
	if int(num) != amount {
		log.Logger.Error("Error  InsertMulti wallet history info  num:%+v != len:%+v ", num, amount)
		err := o.Rollback()
		if err != nil {
			log.Logger.Debug("DEBUG: collectWalletData orm transation rollback error: %+v", err)
		}
		return
	}
	err = o.Commit()
	if err != nil {
		log.Logger.Debug("DEBUG: collectWalletData orm transation Commit error: %+v", err)
		return
	}


}

func CalculateWalletProfit() {
	log.Logger.Debug("DEBUG: calculateWalletProfit()")
	//事务
	o := orm.NewOrm()
	err := o.Begin()
	if err != nil {
		log.Logger.Debug("DEBUG: calculateWalletProfit orm transation begin error: %+v", err)
		return
	}

	var walletProInfos []models.WalletProfitInfo
	//query wallet_fund_settlement  max settlement date
	sql := "select settlement_type,max(settlement_date)  settlement_date from fly_wallet_fund_settlement_persistence_data group by settlement_type"
	_, err = o.Raw(sql).QueryRows(&walletProInfos)
	if err != nil {
		log.Logger.Error("ERROR: calculateWalletProfit() QueryRows err=%v", err)
		err := o.Rollback()
		if err != nil {
			log.Logger.Debug("DEBUG: calculateWalletProfit orm transation rollback error: %+v", err)
			return
		}
		return
	} else {
		log.Logger.Debug("DEBUG: calculateWalletProfit walletProInfos=%+v", walletProInfos)
	}
	mapWalletFuncSettlementStatus := make(map[string]string)

	//记录type-date
	for i := 0; i < len(walletProInfos); i++ {
		SettlementType := walletProInfos[i].SettlementType
		SettlementDate := walletProInfos[i].SettlementDate
		mapWalletFuncSettlementStatus[SettlementType] = SettlementDate
	}
	walletProInfos = make([]models.WalletProfitInfo, 0)
	ws := make([]models.WalletBaseinfo, 0)
	//build wallet fund settlement daliy data
	if wfsd, err := calculate(ws, mapWalletFuncSettlementStatus["1"]); err != nil {
		log.Logger.Error("ERROR: DoWalletFundSettlement(), err=%v", err)

		err := o.Rollback()
		if err != nil {
			log.Logger.Debug("DEBUG: calculateWalletProfit orm transation rollback error: %+v", err)
			return
		}
		return
	} else {
		//log.Logger.Debug("DEBUG: DoWalletFundSettlement(), wfsd=%v", wfsd)
		if len(wfsd) > 0 {
			walletProInfos = append(walletProInfos, wfsd...)
		}
	}

	if len(walletProInfos) > 0 {
		num, err := o.InsertMulti(4, walletProInfos)
		if err != nil {
			log.Logger.Error("ERROR: DoWalletFundSettlement(), InsertMulti error:%+v; sueccess num:%+v", err, num)
			err := o.Rollback()
			if err != nil {
				log.Logger.Debug("DEBUG: calculateWalletProfit orm transation rollback error: %+v", err)
				return
			}
			return
		}

	} else {
		log.Logger.Debug("DEBUG: DoWalletFundSettlement() calculate not have resp")
		err := o.Rollback()
		if err != nil {
			log.Logger.Debug("DEBUG: calculateWalletProfit orm transation rollback error: %+v", err)
			return
		}
	}
	err = o.Commit()
	if err != nil {
		log.Logger.Debug("DEBUG: calculateWalletProfit orm transation commit error: %+v", err)
		return
	}
}

func calculate(ws []models.WalletBaseinfo, latestSettlementDate string) ([]models.WalletProfitInfo, error) {
	log.Logger.Debug("DEBUG: DoWalletFundSettlementDaliyData() latestSettlementDate: %+v", latestSettlementDate)

	var rsp []models.WalletProfitInfo

	var tNextSettlementDate time.Time
	//获取昨天的日期
	lastDate := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	//判断查得的日期
	if len(latestSettlementDate) > 0 {
		if latestSettlementDate >= lastDate {
			log.Logger.Debug("DEBUG: DoWalletFundSettlementDaliyData(), latestSettlement=%v, lastMonth=%v, return", latestSettlementDate, lastDate)
			return nil, nil
		}
		if tLatestSettlementDate, err := time.Parse("2006-01-02", latestSettlementDate); err != nil {
			log.Logger.Error("ERROR: DoWalletFundSettlementDaliyData(), err=%v", err)
			return nil, err
		} else {
			//下次处理的日期
			tNextSettlementDate = tLatestSettlementDate.AddDate(0, 0, 2)
		}
	} else {
		//如果之前没有,下次为昨天
		tNextSettlementDate = time.Now()
	}
	//
	begin := tNextSettlementDate
	currentDate := time.Now().Format("2006-01-02")
	for {
		nextDate := begin.AddDate(0, 0, -1).Format("2006-01-02")
		if nextDate == currentDate {
			break
		}
		dayStartDateTime := begin.AddDate(0, 0, -1).Format("2006-01-02 15:04:05")
		dayEndDateTime := begin.Format("2006-01-02 15:04:05")
		monthStartDateTime := begin.AddDate(0, -1, 0).Format("2006-01-02 15:04:05")
		monthEndDateTime := begin.Format("2006-01-02 15:04:05")
		quarterStartDateTime := begin.AddDate(0, -3, 0).Format("2006-01-02 15:04:05")
		quarterEndDateTime := begin.Format("2006-01-02 15:04:05")
		yearStartDateTime := begin.AddDate(-1, 0, 0).Format("2006-01-02 15:04:05")
		yearEndDateTime := begin.Format("2006-01-02 15:04:05")

		//遍历walletbaseinfo
		for i := 0; i < len(ws); i++ {
			//day
			dayStartFIL, dayStartAttoFil, err := getBalance(dayStartDateTime, ws[i].WalletId)
			if err != nil {
				log.Logger.Error("ERROR: DoWalletFundSettlementDaliyData(), err=%v", err)
				return nil, err
			}
			dayEndFIL, dayEndAttoFil, err := getBalance(dayEndDateTime, ws[i].WalletId)
			if err != nil {
				log.Logger.Error("ERROR: DoWalletFundSettlementDaliyData(), err=%v", err)
				return nil, err
			}
			//month
			monthStartFIL, monthStartAttoFil, err := getBalance(monthStartDateTime, ws[i].WalletId)
			if err != nil {
				log.Logger.Error("ERROR: DoWalletFundSettlementDaliyData(), err=%v", err)
				return nil, err
			}
			monthEndFIL, monthEndAttoFil, err := getBalance(monthEndDateTime, ws[i].WalletId)
			if err != nil {
				log.Logger.Error("ERROR: DoWalletFundSettlementDaliyData(), err=%v", err)
				return nil, err
			}
			//quarter
			quarterStartFIL, quarterStartAttoFil, err := getBalance(quarterStartDateTime, ws[i].WalletId)
			if err != nil {
				log.Logger.Error("ERROR: DoWalletFundSettlementDaliyData(), err=%v", err)
				return nil, err
			}
			quarterEndFIL, quarterEndAttoFil, err := getBalance(quarterEndDateTime, ws[i].WalletId)
			if err != nil {
				log.Logger.Error("ERROR: DoWalletFundSettlementDaliyData(), err=%v", err)
				return nil, err
			}
			//year
			yearStartFIL, yearStartAttoFil, err := getBalance(yearStartDateTime, ws[i].WalletId)
			if err != nil {
				log.Logger.Error("ERROR: DoWalletFundSettlementDaliyData(), err=%v", err)
				return nil, err
			}
			yearEndFIL, yearEndAttoFil, err := getBalance(yearEndDateTime, ws[i].WalletId)
			if err != nil {
				log.Logger.Error("ERROR: DoWalletFundSettlementDaliyData(), err=%v", err)
				return nil, err
			}
			//day
			amountDayFIL := bit.StringSub(dayEndFIL, dayStartFIL)
			amountDayAttoFil := bit.StringSub(dayEndAttoFil, dayStartAttoFil)
			//month
			amountMonthFIL := bit.StringSub(monthEndFIL, monthStartFIL)
			amountMonthAttoFil := bit.StringSub(monthEndAttoFil, monthStartAttoFil)
			//quarter
			amountQuarterFIL := bit.StringSub(quarterEndFIL, quarterStartFIL)
			amountQuarterAttoFil := bit.StringSub(quarterEndAttoFil, quarterStartAttoFil)
			//year
			amountYearFIL := bit.StringSub(yearEndFIL, yearStartFIL)
			amountYearAttoFil := bit.StringSub(yearEndAttoFil, yearStartAttoFil)

			log.Logger.Debug("DEBUG: DoWalletFundSettlementDaliyData(), amountFil=%+v amountAttoFil: %+v", amountDayFIL, amountDayAttoFil)
			//day
			dayStartAmount, dayEndAmount := getAmount(dayStartFIL, dayStartAttoFil, dayEndFIL, dayEndAttoFil)
			monthStartAmount, monthEndAmount := getAmount(monthStartFIL, monthStartAttoFil, monthEndFIL, monthEndAttoFil)
			quarterStartAmount, quarterEndAmount := getAmount(quarterStartFIL, quarterStartAttoFil, quarterEndFIL, quarterEndAttoFil)
			yearStartAmount, yearEndAmount := getAmount(yearStartFIL, yearStartAttoFil, yearEndFIL, yearEndAttoFil)

			//	mId := mapWalletToMiner[ws[i].WalletId]
			dayWSD := handleResult(dayStartAmount, dayEndAmount, ws[i].WalletId, "1", "0", nextDate, amountDayFIL, amountDayAttoFil)
			monthWSD := handleResult(monthStartAmount, monthEndAmount, ws[i].WalletId, "2", "0", nextDate, amountMonthFIL, amountMonthAttoFil)
			quarterWSD := handleResult(quarterStartAmount, quarterEndAmount, ws[i].WalletId, "3", "0", nextDate, amountQuarterFIL, amountQuarterAttoFil)
			yearWSD := handleResult(yearStartAmount, yearEndAmount, ws[i].WalletId, "4", "0", nextDate, amountYearFIL, amountYearAttoFil)
			rsp = append(rsp, dayWSD)
			rsp = append(rsp, monthWSD)
			rsp = append(rsp, quarterWSD)
			rsp = append(rsp, yearWSD)
		}
		begin = begin.AddDate(0, 0, 1)
	}
	return rsp, nil
}

func getBalance(time string, walletId string) (string, string, error) {
	Fil := "0"
	attoFil := "0"
	o := orm.NewOrm()
	walletHistoryInfo := models.WalletHistoryData{}
	num, err := o.QueryTable("fly_wallet_history_data").Filter("wallet_id", walletId).Filter("create_time_lte", time).OrderBy("-create_time").All(&walletHistoryInfo)
	if err != nil {
		return Fil, attoFil, err
	}
	if num == 0 {
		return Fil, attoFil, errors.New("num = 0")
	}
	//根据walleid,日期. 查询wallethistorydata
	Fil = walletHistoryInfo.BalanceFil
	attoFil = walletHistoryInfo.BalanceAttofil

	log.Logger.Debug("DEBUG: DoWalletFundSettlementDaliyData() getBalance FIL:%+v;Fil:%+v", Fil, attoFil)
	return Fil, attoFil, nil
}

func handleResult(startAmount, endAmount, walletId, selltlementType, status, nextDate, amountFIL, amountAttoFil string) models.WalletProfitInfo {

	return models.WalletProfitInfo{
		//Id:             id.NewPrimaryId(db.TBL_WALLET_FUND_SETTLEMENT_PERSISTENCE_DATA_PREFIX),
		SettlementType: selltlementType,
		SettlementDate: nextDate,
		//	ComputerRoomId: computerRoomId,
		WalletId: walletId,
		//MinerId:        minerId,
		//StartAmount:    bit.TransFilToFIL(StartAmount),
		StartAmount: startAmount,
		//	EndAmount:      bit.TransFilToFIL(EndAmount),
		EndAmount: endAmount,
		//Amount:         bit.TransFilToFIL(bit.StringFilAttofil(amountFil, amountAttoFil)),
		Amount:     bit.StringFilAttofil(amountFIL, amountAttoFil),
		CreateTime: time.Now().Unix(),
		Status:     status,
	}
}

func getAmount(startFIL, startAttoFil, endFIL, endAttoFil string) (string, string) {
	var startAmount string = "0"
	if len(startFIL) > 0 {
		startAmount = startFIL
	}
	if len(startAttoFil) > 0 && startAttoFil != "0" {
		startAmount = fmt.Sprintf("%s.%s", startAmount, bit.IntTo18BitDecimal(startAttoFil))
	}

	var endAmount string = "0"
	if len(endFIL) > 0 {
		endAmount = endFIL
	}
	if len(endAttoFil) > 0 && endAttoFil != "0" {
		endAmount = fmt.Sprintf("%s.%s", endAmount, bit.IntTo18BitDecimal(endAttoFil))
	}
	return startAmount, endAmount
}

func getWalletinfo(walletId string) (fil, attoFil string) {
	lotusHost := beego.AppConfig.String("lotusHost")
	requestHeader := http.Header{}
	nodeApi, closer, err := lotusClient.NewFullNodeRPC(context.Background(), lotusHost, requestHeader)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer closer()
	//balance 处理 todo,如何组装address.Address
	walletAddr,err:=address.NewFromString(walletId)
	if err != nil {
		log.Logger.Error("Error  WalletBalance NewFromString error:%+v", err)
		return
	}
	//log.Logger.Debug("Debug  WalletBalance NewFromString walletAddr :%+v", walletAddr)

	resp, err := nodeApi.WalletBalance(context.Background(),walletAddr)
	if err != nil {
		log.Logger.Error("Error  WalletBalance error:%+v", err)
		return
	}
	balance := resp.String()
	amount:= bit.TransFilToFIL(balance)
	filSlice:=strings.Split(amount,".")
	fil=filSlice[0]
	attoFil=filSlice[1]
	return
}
