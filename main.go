package main

import (
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"os/signal"
	"profit-allocation/controllers"
	"profit-allocation/lotus"
	"profit-allocation/models"
	//_ "profit-allocation/routers"
	"syscall"
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

	//reward.TetsGetInfo()
	go lotus.Setup()
	var shutdownCh <-chan struct{}
	sigCh := make(chan os.Signal, 2)
	shutdownDone := make(chan struct{})
	go func() {
		select {
		case sig := <-sigCh:
			fmt.Println("received shutdown", "signal", sig)
		case <-shutdownCh:
			fmt.Println("received shutdown")
		}

		fmt.Println("Shutting down...")
		close(shutdownDone)
	}()
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	web.Router("/firefly/profit/total_reward_info", &controllers.RewardController{}, "get:GetRewardAndPledge")

	web.Router("/firefly/profit/total_messages_gas_info", &controllers.RewardController{}, "get:GetMessagesGas")

	web.Router("/firefly/profit/total_miner_info", &controllers.RewardController{}, "get:GetMinerInfo")

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
