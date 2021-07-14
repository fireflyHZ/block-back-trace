package models

import (
	"errors"
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	logging "github.com/ipfs/go-log/v2"
)

var (
	Wallets        = make(map[string]string)
	Miners         = make(map[string]int)
	LotusHost      string
	LotusSignToken string
)

var log = logging.Logger("models")

func InitData() error {
	minerInfos := make([]MinerInfo, 0)
	o := orm.NewOrm()
	num, err := o.QueryTable("fly_miner_info").All(&minerInfos)
	if err != nil {
		log.Errorf("get miner and wallet relation info err:%+v\n", err)
		return err
	}
	if num == 0 {
		log.Error("get miner and wallet relation info's number is 0")
		return errors.New("get miner and wallet relation info's number is 0")
	}
	miners := make(map[string]int)
	//wallets := make(map[string]int)
	for _, info := range minerInfos {
		miners[info.MinerId] = 1
		//	wallets[info.WalletId] = 2
	}
	Miners = miners
	//Wallets = wallets
	LotusHost, err = web.AppConfig.String("lotusHost")
	if err != nil {
		log.Errorf("get lotusHost  err:%+v\n", err)
		return err
	}
	LotusSignToken, err = web.AppConfig.String("LotusSignToken")
	if err != nil {
		log.Errorf("get lotusHost  err:%+v\n", err)
		return err
	}
	return nil
}
