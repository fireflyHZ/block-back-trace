package reward

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/lotus/api"
	lotusClient "github.com/filecoin-project/lotus/api/client"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/vm"
	"github.com/filecoin-project/specs-actors/v2/actors/builtin"
	cid "github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"
	"net/http"
	"profit-allocation/models"
	"profit-allocation/tool/bit"
	"profit-allocation/tool/sync"
	"strconv"
	"time"
)

var rewardLog = logging.Logger("reward-log")
var Client api.FullNode

func CollectTotalRerwardAndPledge() {
	defer sync.Wg.Done()
	var rewardBlockHeight int64
	if height, err := queryListenRewardNetStatus(); err != nil {
		rewardLog.Errorf(" collect total rerward and pledge, err=%v", err)
		return
	} else {
		rewardBlockHeight = height
	}
	rewardLog.Infof(" collect total rerward and pledge rewardBlockHeight:%v ", rewardBlockHeight)
	chainHeadResp, err := collectLotusChainHeadBlock()
	if err != nil {
		rewardLog.Errorf("collect lotus chain head block err=%v", err)
		return
	}

	blockHeight := int64(chainHeadResp.Height())
	if blockHeight <= rewardBlockHeight+11 {
		return
	}

	if blockHeight-rewardBlockHeight > 200 {
		//	rewardLog.Debug("DEBUG: collectLotusChainBlockRunData()  >200")

		h, err := getRewardAndPledge(rewardBlockHeight+200, rewardBlockHeight)
		if err != nil {
			rewardLog.Errorf("get reward and pledge handle > 200 err:%+v", err)
			return
		}
		rewardBlockHeight = h
	} else {
		h, err := getRewardAndPledge(blockHeight, rewardBlockHeight)
		if err != nil {
			rewardLog.Errorf("get reward and pledge handle < 200 err:%+v", err)
			return
		}
		rewardBlockHeight = h
	}

	err = updateListenRewardNetStatus(rewardBlockHeight)
	if err != nil {
		rewardLog.Errorf("update net run data height:%+v err :%+v\n", rewardBlockHeight, err)
	}
}

func queryListenRewardNetStatus() (height int64, err error) {
	o := orm.NewOrm()
	netRunData := new(models.ListenRewardNetStatus)
	n, err := o.QueryTable("fly_listen_reward_net_status").All(netRunData)
	if err != nil {
		return
	}
	if n == 0 {
		height = 460080
		return
	} else {
		height = netRunData.ReceiveBlockHeight
	}
	return
}

func updateListenRewardNetStatus(height int64) (err error) {
	o := orm.NewOrm()
	netRunData := new(models.ListenRewardNetStatus)
	n, err := o.QueryTable("fly_listen_reward_net_status").All(netRunData)
	if err != nil {
		return
	}
	if n == 0 {
		netRunData.ReceiveBlockHeight = height
		netRunData.CreateTime = time.Now()
		netRunData.UpdateTime = time.Now()
		_, err = o.Insert(netRunData)
		if err != nil {
			return err
		}

	} else {
		netRunData.ReceiveBlockHeight = height
		netRunData.UpdateTime = time.Now()
		_, err = o.Update(netRunData)
		if err != nil {
			return err
		}
	}
	return nil
}

func getRewardAndPledge(dealBlcokHeight int64, end int64) (int64, error) {

	chainHeightHandle, err := getChainHeadByHeight(end)
	if err != nil {
		rewardLog.Errorf("get chain head by height :%+v err :%+v", dealBlcokHeight-1, err)
		return end, err
	}

	dh := dealBlcokHeight

	for i := end; i < dealBlcokHeight; i++ {
		chainHeightAfter, err := getChainHeadByHeight(i + 1)
		if err != nil {
			rewardLog.Errorf("get after chain head by height:%+v err=%+v", dealBlcokHeight, err)
			return i, err
		}

		blockMessageResp, err := getParentsBlockMessage(chainHeightAfter.Cids()[0])
		if err != nil {
			rewardLog.Errorf("get parents block message cid %s  err=%v", chainHeightAfter.Cids()[0].String(), err)
			return i, err
		}

		blocks := chainHeightHandle.Blocks()
		bFlag := true
		for index, block := range blocks {

			if inMiners(block.Miner.String()) {
				err = calculateRewardAndPledge(index, blocks, chainHeightHandle.Cids(), chainHeightHandle.Key(), chainHeightAfter.Cids()[0], blockMessageResp)
				if err != nil {
					rewardLog.Errorf("calculate reward and pledge height:%+v err=%+v", end, err)
					return i, err
				}
				bFlag = false
			}

		}

		//pledge && power
		//没出块，且当前高度有块
		if bFlag && len(blocks) > 0 {
			t := time.Unix(int64(blocks[0].Timestamp), 0)
			if isExecutingPoint(t) {
				err = calculatePowerAndPledge(blocks[0].Height, chainHeightHandle.Key(), int64(blocks[0].Timestamp))
				if err != nil {
					rewardLog.Errorf("calculate power and pledge height:%+v err=%+v", end, err)
					return i, err
				}
			}
		}

		//parentTipsetKey := chainHeightHandle.Parents()
		chainHeightHandle = chainHeightAfter
		bFlag = true
	}
	return dh, nil
}

