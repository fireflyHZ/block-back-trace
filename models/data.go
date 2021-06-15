package models

import (
	"errors"
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	logging "github.com/ipfs/go-log/v2"
)

var (
	Wallets        map[string]int
	Miners         map[string]int
	LotusHost      string
	LotusSignToken string
)

var log = logging.Logger("models")

func InitData() error {
	minerAndWalletRelations := make([]MinerAndWalletRelation, 0)
	o := orm.NewOrm()
	num, err := o.QueryTable("fly_miner_and_wallet_relation").All(&minerAndWalletRelations)
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
	for _, info := range minerAndWalletRelations {
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
