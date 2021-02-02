package reward

import (
	"bytes"
	"context"
	"errors"
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
	"time"
)

var rewardLog = logging.Logger("reward-log")
var Client api.FullNode

func CollectTotalRerwardAndPledge() {
	defer sync.Wg.Done()
	var rewardBlockHeight = 0
	if height, err := queryListenRewardNetStatus(); err != nil {
		rewardLog.Error("ERROR: collectLotusChainBlockRunData(), err=%v", err)
		return
	} else {
		rewardBlockHeight = height
	}
	rewardLog.Debug("DEBUG: collectLotusChainBlockRunData(),  rewardBlockHeight:%v ", rewardBlockHeight)
	chainHeadResp, err := collectLotusChainHeadBlock()
	if err != nil {
		rewardLog.Error("ERROR: collectLotusChainBlockRunData()  collectLotusChainHeadBlock err=%v", err)
		return
	}

	blockHeight := int(chainHeadResp.Height())
	if blockHeight <= rewardBlockHeight+11 {
		return
	}

	if blockHeight-rewardBlockHeight > 200 {
		//	rewardLog.Debug("DEBUG: collectLotusChainBlockRunData()  >200")

		h, err := getRewardAndPledge(rewardBlockHeight+200, rewardBlockHeight)
		if err != nil {
			rewardLog.Error("ERROR: collectLotusChainBlockRunData() handleRequestInfo >200 err:%+v", err)
			return
		}
		rewardBlockHeight = h
		//rewardLog.Debug("======== >500 ok")
	} else {
		//rewardLog.Debug("DEBUG: collectLotusChainBlockRunData()  <200")

		h, err := getRewardAndPledge(blockHeight, rewardBlockHeight)
		if err != nil {
			rewardLog.Error("ERROR: collectLotusChainBlockRunData() handleRequestInfo <=200 err:%+v", err)
			return
		}
		rewardBlockHeight = h
		//rewardLog.Debug("======== <500 ok")
	}

	err = updateListenRewardNetStatus(rewardBlockHeight)
	if err != nil {
		rewardLog.Error("updateNetRunData height:%+v err :%+v\n", DealMessageBlockHeight, err)
	}
}

func queryListenRewardNetStatus() (height int, err error) {
	o := orm.NewOrm()
	netRunData := new(models.ListenRewardNetStatus)
	n, err := o.QueryTable("fly_listen_reward_net_status").All(netRunData)
	if err != nil {
		return
	}
	if n == 0 {
		height = 413300
		return
	} else {
		height = netRunData.ReceiveBlockHeight
	}
	return
}