func isExecutingPoint(nowDatetime time.Time) bool {
	_, nowMin, _ := nowDatetime.Clock()
	if nowMin == 59 {
		return true
	} else {
		return false
	}
}

func calculatePowerAndPledge(height abi.ChainEpoch, tipsetKey types.TipSetKey, tStamp int64) error {
	t := time.Unix(tStamp, 0)
	tStr := t.Format("2006-01-02")
	epoch := int64(height)
	//获取质押
	o := orm.NewOrm()
	txOrm, err := o.Begin()
	if err != nil {
		rewardLog.Errorf("collectWalletData orm transation begin error: %+v", err)
		return err
	}
	for _, miner := range models.Miners {

		available, preCommit, vesting, pleage, sectorCount, err := GetMienrPleage(miner, height)
		if err != nil {
			if fmt.Sprintf("%+v", err) == "actor not found" {
				rewardLog.Warnf("GetMienrPleage ParseFloat miner:%+v height:%+v err:%+v", miner, height, err)
				continue
			}
			rewardLog.Errorf("GetMienrPleage ParseFloat miner:%+v height:%+v err:%+v", miner, height, err)
			err := txOrm.Rollback()
			if err != nil {
				rewardLog.Errorf("collectWalletData orm transation rollback error: %+v", err)
			}
			return err
		}
		power, precentage, err := getMinerPower(miner, tipsetKey)
		if err != nil {
			if fmt.Sprintf("%+v", err) == "actor not found" {
				rewardLog.Warnf("GetMienrPleage getMinerPower miner:%+v height:%+v err:%+v", miner, height, err)
				continue
			}
			rewardLog.Errorf(" GetMienrPleage getMinerPower miner:%+v height:%+v err:%+v", miner, height, err)
			err := txOrm.Rollback()
			if err != nil {
				rewardLog.Errorf("collectWalletData orm transation rollback error: %+v", err)
			}
			return err
		}
		//	rewardLog.Debug("------gas:%+v,mine:%+v,penalty:%+v,value:%+v", gas, mine, penalty, value)
		err = putMinerPowerStatus(txOrm, miner, power, available, preCommit, vesting, pleage, precentage, sectorCount, epoch, t)
		if err != nil {
			rewardLog.Errorf(" GetMienrPleage putMinerPowerStatus miner:%+v height:%+v err:%+v", miner, height, err)
			err := txOrm.Rollback()
			if err != nil {
				rewardLog.Errorf("collectWalletData orm transation rollback error: %+v", err)
			}
			return err
		}
		//收益分配
		minerInfo := new(models.MinerInfo)
		n, err := txOrm.QueryTable("fly_miner_info").Filter("miner_id", miner).All(minerInfo)

		if err != nil {
			rewardLog.Errorf("QueryTable minerInfo:%+v err:%+v num:%+v ", miner, err, n)
			err := txOrm.Rollback()
			if err != nil {
				rewardLog.Errorf("collectMinerData orm transation rollback error: %+v", err)
			}
			return err
		}

		oldPleage := minerInfo.Pleage
		oldPower := minerInfo.QualityPower
		oldVesting := minerInfo.Vesting
		//rewardLog.Debug("-=-=-=-=-=-power:%+v old:%+v",power,oldPower)

		if n == 0 {
			return errors.New("get miner power  error")
		} else {
			//更新miner info
			minerInfo.Pleage = pleage
			minerInfo.QualityPower = power
			minerInfo.Vesting = vesting
			minerInfo.UpdateTime = t

			_, err := txOrm.Update(minerInfo)
			if err != nil {
				rewardLog.Errorf("Update minerInfo miner:%+v  err:%+v ", miner, err)
				err := txOrm.Rollback()
				if err != nil {
					rewardLog.Errorf("collectMinerData orm transation rollback error: %+v", err)
				}
				return err
			}
		}
		rewardInfos := make([]models.MinerStatusAndDailyChange, 0)
		//入库
		n, err = o.Raw("select * from fly_miner_status_and_daily_change where miner_id=? and update_time::date=to_date(?,'YYYY-MM-DD')", miner, tStr).QueryRows(&rewardInfos)
		//n, err = txOrm.QueryTable("fly_reward_info").Filter("miner_id", miner).Filter("update_time", tStr).All(rewardInfo)
		if err != nil {
			rewardLog.Errorf("QueryTable rewardInfo:%+v err:%+v num:%+v time:%+v", miner, err, n, tStr)
			err := txOrm.Rollback()
			if err != nil {
				rewardLog.Errorf("calculatePowerAndPledge orm transation rollback error: %+v", err)
			}
			return err
		}
		rewardInfo := new(models.MinerStatusAndDailyChange)
		if n == 0 {
			//rewardInfo.Time = tStr
			rewardInfo.MinerId = miner
			rewardInfo.Pledge = pleage - oldPleage
			rewardInfo.Power = power - oldPower
			rewardInfo.Vesting = vesting - oldVesting
			rewardInfo.Reward = 0
			rewardInfo.Epoch = epoch
			rewardInfo.UpdateTime = t

			_, err = txOrm.Insert(rewardInfo)
			if err != nil {
				rewardLog.Errorf("Insert miner:%+v time:%+v err:%+v ", miner, tStr, err)
				err := txOrm.Rollback()
				if err != nil {
					rewardLog.Errorf("calculatePowerAndPledge orm transation rollback error: %+v", err)
				}
				return err
			}
		} else {
			rewardInfo = &rewardInfos[0]
			if rewardInfo.Epoch < epoch {

				rewardInfo.Pledge += pleage - oldPleage
				rewardInfo.Power += power - oldPower
				rewardInfo.Vesting += vesting - oldVesting
				rewardInfo.Epoch = epoch
				rewardInfo.UpdateTime = t
				_, err := txOrm.Update(rewardInfo)
				if err != nil {
					rewardLog.Errorf("Update miner:%+v time:%+v err:%+v ", miner, t, err)
					err := txOrm.Rollback()
					if err != nil {
						rewardLog.Errorf("calculatePowerAndPledge orm transation rollback error: %+v", err)
					}
					return err
				}
			}

		}
		rewardLog.Debugf("miner:%+v epoch:%+v  power:%+v  pledge:%+v", miner, epoch, rewardInfo.Power, rewardInfo.Pledge)

	}
	err = updateListenRewardNetStatus(epoch + 1)
	if err != nil {
		rewardLog.Errorf("Update net run data tmp  err:%+v height:%+v ", err, epoch)
		err := txOrm.Rollback()
		if err != nil {
			rewardLog.Errorf("calculatePowerAndPledge orm transation rollback error: %+v", err)
		}
		return err
	}
	err = txOrm.Commit()
	if err != nil {
		rewardLog.Errorf("calculatePowerAndPledge orm transation commit error: %+v", err)
		return err
	}
	rewardLog.Debug("calculate complete epoch:%+v ", epoch)
	return nil
}

