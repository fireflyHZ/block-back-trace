package lotus

import (
	"fmt"
	"github.com/astaxie/beego/orm"
	"profit-allocation/lotus/reward"
	"profit-allocation/lotus/wallet"
	"profit-allocation/models"
	"profit-allocation/tool/log"
	"profit-allocation/tool/sync"
	"strconv"
	"time"
)

func Setup() {
	walletProfit := time.NewTicker(time.Second * time.Duration(3600))
	collectTime := time.NewTicker(time.Second * time.Duration(30))
	defer walletProfit.Stop()
	defer collectTime.Stop()

	//完成数据初始化

	initOrderData()
	//---------------------------------------------
	for {
		select {
		case nowDatetime := <-walletProfit.C:
			//判断每天0点进行数据获取
			if isExecutingPoint(nowDatetime) {
				wallet.CalculateWalletProfit()
			}
		case <-collectTime.C:
			loop()
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
	go wallet.CollectWalletData()
	go reward.CollectLotusChainBlockRunData()
	sync.Wg.Wait()
}

func initOrderData() {
	var goodNum int64
	var orderNum int64
	var total int
	o := orm.NewOrm()
	orderInfos:=make([]models.OrderInfo,0)
	n,err:=o.QueryTable("fly_order_info").All(&orderInfos)
	if err != nil {
		fmt.Println("11111 QueryTable fly_order_info", err)
	}
	if n==0{
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
	netRunData:=new(models.NetRunDataPro)
	n,err=o.QueryTable("fly_net_run_data_pro").All(netRunData)
	if err != nil {
		fmt.Println("11111 QueryTable fly_net_run_data_pro", err)
	}
	if n==0{
		netRunData.TotalShare=total
		netRunData.AllShare=50000000
		netRunData.ReceiveBlockHeight=148888
		n,err=o.Insert(netRunData)
		if err != nil {
			fmt.Println("insert netrundata err:",err)
		}
	}


	minerInfo:=make([]models.MinerInfo,0)
	n,err=o.QueryTable("fly_miner_info").All(&minerInfo)
	if err != nil {
		fmt.Println("11111 QueryTable fly_net_run_data_pro", err)
	}
	pleagef02420, err := strconv.ParseFloat("32958.213756507100668595", 64)
	pleagef021695, err := strconv.ParseFloat("1754.011856122781753658", 64)
	pleagef021704, err := strconv.ParseFloat("505.89973939318149791", 64)
	if err != nil {
		log.Logger.Error("ParseFloat err:%+v",err)
	}
	if n==0{
		miner1:=models.MinerInfo{
			MinerId:      "f02420",
			QualityPower: 1855.65625,
			Pleage: pleagef02420,
			CreateTime:   0,
			UpdateTime:   0,
		}
		miner2:=models.MinerInfo{
			MinerId:      "f021695",
			QualityPower: 187.59375,
			Pleage: pleagef021695,
			CreateTime:   0,
			UpdateTime:   0,
		}
		miner3:=models.MinerInfo{
			MinerId:      "f021704",
			QualityPower: 51.53125,
			Pleage: pleagef021704,
			CreateTime:   0,
			UpdateTime:   0,
		}
		minerInfo=append(minerInfo,miner1)
		minerInfo=append(minerInfo,miner2)
		minerInfo=append(minerInfo,miner3)
		//minerInfo=append(minerInfo,miner1)
		n,err=o.InsertMulti(3,minerInfo)
		if err != nil {
			fmt.Println("insert netrundata err:",err)
		}
	}

}
