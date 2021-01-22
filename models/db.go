package models

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"profit-allocation/tool/log"
)

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

// beego数据库连接初始化
func init() {
	// 注册数据库驱动
	if err := orm.RegisterDriver("mysql", orm.DRMySQL); err != nil {
		log.Logger.Error("Error RegisterDriver err:%+v", err)
		return
	}
	userName := beego.AppConfig.String("mysqluser")
	password := beego.AppConfig.String("mysqlpass")
	address := beego.AppConfig.String("mysqlurls")
	dbName := beego.AppConfig.String("mysqldb")
	url := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&allowOldPasswords=1",
		userName, password, address, dbName)

	// 注册数据库
	if err := orm.RegisterDataBase("default", "mysql", url); err != nil {
		//log.Logger.Error("Error RegisterDataBase err:%+v",err)
		fmt.Println("RegisterDataBase err:", err)
		return
	}

	orm.RegisterModelWithPrefix("fly_",
		new(RewardInfo),
		new(MessageRewardInfo),
		new(ExpendInfo),
		new(WalletBaseinfo),
		new(WalletHistoryData),
		new(WalletProfitInfo),
		new(MinerInfo),
		new(NetRunDataPro),
		new(ExpendMessages),
		new(RewardMessages),
		new(MineBlocks),
		new(MineMessages),
		new(UserInfo),
		new(UserBlockRewardInfo),
		new(UserDailyRewardInfo),
		new(OrderInfo),
		new(OrderBlockRewardInfo),
		new(OrderDailyRewardInfo),
		new(Orders),
		new(SettlePlan),
		new(OrderGoods),
		new(VestingInfo),
		new(UserFilDaily),
		new(UserFilPledge),
		new(Transfer),
		//---------------
		new(MinerInfoTmp),
		new(NetRunDataProTmp),
		new(RewardInfoTmp),
		new(MsgGasNetRunDataProTmp),
		//-----------------------
		new(MinerPowerStatus),
		new(MinerAndWalletRelation),
		new(ExpendInfoTmp),
		new(ExpendMessagesTmp),
	)
	if err := orm.RunSyncdb("default", false, true); err != nil {
		log.Logger.Error("Error RunSyncdb err:%+v", err)
		return
	}

	/*users:=make( []UserInfo,5)
	for i:=0;i<5;i++ {
		users[i].UserId=i+1
		users[i].Share=0.2
		users[i].Power=11
	}
	o:=orm.NewOrm()
	num,err:=o.InsertMulti(5,users)
	fmt.Println(num,err)
	*/

}
