package controllers

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"profit-allocation/models"
	"profit-allocation/tool/data"
	"profit-allocation/tool/log"
)

type OrderController struct {
	beego.Controller
}

type OrdersInfo struct {
	Infos []*models.OrderInfo
	AllShare int
}

func (c *OrderController) OrderInfo() {
	ordersInfo := new(OrdersInfo)
	//c.GetString("sdf")
	//fmt.Printf("-------body info :%+v\n",string(c.Ctx.Input.RequestBody))

	err := json.Unmarshal(c.Ctx.Input.RequestBody, ordersInfo)
	if err != nil {
		log.Logger.Error("user info unmarshal err:%+v", err)
		return
	}
	//fmt.Printf("-------orders info :%+v\n",ordersInfo)
	o := orm.NewOrm()

	for _, info := range ordersInfo.Infos {
		order := new(models.OrderInfo)
		n, err := o.QueryTable("fly_order_info").Filter("order_id", info.OrderId).All(order)
		if err != nil {
			log.Logger.Error("query user :%+v info  err:%+v", info.UserId, err)
			resp := data.PostResp{
				Code: "failed",
				Msg:  "query user info error",
			}
			c.Data["result"] = resp
			return
		}
		if n == 0 {
			_, err := o.Insert(info)
			if err != nil {
				log.Logger.Error("insert user :%+v info  err:%+v", info.UserId, err)
				resp := data.PostResp{
					Code: "failed",
					Msg:  "insert user info error",
				}
				c.Data["result"] = resp
				return
			}
		} else {
			order.Share = info.Share
			_, err := o.Update(order)
			if err != nil {
				log.Logger.Error("update user :%+v info  err:%+v", info.UserId, err)
				resp := data.PostResp{
					Code: "failed",
					Msg:  "update user info error",
				}
				c.Data["result"] = resp
				return
			}
		}
	}
	//更新总份额
	orders:=make([]models.OrderInfo,0)
	_, err = o.QueryTable("fly_order_info").All(&orders)
	if err != nil {
		log.Logger.Error("query orders info  err:%+v", err)
		resp := data.PostResp{
			Code: "failed",
			Msg:  "query all orders info error",
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	}
	total:=0
	for _,order:=range orders{
	total+=order.Share
	}

	//fmt.Println("----total ",total)
	netRunData := new(models.NetRunDataPro)
	_, err = o.QueryTable("fly_net_run_data_pro").All(netRunData)
	if err != nil {
		log.Logger.Error("query net run data info  err:%+v", err)
		resp := data.PostResp{
			Code: "failed",
			Msg:  "query net run data info error",
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	}
	netRunData.AllShare=ordersInfo.AllShare
	netRunData.TotalShare=total
	_,err=o.Update(netRunData)
	if err != nil {
		log.Logger.Error("update  net run data info  err:%+v", err)
		resp := data.PostResp{
			Code: "failed",
			Msg:  "update net run data info error",
		}
		c.Data["json"] = &resp
		c.ServeJSON()
		return
	}
	resp := data.PostResp{
		Code: "ok",
		Msg:  "update order info success",
	}
	c.Data["json"] = &resp
	c.ServeJSON()
	return

}
