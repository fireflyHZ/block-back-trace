package main

import (
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	_ "github.com/go-sql-driver/mysql"
	"profit-allocation/lotus/reward"
	"profit-allocation/models"
	_ "profit-allocation/routers"
)

func main() {

	//if err := initDatabase(); err != nil {
	//	fmt.Println("init database error:", err)
	//	return
	//}
	//if err := models.InitData(); err != nil {
	//	fmt.Println("init data error:", err)
	//	return
	//}

	reward.TetsGetInfo()
	//go lotus.Setup()
	//controllers.InitData()
	//p,_:=reward.GetMienrPleage("f021704",195059)
	//fmt.Println("-----",p)

	web.Run()
}

//初始化mysql
func initDatabase() error {
	// 注册数据库驱动
	if err := orm.RegisterDriver("mysql", orm.DRMySQL); err != nil {
		return err
	}

	url, err := web.AppConfig.String("mysql")
	if err != nil {
		return err
	}

	// 注册数据库
	if err := orm.RegisterDataBase("default", "mysql", url); err != nil {
		return err
	}

	orm.RegisterModelWithPrefix("fly_",
		new(models.ListenMsgGasNetStatus),
		new(models.ListenRewardNetStatus),
		new(models.RewardInfo),
		new(models.ExpendInfo),
		new(models.MinerInfo),
		new(models.ExpendMessages),
		new(models.MineBlocks),
		new(models.MineMessages),
		new(models.MinerPowerStatus),
		new(models.MinerAndWalletRelation),
	)
	if err := orm.RunSyncdb("default", false, true); err != nil {
		return err
	}

	return nil
}