func calculateRewardAndPledge(index int, blocks []*types.BlockHeader, blockCid []cid.Cid, tipsetKey types.TipSetKey, blockAfter cid.Cid, messages []api.Message) error {
	//获取minerid
	miner := blocks[index].Miner.String()
	o := orm.NewOrm()
	//查询数据
	txOrm, err := o.Begin()
	if err != nil {
		rewardLog.Errorf("collectWalletData orm transation begin error: %+v", err)
		return err
	}
	t := time.Unix(int64(blocks[0].Timestamp), 0)
	tStr := t.Format("2006-01-02")
	//epoch := blocks[0].Height.String()
	epoch := int64(blocks[0].Height)
	//rewardLog.Debug("Debug collectMinertData height:%+v", epoch)
	winCount := blocks[index].ElectionProof.WinCount
	value, power, percentage, err := calculateReward(index, blocks[index].Miner, blockCid, tipsetKey, blocks[index].ParentBaseFee, winCount, blocks[index], blockAfter, messages, epoch)
	//	rewardLog.Debug("------gas:%+v,mine:%+v,penalty:%+v,value:%+v", gas, mine, penalty, value)
	if err != nil {
		errTx := txOrm.Rollback()
		if errTx != nil {
			rewardLog.Errorf("collectWalletData orm transation rollback error: %+v", errTx)
		}
		return err
	}

	//获取质押
	available, preCommit, vesting, pleage, sectorCounts, err := GetMienrPleage(miner, blocks[0].Height)
	if err != nil {
		rewardLog.Errorf("GetMienrPleage ParseFloat err:%+v", err)
		errTx := txOrm.Rollback()
		if errTx != nil {
			rewardLog.Debug("collectWalletData orm transation rollback error: %+v", errTx)
		}
		return err
	}
	//	rewardLog.Debug("------gas:%+v,mine:%+v,penalty:%+v,value:%+v", gas, mine, penalty, value)
	//新增存储当天miner power状态、MinerPowerStatus
	err = putMinerPowerStatus(txOrm, miner, power, available, preCommit, vesting, pleage, percentage, sectorCounts, epoch, t)
	if err != nil {
		errTx := txOrm.Rollback()
		if errTx != nil {
			rewardLog.Debug("collectWalletData orm transation rollback error: %+v", errTx)
		}
		return err
	}
	//收益分配
	minerInfo := new(models.MinerInfo)
	n, err := txOrm.QueryTable("fly_miner_info").Filter("miner_id", miner).All(minerInfo)

	if err != nil {
		rewardLog.Errorf("QueryTable minerInfo:%+v err:%+v num:%+v ", miner, err, n)
		errTx := txOrm.Rollback()
		if errTx != nil {
			rewardLog.Debug("collectMinerData orm transation rollback error: %+v", errTx)
		}
		return err
	}

	oldPleage := minerInfo.Pleage
	oldPower := minerInfo.QualityPower
	oldVesting := minerInfo.Vesting
	//rewardLog.Debug("-=-=-=-=-=-power:%+v old:%+v",power,oldPower)

	if n == 0 {
		return errors.New("get miner power  error")
	} else {
		//更新miner info
		minerInfo.Pleage = pleage
		minerInfo.QualityPower = power
		minerInfo.Vesting = vesting
		minerInfo.UpdateTime = t

		_, err := txOrm.Update(minerInfo)
		if err != nil {
			rewardLog.Errorf("Update minerInfo miner:%+v  err:%+v ", miner, err)
			errTx := txOrm.Rollback()
			if errTx != nil {
				rewardLog.Debug("collectMinerData orm transation rollback error: %+v", errTx)
			}
			return err
		}
	}
	//记录出块
	mineBlock := new(models.MineBlocks)
	n, err = txOrm.QueryTable("fly_mine_blocks").Filter("miner_id", miner).Filter("epoch", epoch).All(mineBlock)
	if err != nil {
		rewardLog.Errorf("QueryTable mine blocks :%+v err:%+v num:%+v ", miner, err, n)
		errTx := txOrm.Rollback()
		if errTx != nil {
			rewardLog.Debug("collectMinerData orm transation rollback error: %+v", errTx)
		}
		return err
	}
	if n == 0 {
		mineBlock.MinerId = miner
		mineBlock.Epoch = epoch
		mineBlock.Reward = value
		mineBlock.WinCount = winCount
		mineBlock.CreateTime = t

		_, err = txOrm.Insert(mineBlock)
		if err != nil {
			rewardLog.Errorf("Update minerInfo miner:%+v  err:%+v ", miner, err)
			errTx := txOrm.Rollback()
			if errTx != nil {
				rewardLog.Debug("collectMinerData orm transation rollback error: %+v", errTx)
			}
			return err
		}
	} else {
		rewardLog.Warnf(" blocks has mined  :%+v epoch:%+v num:%+v ", miner, epoch, n)
		errTx := txOrm.Rollback()
		if errTx != nil {
			rewardLog.Debug("GcollectMinerData orm transation rollback error: %+v", errTx)
		}
		return nil
	}

	rewardInfos := make([]models.MinerStatusAndDailyChange, 0)
	//入库
	n, err = o.Raw("select * from fly_miner_status_and_daily_change where miner_id=? and update_time::date=to_date(?,'YYYY-MM-DD')", miner, tStr).QueryRows(&rewardInfos)
	//n, err = txOrm.QueryTable("fly_reward_info").Filter("miner_id", miner).Filter("time", tStr).All(rewardInfo)
	if err != nil {
		rewardLog.Errorf("QueryTable rewardInfo:%+v err:%+v num:%+v time:%+v", miner, err, n, tStr)
		errTx := txOrm.Rollback()
		if errTx != nil {
			rewardLog.Debug("collectWalletData orm transation rollback error: %+v", errTx)
		}
		return err
	}
	rewardInfo := new(models.MinerStatusAndDailyChange)
	if n == 0 {
		//记录块收益
		//rewardInfo.Time = tStr
		rewardInfo.Epoch = epoch
		rewardInfo.MinerId = miner
		rewardInfo.Pledge = pleage - oldPleage
		rewardInfo.Power = power - oldPower
		rewardInfo.Vesting = vesting - oldVesting
		rewardInfo.Reward = value
		rewardInfo.TotalReward += value
		rewardInfo.BlockNum = 1
		rewardInfo.TotalBlockNum += 1
		rewardInfo.WinCounts = winCount
		rewardInfo.TotalWinCounts += winCount
		rewardInfo.UpdateTime = t

		_, err = txOrm.Insert(rewardInfo)
		if err != nil {
			rewardLog.Errorf("Insert miner:%+v time:%+v err:%+v ", miner, t, err)
			errTx := txOrm.Rollback()
			if errTx != nil {
				rewardLog.Errorf("collectWalletData orm transation rollback error: %+v", errTx)
			}
			return err
		}
	} else {
		//记录块收益
		rewardInfo = &rewardInfos[0]
		//更新walletinfo
		if rewardInfo.Epoch < epoch {

			rewardInfo.Pledge += pleage - oldPleage
			rewardInfo.Power += power - oldPower
			rewardInfo.Vesting += vesting - oldVesting
			rewardInfo.Reward += value
			rewardInfo.TotalReward += value
			rewardInfo.Epoch = epoch
			rewardInfo.BlockNum += 1
			rewardInfo.TotalBlockNum += 1
			rewardInfo.WinCounts += winCount
			rewardInfo.TotalWinCounts += winCount
			rewardInfo.UpdateTime = t
			_, err := txOrm.Update(rewardInfo)
			if err != nil {
				rewardLog.Errorf("Update miner:%+v time:%+v err:%+v ", miner, t, err)
				errTx := txOrm.Rollback()
				if errTx != nil {
					rewardLog.Errorf("collectWalletData orm transation rollback error: %+v", errTx)
				}
				return err
			}
		}

	}
	err = updateListenRewardNetStatus(epoch + 1)
	if err != nil {
		rewardLog.Errorf("Update net run data tmp  err:%+v height:%+v ", err, epoch)
		errTx := txOrm.Rollback()
		if errTx != nil {
			rewardLog.Errorf("collectWalletData orm transation rollback error: %+v", errTx)
		}
		return err
	}
	err = txOrm.Commit()
	if err != nil {
		rewardLog.Errorf("collectWalletData orm transation commit error: %+v", err)
		return err
	}
	rewardLog.Debugf("miner:%+v epoch:%+v reward:%+v total:%+v", miner, epoch, value, rewardInfo.Reward)
	return nil
}

