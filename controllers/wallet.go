package controllers

import (
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	logging "github.com/ipfs/go-log/v2"
	"profit-allocation/models"
)

type WalletController struct {
	web.Controller
}

var walletLog = logging.Logger("wallet-ctr-log")

func (c *WalletController) QueryWalletBalance() {
	walletId := c.GetString("wallet_id")
	if walletId == "" {
		resp := models.WalletInfoResp{
			Code:          "faile",
			Msg:           "wallet id is nil",
			WalletBalance: "",
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	}
	walletLog.Debug("DEBUG: QueryWalletBalance()  walletId:%+v", walletId)
	walletInfo := make([]models.WalletBaseinfo, 0)
	o := orm.NewOrm()
	num, err := o.QueryTable("wallet_baseinfo").Filter("wallet_id", walletId).All(&walletInfo)
	if err != nil || num == 0 {
		resp := models.WalletInfoResp{
			Code:          "faile",
			Msg:           "haven't this wallet info",
			WalletBalance: "",
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	} else {
		resp := models.WalletInfoResp{
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
		resp := models.WalletProfitResp{
			Code:   "faile",
			Msg:    "wallet id is nil",
			Amount: "",
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	}
	if date == "" {
		resp := models.WalletProfitResp{
			Code:   "faile",
			Msg:    "settlement_date  is nil",
			Amount: "",
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	}
	if typ == "" {
		resp := models.WalletProfitResp{
			Code:   "faile",
			Msg:    "settlement_type  is nil",
			Amount: "",
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	}

	walletLog.Debug("DEBUG: QueryWalletProfit() walletId: %+v -- date: %+v -- type: %+v", walletId, date, typ)
	walletProfitInfo := make([]models.WalletProfitInfo, 0)
	o := orm.NewOrm()
	q := o.QueryTable("fly_wallet_profit_info").Filter("wallet_id", walletId)
	num, err := q.Filter("settlement_date", date).Filter("settlement_type", typ).All(&walletProfitInfo)

	if err != nil || num == 0 {
		resp := models.WalletProfitResp{
			Code: "faile",
			Msg:  "haven't this wallet info",
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	} else {
		resp := models.WalletProfitResp{
			Code:   "ok",
			Msg:    "successful",
			Amount: walletProfitInfo[0].Amount,
		}
		c.Data["json"] = &resp
		c.ServeJSON()
	}
	return
}
