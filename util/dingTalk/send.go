package dingTalk

import (
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"github.com/blinkbean/dingtalk"
	"profit-allocation/models"
)

func parseDingTalkData(miner, name string, epoch int64, winCount int64, t string) string {

	return fmt.Sprintf("#### [矿工丢块告警](%s)：\n\t- 矿工编号: %s\n\t- 矿工名称: %s\n\t- 丢失高度: %d\n\t- 丢失块数: %d\n\t- 发生时间: %s", models.GrafanaLink, miner, name, epoch, winCount, t)
}

func SendDingtalkData(miner string, epoch int64, winCount int64, t string) error {
	o := orm.NewOrm()
	mi := new(models.MinerInfo)
	_, err := o.QueryTable("fly_miner_info").Filter("miner_id", miner).All(mi)
	if err != nil {
		return err
	}
	msg := parseDingTalkData(miner, mi.Name, epoch, winCount, t)
	bot := dingtalk.InitDingTalkWithSecret(models.DingTalkToken, models.DingTalkSecret)

	if err := bot.SendMarkDownMessage("### 出块丢失：\n", msg); err != nil {
		return err
	}
	return nil
}

func TestSendDingTalk() {
	err := SendDingtalkData("f02420", 123, 0, "2008-09-27")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("ok")
}