func calculateReward(index int, miner address.Address, blockCid []cid.Cid, tipsetKey types.TipSetKey, basefee abi.TokenAmount, winCount int64, header *types.BlockHeader, blockAfter cid.Cid, msgs []api.Message, height int64) (float64, float64, float64, error) {
	o := orm.NewOrm()
	totalGas := abi.NewTokenAmount(0)
	mineReward := abi.NewTokenAmount(0)
	totalPenalty := abi.NewTokenAmount(0)
	//requestHeader := http.Header{}
	ctx := context.Background()
	rewardMap := make(map[string]gasAndPenalty)
	allRewardMap := make(map[string]gasAndPenalty)
	base := gasAndPenalty{
		gas:     abi.NewTokenAmount(0),
		penalty: abi.NewTokenAmount(0),
	}

	for i := index; i >= 0; i-- {
		if i == index {
			messages, err := Client.ChainGetBlockMessages(context.Background(), blockCid[i])
			if err != nil {
				rewardLog.Errorf("getRewardInfo ChainGetBlockMessages err:%+v", err)
				return 0, 0, 0, err
			}
			for _, message := range messages.BlsMessages {
				rewardMap[message.Cid().String()] = base
			}
		} else {
			messages, err := Client.ChainGetBlockMessages(context.Background(), blockCid[i])
			if err != nil {
				rewardLog.Errorf("getRewardInfo ChainGetBlockMessages err:%+v", err)
				return 0, 0, 0, err
			}
			for _, message := range messages.BlsMessages {
				_, ok := rewardMap[message.Cid().String()]
				if ok {
					delete(rewardMap, message.Cid().String())
				}
			}
		}

	}

	for i, message := range msgs {
		//rewardLog.Debug("======i:%+v msgID:%+v len:%+v", i, message.Cid.String(), len(msgs))
		gasout, err := getGasout(blockAfter, message.Message, basefee, i, height)
		if err != nil {
			return 0, 0, 0, err
		}
		//	rewardLog.Debug("7777777 gas:%+v",gasout)
		gasPenalty := gasAndPenalty{
			gas:     gasout.MinerTip,
			penalty: gasout.MinerPenalty,
		}
		allRewardMap[message.Cid.String()] = gasPenalty
	}

	for msgId, _ := range rewardMap {
		//记录收益的message
		var msgGas abi.TokenAmount
		var msgPenalty abi.TokenAmount
		if gas, ok := allRewardMap[msgId]; ok {
			msgGas = gas.gas
			msgPenalty = gas.penalty
		} else {
			msgGas = rewardMap[msgId].gas
			msgPenalty = rewardMap[msgId].penalty
		}

		totalGas = big.Add(msgGas, totalGas)
		totalPenalty = big.Add(msgPenalty, totalPenalty)
		gasFlo, err := strconv.ParseFloat(bit.TransFilToFIL(msgGas.String()), 64)
		if err != nil {
			rewardForLog.Errorf("parse gas to float err:%+v", err)
			return 0, 0, 0, err
		}
		penFlo, err := strconv.ParseFloat(bit.TransFilToFIL(msgPenalty.String()), 64)
		if err != nil {
			rewardForLog.Errorf("parse penalty to float err:%+v", err)
			return 0, 0, 0, err
		}
		mineMsg := new(models.MineMessages)
		mineMsg.MinerId = miner.String()
		mineMsg.MessageId = msgId
		mineMsg.Gas = gasFlo
		mineMsg.Penalty = penFlo
		mineMsg.Epoch = int64(header.Height)
		mineMsg.CreateTime = time.Unix(int64(header.Timestamp), 0)

		_, err = o.Insert(mineMsg)
		if err != nil {
			rewardForLog.Errorf("inert msg:%+v err:%+v", msgId, err)
			return 0, 0, 0, err
		}
	}

	power, err := Client.StateMinerPower(ctx, miner, tipsetKey)
	if err != nil {
		rewardLog.Errorf("StateMinerPower err:%+v", err)
		return 0, 0, 0, err
	}
	var f float64 = 1024
	minerPower := float64(power.MinerPower.QualityAdjPower.Int64()) / f / f / f / f
	percentage := float64(power.MinerPower.QualityAdjPower.Int64()) / float64(power.TotalPower.QualityAdjPower.Int64())
	rewardActor, err := Client.StateGetActor(ctx, builtin.RewardActorAddr, tipsetKey)
	if err != nil {
		rewardLog.Errorf("StateGetActor err:%+v", err)
		return 0, 0, 0, err
	}

	rewardStateRaw, err := Client.ChainReadObj(ctx, rewardActor.Head)
	if err != nil {
		rewardLog.Errorf("ChainReadObj err:%+v", err)
		return 0, 0, 0, err
	}

	r := bytes.NewReader(rewardStateRaw)
	rewardActorState := unmarshalState(r)

	mineReward = big.Div(rewardActorState.ThisEpochReward, abi.NewTokenAmount(5))
	mineReward = big.Mul(mineReward, abi.NewTokenAmount(winCount))
	//rewardLog.Infof("miner reward :%+v total gas :%+v total penalty:%+v",mineReward,totalGas,totalPenalty )
	value := big.Sub(big.Add(mineReward, totalGas), totalPenalty)
	if value.LessThan(abi.NewTokenAmount(0)) {
		value = abi.NewTokenAmount(0)
	}
	valueStr := bit.TransFilToFIL(value.String())
	totalValue, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		rewardLog.Errorf("parse value to float err:%+v", err)
		return 0, 0, 0, err
	}

	return totalValue, minerPower, percentage, nil
}

