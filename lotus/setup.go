package lotus

import (
	"fmt"
	"github.com/astaxie/beego/orm"
	"profit-allocation/lotus/reward"
	"profit-allocation/models"
	"profit-allocation/tool/log"
	"profit-allocation/tool/sync"
	"strconv"
	"time"
)

//var HandleChannel = make(chan int)

func Setup() {
	//walletProfit := time.NewTicker(time.Second * time.Duration(3600))
	collectTime := time.NewTicker(time.Second * time.Duration(30))
	//defer walletProfit.Stop()
	defer collectTime.Stop()

	//完成数据初始化

	//initOrderData()
	initTmpData()

	//---------------------------------------------
	//HandleChannel<-0
	for {
		select {
		//case nowDatetime := <-walletProfit.C:
		//	//判断每天0点进行数据获取
		//	if isExecutingPoint(nowDatetime) {
		//		wallet.CalculateWalletProfit()
		//		//reward.CalculateUserFund()
		//	}
		case <-collectTime.C:
			loop()
		//case <-HandleChannel:
		//	time.Sleep(time.Second*20)
		//	reward.CollectTotalRerwardAndPledge()
		}

	}
}
func isExecutingPoint(nowDatetime time.Time) bool {
	nowH, _, _ := nowDatetime.Clock()
	if nowH == 0 {
		return true
	} else {
		return false
	}
}

func loop() {
	sync.Wg.Add(2)
//	go wallet.CollectWalletData()
//	go reward.CollectLotusChainBlockRunData()
	//================
	go reward.CollectTotalRerwardAndPledge()
	go reward.CalculateMsgGasData()
	sync.Wg.Wait()
}

func initOrderData() {
	var goodNum int64
	var orderNum int64
	var total int
	o := orm.NewOrm()
	orderInfos := make([]models.OrderInfo, 0)
	n, err := o.QueryTable("fly_order_info").All(&orderInfos)
	if err != nil {
		fmt.Println("11111 QueryTable fly_order_info", err)
	}
	if n == 0 {
		settlePlans := make([]models.SettlePlan, 0)
		n, err := o.QueryTable("fly_settle_plan").All(&settlePlans)
		if err != nil {
			fmt.Println("11111", err)
		}

		for _, settlePlan := range settlePlans {
			orderGoods := make([]models.OrderGoods, 0)
			n, err = o.QueryTable("fly_order_goods").Filter("article_id", settlePlan.ArticleId).All(&orderGoods)
			if err != nil {
				fmt.Println("22222", err)
			}
			goodNum += n
			//fmt.Printf("order goods %+v\n", orderGoods)
			for _, orderGood := range orderGoods {
				orders := new(models.Orders)
				n, err = o.QueryTable("fly_orders").Filter("id", orderGood.OrderId).Filter("status__in", 2, 3).All(orders)
				if err != nil || n == 0 {
					fmt.Println("333333", err, "n", n)
					continue
				}
				orderNum += n

				orderInfo := new(models.OrderInfo)
				orderInfo.UserId = orders.UserId
				orderInfo.OrderId = orders.Id
				orderInfo.Power = 12
				orderInfo.Share = orderGood.Quantity * int(settlePlan.Quantity)
				total += (orderGood.Quantity) * int(settlePlan.Quantity)
				_, err = o.Insert(orderInfo)
				if err != nil || n == 0 {
					fmt.Println("insert ", orders.UserId, err)
				}
			}
		}
	}
	fmt.Printf("total:%+v\n  good：%+v\n order:%+v\n", total, goodNum, orderNum)
	netRunData := new(models.NetRunDataPro)
	n, err = o.QueryTable("fly_net_run_data_pro").All(netRunData)
	if err != nil {
		fmt.Println("11111 QueryTable fly_net_run_data_pro", err)
	}
	if n == 0 {
		netRunData.TotalShare = total
		netRunData.AllShare = 50000000
		netRunData.ReceiveBlockHeight = 148888
		n, err = o.Insert(netRunData)
		if err != nil {
			fmt.Println("insert netrundata err:", err)
		}
	}

	minerInfo := make([]models.MinerInfo, 0)
	n, err = o.QueryTable("fly_miner_info").All(&minerInfo)
	if err != nil {
		fmt.Println("11111 QueryTable fly_net_run_data_pro", err)
	}
	pleagef02420, err := strconv.ParseFloat("32958.213756507100668595", 64)
	pleagef021695, err := strconv.ParseFloat("1754.011856122781753658", 64)
	pleagef021704, err := strconv.ParseFloat("505.89973939318149791", 64)
	if err != nil {
		log.Logger.Error("ParseFloat err:%+v", err)
	}
	if n == 0 {
		miner1 := models.MinerInfo{
			MinerId:      "f02420",
			QualityPower: 1855.65625,
			Pleage:       pleagef02420,
			CreateTime:   0,
			UpdateTime:   0,
		}
		miner2 := models.MinerInfo{
			MinerId:      "f021695",
			QualityPower: 187.59375,
			Pleage:       pleagef021695,
			CreateTime:   0,
			UpdateTime:   0,
		}
		miner3 := models.MinerInfo{
			MinerId:      "f021704",
			QualityPower: 51.53125,
			Pleage:       pleagef021704,
			CreateTime:   0,
			UpdateTime:   0,
		}
		minerInfo = append(minerInfo, miner1)
		minerInfo = append(minerInfo, miner2)
		minerInfo = append(minerInfo, miner3)
		//minerInfo=append(minerInfo,miner1)
		n, err = o.InsertMulti(3, minerInfo)
		if err != nil {
			fmt.Println("insert netrundata err:", err)
		}
	}

	//初始化userInfo表
	userInfos := make([]models.UserInfo, 0)
	n, err = o.QueryTable("fly_user_info").All(&userInfos)
	if err != nil {
		fmt.Println("11111 QueryTable fly_user_info", err)
	}
	if n == 0 {
		orderInfos := make([]models.OrderInfo, 0)
		_, err := o.QueryTable("fly_order_info").All(&orderInfos)
		if err != nil {
			fmt.Println("11111 QueryTable fly_order_info", err)
		}
		for _, order := range orderInfos {
			uInfo := new(models.UserInfo)
			n, err := o.QueryTable("fly_user_info").Filter("user_id", order.UserId).All(uInfo)
			if err != nil {
				fmt.Println("11111 QueryTable fly_user_info", err)
			}
			if n == 0 {
				userFilDaily := new(models.UserFilDaily)
				_, err := o.QueryTable("fly_user_fil_daily").Filter("user_id", order.UserId).All(userFilDaily)
				if err != nil {
					fmt.Println("11111 QueryTable fly_user_fil_daily", order.UserId, err)
				}
				userFilPledge := new(models.UserFilPledge)
				_, err = o.QueryTable("fly_user_fil_pledge").Filter("user_id", order.UserId).All(userFilPledge)
				if err != nil {
					fmt.Println("11111 QueryTable fly_user_fil_pleage", order.UserId, err)
				}
				uInfo.UserId = order.UserId
				//uInfo.Share=order.Share
				uInfo.Reward = userFilDaily.FilAmount + userFilPledge.FilPledge
				//uInfo.Available=userFilDaily.FilAmount*0.25
				uInfo.Available = 0
				//uInfo.Vesting=userFilDaily.FilAmount*0.75+userFilPledge.FilPledge
				uInfo.Vesting = userFilDaily.FilAmount + userFilPledge.FilPledge
				_, err = o.Insert(uInfo)
				if err != nil {
					fmt.Printf("11111 insert user:%+v to fly_user_info err:%+v \n", uInfo.UserId, err)
				}
				vesting := models.VestingInfo{
					UserId: order.UserId,
					//Vesting:   userFilDaily.FilAmount * 0.75,
					Vesting: userFilDaily.FilAmount,
					//Release:   userFilDaily.FilAmount * 0.75 / 180,
					Release:   userFilDaily.FilAmount / 180,
					Times:     0,
					StartTime: "2020-10-15",
				}
				_, err = o.Insert(&vesting)
				if err != nil {
					fmt.Println("11111 insert Insret Table vesting info err: ", err)
					//return
				}

			}
		}
	}

}

