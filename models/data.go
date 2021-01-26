package models

import (
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	logging "github.com/ipfs/go-log/v2"
	"strings"
)

var (
	Wallets   []string
	Miners    []string
	O         orm.Ormer
	LotusHost string
)

var log = logging.Logger("models")

func InitData() error {
	minersStr, err := web.AppConfig.String("miners")
	if err != nil {
		log.Errorf("get miners  err:%+v\n", err)
		return err
	}
	Miners = strings.Split(minersStr, ",")
	walletsStr, err := web.AppConfig.String("wallets")
	if err != nil {
		log.Errorf("get wallets  err:%+v\n", err)
		return err
	}
	Wallets = strings.Split(walletsStr, ",")
	LotusHost, err = web.AppConfig.String("lotusHost")
	if err != nil {
		log.Errorf("get lotusHost  err:%+v\n", err)
		return err
	}
	O = orm.NewOrm()
	return nil
}
