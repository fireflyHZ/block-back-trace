package main

import (
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/filter/cors"
	_ "github.com/lib/pq"
	"profit-allocation/controllers"
	"profit-allocation/lotus/power"
	"profit-allocation/models"
)

func main() {

	if err := initDatabase(); err != nil {
		fmt.Println("init database error:", err)
		return
	}
	if err := models.InitData(); err != nil {
		fmt.Println("init data error:", err)
		return
	}

	power.PartitionCheck()
	//go lotus.Setup()

	web.InsertFilter("*", web.BeforeRouter, cors.Allow(&cors.Options{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		AllowCredentials: true,
	}))

	web.Router("/firefly/profit/total_reward_info", &controllers.RewardController{}, "get:GetRewardAndPledge")

	web.Router("/firefly/profit/total_messages_gas_info", &controllers.RewardController{}, "get:GetMessagesGas")

	web.Router("/firefly/profit/total_miner_info", &controllers.RewardController{}, "get:GetMinerInfo")
	web.Router("/firefly/profit/block", &controllers.BlockController{}, "get:GetMinerMineBlockPercentage")
	web.Router("/firefly/profit/luck", &controllers.BlockController{}, "get:GetMinersLuck")
	web.Router("/firefly/profit/balance", &controllers.MinerController{}, "get:GetMinerBalance")

	web.Run()
}

//初始化mysql
func initDatabase() error {
	// 注册数据库驱动
	if err := orm.RegisterDriver("postgres", orm.DRPostgres); err != nil {
		return err
	}

	url, err := web.AppConfig.String("postgres")
	if err != nil {
		return err
	}
	maxIdle := 200
	maxConn := 200
	// 注册数据库
	if err := orm.RegisterDataBase("default", "postgres", url, orm.MaxIdleConnections(maxIdle), orm.MaxOpenConnections(maxConn)); err != nil {
		return err
	}

	orm.RegisterModelWithPrefix("fly_",
		new(models.ListenMsgGasNetStatus),
		new(models.ListenRewardNetStatus),
		new(models.ExpendInfo),
		new(models.MinerInfo),
		new(models.ExpendMessages),
		new(models.PreAndProveMessages),
		new(models.MineMessages),
		new(models.MinerStatusAndDailyChange),
		new(models.MinerAndWalletRelation),
		new(models.MineBlockRight),
		new(models.AllMinersMined),
		new(models.AllMinersPower),
		new(models.WalletInfo),
		new(models.ReceiveMessages),
	)
	if err := orm.RunSyncdb("default", false, true); err != nil {
		return err
	}
	//lotus.InitMinerData()
	return nil
}
