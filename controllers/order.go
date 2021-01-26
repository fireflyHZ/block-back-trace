package controllers

import (
	"encoding/json"
	"github.com/beego/beego/v2/server/web"
	logging "github.com/ipfs/go-log/v2"
	"profit-allocation/models"
)

type OrderController struct {
	web.Controller
}

type OrdersInfo struct {
	Infos    []*models.OrderInfo
	AllShare int
}

var orderLog = logging.Logger("order-ctr-log")

func (c *OrderController) OrderInfo() {
	ordersInfo := new(OrdersInfo)
	//c.GetString("sdf")
	//fmt.Printf("-------body info :%+v\n",string(c.Ctx.Input.RequestBody))

	err := json.Unmarshal(c.Ctx.Input.RequestBody, ordersInfo)
	if err != nil {
		orderLog.Error("user info unmarshal err:%+v", err)
		return
	}
	//fmt.Printf("-------orders info :%+v\n",ordersInfo)

	for _, info := range ordersInfo.Infos {
		order := new(models.OrderInfo)
		n, err := models.O.QueryTable("fly_order_info").Filter("order_id", info.OrderId).All(order)
		if err != nil {
			orderLog.Error("query user :%+v info  err:%+v", info.UserId, err)
			resp := models.PostResp{
				Code: "failed",
				Msg:  "query user info error",
			}
			c.Data["result"] = resp
			return
		}
		if n == 0 {
			_, err := models.O.Insert(info)
			if err != nil {
				orderLog.Error("insert user :%+v info  err:%+v", info.UserId, err)
				resp := models.PostResp{
					Code: "failed",
					Msg:  "insert user info error",
				}
				c.Data["result"] = resp
				return
			}
		} else {
			order.Share = info.Share
			_, err := models.O.Update(order)
			if err != nil {
				orderLog.Error("update user :%+v info  err:%+v", info.UserId, err)
				resp := models.PostResp{
					Code: "failed",
					Msg:  "update user info error",
				}
				c.Data["result"] = resp
				return
			}
		}
	}
	//更新总份额
	orders := make([]models.OrderInfo, 0)
	_, err = models.O.QueryTable("fly_order_info").All(&orders)
	if err != nil {
		orderLog.Error("query orders info  err:%+v", err)
		resp := models.PostResp{
			Code: "failed",
			Msg:  "query all orders info error",
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	}
	total := 0
	for _, order := range orders {
		total += order.Share
	}

	//fmt.Println("----total ",total)
	netRunData := new(models.NetRunDataPro)
	_, err = models.O.QueryTable("fly_net_run_data_pro").All(netRunData)
	if err != nil {
		orderLog.Error("query net run data info  err:%+v", err)
		resp := models.PostResp{
			Code: "failed",
			Msg:  "query net run data info error",
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	}
	netRunData.AllShare = ordersInfo.AllShare
	netRunData.TotalShare = total
	_, err = models.O.Update(netRunData)
	if err != nil {
		orderLog.Error("update  net run data info  err:%+v", err)
		resp := models.PostResp{
			Code: "failed",
			Msg:  "update net run data info error",
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	}
	resp := models.PostResp{
		Code: "ok",
		Msg:  "update order info success",
	}
	c.Data["json"] = &resp
	c.ServeJSON()
	return

}
