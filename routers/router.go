package routers

import (
	"github.com/beego/beego/v2/server/web"
	"profit-allocation/controllers"
)

func init() {

	ns :=
		web.NewNamespace("/firefly",

			web.NSNamespace("/profit",
				// /firefly/profit/wallet_balance
				web.NSRouter("/wallet_balance", &controllers.WalletController{}, "get:QueryWalletBalance"),
				// /firefly/profit/wallet_profit
				web.NSRouter("/wallet_profit", &controllers.WalletController{}, "get:QueryWalletProfit"),
				// /firefly/profit/reward_info

				// /firefly/profit/order_info
				web.NSRouter("/order_info", &controllers.OrderController{}, "post:OrderInfo"),
				// /firefly/profit/user_info
				web.NSRouter("/user_info", &controllers.UserController{}, "get:GetUserInfo"),
				// /firefly/profit/user_daily_info
				web.NSRouter("/user_daily_info", &controllers.UserController{}, "get:GetUserDailyInfo"),
				// /firefly/profit/total_reward_info
				web.NSRouter("/total_reward_info", &controllers.RewardTmpController{}, "get:GetRewardAndPledge"),
				web.NSRouter("/total_messages_gas_info", &controllers.RewardTmpController{}, "get:GetMessagesGas"),
				web.NSRouter("/total_miner_info", &controllers.RewardTmpController{}, "get:GetMinerInfo"),
			),
		)

	web.AddNamespace(ns)
}