func getMinerPower(minerStr string, tipsetKey types.TipSetKey) (float64, float64, error) {
	//requestHeader := http.Header{}
	ctx := context.Background()
	miner, err := address.NewFromString(minerStr)
	if err != nil {
		rewardLog.Errorf("getMinerPower NewFromString err:%+v", err)
		return 0, 0, err
	}

	power, err := Client.StateMinerPower(ctx, miner, tipsetKey)
	if err != nil {
		rewardLog.Errorf("StateMinerPower err:%+v", err)
		return 0, 0, err
	}
	var f float64 = 1024
	minerPower := float64(power.MinerPower.QualityAdjPower.Int64()) / f / f / f / f
	percentage := float64(power.MinerPower.QualityAdjPower.Int64()) / float64(power.TotalPower.QualityAdjPower.Int64())
	return minerPower, percentage, nil
}

func TestCalculateReward() {

	var i int64 = 252707
	chainHeightHandle, err := getChainHeadByHeight(252707)
	if err != nil {
		rewardLog.Errorf("handleRequestInfo() getChainHeadByHeight err=%+v", err)
		return
	}
	var re float64
	n := 0
	for {

		if "2020-11-22" <= time.Unix(int64(chainHeightHandle.Blocks()[0].Timestamp), 0).Format("2006-01-02") {
			break
		}
		chainHeightAfter, err := getChainHeadByHeight(i + 1)
		if err != nil {
			rewardLog.Errorf("handleRequestInfo() getChainHeadByHeight height:%+v err=%+v", i, err)
			return
		}
		blockMessageResp, err := getParentsBlockMessage(chainHeightAfter.Cids()[0])
		if err != nil {
			rewardLog.Errorf("handleRequestInfo() getParentsBlockMessage cid %s  err=%v", chainHeightAfter.Cids()[0].String(), err)
			return
		}

		blocks := chainHeightHandle.Blocks()
		for index, block := range blocks {
			if inMiners(block.Miner.String()) {
				n++
				v, err := calculateRewardAndPledgeTest(index, blocks, chainHeightHandle.Cids(), chainHeightHandle.Key(), chainHeightAfter.Cids()[0], blockMessageResp)
				if err != nil {
					rewardLog.Errorf("handleRequestInfo() calculateMineReward height:%+v err=%+v", i, err)
					return
				}
				re += v
				rewardLog.Debug("height:%+v total reward:%+v,count:%+v", i, re, n)
			}
		}
		chainHeightHandle = chainHeightAfter
		i++

	}
	rewardLog.Debug("total ok")
}

