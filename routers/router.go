package routers

import (
	"github.com/astaxie/beego"
	"profit-allocation/controllers"
)

func init() {

	ns :=
		beego.NewNamespace("/firefly",

			beego.NSNamespace("/profit",
				// /firefly/profit/wallet_balance
				beego.NSRouter("/wallet_balance", &controllers.WalletController{}, "get:QueryWalletBalance"),
				// /firefly/profit/wallet_profit
				beego.NSRouter("/wallet_profit", &controllers.WalletController{}, "get:QueryWalletProfit"),
				// /firefly/profit/reward_info
				beego.NSRouter("/reward_info", &controllers.RewardController{}, "get:QueryRewardInfo"),
				// /firefly/profit/order_reward_info
				beego.NSRouter("/order_reward_info", &controllers.RewardController{}, "get:QueryOrderDailyReward"),
				// /firefly/profit/order_info
				beego.NSRouter("/order_info", &controllers.OrderController{}, "post:OrderInfo"),
				// /firefly/profit/user_info
				beego.NSRouter("/user_info", &controllers.UserController{}, "get:GetUserInfo"),
				// /firefly/profit/user_daily_info
				beego.NSRouter("/user_daily_info", &controllers.UserController{}, "get:GetUserDailyInfo"),
				// /firefly/profit/total_reward_info
				beego.NSRouter("/total_reward_info", &controllers.RewardTmpController{}, "get:GetRewardAndPledge"),
				beego.NSRouter("/total_messages_gas_info", &controllers.RewardTmpController{}, "get:GetMessagesGas"),
				beego.NSRouter("/total_miner_info", &controllers.RewardTmpController{}, "get:GetMinerInfo"),
			),
		)

	beego.AddNamespace(ns)
}
