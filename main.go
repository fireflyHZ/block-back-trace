package main

import (
	"errors"
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	_ "github.com/lib/pq"
	"os"
	"profit-allocation/lotus/mine"
	"profit-allocation/models"
)

func main() {
	if err := initDatabase(); err != nil {
		fmt.Println("init database error:", err)
		return
	}
	mine.CalculateMineRight()
	//reward.Wakaka()
}

//初始化mysql
func initDatabase() error {
	// 注册数据库驱动
	if err := orm.RegisterDriver("postgres", orm.DRPostgres); err != nil {
		return err
	}

	url := os.Getenv("POSTGRES")
	if url == "" {
		return errors.New("get POSTGRES error")
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
		new(models.CalculateMineRightStatus),
		new(models.MineRight),
	)
	if err := orm.RunSyncdb("default", false, true); err != nil {
		return err
	}

	return nil
}