func initTmpData() {
	o := orm.NewOrm()
	minerInfo := make([]models.MinerInfoTmp, 0)
	n, err := o.QueryTable("fly_miner_info_tmp").All(&minerInfo)
	if err != nil {
		fmt.Println("11111 QueryTable fly_net_run_data_pro", err)
	}
	pleagef02420, err := strconv.ParseFloat("40740.0792792743", 64)
	pleagef021695, err := strconv.ParseFloat("1752.1556517147642", 64)
	pleagef021704, err := strconv.ParseFloat("1740.0369707424757", 64)
	if err != nil {
		log.Logger.Error("ParseFloat err:%+v", err)
	}
	if n == 0 {
		miner1 := models.MinerInfoTmp{
			MinerId:      "f02420",
			QualityPower: 3064.53125,
			Pleage:       pleagef02420,
			CreateTime:   0,
			UpdateTime:   0,
		}
		miner2 := models.MinerInfoTmp{
			MinerId:      "f021695",
			QualityPower: 199.03125,
			Pleage:       pleagef021695,
			CreateTime:   0,
			UpdateTime:   0,
		}
		miner3 := models.MinerInfoTmp{
			MinerId:      "f021704",
			QualityPower: 255.78125,
			Pleage:       pleagef021704,
			CreateTime:   0,
			UpdateTime:   0,
		}
		minerInfo = append(minerInfo, miner1)
		minerInfo = append(minerInfo, miner2)
		minerInfo = append(minerInfo, miner3)
		//minerInfo=append(minerInfo,miner1)
		n, err = o.InsertMulti(3, minerInfo)
		if err != nil {
			fmt.Println("insert netrundata err:", err)
		}
	}

	netRunData := new(models.NetRunDataProTmp)
	n, err = o.QueryTable("fly_net_run_data_pro_tmp").All(netRunData)
	if err != nil {
		fmt.Println("11111 QueryTable fly_net_run_data_pro", err)
	}
	if n == 0 {
		netRunData.ReceiveBlockHeight = 221040
		n, err = o.Insert(netRunData)
		if err != nil {
			fmt.Println("insert netrundata err:", err)
		}
	}
	msgGasNetRunData := new(models.MsgGasNetRunDataProTmp)
	n, err = o.QueryTable("fly_msg_gas_net_run_data_pro_tmp").All(msgGasNetRunData)
	if err != nil {
		fmt.Println("11111 QueryTable fly_net_run_data_pro", err)
	}
	if n == 0 {
		msgGasNetRunData.ReceiveBlockHeight = 221040
		n, err = o.Insert(msgGasNetRunData)
		if err != nil {
			fmt.Println("insert msgGasNetRunData err:", err)
		}
	}
}
