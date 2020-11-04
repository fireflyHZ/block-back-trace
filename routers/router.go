package routers

import (
	"github.com/astaxie/beego"
	"profit-allocation/controllers"
)

func init() {
	//beego.Router("/firefly/profit/wallet_balance", &controllers.WalletController{}, "get:QueryWalletBalance")
	//beego.Router("/firefly/profit/wallet_profit", &controllers.WalletController{}, "get:QueryWalletProfit")
	//beego.Router("/firefly/profit/reward_info", &controllers.RewardController{}, "get:QueryRewardInfo")

	ns :=
		beego.NewNamespace("/firefly",

			beego.NSNamespace("/profit",
				// /firefly/profit/wallet_balance
				beego.NSRouter("/wallet_balance",  &controllers.WalletController{}, "get:QueryWalletBalance"),
				// /firefly/profit/wallet_profit
				beego.NSRouter("/wallet_profit",  &controllers.WalletController{}, "get:QueryWalletProfit"),
				// /firefly/profit/reward_info
				beego.NSRouter("/reward_info", &controllers.RewardController{}, "get:QueryRewardInfo"),
				beego.NSRouter("/order_reward_info", &controllers.RewardController{}, "get:QueryOrderDailyReward"),
				// /firefly/profit/user_info
				beego.NSRouter("/order_info", &controllers.OrderController{}, "post:OrderInfo"),
			),
		)

	beego.AddNamespace(ns)
}