func updateListenRewardNetStatus(height int) (err error) {
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

func getRewardAndPledge(dealBlcokHeight int, end int) (int, error) {

	chainHeightHandle, err := getChainHeadByHeight(end)
	if err != nil {
		rewardLog.Error("ERROR: handleRequestInfo() getChainHeadByHeight height:%+v err=%+v", dealBlcokHeight-1, err)
		return end, err
	}

	dh := dealBlcokHeight

	for i := end; i < dealBlcokHeight; i++ {
		chainHeightAfter, err := getChainHeadByHeight(i + 1)
		if err != nil {
			rewardLog.Error("ERROR: handleRequestInfo() getChainHeadByHeight height:%+v err=%+v", dealBlcokHeight, err)
			return i, err
		}

		blockMessageResp, err := getParentsBlockMessage(chainHeightAfter.Cids()[0])
		if err != nil {
			rewardLog.Error("ERROR: handleRequestInfo() getParentsBlockMessage cid %s  err=%v", chainHeightAfter.Cids()[0].String(), err)
			return i, err
		}

		blocks := chainHeightHandle.Blocks()
		bFlag := true
		for index, block := range blocks {

			if inMiners(block.Miner.String()) {
				err = calculateRewardAndPledge(index, blocks, chainHeightHandle.Cids(), chainHeightHandle.Key(), chainHeightAfter.Cids()[0], blockMessageResp)
				if err != nil {
					rewardLog.Error("ERROR: handleRequestInfo() calculateMineReward height:%+v err=%+v", end, err)
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
					rewardLog.Error("ERROR: handleRequestInfo() calculatePowerAndPledge height:%+v err=%+v", end, err)
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
	epoch := int(height)
	//获取质押
	o := orm.NewOrm()
	txOrm, err := o.Begin()
	if err != nil {
		rewardLog.Debug("DEBUG: collectWalletData orm transation begin error: %+v", err)
		return err
	}
	for _, miner := range models.Miners {

		available, preCommit, vesting, pleage, err := GetMienrPleage(miner, height)
		if err != nil {
			rewardLog.Error("ERROR GetMienrPleage ParseFloat miner:%+v height:%+v err:%+v", miner, height, err)
			err := txOrm.Rollback()
			if err != nil {
				rewardLog.Debug("DEBUG: collectWalletData orm transation rollback error: %+v", err)
			}
			return err
		}
		power, err := getMinerPower(miner, tipsetKey)
		if err != nil {
			rewardLog.Error("ERROR GetMienrPleage getMinerPower miner:%+v height:%+v err:%+v", miner, height, err)
			err := txOrm.Rollback()
			if err != nil {
				rewardLog.Debug("DEBUG: collectWalletData orm transation rollback error: %+v", err)
			}
			return err
		}
		//	rewardLog.Debug("------gas:%+v,mine:%+v,penalty:%+v,value:%+v", gas, mine, penalty, value)
		err = putMinerPowerStatus(txOrm, miner, power, available, preCommit, vesting, pleage, tStr)
		if err != nil {
			rewardLog.Error("ERROR GetMienrPleage putMinerPowerStatus miner:%+v height:%+v err:%+v", miner, height, err)
			err := txOrm.Rollback()
			if err != nil {
				rewardLog.Debug("DEBUG: collectWalletData orm transation rollback error: %+v", err)
			}
			return err
		}
		//收益分配
		minerInfo := new(models.MinerInfo)
		n, err := txOrm.QueryTable("fly_miner_info").Filter("miner_id", miner).All(minerInfo)

		if err != nil {
			rewardLog.Error("Error  QueryTable minerInfo:%+v err:%+v num:%+v ", miner, err, n)
			err := txOrm.Rollback()
			if err != nil {
				rewardLog.Debug("DEBUG: collectMinerData orm transation rollback error: %+v", err)
			}
			return err
		}

		oldPleage := minerInfo.Pleage
		oldPower := minerInfo.QualityPower
		//rewardLog.Debug("-=-=-=-=-=-power:%+v old:%+v",power,oldPower)

		if n == 0 {
			return errors.New("get miner power  error")
		} else {
			//更新miner info
			minerInfo.Pleage = pleage
			minerInfo.QualityPower = power
			minerInfo.UpdateTime = t

			_, err := txOrm.Update(minerInfo)
			if err != nil {
				rewardLog.Error("Error  Update minerInfo miner:%+v  err:%+v ", miner, err)
				err := txOrm.Rollback()
				if err != nil {
					rewardLog.Debug("DEBUG: collectMinerData orm transation rollback error: %+v", err)
				}
				return err
			}
		}
		rewardInfo := new(models.RewardInfo)
		//入库
		n, err = txOrm.QueryTable("fly_reward_info").Filter("miner_id", miner).Filter("time", tStr).All(rewardInfo)
		if err != nil {
			rewardLog.Error("Error  QueryTable rewardInfo:%+v err:%+v num:%+v time:%+v", miner, err, n, tStr)
			err := txOrm.Rollback()
			if err != nil {
				rewardLog.Debug("DEBUG: calculatePowerAndPledge orm transation rollback error: %+v", err)
			}
			return err
		}
		if n == 0 {
			rewardInfo.Time = tStr
			rewardInfo.MinerId = miner
			rewardInfo.Pledge = pleage - oldPleage
			rewardInfo.Power = power - oldPower
			rewardInfo.Value = "0.0"
			rewardInfo.Epoch = epoch
			rewardInfo.UpdateTime = t

			_, err = txOrm.Insert(rewardInfo)
			if err != nil {
				rewardLog.Error("Error  Insert miner:%+v time:%+v err:%+v ", miner, tStr, err)
				err := txOrm.Rollback()
				if err != nil {
					rewardLog.Error("Error: calculatePowerAndPledge orm transation rollback error: %+v", err)
				}
				return err
			}
		} else {

			if rewardInfo.Epoch < epoch {

				rewardInfo.Pledge += pleage - oldPleage
				rewardInfo.Power += power - oldPower
				//rewardInfo.Value = bit.CalculateReward(rewardInfo.Value, value)
				rewardInfo.Epoch = epoch
				rewardInfo.UpdateTime = t
				_, err := txOrm.Update(rewardInfo)
				if err != nil {
					rewardLog.Error("Error  Update miner:%+v time:%+v err:%+v ", miner, t, err)
					err := txOrm.Rollback()
					if err != nil {
						rewardLog.Debug("DEBUG: calculatePowerAndPledge orm transation rollback error: %+v", err)
					}
					return err
				}
			}

		}
		rewardLog.Debug("Debug miner:%+v epoch:%+v  power:%+v  pledge:%+v", miner, epoch, rewardInfo.Power, rewardInfo.Pledge)

	}
	err = updateListenRewardNetStatus(epoch + 1)
	if err != nil {
		rewardLog.Error("Error  Update net run data tmp  err:%+v height:%+v ", err, epoch)
		err := txOrm.Rollback()
		if err != nil {
			rewardLog.Debug("DEBUG: calculatePowerAndPledge orm transation rollback error: %+v", err)
		}
		return err
	}
	err = txOrm.Commit()
	if err != nil {
		rewardLog.Debug("DEBUG: calculatePowerAndPledge orm transation commit error: %+v", err)
		return err
	}
	rewardLog.Debug("Debug  calculate complete epoch:%+v ", epoch)
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
	epoch := int(blocks[0].Height)
	//rewardLog.Debug("Debug collectMinertData height:%+v", epoch)
	winCount := blocks[index].ElectionProof.WinCount
	value, power, err := calculateReward(index, blocks[index].Miner, blockCid, tipsetKey, blocks[index].ParentBaseFee, winCount, blocks[index], blockAfter, messages)
	//	rewardLog.Debug("------gas:%+v,mine:%+v,penalty:%+v,value:%+v", gas, mine, penalty, value)
	if err != nil {
		errTx := txOrm.Rollback()
		if errTx != nil {
			rewardLog.Errorf("collectWalletData orm transation rollback error: %+v", errTx)
		}
		return err
	}

	//获取质押
	available, preCommit, vesting, pleage, err := GetMienrPleage(miner, blocks[0].Height)
	if err != nil {
		rewardLog.Error("ERROR GetMienrPleage ParseFloat err:%+v", err)
		errTx := txOrm.Rollback()
		if errTx != nil {
			rewardLog.Debug("DEBUG: collectWalletData orm transation rollback error: %+v", errTx)
		}
		return err
	}
	//	rewardLog.Debug("------gas:%+v,mine:%+v,penalty:%+v,value:%+v", gas, mine, penalty, value)
	//新增存储当天miner power状态、MinerPowerStatus
	err = putMinerPowerStatus(txOrm, miner, power, available, preCommit, vesting, pleage, tStr)
	if err != nil {
		errTx := txOrm.Rollback()
		if errTx != nil {
			rewardLog.Debug("DEBUG: collectWalletData orm transation rollback error: %+v", errTx)
		}
		return err
	}
	//收益分配
	minerInfo := new(models.MinerInfo)
	n, err := txOrm.QueryTable("fly_miner_info").Filter("miner_id", miner).All(minerInfo)

	if err != nil {
		rewardLog.Error("Error  QueryTable minerInfo:%+v err:%+v num:%+v ", miner, err, n)
		errTx := txOrm.Rollback()
		if errTx != nil {
			rewardLog.Debug("DEBUG: collectMinerData orm transation rollback error: %+v", errTx)
		}
		return err
	}

	oldPleage := minerInfo.Pleage
	oldPower := minerInfo.QualityPower
	//rewardLog.Debug("-=-=-=-=-=-power:%+v old:%+v",power,oldPower)

	if n == 0 {
		return errors.New("get miner power  error")
	} else {
		//更新miner info
		minerInfo.Pleage = pleage
		minerInfo.QualityPower = power
		minerInfo.UpdateTime = t

		_, err := txOrm.Update(minerInfo)
		if err != nil {
			rewardLog.Error("Error  Update minerInfo miner:%+v  err:%+v ", miner, err)
			errTx := txOrm.Rollback()
			if errTx != nil {
				rewardLog.Debug("DEBUG: collectMinerData orm transation rollback error: %+v", errTx)
			}
			return err
		}
	}
	//记录出块
	mineBlock := models.MineBlocks{
		//Id:         0,
		MinerId:  miner,
		Epoch:    epoch,
		Reward:   value,
		Gas:      "",
		Penalty:  "",
		Value:    "",
		Power:    0,
		WinCount: winCount,
		//Time:       tStr,
		CreateTime: t,
	}
	_, err = txOrm.Insert(&mineBlock)
	if err != nil {
		rewardLog.Error("Error  Update minerInfo miner:%+v  err:%+v ", miner, err)
		errTx := txOrm.Rollback()
		if errTx != nil {
			rewardLog.Debug("DEBUG: collectMinerData orm transation rollback error: %+v", errTx)
		}
		return err
	}

	rewardInfo := new(models.RewardInfo)
	//入库
	n, err = txOrm.QueryTable("fly_reward_info").Filter("miner_id", miner).Filter("time", tStr).All(rewardInfo)
	if err != nil {
		rewardLog.Error("Error  QueryTable rewardInfo:%+v err:%+v num:%+v time:%+v", miner, err, n, tStr)
		errTx := txOrm.Rollback()
		if errTx != nil {
			rewardLog.Debug("DEBUG: collectWalletData orm transation rollback error: %+v", errTx)
		}
		return err
	}
	if n == 0 {
		//记录块收益
		rewardInfo.Time = tStr
		rewardInfo.MinerId = miner
		rewardInfo.Pledge = pleage - oldPleage
		rewardInfo.Power = power - oldPower
		rewardInfo.Value = value
		rewardInfo.Epoch = epoch
		rewardInfo.BlockNum = 1
		rewardInfo.WinCounts = winCount
		rewardInfo.UpdateTime = t

		_, err = txOrm.Insert(rewardInfo)
		if err != nil {
			rewardLog.Error("Error  Insert miner:%+v time:%+v err:%+v ", miner, t, err)
			errTx := txOrm.Rollback()
			if errTx != nil {
				rewardLog.Error("Error: collectWalletData orm transation rollback error: %+v", errTx)
			}
			return err
		}
	} else {
		//记录块收益
		//更新walletinfo
		if rewardInfo.Epoch < epoch {

			rewardInfo.Pledge += pleage - oldPleage
			rewardInfo.Power += power - oldPower
			rewardInfo.Value = bit.CalculateReward(rewardInfo.Value, value)
			rewardInfo.Epoch = epoch
			rewardInfo.BlockNum += 1
			rewardInfo.WinCounts += winCount
			rewardInfo.UpdateTime = t
			_, err := txOrm.Update(rewardInfo)
			if err != nil {
				rewardLog.Error("Error  Update miner:%+v time:%+v err:%+v ", miner, t, err)
				errTx := txOrm.Rollback()
				if errTx != nil {
					rewardLog.Debug("DEBUG: collectWalletData orm transation rollback error: %+v", errTx)
				}
				return err
			}
		}

	}
	err = updateListenRewardNetStatus(epoch + 1)
	if err != nil {
		rewardLog.Error("Error  Update net run data tmp  err:%+v height:%+v ", err, epoch)
		errTx := txOrm.Rollback()
		if errTx != nil {
			rewardLog.Debug("DEBUG: collectWalletData orm transation rollback error: %+v", errTx)
		}
		return err
	}
	err = txOrm.Commit()
	if err != nil {
		rewardLog.Debug("DEBUG: collectWalletData orm transation commit error: %+v", err)
		return err
	}
	rewardLog.Debug("Debug miner:%+v epoch:%+v reward:%+v total:%+v", miner, epoch, value, rewardInfo.Value)
	return nil
}

func calculateReward(index int, miner address.Address, blockCid []cid.Cid, tipsetKey types.TipSetKey, basefee abi.TokenAmount, winCount int64, header *types.BlockHeader, blockAfter cid.Cid, msgs []api.Message) (string, float64, error) {
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
	//var totalGas string
	//var totalValue string
	//var mineReward string
	//nodeApi, closer, err := lotusClient.NewFullNodeRPC(context.Background(), models.LotusHost, requestHeader)
	//if err != nil {
	//	fmt.Println(err)
	//	return "0.0", 0, err
	//}
	//defer closer()
	for i := index; i >= 0; i-- {
		if i == index {
			messages, err := Client.ChainGetBlockMessages(context.Background(), blockCid[i])
			if err != nil {
				rewardLog.Error("Error getRewardInfo ChainGetBlockMessages err:%+v", err)
				return "0.0", 0, err
			}
			for _, message := range messages.BlsMessages {
				rewardMap[message.Cid().String()] = base
			}
		} else {
			messages, err := Client.ChainGetBlockMessages(context.Background(), blockCid[i])
			if err != nil {
				rewardLog.Error("Error getRewardInfo ChainGetBlockMessages err:%+v", err)
				return "0.0", 0, err
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
		gasout, err := getGasout(blockAfter, message.Message, basefee, i)
		if err != nil {
			return "0.0", 0, err
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
		mineMsg := new(models.MineMessages)
		mineMsg.MinerId = miner.String()
		mineMsg.MessageId = msgId
		mineMsg.Gas = msgGas.String()
		mineMsg.Penalty = msgPenalty.String()
		mineMsg.Epoch = header.Height.String()
		mineMsg.CreateTime = time.Unix(int64(header.Timestamp), 0)

		_, err := o.Insert(mineMsg)
		if err != nil {
			rewardForLog.Error("Error inert msg:%+v err:%+v", msgId, err)
			continue
		}
	}

	power, err := Client.StateMinerPower(ctx, miner, tipsetKey)
	if err != nil {
		rewardLog.Error("StateMinerPower err:%+v", err)
		return "0.0", 0, err
	}
	var f float64 = 1024
	minerPower := float64(power.MinerPower.QualityAdjPower.Int64()) / f / f / f / f

	rewardActor, err := Client.StateGetActor(ctx, builtin.RewardActorAddr, tipsetKey)
	if err != nil {
		rewardLog.Error("StateGetActor err:%+v", err)
		return "0.0", 0, err
	}

	rewardStateRaw, err := Client.ChainReadObj(ctx, rewardActor.Head)
	if err != nil {
		rewardLog.Error("ChainReadObj err:%+v", err)
		return "0.0", 0, err
	}

	r := bytes.NewReader(rewardStateRaw)
	rewardActorState := unmarshalState(r)

	mineReward = big.Div(rewardActorState.ThisEpochReward, abi.NewTokenAmount(5))
	mineReward = big.Mul(mineReward, abi.NewTokenAmount(winCount))

	value := big.Sub(big.Add(mineReward, totalGas), totalPenalty)
	if value.LessThan(abi.NewTokenAmount(0)) {
		value = abi.NewTokenAmount(0)
	}
	totalValue := bit.TransFilToFIL(value.String())

	return totalValue, minerPower, nil
}

func getMinerPower(minerStr string, tipsetKey types.TipSetKey) (float64, error) {
	//requestHeader := http.Header{}
	ctx := context.Background()
	miner, err := address.NewFromString(minerStr)
	if err != nil {
		rewardLog.Error("getMinerPower NewFromString err:%+v", err)
		return 0, err
	}
	//nodeApi, closer, err := lotusClient.NewFullNodeRPC(context.Background(), models.LotusHost, requestHeader)
	//if err != nil {
	//	rewardLog.Error("getMinerPower NewFullNodeRPC err:%+v", err)
	//	return 0, err
	//}
	//defer closer()
	power, err := Client.StateMinerPower(ctx, miner, tipsetKey)
	if err != nil {
		rewardLog.Error("StateMinerPower err:%+v", err)
		return 0, err
	}
	var f float64 = 1024
	minerPower := float64(power.MinerPower.QualityAdjPower.Int64()) / f / f / f / f
	return minerPower, nil
}

func TestCalculateReward() {

	i := 252707
	chainHeightHandle, err := getChainHeadByHeight(252707)
	if err != nil {
		rewardLog.Error("ERROR: handleRequestInfo() getChainHeadByHeight err=%+v", err)
		return
	}
	re := "0.0"
	n := 0
	for {

		if "2020-11-22" <= time.Unix(int64(chainHeightHandle.Blocks()[0].Timestamp), 0).Format("2006-01-02") {
			break
		}
		chainHeightAfter, err := getChainHeadByHeight(i + 1)
		if err != nil {
			rewardLog.Error("ERROR: handleRequestInfo() getChainHeadByHeight height:%+v err=%+v", i, err)
			return
		}
		//chainHeightHandle, err := getChainHeadByHeight(i)
		//if err != nil {
		//	rewardLog.Error("ERROR: handleRequestInfo() getChainHeadByHeight height:%+v err=%+v", dealBlcokHeight-1, err)
		//	return end, err
		//}
		blockMessageResp, err := getParentsBlockMessage(chainHeightAfter.Cids()[0])
		if err != nil {
			rewardLog.Error("ERROR: handleRequestInfo() getParentsBlockMessage cid %s  err=%v", chainHeightAfter.Cids()[0].String(), err)
			return
		}

		blocks := chainHeightHandle.Blocks()
		for index, block := range blocks {
			if inMiners(block.Miner.String()) {
				n++
				v, err := calculateRewardAndPledgeTest(index, blocks, chainHeightHandle.Cids(), chainHeightHandle.Key(), chainHeightAfter.Cids()[0], blockMessageResp)
				if err != nil {
					rewardLog.Error("ERROR: handleRequestInfo() calculateMineReward height:%+v err=%+v", i, err)
					return
				}
				re = bit.CalculateReward(v, re)
				rewardLog.Debug("Debug height:%+v total reward:%+v,count:%+v", i, re, n)
			}
		}
		chainHeightHandle = chainHeightAfter
		i++

	}
	rewardLog.Debug("Debug total ok")
}

func calculateRewardAndPledgeTest(index int, blocks []*types.BlockHeader, blockCid []cid.Cid, tipsetKey types.TipSetKey, blockAfter cid.Cid, messages []api.Message) (string, error) {

	epoch := int(blocks[0].Height)
	//rewardLog.Debug("Debug collectMinertData height:%+v", epoch)
	winCount := blocks[index].ElectionProof.WinCount
	value, _, err := calculateReward(index, blocks[index].Miner, blockCid, tipsetKey, blocks[index].ParentBaseFee, winCount, blocks[index], blockAfter, messages)
	if err != nil {
		return "0.0", err
	}
	rewardLog.Debug("------miner:%+v,epoch:%+v,value:%+v,wincount:%+v", blocks[index].Miner, epoch, value, winCount)

	return value, nil
}

func putMinerPowerStatus(o orm.TxOrmer, miner string, power, available, preCommit, vesting, pleage float64, t string) error {
	minerPowerStatus := new(models.MinerPowerStatus)
	num, err := o.QueryTable("fly_miner_power_status").Filter("miner_id", miner).Filter("time", t).All(minerPowerStatus)
	if err != nil {
		rewardLog.Error("Error  QueryTable miner power status :%+v err:%+v num:%+v ", miner, err, num)
		err := o.Rollback()
		if err != nil {
			rewardLog.Debug("DEBUG: collectMinerData orm transation rollback error: %+v", err)
		}
		return err
	}
	if num == 0 {
		minerPowerStatus.Power = power
		minerPowerStatus.Available = available
		minerPowerStatus.Pleage = pleage
		minerPowerStatus.PreCommit = preCommit
		minerPowerStatus.Vesting = vesting
		minerPowerStatus.Time = t
		minerPowerStatus.MinerId = miner
		_, err = o.Insert(minerPowerStatus)
		if err != nil {
			rewardLog.Error("Error  InsertTable miner power status :%+v err:%+v ", miner, err)
			err := o.Rollback()
			if err != nil {
				rewardLog.Debug("DEBUG: collectMinerData orm transation rollback error: %+v", err)
			}
			return err
		}
	} else {
		minerPowerStatus.Power = power
		minerPowerStatus.Available = available
		minerPowerStatus.Pleage = pleage
		minerPowerStatus.PreCommit = preCommit
		minerPowerStatus.Vesting = vesting
		_, err = o.Update(minerPowerStatus)
		if err != nil {
			rewardLog.Error("Error  UpdateTable miner power status :%+v err:%+v ", miner, err)
			err := o.Rollback()
			if err != nil {
				rewardLog.Debug("DEBUG: collectMinerData orm transation rollback error: %+v", err)
			}
			return err
		}
	}
	return nil
}

func getChainHeadByHeight(height int) (tipset *types.TipSet, err error) {
	//requestHeader := http.Header{}
	//nodeApi, closer, err := lotusClient.NewFullNodeRPC(context.Background(), models.LotusHost, requestHeader)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//defer closer()

	epoch := abi.ChainEpoch(height)
	tipsetKey := types.NewTipSetKey()
	tipset, err = Client.ChainGetTipSetByHeight(context.Background(), epoch, tipsetKey)

	return
}

func collectLotusChainHeadBlock() (tipset *types.TipSet, err error) {
	//requestHeader := http.Header{}
	//nodeApi, closer, err := lotusClient.NewFullNodeRPC(context.Background(), models.LotusHost, requestHeader)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//defer closer()

	tipset, err = Client.ChainHead(context.Background())

	return
}

func getParentsBlockMessage(cid cid.Cid) (messages []api.Message, err error) {
	//requestHeader := http.Header{}
	//nodeApi, closer, err := lotusClient.NewFullNodeRPC(context.Background(), models.LotusHost, requestHeader)
	//if err != nil {
	//	return
	//}
	//defer closer()
	messages, err = Client.ChainGetParentMessages(context.Background(), cid)
	return
}

func getGasout(blockCid cid.Cid, messages *types.Message, basefee abi.TokenAmount, i int) (gasout vm.GasOutputs, err error) {
	charge := true
	//requestHeader := http.Header{}
	ctx := context.Background()

	//nodeApi, closer, err := lotusClient.NewFullNodeRPC(ctx, models.LotusHost, requestHeader)
	//if err != nil {
	//	rewardForLog.Error("getGasout  NewFullNodeRPC err:%+v", err)
	//	return
	//}
	//defer closer()

	resp, err := Client.ChainGetParentReceipts(ctx, blockCid)

	if err != nil {
		rewardForLog.Error("getGasout  ChainGetParentReceipts err:%+v", err)
		return
	}
	if messages.Method == 5 {
		charge = false
	}
	gasout = vm.ComputeGasOutputs(resp[i].GasUsed, messages.GasLimit, basefee, messages.GasFeeCap, messages.GasPremium, charge)
	return
}

func CreateLotusClient() {
	var err error
	requestHeader := http.Header{}
	requestHeader.Add("Content-Type", "application/json")
	rewardLog.Debugf("lotus host:%v", models.LotusHost)
	Client, _, err = lotusClient.NewFullNodeRPC(context.Background(), models.LotusHost, requestHeader)
	if err != nil {
		rewardLog.Errorf("create lotus client", err)
		return
	}
}