func calculateRewardAndPledgeTest(index int, blocks []*types.BlockHeader, blockCid []cid.Cid, tipsetKey types.TipSetKey, blockAfter cid.Cid, messages []api.Message) (float64, error) {
	epoch := int64(blocks[0].Height)
	winCount := blocks[index].ElectionProof.WinCount
	value, _, _, err := calculateReward(index, blocks[index].Miner, blockCid, tipsetKey, blocks[index].ParentBaseFee, winCount, blocks[index], blockAfter, messages, epoch)
	if err != nil {
		return 0, err
	}
	rewardLog.Debug("miner:%+v,epoch:%+v,value:%+v,wincount:%+v", blocks[index].Miner, epoch, value, winCount)

	return value, nil
}

func putMinerPowerStatus(o orm.TxOrmer, miner string, power, available, preCommit, vesting, pleage, powerPercentage float64, sectorCounts *api.MinerSectors, epoch int64, t time.Time) error {
	minerPowerStatuss := make([]models.MinerStatusAndDailyChange, 0)
	num, err := o.Raw("select * from fly_miner_status_and_daily_change where miner_id=? and update_time::date=to_date(?,'YYYY-MM-DD')", miner, t.Format("2006-01-02")).QueryRows(&minerPowerStatuss)

	//num, err := o.QueryTable("fly_miner_status_and_daily_change").Filter("miner_id", miner).Filter("time", t).All(minerPowerStatus)
	if err != nil {
		rewardLog.Errorf("QueryTable miner power status :%+v err:%+v num:%+v ", miner, err, num)
		err := o.Rollback()
		if err != nil {
			rewardLog.Debug("collectMinerData orm transation rollback error: %+v", err)
		}
		return err
	}
	minerPowerStatus := new(models.MinerStatusAndDailyChange)
	if num == 0 {
		minerPowerStatus.Epoch = epoch
		minerPowerStatus.TotalPower = power
		minerPowerStatus.TotalAvailable = available
		minerPowerStatus.TotalPledge = pleage
		minerPowerStatus.TotalPreCommit = preCommit
		minerPowerStatus.TotalVesting = vesting
		minerPowerStatus.UpdateTime = t
		minerPowerStatus.MinerId = miner
		minerPowerStatus.LiveSectorsNumber = sectorCounts.Live
		minerPowerStatus.ActiveSectorsNumber = sectorCounts.Active
		minerPowerStatus.FaultySectorsNumber = sectorCounts.Faulty
		minerPowerStatus.PowerPercentage = powerPercentage
		_, err = o.Insert(minerPowerStatus)
		if err != nil {
			rewardLog.Errorf("InsertTable miner power status :%+v err:%+v ", miner, err)
			err := o.Rollback()
			if err != nil {
				rewardLog.Errorf("collectMinerData orm transation rollback error: %+v", err)
			}
			return err
		}
	} else {
		minerPowerStatus = &minerPowerStatuss[0]
		minerPowerStatus.Epoch = epoch
		minerPowerStatus.TotalPower = power
		minerPowerStatus.TotalAvailable = available
		minerPowerStatus.TotalPledge = pleage
		minerPowerStatus.TotalPreCommit = preCommit
		minerPowerStatus.TotalVesting = vesting
		minerPowerStatus.LiveSectorsNumber = sectorCounts.Live
		minerPowerStatus.ActiveSectorsNumber = sectorCounts.Active
		minerPowerStatus.FaultySectorsNumber = sectorCounts.Faulty
		minerPowerStatus.PowerPercentage = powerPercentage
		_, err = o.Update(minerPowerStatus)
		if err != nil {
			rewardLog.Errorf("UpdateTable miner power status :%+v err:%+v ", miner, err)
			err := o.Rollback()
			if err != nil {
				rewardLog.Errorf("collectMinerData orm transation rollback error: %+v", err)
			}
			return err
		}
	}
	return nil
}

