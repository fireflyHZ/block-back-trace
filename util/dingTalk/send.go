package dingTalk

import (
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"github.com/blinkbean/dingtalk"
	"profit-allocation/models"
	"strings"
)

func parseMineDingTalkData(miner, name string, epoch int64, winCount int64, t string) string {

	return fmt.Sprintf("#### [矿工丢块告警](%s)：\n\t- 矿工编号:   %s(%s)\n\t- 丢失高度:   %d\n\t- 丢失块数:   %d\n\t- 发生时间:   %s", models.GrafanaLink, miner, name, epoch, winCount, t)
}

func SendMineDingtalkData(miner string, epoch int64, winCount int64, t string) error {
	o := orm.NewOrm()
	mi := new(models.MinerInfo)
	_, err := o.QueryTable("fly_miner_info").Filter("miner_id", miner).All(mi)
	if err != nil {
		return err
	}
	msg := parseMineDingTalkData(miner, mi.Name, epoch, winCount, t)
	bot := dingtalk.InitDingTalkWithSecret(models.DingTalkToken, models.DingTalkSecret)

	if err := bot.SendMarkDownMessage("### 出块丢失：\n", msg); err != nil {
		return err
	}
	return nil
}

func parsePowerDingTalkData(mpi *models.MinerPartitionInfo, name string) string {
	tipset := strings.Split(mpi.TipsetTime, " ")
	open := strings.Split(mpi.DeadlineOpenTime, " ")
	closeT := strings.Split(mpi.DeadlineCloseTime, " ")
	return fmt.Sprintf("#### 时空证明告警：\n\t"+
		"- 矿工编号:   %s(%s)\n\t"+
		"- 网络当前高度:   %s\n\t"+
		"- 消耗时间:   %s\n\t"+
		"- 开启时间:   %s\n\t"+
		"- 关闭时间:   %s\n\t"+
		"- 窗口编号:   %d\n\t"+
		"- 扇区数量:   %d\n\t"+
		"- 错误扇区数量:   %d\n\t"+
		"- AllPartitions:   %d\n\t"+
		"- ProvenPartitions:   %d\n\t",
		mpi.Maddr, name, tipset[1], mpi.DeadlineElapsedTime,
		open[1], closeT[1], mpi.DeadlineIndex, mpi.AllSectors,
		mpi.FaultSectors, mpi.AllPartitions, mpi.ProvenPartitions)
}

func SendPowerDingtalkData(mpi *models.MinerPartitionInfo) error {
	o := orm.NewOrm()
	mi := new(models.MinerInfo)
	_, err := o.QueryTable("fly_miner_info").Filter("miner_id", mpi.Maddr).All(mi)
	if err != nil {
		fmt.Println(err)
		return err
	}
	msg := parsePowerDingTalkData(mpi, mi.Name)
	bot := dingtalk.InitDingTalkWithSecret(models.DingTalkToken, models.DingTalkSecret)
	if err := bot.SendMarkDownMessage("### 时空证明告警：\n", msg); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func parseCheckPowerErrorDingTalkData(errStr string) string {
	return fmt.Sprintf("#### 时空证明告警：\n\t"+
		"- 检查错误:   %s\n\t", errStr)
}

func SendCheckPowerErrorDingtalkData(errStr string) error {
	msg := parseCheckPowerErrorDingTalkData(errStr)
	bot := dingtalk.InitDingTalkWithSecret(models.DingTalkToken, models.DingTalkSecret)

	if err := bot.SendMarkDownMessage("### 时空证明告警：\n", msg); err != nil {
		return err
	}
	return nil
}

func SendTestPowerDingtalkData(mpi *models.MinerPartitionInfo) error {
	o := orm.NewOrm()
	mi := new(models.MinerInfo)
	_, err := o.QueryTable("fly_miner_info").Filter("miner_id", mpi.Maddr).All(mi)
	if err != nil {
		fmt.Println(err)
		return err
	}
	msg := parseTestPowerDingTalkData(mpi, mi.Name)
	bot := dingtalk.InitDingTalkWithSecret(models.DingTalkToken, models.DingTalkSecret)
	if err := bot.SendMarkDownMessage("### 时空证明告警：\n", msg); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func parseTestPowerDingTalkData(mpi *models.MinerPartitionInfo, name string) string {
	tipset := strings.Split(mpi.TipsetTime, " ")
	open := strings.Split(mpi.DeadlineOpenTime, " ")
	closeT := strings.Split(mpi.DeadlineCloseTime, " ")
	return fmt.Sprintf("#### 时空证明告警(test)：\n\t"+
		"- 矿工编号:   %s(%s)\n\t"+
		"- 网络当前高度:   %s\n\t"+
		"- 消耗时间:   %s\n\t"+
		"- 开启时间:   %s\n\t"+
		"- 关闭时间:   %s\n\t"+
		"- 窗口编号:   %d\n\t"+
		"- 扇区数量:   %d\n\t"+
		"- 错误扇区数量:   %d\n\t"+
		"- AllPartitions:   %d\n\t"+
		"- ProvenPartitions:   %d\n\t",
		mpi.Maddr, name, tipset[1], mpi.DeadlineElapsedTime,
		open[1], closeT[1], mpi.DeadlineIndex, mpi.AllSectors,
		mpi.FaultSectors, mpi.AllPartitions, mpi.ProvenPartitions)
}

func TestSendDingTalk() {
	err := SendMineDingtalkData("f02420", 123, 0, "2008-09-27")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("ok")
}
