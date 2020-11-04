package tool

import (
	"github.com/astaxie/beego"
	"strings"
)

var Wallets []string
var Miners []string

func init()  {
	minersStr := beego.AppConfig.String("miners")
	Miners= strings.Split(minersStr, ",")
	walletsStr := beego.AppConfig.String("wallets")
	Wallets = strings.Split(walletsStr, ",")
}