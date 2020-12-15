package controllers

import (
	"github.com/astaxie/beego/orm"
	"profit-allocation/models"
	"profit-allocation/tool/log"
)

func InitData()  {
	re:=new(models.MinerAndWalletRelation)
	re.MinerId="f02420"
	re.WalletId="f3va7lv4wkcfq5mmqirr4pyrogtnuknw2hma5y6luwbx6iv4qcwgrvzyn2zljgbgtmv7lxr3jsa4eo2az3kqra"
	o:=orm.NewOrm()
	_,err:=o.Insert(re)
	if err != nil {
		log.Logger.Error("insert miner:%+v wallet:%+v err:%+v",re.MinerId,re.WalletId,err)
		return
	}
	log.Logger.Debug("insert miner:%+v wallet:%+v success",re.MinerId,re.WalletId)
}