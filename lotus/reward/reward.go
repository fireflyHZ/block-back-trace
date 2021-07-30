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
	"github.com/filecoin-project/lotus/chain/gen"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/vm"
	"github.com/filecoin-project/specs-actors/v2/actors/builtin"
	cid "github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"
	"profit-allocation/lotus/client"
	"profit-allocation/models"
	"profit-allocation/tool/bit"
	"profit-allocation/tool/sync"
	"strconv"
	"time"
)

var rewardLog = logging.Logger("reward-log")

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

	if blockHeight-rewardBlockHeight > 60000 {
		//	rewardLog.Debug("DEBUG: collectLotusChainBlockRunData()  >200")

		h, err := getRewardAndPledge(rewardBlockHeight+60000, rewardBlockHeight)
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
		height = 715000
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
		//查询数据
		chainHeightAfter, err := getChainHeadByHeight(i + 1)
		if err != nil {
			rewardLog.Errorf("get after chain head by height:%+v err=%+v", i, err)
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
		//计算出块权
		err = recordMineBlockRight(chainHeightHandle)
		if err != nil {
			rewardLog.Errorf("record mine block right height:%+v err=%+v", i, err)
			return i, err
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
	if nowMin == 59 || nowMin == 30 {
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
	for miner, _ := range models.Miners {
		//获取出块权数量
		rightNum, err := getBlockRightNum(miner, t)
		if err != nil {
			rewardLog.Errorf("collectWalletData orm transation begin error: %+v", err)
			return err
		}
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
		power, percentage, err := getMinerPower(miner, tipsetKey)
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
		//err = putMinerPowerStatus(txOrm, miner, power, available, preCommit, vesting, pleage, percentage, sectorCount, epoch, t)
		//if err != nil {
		//	rewardLog.Errorf(" GetMienrPleage putMinerPowerStatus miner:%+v height:%+v err:%+v", miner, height, err)
		//	err := txOrm.Rollback()
		//	if err != nil {
		//		rewardLog.Errorf("collectWalletData orm transation rollback error: %+v", err)
		//	}
		//	return err
		//}
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
		//rewardInfos := make([]models.MinerStatusAndDailyChange, 0)
		rewardInfo := new(models.MinerStatusAndDailyChange)

		//入库
		//n, err = o.Raw("select * from fly_miner_status_and_daily_change where miner_id=? and time=to_date(?,'YYYY-MM-DD')", miner, tStr).QueryRows(&rewardInfos)
		n, err = txOrm.QueryTable("fly_miner_status_and_daily_change").Filter("miner_id", miner).Filter("time", t).All(rewardInfo)
		if err != nil {
			rewardLog.Errorf("QueryTable rewardInfo:%+v err:%+v num:%+v time:%+v", miner, err, n, tStr)
			err := txOrm.Rollback()
			if err != nil {
				rewardLog.Errorf("calculatePowerAndPledge orm transation rollback error: %+v", err)
			}
			return err
		}
		if n == 0 {
			//rewardInfo.Time = tStr
			rewardInfo.MinerId = miner
			rewardInfo.Pledge = pleage - oldPleage
			rewardInfo.Power = power - oldPower
			rewardInfo.Vesting = vesting - oldVesting
			rewardInfo.Reward = 0
			rewardInfo.Epoch = epoch
			rewardInfo.MinedPercentage = 1
			rewardInfo.Time = t
			rewardInfo.UpdateTime = t
			//total
			rewardInfo.TotalPower = power
			rewardInfo.TotalAvailable = available
			rewardInfo.TotalPledge = pleage
			rewardInfo.TotalPreCommit = preCommit
			rewardInfo.TotalVesting = vesting
			rewardInfo.LiveSectorsNumber = sectorCount.Live
			rewardInfo.ActiveSectorsNumber = sectorCount.Active
			rewardInfo.FaultySectorsNumber = sectorCount.Faulty
			rewardInfo.PowerPercentage = percentage

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
			rewardLog.Infof("now  miner:%+v epoch:%+v ", miner, rewardInfo.Epoch)
			//rewardInfo = &rewardInfos[0]
			if rewardInfo.Epoch < epoch {

				rewardInfo.Pledge += pleage - oldPleage
				rewardInfo.Power += power - oldPower
				rewardInfo.Vesting += vesting - oldVesting
				rewardInfo.Epoch = epoch
				if rightNum == 0 {
					rewardInfo.MinedPercentage = 1
				} else {
					rewardInfo.MinedPercentage = float64(rewardInfo.BlockNum) / float64(rightNum)
				}
				rewardInfo.UpdateTime = t
				rewardInfo.TotalPower = power
				rewardInfo.TotalAvailable = available
				rewardInfo.TotalPledge = pleage
				rewardInfo.TotalPreCommit = preCommit
				rewardInfo.TotalVesting = vesting
				rewardInfo.LiveSectorsNumber = sectorCount.Live
				rewardInfo.ActiveSectorsNumber = sectorCount.Active
				rewardInfo.FaultySectorsNumber = sectorCount.Faulty
				rewardInfo.PowerPercentage = percentage
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
		rewardLog.Infof("miner:%+v epoch:%+v  power:%+v  pledge:%+v", miner, epoch, rewardInfo.Power, rewardInfo.Pledge)

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
	rewardLog.Infof("calculate complete epoch:%+v ", epoch)
	return nil
}

func calculateRewardAndPledge(index int, blocks []*types.BlockHeader, blockCid []cid.Cid, tipsetKey types.TipSetKey, blockAfter cid.Cid, messages []api.Message) error {
	o := orm.NewOrm()
	txOrm, err := o.Begin()
	if err != nil {
		rewardLog.Errorf("collectWalletData orm begin error: %+v", err)
		return err
	}
	//获取minerid
	miner := blocks[index].Miner.String()
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
			rewardLog.Errorf("collectWalletData orm transation rollback error: %+v", errTx)
		}
		return err
	}
	//	rewardLog.Debug("------gas:%+v,mine:%+v,penalty:%+v,value:%+v", gas, mine, penalty, value)
	//新增存储当天miner power状态、MinerPowerStatus
	//err = putMinerPowerStatus(txOrm, miner, power, available, preCommit, vesting, pleage, percentage, sectorCounts, epoch, t)
	//if err != nil {
	//	errTx := txOrm.Rollback()
	//	if errTx != nil {
	//		rewardLog.Errorf("collectWalletData orm transation rollback error: %+v", errTx)
	//	}
	//	return err
	//}
	//收益分配
	minerInfo := new(models.MinerInfo)
	n, err := txOrm.QueryTable("fly_miner_info").Filter("miner_id", miner).All(minerInfo)

	if err != nil {
		rewardLog.Errorf("QueryTable minerInfo:%+v err:%+v num:%+v ", miner, err, n)
		errTx := txOrm.Rollback()
		if errTx != nil {
			rewardLog.Errorf("collectMinerData orm transation rollback error: %+v", errTx)
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
				rewardLog.Errorf("collectMinerData orm transation rollback error: %+v", errTx)
			}
			return err
		}
	}
	//记录出块
	//更新出块权
	err = updateMineBlockRight(int64(blocks[0].Height), miner, t, value, winCount)
	if err != nil {
		rewardLog.Errorf("update mine block right  error: %+v", err)
		return err
	}
	//计算出块权数量
	rightNum, err := getBlockRightNum(miner, t)
	if err != nil {
		rewardLog.Errorf("get block right num error: %+v", err)
		return err
	}
	rewardLog.Infof("miner %+v have %+v mine right on %+v", miner, rightNum, t.String())

	rewardInfo := new(models.MinerStatusAndDailyChange)
	//rewardInfos := make([]models.MinerStatusAndDailyChange, 0)
	//入库
	//n, err = o.Raw("select * from fly_miner_status_and_daily_change where miner_id=? and time=to_date(?,'YYYY-MM-DD')", miner, tStr).QueryRows(&rewardInfos)
	n, err = txOrm.QueryTable("fly_miner_status_and_daily_change").Filter("miner_id", miner).Filter("time", t).All(rewardInfo)
	if err != nil {
		rewardLog.Errorf("QueryTable rewardInfo:%+v err:%+v num:%+v time:%+v", miner, err, n, tStr)
		errTx := txOrm.Rollback()
		if errTx != nil {
			rewardLog.Errorf("collectWalletData orm transation rollback error: %+v", errTx)
		}
		return err
	}
	//rewardLog.Infof("miner %+v reward info %+v mine right on %+v", miner, rewardInfo, epoch)
	//rewardLog.Infof("n:%+v miner %+v reward info %+v mine right on %+v", n, miner, rewardInfo, epoch)
	if n == 0 {
		//记录块收益
		rewardInfo.Epoch = epoch
		rewardInfo.MinerId = miner
		rewardInfo.Pledge = pleage - oldPleage
		rewardInfo.Power = power - oldPower
		rewardInfo.Vesting = vesting - oldVesting
		rewardInfo.Reward = value
		rewardInfo.TotalReward += value
		rewardInfo.BlockNum = 1
		rewardInfo.MinedPercentage = float64(rewardInfo.BlockNum) / float64(rightNum)
		rewardInfo.TotalBlockNum += 1
		rewardInfo.WinCounts = winCount
		rewardInfo.TotalWinCounts += winCount
		rewardInfo.Time = t
		rewardInfo.UpdateTime = t
		//total
		rewardInfo.TotalPower = power
		rewardInfo.TotalAvailable = available
		rewardInfo.TotalPledge = pleage
		rewardInfo.TotalPreCommit = preCommit
		rewardInfo.TotalVesting = vesting
		rewardInfo.LiveSectorsNumber = sectorCounts.Live
		rewardInfo.ActiveSectorsNumber = sectorCounts.Active
		rewardInfo.FaultySectorsNumber = sectorCounts.Faulty
		rewardInfo.PowerPercentage = percentage
		rewardLog.Infof("miner %+v reward info %+v mine right on %+v", miner, rewardInfo, epoch)
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
		rewardLog.Infof("now  miner:%+v epoch:%+v ", miner, rewardInfo.Epoch)
		//rewardInfo = &rewardInfos[0]
		//更新walletinfo
		if rewardInfo.Epoch < epoch {
			rewardInfo.Pledge += pleage - oldPleage
			rewardInfo.Power += power - oldPower
			rewardInfo.Vesting += vesting - oldVesting
			rewardInfo.Reward += value
			rewardInfo.TotalReward += value
			rewardInfo.Epoch = epoch
			rewardInfo.BlockNum += 1
			rewardInfo.MinedPercentage = float64(rewardInfo.BlockNum) / float64(rightNum)
			rewardInfo.TotalBlockNum += 1
			rewardInfo.WinCounts += winCount
			rewardInfo.TotalWinCounts += winCount
			rewardInfo.UpdateTime = t

			rewardInfo.TotalPower = power
			rewardInfo.TotalAvailable = available
			rewardInfo.TotalPledge = pleage
			rewardInfo.TotalPreCommit = preCommit
			rewardInfo.TotalVesting = vesting
			rewardInfo.LiveSectorsNumber = sectorCounts.Live
			rewardInfo.ActiveSectorsNumber = sectorCounts.Active
			rewardInfo.FaultySectorsNumber = sectorCounts.Faulty
			rewardInfo.PowerPercentage = percentage
			rewardLog.Infof("miner %+v reward info %+v mine right on %+v", miner, rewardInfo, epoch)
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
	rewardLog.Infof("miner:%+v epoch:%+v reward:%+v", miner, epoch, value)
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
			messages, err := client.Client.ChainGetBlockMessages(context.Background(), blockCid[i])
			if err != nil {
				rewardLog.Errorf("getRewardInfo ChainGetBlockMessages err:%+v", err)
				return 0, 0, 0, err
			}
			for _, message := range messages.BlsMessages {
				rewardMap[message.Cid().String()] = base
			}
		} else {
			messages, err := client.Client.ChainGetBlockMessages(context.Background(), blockCid[i])
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

	power, err := client.Client.StateMinerPower(ctx, miner, tipsetKey)
	if err != nil {
		rewardLog.Errorf("StateMinerPower err:%+v", err)
		return 0, 0, 0, err
	}
	var f float64 = 1024
	minerPower := float64(power.MinerPower.QualityAdjPower.Int64()) / f / f / f / f
	percentage := float64(power.MinerPower.QualityAdjPower.Int64()) / float64(power.TotalPower.QualityAdjPower.Int64())
	rewardActor, err := client.Client.StateGetActor(ctx, builtin.RewardActorAddr, tipsetKey)
	if err != nil {
		rewardLog.Errorf("StateGetActor err:%+v", err)
		return 0, 0, 0, err
	}

	rewardStateRaw, err := client.Client.ChainReadObj(ctx, rewardActor.Head)
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

	power, err := client.Client.StateMinerPower(ctx, miner, tipsetKey)
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
	//	minerPowerStatuss := make([]models.MinerStatusAndDailyChange, 0)
	minerPowerStatus := new(models.MinerStatusAndDailyChange)
	//num, err := o.Raw("select * from fly_miner_status_and_daily_change where miner_id=? and time=to_date(?,'YYYY-MM-DD')", miner, t.Format("2006-01-02")).QueryRows(&minerPowerStatuss)
	num, err := o.QueryTable("fly_miner_status_and_daily_change").Filter("miner_id", miner).Filter("time", t).All(minerPowerStatus)
	if err != nil {
		rewardLog.Errorf("QueryTable miner power status :%+v err:%+v num:%+v ", miner, err, num)
		err := o.Rollback()
		if err != nil {
			rewardLog.Debug("collectMinerData orm transation rollback error: %+v", err)
		}
		return err
	}
	if num == 0 {
		//minerPowerStatus.Epoch = epoch
		minerPowerStatus.TotalPower = power
		minerPowerStatus.TotalAvailable = available
		minerPowerStatus.TotalPledge = pleage
		minerPowerStatus.TotalPreCommit = preCommit
		minerPowerStatus.TotalVesting = vesting
		minerPowerStatus.Time = t
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
		//	minerPowerStatus = &minerPowerStatuss[0]
		//minerPowerStatus.Epoch = epoch
		minerPowerStatus.TotalPower = power
		minerPowerStatus.TotalAvailable = available
		minerPowerStatus.TotalPledge = pleage
		minerPowerStatus.TotalPreCommit = preCommit
		minerPowerStatus.TotalVesting = vesting
		minerPowerStatus.LiveSectorsNumber = sectorCounts.Live
		minerPowerStatus.ActiveSectorsNumber = sectorCounts.Active
		minerPowerStatus.FaultySectorsNumber = sectorCounts.Faulty
		minerPowerStatus.PowerPercentage = powerPercentage
		minerPowerStatus.UpdateTime = t
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
	tipset, err = client.Client.ChainGetTipSetByHeight(context.Background(), epoch, tipsetKey)

	return
}

func collectLotusChainHeadBlock() (tipset *types.TipSet, err error) {
	tipset, err = client.Client.ChainHead(context.Background())
	return
}

func getParentsBlockMessage(cid cid.Cid) (messages []api.Message, err error) {
	messages, err = client.Client.ChainGetParentMessages(context.Background(), cid)
	return
}

func getGasout(blockCid cid.Cid, messages *types.Message, basefee abi.TokenAmount, i int, height int64) (gasout vm.GasOutputs, err error) {
	charge := true
	ctx := context.Background()
	resp, err := client.Client.ChainGetParentReceipts(ctx, blockCid)

	if err != nil {
		rewardLog.Errorf("getGasout  ChainGetParentReceipts err:%+v", err)
		return
	}
	if messages.Method == 5 && height > 343200 {
		charge = false
	}
	gasout = vm.ComputeGasOutputs(resp[i].GasUsed, messages.GasLimit, basefee, messages.GasFeeCap, messages.GasPremium, charge)
	return
}

func recordMineBlockRight(tipset *types.TipSet) error {

	for miner, _ := range models.Miners {
		minerAddr, _ := address.NewFromString(miner)
		if ok, winCount := calculateMinerRight(tipset.Height()-1, minerAddr); ok {
			mbr := new(models.MineBlockRight)
			mbr.MinerId = miner
			mbr.Epoch = int64(tipset.Height())
			mbr.Missed = true
			mbr.WinCount = winCount
			mbr.Time = time.Unix(int64(tipset.MinTimestamp()), 0)
			mbr.UpdateTime = time.Unix(int64(tipset.MinTimestamp()), 0)
			err := mbr.Insert()
			if err != nil {
				rewardLog.Errorf("calculate miner right err:%+v", err)
				return err
			}
			rewardLog.Infof("miner %+v have a mine right in epoch %+v", miner, tipset.Height())
		}

	}
	return nil
}

func calculateMinerRight(h abi.ChainEpoch, miner address.Address) (bool, int64) {
	ctx := context.Background()
	round := h + 1
	for {
		tp, err := client.Client.ChainGetTipSetByHeight(ctx, h, types.NewTipSetKey())
		if err != nil {
			rewardLog.Warnf("ChainGetTipSetByHeight err:%+v", err)
			time.Sleep(time.Second * 3)
			continue
		}

		mbi, err := client.Client.MinerGetBaseInfo(ctx, miner, round, tp.Key())
		if err != nil {
			rewardLog.Warnf("MinerGetBaseInfo miner:%+v err:%+v", miner, err)
			time.Sleep(time.Second * 3)
			continue
		}

		if mbi == nil {
			time.Sleep(time.Second * 3)
			continue
		}
		if !mbi.EligibleForMining {
			// slashed or just have no power yet
			return false, 0
		}

		beaconPrev := mbi.PrevBeaconEntry
		bvals := mbi.BeaconEntries

		rbase := beaconPrev
		if len(bvals) > 0 {
			rbase = bvals[len(bvals)-1]
		}

		p, err := gen.IsRoundWinner(ctx, tp, round, miner, rbase, mbi, client.SignClient)
		if err != nil {
			rewardLog.Warnf("IsRoundWinner miner%+v err:%+v", miner, err)
			return false, 0
		}

		if p == nil {
			return false, 0
		}
		return true, p.WinCount
	}
}

func updateMineBlockRight(epoch int64, miner string, t time.Time, value float64, winCount int64) error {
	mbr := new(models.MineBlockRight)
	mbr.MinerId = miner
	mbr.Epoch = epoch
	mbr.Missed = false
	mbr.Reward = value
	mbr.WinCount = winCount
	mbr.Time = t
	mbr.UpdateTime = t
	//err := mbr.Update(t, value, winCount)
	err := mbr.Insert()
	if err != nil {
		rewardLog.Errorf("calculate miner right err:%+v", err)
		return err
	}

	return nil
}
func getBlockRightNum(miner string, t time.Time) (int64, error) {
	mbrs := make([]models.MineBlockRight, 0)
	o := orm.NewOrm()
	num, err := o.QueryTable("fly_mine_block_right").Filter("miner_id", miner).Filter("time", t).All(&mbrs)
	if err != nil {
		return 0, err
	}
	return num, nil
}
