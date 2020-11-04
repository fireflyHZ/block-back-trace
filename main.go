package main

import (
	"github.com/astaxie/beego"
	"profit-allocation/lotus"

	_ "profit-allocation/models"
	_ "profit-allocation/routers"
	_ "profit-allocation/tool"
	"profit-allocation/tool/log"
)

func main() {
	log.Init("profit.log", "debug")
	//reward.TetsGetInfo()
	go lotus.Setup()
	//reward.GetMienrPleage("f021704")
	beego.Run()
}