func getChainHeadByHeight(height int64) (tipset *types.TipSet, err error) {
	epoch := abi.ChainEpoch(height)
	tipsetKey := types.NewTipSetKey()
	tipset, err = Client.ChainGetTipSetByHeight(context.Background(), epoch, tipsetKey)

	return
}

func collectLotusChainHeadBlock() (tipset *types.TipSet, err error) {
	tipset, err = Client.ChainHead(context.Background())
	return
}

func getParentsBlockMessage(cid cid.Cid) (messages []api.Message, err error) {
	messages, err = Client.ChainGetParentMessages(context.Background(), cid)
	return
}

func getGasout(blockCid cid.Cid, messages *types.Message, basefee abi.TokenAmount, i int, height int64) (gasout vm.GasOutputs, err error) {
	charge := true
	ctx := context.Background()
	resp, err := Client.ChainGetParentReceipts(ctx, blockCid)

	if err != nil {
		rewardForLog.Errorf("getGasout  ChainGetParentReceipts err:%+v", err)
		return
	}
	if messages.Method == 5 && height > 343200 {
		charge = false
	}
	gasout = vm.ComputeGasOutputs(resp[i].GasUsed, messages.GasLimit, basefee, messages.GasFeeCap, messages.GasPremium, charge)
	return
}

func CreateLotusClient() {
	var err error
	requestHeader := http.Header{}
	requestHeader.Add("Content-Type", "application/json")
	Client, _, err = lotusClient.NewFullNodeRPC(context.Background(), models.LotusHost, requestHeader)
	if err != nil {
		rewardLog.Errorf("create lotus client%+v,host:%+v", err, models.LotusHost)
		return
	}
}
