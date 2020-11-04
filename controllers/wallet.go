package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"profit-allocation/models"
	"profit-allocation/tool/data"
	"profit-allocation/tool/log"
)

type WalletController struct {
	beego.Controller
}

func (c *WalletController) QueryWalletBalance() {
	walletId := c.GetString("wallet_id")
	if walletId == "" {
		resp := data.WalletInfoResp{
			Code:          "faile",
			Msg:           "wallet id is nil",
			WalletBalance: "",
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	}
	log.Logger.Debug("DEBUG: QueryWalletBalance()  walletId:%+v", walletId)
	walletInfo := make([]models.WalletBaseinfo, 0)
	o := orm.NewOrm()
	num, err := o.QueryTable("wallet_baseinfo").Filter("wallet_id", walletId).All(&walletInfo)
	if err != nil || num == 0 {
		resp := data.WalletInfoResp{
			Code:          "faile",
			Msg:           "haven't this wallet info",
			WalletBalance: "",
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	} else {
		resp := data.WalletInfoResp{
			Code:          "ok",
			Msg:           "successful",
			WalletBalance: walletInfo[0].BalanceFil,
		}
		c.Data["json"] = &resp
		c.ServeJSON()
	}

}

func (c *WalletController) QueryWalletProfit() {
	walletId := c.GetString("wallet_id")
	date := c.GetString("settlement_date")
	typ := c.GetString("settlement_type")
	if walletId == "" {
		resp := data.WalletProfitResp{
			Code:          "faile",
			Msg:           "wallet id is nil",
			Amount: "",
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	}
	if date == "" {
		resp := data.WalletProfitResp{
			Code:          "faile",
			Msg:           "settlement_date  is nil",
			Amount: "",
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	}
	if typ == "" {
		resp := data.WalletProfitResp{
			Code:          "faile",
			Msg:           "settlement_type  is nil",
			Amount: "",
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	}

	log.Logger.Debug("DEBUG: QueryWalletProfit() walletId: %+v -- date: %+v -- type: %+v", walletId, date, typ)
	walletProfitInfo := make([]models.WalletProfitInfo, 0)
	o := orm.NewOrm()
	q := o.QueryTable("fly_wallet_profit_info").Filter("wallet_id", walletId)
	num,err:=q.Filter("settlement_date",date).Filter("settlement_type",typ).All(&walletProfitInfo)

	if err != nil || num == 0 {
		resp := data.WalletProfitResp{
			Code:          "faile",
			Msg: "haven't this wallet info",
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	} else {
		resp := data.WalletProfitResp{
			Code:          "ok",
			Msg:           "successful",
			Amount: walletProfitInfo[0].Amount,
		}
		c.Data["json"] = &resp
		c.ServeJSON()
	}
	return
}
