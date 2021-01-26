package lotus

import (
	"fmt"
	logging "github.com/ipfs/go-log/v2"
	"profit-allocation/lotus/reward"
	"profit-allocation/models"
	"profit-allocation/tool/sync"
	"strconv"
	"time"
)

var setupLog = logging.Logger("lotus-setup")

func Setup() {
	reward.CreateLotusClient()
	collectTime := time.NewTicker(time.Second * time.Duration(30))

	defer collectTime.Stop()

	//完成数据初始化
	initTmpData()
	for {
		select {
		case <-collectTime.C:
			loop()
		}

	}
}

func loop() {
	sync.Wg.Add(2)
	go reward.CalculateMsgGasData()
	go reward.CollectTotalRerwardAndPledge()
	sync.Wg.Wait()
}

func initTmpData() {

	minerInfo := make([]models.MinerInfo, 0)
	n, err := models.O.QueryTable("fly_miner_info").All(&minerInfo)
	if err != nil {
		fmt.Println("11111 QueryTable fly_net_run_data_pro", err)
	}
	pleagef02420, err := strconv.ParseFloat("55890.17124143092", 64)
	pleagef021695, err := strconv.ParseFloat("1752.1556517147642", 64)
	pleagef021704, err := strconv.ParseFloat("1979.057228561", 64)
	if err != nil {
		setupLog.Error("ParseFloat err:%+v", err)
	}
	if n == 0 {
		miner1 := models.MinerInfo{
			MinerId:      "f02420",
			QualityPower: 5200.78125,
			Pleage:       pleagef02420,
			CreateTime:   time.Now(),
			//UpdateTime:   0,
		}
		miner2 := models.MinerInfo{
			MinerId:      "f021695",
			QualityPower: 199.03125,
			Pleage:       pleagef021695,
			CreateTime:   time.Now(),
			//UpdateTime:   0,
		}
		miner3 := models.MinerInfo{
			MinerId:      "f021704",
			QualityPower: 301.0625,
			Pleage:       pleagef021704,
			CreateTime:   time.Now(),
			//	UpdateTime:   0,
		}
		minerInfo = append(minerInfo, miner1)
		minerInfo = append(minerInfo, miner2)
		minerInfo = append(minerInfo, miner3)
		//minerInfo=append(minerInfo,miner1)
		n, err = models.O.InsertMulti(3, minerInfo)
		if err != nil {
			fmt.Println("insert netrundata err:", err)
		}
	}

	//netRunData := new(models.NetRunDataProTmp)
	//n, err = models.O.QueryTable("fly_net_run_data_pro_tmp").All(netRunData)
	//if err != nil {
	//	fmt.Println("11111 QueryTable fly_net_run_data_pro", err)
	//}
	//if n == 0 {
	//	netRunData.ReceiveBlockHeight = 344500
	//	n, err = models.O.Insert(netRunData)
	//	if err != nil {
	//		fmt.Println("insert netrundata err:", err)
	//	}
	//}
	//msgGasNetRunData := new(models.MsgGasNetRunDataProTmp)
	//n, err = models.O.QueryTable("fly_msg_gas_net_run_data_pro_tmp").All(msgGasNetRunData)
	//if err != nil {
	//	fmt.Println("11111 QueryTable fly_net_run_data_pro", err)
	//}
	//if n == 0 {
	//	msgGasNetRunData.ReceiveBlockHeight = 344500
	//	n, err = models.O.Insert(msgGasNetRunData)
	//	if err != nil {
	//		fmt.Println("insert msgGasNetRunData err:", err)
	//	}
	//}
	minerAndWalletRelations := make([]models.MinerAndWalletRelation, 0)
	n, err = models.O.QueryTable("fly_miner_and_wallet_relation").All(&minerAndWalletRelations)
	if err != nil {
		fmt.Println("11111 QueryTable fly_net_run_data_pro", err)
	}
	if n == 0 {
		minerAndWalletRelation1 := models.MinerAndWalletRelation{
			MinerId:  "f02420",
			WalletId: "f3wykhltf7g7guch6sz6u6hs4gdmvxrz2likki7aaf7th23jmofrswbd6rmlhbxx4urt6ycbtmlhitgmynky3a",
		}
		minerAndWalletRelation2 := models.MinerAndWalletRelation{
			MinerId:  "f02420",
			WalletId: "f3vfgq65omcht6hbmlwe2g7mowf334zyoa6zcqm543vtmb3uqpnpei4bwhbmo2qi3qntrfiojhcnpciakea6ma",
		}
		minerAndWalletRelation3 := models.MinerAndWalletRelation{
			MinerId:  "f02420",
			WalletId: "f3rmhlmqfaph6czwiqwlg3kfjgejugt5thcviowlmt3l42464q25ptk3znphuuiwrdbyumun3ui7q2gut7v2da",
		}
		minerAndWalletRelation4 := models.MinerAndWalletRelation{
			MinerId:  "f02420",
			WalletId: "f3va7lv4wkcfq5mmqirr4pyrogtnuknw2hma5y6luwbx6iv4qcwgrvzyn2zljgbgtmv7lxr3jsa4eo2az3kqra",
		}
		minerAndWalletRelation5 := models.MinerAndWalletRelation{
			MinerId:  "f021695",
			WalletId: "f3qqdp53ooe4xvqwt4dmoixb6ej6jgmk7zbkjaiujfmfmuyrpenewqre6tlokcxnwp7zpmq3ohlw2wheqir2ga",
		}
		minerAndWalletRelation6 := models.MinerAndWalletRelation{
			MinerId:  "f021695",
			WalletId: "f3wqijosc44y6a6nckbobrwmq6cocoja3lgrly462z3sjwigyi6pzltourrk4lk4jkt332yr5k4xb6mxmct25a",
		}
		minerAndWalletRelation7 := models.MinerAndWalletRelation{
			MinerId:  "f021704",
			WalletId: "f3spvlhfuga45prd7fg7dswphgm4hotpxmydyzpjloy2rekpyfnwpbdnd7wuyael2pryb3xztp4k56ju3ib5sq",
		}
		minerAndWalletRelation8 := models.MinerAndWalletRelation{
			MinerId:  "f021704",
			WalletId: "f3skdqsai23rhavva77g7nkr736j7mjql53xv7362ovlw7o3yz334ajchyb7fir35cnutijfusp6mngobyjvya",
		}
		minerAndWalletRelation9 := models.MinerAndWalletRelation{
			MinerId:  "f021704",
			WalletId: "f3sc6mo6jiwxwwgsx4gwz5vbpcn4p6ejybgogocfntujmjaibluzm6ngj7qqj72gck7rtuibtgsow6ttuq43dq",
		}
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation1)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation2)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation3)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation4)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation5)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation6)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation7)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation8)
		minerAndWalletRelations = append(minerAndWalletRelations, minerAndWalletRelation9)
		n, err = models.O.InsertMulti(9, minerAndWalletRelations)
		if err != nil {
			fmt.Println("insert minerAndWalletRelations err:", err)
		}
	}
}
