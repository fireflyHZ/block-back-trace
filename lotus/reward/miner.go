package reward

import (
	"context"
	"github.com/beego/beego/v2/client/orm"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/lotus/chain/types"
	"math"
	"profit-allocation/lotus/client"
	"profit-allocation/models"
	"profit-allocation/tool/bit"
	"strconv"
	"time"
)

type MinerBalance struct {
	Address address.Address `json:"address"`
	Balance float64         `json:"balance"`
}

func QueryMinerAddressBalance(minerId string) (addrs map[string][]*MinerBalance, err error) {
	addrs, err = getMinerAddress(minerId)
	if err != nil {
		msgLog.Errorf("Get miner address error, miner:%+v err:%+v", minerId, err)
		return addrs, err
	}
	ctx := context.Background()
	for t, mbs := range addrs {
		for _, mb := range mbs {
			balance, err := client.Client.WalletBalance(ctx, mb.Address)
			if err != nil {
				msgLog.Errorf("Get address balance error, miner:%+v type:%+v address:%+v err:%+v", minerId, t, mb.Address, err)
				return addrs, err
			}
			balanceStr := bit.TransFilToFIL(balance.String())
			balanceFloat, err := strconv.ParseFloat(balanceStr, 64)
			mb.Balance = balanceFloat
		}
	}
	return addrs, err
}

func getMinerAddress(minerId string) (minerAddrs map[string][]*MinerBalance, err error) {
	minerAddrs = make(map[string][]*MinerBalance, 0)
	ctx := context.Background()
	tipset, err := client.Client.ChainHead(ctx)
	if err != nil {
		msgLog.Errorf("Get chain head error, miner:%+v err:%+v", minerId, err)
		return minerAddrs, err
	}
	minerAddr, err := address.NewFromString(minerId)
	if err != nil {
		msgLog.Errorf("NewFromString miner:%+v err:%+v", minerId, err)
		return minerAddrs, err
	}
	mi, err := client.Client.StateMinerInfo(ctx, minerAddr, tipset.Key())
	if err != nil {
		msgLog.Errorf("StateMinerInfo miner:%+v err:%+v", minerId, err)
		return minerAddrs, err
	}
	//owner address
	ownerAddr, err := client.Client.StateAccountKey(ctx, mi.Owner, types.EmptyTSK)
	if err != nil {
		msgLog.Errorf("state account owner key miner:%+v err:%+v", minerId, err)
		return minerAddrs, err
	}
	owner := []*MinerBalance{{
		Address: ownerAddr,
		Balance: 0,
	}}
	minerAddrs["owner"] = owner

	//worker address
	wokerAddr, err := client.Client.StateAccountKey(ctx, mi.Worker, types.EmptyTSK)
	if err != nil {
		msgLog.Errorf("state account worker key miner:%+v err:%+v", minerId, err)
		return minerAddrs, err
	}
	worker := []*MinerBalance{{
		Address: wokerAddr,
		Balance: 0,
	}}
	minerAddrs["worker"] = worker

	//control address
	for _, controlAddr := range mi.ControlAddresses {
		contorl, err := client.Client.StateAccountKey(ctx, controlAddr, types.EmptyTSK)
		if err != nil {
			msgLog.Errorf("state account control key miner:%+v err%+v:", minerId, err)
			return minerAddrs, err
		}
		cont := []*MinerBalance{{
			Address: contorl,
			Balance: 0,
		}}
		minerAddrs["control"] = cont
	}
	return minerAddrs, nil
}

func RecordAllMiners() {
	o := orm.NewOrm()
	ctx := context.Background()
Retry:
	for {
		head, err := client.Client.ChainHead(ctx)
		if err != nil {
			msgLog.Errorf("Get chain head error:%+v", err)
			sleep()
			continue
		}
		ms, err := client.Client.StateListMiners(ctx, head.Key())
		if err != nil {
			msgLog.Errorf("Get chain all miners error:%+v", err)
			sleep()
			continue
		}
		for _, miner := range ms {
			power, err := client.Client.StateMinerPower(ctx, miner, head.Key())
			if err != nil {
				msgLog.Errorf("Get chain miner:%+v power error:%+v", miner, err)
				sleep()
				goto Retry
			}
			GiB := math.Pow(1024, 3)
			gPower := power.MinerPower.QualityAdjPower.Int.Div(power.MinerPower.QualityAdjPower.Int, big.NewInt(int64(GiB)).Int)
			tPower := float64(gPower.Int64()) / 1024
			mi := new(models.AllMinersPower)
			num, err := o.QueryTable("fly_all_miners_power").Filter("miner_id", miner.String()).All(mi)
			if err != nil {
				msgLog.Errorf("Query table miner:%+v power error:%+v", miner, err)
				sleep()
				goto Retry
			}
			if num == 0 {
				mi.MinerId = miner.String()
				mi.Power = tPower
				mi.UpdateTime = time.Unix(int64(head.MinTimestamp()), 0)
				_, err := o.Insert(mi)
				if err != nil {
					msgLog.Errorf("insert table miner:%+v power error:%+v", miner, err)
					sleep()
					goto Retry
				}
			} else {
				mi.Power = tPower
				mi.UpdateTime = time.Unix(int64(head.MinTimestamp()), 0)
				_, err := o.Update(mi)
				if err != nil {
					msgLog.Errorf("update table miner:%+v power error:%+v", miner, err)
					sleep()
					goto Retry
				}
			}
		}
		time.Sleep(time.Hour * 6)
	}

}
func sleep() {
	time.Sleep(time.Second * 60)
}
