package reward

import (
	"bytes"
	"context"
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/vm"
	"github.com/filecoin-project/specs-actors/v2/actors/builtin"
	"github.com/filecoin-project/specs-actors/v2/actors/builtin/miner"
	cid "github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"
	"math"
	"profit-allocation/models"
	"profit-allocation/tool/sync"
	"strconv"
	"time"
)

var msgLog = logging.Logger("message-log")

func CalculateMsgGasData() {
	defer sync.Wg.Done()
	CostFromHeight, err := queryMsgGasNetStatus()
	if err != nil {
		msgLog.Errorf("ERROR: collectLotusChainBlockRunData(), err=%v", err)
		return
	}
	msgLog.Debugf("DEBUG: CalculateMsgGasData(),  CostFromHeight:%v ", CostFromHeight)

	chainHeadResp, err := collectLotusChainHeadBlock()
	if err != nil {
		msgLog.Errorf("ERROR: CalculateMsgGasData()  collectLotusChainHeadBlock err=%v", err)
		return
	}
	blockHeight := int64(chainHeadResp.Height())
	if blockHeight <= CostFromHeight+11 {
		return
	}

	if blockHeight-CostFromHeight > 50 {
		h, err := handleMsgGasInfo(CostFromHeight+50, CostFromHeight)
		if err != nil {
			msgLog.Errorf("ERROR: CalculateMsgGasData() handleRequestInfo >50 err:%+v", err)
			return
		}
		CostFromHeight = h
	} else {
		h, err := handleMsgGasInfo(blockHeight, CostFromHeight)
		if err != nil {
			msgLog.Errorf("ERROR: CalculateMsgGasData() handleRequestInfo <=50 err:%+v", err)
			return
		}
		CostFromHeight = h
	}
	err = updateMsgGasNetStatus(CostFromHeight)
	if err != nil {
		msgLog.Errorf("updateMsgGasNetRunDataTmp height:%+v err :%+v", CostFromHeight, err)
	}

}

func queryMsgGasNetStatus() (height int64, err error) {
	o := orm.NewOrm()
	netRunData := new(models.ListenMsgGasNetStatus)
	n, err := o.QueryTable("fly_listen_msg_gas_net_status").All(netRunData)
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

func updateMsgGasNetStatus(height int64) (err error) {
	o := orm.NewOrm()
	netRunData := new(models.ListenMsgGasNetStatus)
	n, err := o.QueryTable("fly_listen_msg_gas_net_status").All(netRunData)
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
		//netRunData.CreateTime=time.Now().Unix()
		netRunData.UpdateTime = time.Now()
		_, err = o.Update(netRunData)
		if err != nil {
			return err
		}
	}
	return nil
}

func handleMsgGasInfo(dealBlcokHeight int64, end int64) (int64, error) {

	chainHeightHandle, err := getChainHeadByHeight(end)
	if err != nil {
		msgLog.Errorf("ERROR: handleMsgGasInfo() getChainHeadByHeight height:%+v err=%+v", dealBlcokHeight-1, err)
		return end, err
	}

	dh := dealBlcokHeight

	for i := end; i < dealBlcokHeight; i++ {
		chainHeightAfter, err := getChainHeadByHeight(i + 1)
		if err != nil {
			msgLog.Errorf("ERROR: handleMsgGasInfo() getChainHeadByHeight height:%+v err=%+v", dealBlcokHeight, err)
			return i, err
		}

		blockMessageResp, err := getParentsBlockMessage(chainHeightAfter.Cids()[0])
		if err != nil {
			msgLog.Errorf("ERROR: handleMsgGasInfo() getParentsBlockMessage cid %s  err=%v", chainHeightAfter.Cids()[0].String(), err)
			return i, err
		}
		blocks := chainHeightHandle.Blocks()
		//计算支出
		err = calculateWalletCost(*blocks[0], blockMessageResp, blocks[0].ParentBaseFee, chainHeightAfter.Cids()[0], chainHeightHandle.Key(), i)
		if err != nil {
			return i, err
		}
		chainHeightHandle = chainHeightAfter

	}
	return dh, nil
}

func calculateWalletCost(block types.BlockHeader, messages []api.Message, basefee abi.TokenAmount, blockAfter cid.Cid, tipsetKey types.TipSetKey, height int64) error {
	messagesCostMap := make(map[string]bool)
	consensusFaultMap := make(map[string]bool)
	for i, message := range messages {
		//计算支出
		if inWallets(message.Message.From.String()) || inMiners(message.Message.From.String()) {
			if messagesCostMap[message.Cid.String()] {
				continue
			}
			gasout, err := getGasout(blockAfter, message.Message, basefee, i, height)
			if err != nil {
				return err
			}
			err = recordCostMessage(gasout, message, block)
			if err != nil {
				return err
			}
			messagesCostMap[message.Cid.String()] = true
		}
		//计算ReportConsensusFault惩罚
		if message.Message.Method == 15 && inMiners(message.Message.To.String()) {
			if consensusFaultMap[message.Cid.String()] {
				continue
			}
			//计算

			burn, value, err := reportConsensusFaultPenalty(tipsetKey, message)
			if err != nil {
				return err
			}

			burnInt, err := strconv.ParseInt(burn, 0, 64)
			if err != nil {
				return err
			}
			valueInt, err := strconv.ParseInt(value, 0, 64)
			if err != nil {
				return err
			}
			zeroTokenAmount := abi.NewTokenAmount(0)
			burnTokenAmount := abi.NewTokenAmount(burnInt)
			valueTokenAmount := abi.NewTokenAmount(valueInt)
			gas := vm.GasOutputs{
				BaseFeeBurn:        burnTokenAmount,
				OverEstimationBurn: zeroTokenAmount,
				MinerPenalty:       zeroTokenAmount,
				MinerTip:           valueTokenAmount,
				Refund:             zeroTokenAmount,
				GasRefund:          0,
				GasBurned:          0,
			}
			err = recordCostMessage(gas, message, block)
			if err != nil {
				return err
			}
			consensusFaultMap[message.Cid.String()] = true
		}

		//记录precommit和provecommit消息
		if inMiners(message.Message.To.String()) {
			if message.Message.Method == 6 || message.Message.Method == 7 {
				//pre
				err := recordPreAndProveCommitMsg(message, height, block.Timestamp, message.Message.Method)
				if err != nil {
					return err
				}
			}
		}

	}
	h, err := strconv.ParseInt(block.Height.String(), 10, 64)
	if err != nil {
		msgLog.Errorf("parse hight:%+v err:%+v", block.Height.String(), err)
	}
	err = updateMsgGasNetStatus(h)
	if err != nil {
		msgLog.Errorf("update hight:%+v err:%+v", h, err)
	}
	return nil
}

func recordCostMessage(gasout vm.GasOutputs, message api.Message, block types.BlockHeader) error {
	var err error
	//获取minerid
	walletId := message.Message.From.String()
	to := message.Message.To.String()
	value := message.Message.Value
	if message.Message.Method == 16 {
		value, err = withdrawMsgValue(message)
		if err != nil {
			msgLog.Errorf("recordCostMessageInfo get withdraw msg:%+v value error: %+v", message.Cid, err)
			return err
		}
	}
	//获取wallet对应的miner
	minerId, err := getMinerByWallte(walletId)
	if err != nil {
		msgLog.Errorf("get miner by wallte :%+v err:%+v", walletId, err)
		return err
	}
	msgId := message.Cid.String()
	o := orm.NewOrm()
	txOmer, err := o.Begin()
	if err != nil {
		msgLog.Errorf("DEBUG: recordCostMessageInfo orm transation begin error: %+v", err)
		return err
	}
	expendMsg := new(models.ExpendMessages)
	n, err := txOmer.QueryTable("fly_expend_messages").Filter("message_id", msgId).All(expendMsg)
	if err != nil {
		msgLog.Errorf("Error  QueryTable rewardInfo:%+v err:%+v num:%+v ", walletId, err, n)
		errTx := txOmer.Rollback()
		if errTx != nil {
			msgLog.Errorf("DEBUG: collectWalletData orm transation rollback error: %+v", errTx)
		}
		return err
	}
	if n != 0 {
		//errTx := txOmer.Rollback()
		//if errTx != nil {
		//	msgLog.Errorf("DEBUG: collectWalletData orm transation rollback error: %+v", errTx)
		//}
		msgLog.Warnf("message is already exist:%+v\n", msgId)
		err = txOmer.Commit()
		if err != nil {
			msgLog.Errorf("DEBUG: recordCostMessageInfo orm transation Commit error: %+v", err)
			return err
		}
		return nil
	}
	t := time.Unix(int64(block.Timestamp), 0)
	epoch := int64(block.Height)
	gas := float64(gasout.MinerTip.Int64()) / math.Pow(10, 18)
	burnFee := float64(gasout.BaseFeeBurn.Int64()) / math.Pow(10, 18)
	overBurn := float64(gasout.OverEstimationBurn.Int64()) / math.Pow(10, 18)
	valueFloat := float64(value.Int64()) / math.Pow(10, 18)
	penalty := float64(gasout.MinerPenalty.Int64()) / math.Pow(10, 18)
	expendMsg.MessageId = msgId
	expendMsg.MinerId = minerId
	expendMsg.WalletId = walletId
	expendMsg.To = to
	expendMsg.Epoch = epoch
	expendMsg.Gas = gas
	expendMsg.BaseBurnFee = burnFee
	expendMsg.OverEstimationBurn = overBurn
	expendMsg.Value = valueFloat
	expendMsg.Penalty = penalty
	expendMsg.Method = uint32(message.Message.Method)
	expendMsg.CreateTime = t

	_, err = txOmer.Insert(expendMsg)
	if err != nil {
		msgLog.Errorf("Error  Insert miner:%+v time:%+v err:%+v ", walletId, t.Format("2006-01-02"), err)
		errTx := txOmer.Rollback()
		if errTx != nil {
			msgLog.Errorf("DEBUG: collectWalletData orm transation rollback error: %+v", errTx)
		}
		return err
	}
	//入库
	expendInfos := make([]models.ExpendInfo, 0)
	n, err = o.Raw("select * from fly_expend_info where wallet_id=? and update_time::date=to_date(?,'YYYY-MM-DD')", walletId, t.Format("2006-01-02")).QueryRows(&expendInfos)
	//n, err = txOmer.QueryTable("fly_expend_info").Filter("wallet_id", walletId).Filter("time", t.Format("2006-01-02")).All(expendInfo)
	if err != nil {
		msgLog.Errorf("Error  QueryTable rewardInfo:%+v err:%+v num:%+v time:%+v", walletId, err, n, t.Format("2006-01-02"))
		errTx := txOmer.Rollback()
		if errTx != nil {
			msgLog.Errorf("DEBUG: collectWalletData orm transation rollback error: %+v", errTx)
		}
		return err
	}
	expendInfo := new(models.ExpendInfo)

	if n == 0 {
		//记录块收益
		expendInfo.MinerId = minerId
		expendInfo.WalletId = walletId
		expendInfo.Epoch = epoch
		expendInfo.Gas = gas
		expendInfo.BaseBurnFee = burnFee
		expendInfo.OverEstimationBurn = overBurn
		expendInfo.Value = valueFloat
		//	expendInfo.Time = t.Format("2006-01-02")
		expendInfo.UpdateTime = t

		_, err = txOmer.Insert(expendInfo)
		if err != nil {
			msgLog.Errorf(" Insert miner:%+v time:%+v err:%+v ", walletId, t, err)
			errTx := txOmer.Rollback()
			if errTx != nil {
				msgLog.Errorf("collectWalletData orm transation rollback error: %+v", errTx)
			}
			return err
		}
	} else {
		expendInfo = &expendInfos[0]
		//记录块收益
		expendInfo.MinerId = minerId
		expendInfo.Epoch = epoch
		expendInfo.Gas += gas
		expendInfo.BaseBurnFee += burnFee
		expendInfo.OverEstimationBurn += overBurn
		expendInfo.Value += valueFloat
		expendInfo.UpdateTime = t
		_, err := txOmer.Update(expendInfo)
		if err != nil {
			msgLog.Errorf(" Update miner:%+v time:%+v err:%+v ", walletId, t, err)
			errTx := txOmer.Rollback()
			if errTx != nil {
				msgLog.Errorf(" collectWalletData orm transation rollback error: %+v", errTx)
			}
			return err
		}
	}
	rewardInfos := make([]models.MinerStatusAndDailyChange, 0)
	//入库
	n, err = o.Raw("select * from fly_miner_status_and_daily_change where miner_id=? and update_time::date=to_date(?,'YYYY-MM-DD')", minerId, t.Format("2006-01-02")).QueryRows(&rewardInfos)
	//n, err = txOrm.QueryTable("fly_reward_info").Filter("miner_id", miner).Filter("time", tStr).All(rewardInfo)
	if err != nil {
		rewardLog.Errorf("Error  QueryTable rewardInfo:%+v err:%+v num:%+v time:%+v", err, n, t.Format("2006-01-02"))
		errTx := txOmer.Rollback()
		if errTx != nil {
			rewardLog.Debug("DEBUG: collectWalletData orm transation rollback error: %+v", errTx)
		}
		return err
	}
	rewardInfo := new(models.MinerStatusAndDailyChange)
	if n == 0 {
		//记录块收益
		//rewardInfo.Time = tStr
		rewardInfo.MinerId = minerId
		rewardInfo.Gas = gas
		rewardInfo.TotalGas += gas
		rewardInfo.Epoch = epoch
		rewardInfo.UpdateTime = t

		_, err = txOmer.Insert(rewardInfo)
		if err != nil {
			rewardLog.Errorf("Error  Insert miner:%+v time:%+v err:%+v ", minerId, t, err)
			errTx := txOmer.Rollback()
			if errTx != nil {
				rewardLog.Errorf("Error: collectWalletData orm transation rollback error: %+v", errTx)
			}
			return err
		}
	} else {
		//记录块收益
		rewardInfo = &rewardInfos[0]
		//更新walletinfo
		if rewardInfo.Epoch < epoch {
			rewardInfo.Gas += gas
			rewardInfo.TotalGas += gas
			rewardInfo.Epoch = epoch
			rewardInfo.UpdateTime = t
			_, err := txOmer.Update(rewardInfo)
			if err != nil {
				rewardLog.Errorf("Error  Update miner:%+v time:%+v err:%+v ", minerId, t, err)
				errTx := txOmer.Rollback()
				if errTx != nil {
					rewardLog.Errorf("DEBUG: collectWalletData orm transation rollback error: %+v", errTx)
				}
				return err
			}
		}

	}

	err = txOmer.Commit()
	if err != nil {
		msgLog.Errorf("DEBUG: recordCostMessageInfo orm transation Commit error: %+v", err)
		return err
	}
	return nil
}

func TestMsg() {
	CreateLotusClient()
	chainHeightHandle, err := getChainHeadByHeight(460080)
	if err != nil {
		fmt.Printf("ERROR: handleRequestInfo() getChainHeadByHeight err=%+v \n", err)
		return
	}
	chainHeightAfter, err := getChainHeadByHeight(460081)
	if err != nil {
		fmt.Printf("ERROR: handleRequestInfo() getChainHeadByHeight  err=%+v \n", err)
		return
	}

	//chainHeightHandle, err := getChainHeadByHeight(i)
	//if err != nil {
	//	msgLog.Errorf("ERROR: handleRequestInfo() getChainHeadByHeight height:%+v err=%+v", dealBlcokHeight-1, err)
	//	return end, err
	//}
	fmt.Println(chainHeightAfter)
	blockMessageResp, err := getParentsBlockMessage(chainHeightAfter.Cids()[0])
	if err != nil {
		fmt.Printf("ERROR: handleRequestInfo() getParentsBlockMessage   err=%v\n", err)
		return
	}
	for _, message := range blockMessageResp {

		if message.Message.Method == 16 && inMiners(message.Message.To.String()) {
			//withdrawMsg(message)
		}

	}

	fmt.Println(len(blockMessageResp))
	err = calculateWalletCostTest(chainHeightHandle.Key(), blockMessageResp)
	if err != nil {
		fmt.Printf("ERROR: handleRequestInfo() calculateWalletCost   err=%v\n", err)
		return
	}

}

func calculateWalletCostTest(tipsetKey types.TipSetKey, messages []api.Message) error {

	for _, message := range messages {

		if message.Message.Method == 15 && inMiners(message.Message.To.String()) {
			burn, value, err := reportConsensusFaultPenalty(tipsetKey, message)
			if err != nil {
				return err
			}
			//	err = recordCostMessage(gasout, message, block)
			fmt.Printf("Debug method == 15  burn:%+v value:%+v \n", burn, value)
		}

	}
	fmt.Println(6)
	return nil
}

func reportConsensusFaultPenalty(tipsetKey types.TipSetKey, msg api.Message) (string, string, error) {
	ctx := context.Background()
	rewardActor, err := Client.StateGetActor(ctx, builtin.RewardActorAddr, tipsetKey)
	if err != nil {
		//fmt.Println("111 err",err)
		msgLog.Errorf("StateGetActor err:%+v", err)
		return "0", "0", err
	}

	rewardStateRaw, err := Client.ChainReadObj(ctx, rewardActor.Head)
	if err != nil {
		msgLog.Errorf("ChainReadObj err:%+v", err)
		return "0", "0", err
	}

	//mas.VestedFunds()
	r := bytes.NewReader(rewardStateRaw)
	rewardActorState := unmarshalState(r)

	//fmt.Printf("%+v\n", rewardActorState.ThisEpochRewardSmoothed.Estimate())
	penaltyFee := miner.ConsensusFaultPenalty(rewardActorState.ThisEpochRewardSmoothed.Estimate())
	//fmt.Printf("%+v\n", penaltyFee)
	//rcfp := new(miner.ReportConsensusFaultParams)
	//b := new(bytes.Buffer)
	//_, err = b.Write(msg.Message.Params)
	////fmt.Printf("msg :%+v",msg.Message)
	////fmt.Println(n, err)
	//if err != nil {
	//	msgLog.Errorf("reportConsensusFaultPenalty Write Message.Params err:%+v", err)
	//	return "0", "0", err
	//}
	//err = rcfp.UnmarshalCBOR(b)
	//if err != nil {
	//	msgLog.Errorf("reportConsensusFaultPenalty rcfp UnmarshalCBOR err:%+v", err)
	//	return "0", "0", err
	//}
	//faultAge := abi.ChainEpoch(1000)
	//slasherReward := miner.RewardForConsensusSlashReport(faultAge, penaltyFee)
	////fmt.Printf("%+v\n%+v\n", penaltyFee.String(), slasherReward.String())
	//burnFee := big.NewInt(0)
	//burnFee.Sub(penaltyFee.Int, slasherReward.Int)
	//burnFeeStr := burnFee.String()
	//if len(burnFeeStr) > 19 {
	//	burnFeeStr = burnFeeStr[:19]
	//}
	//slasherRewardStr := slasherReward.String()
	//if len(slasherRewardStr) > 19 {
	//	slasherRewardStr = slasherRewardStr[:19]
	//}
	//return burnFeeStr, slasherRewardStr, nil

	return penaltyFee.String(), "0", nil
}

func recordPreAndProveCommitMsg(msg api.Message, epoch int64, timeStamp uint64, method abi.MethodNum) error {
	m := new(models.PreAndProveMessages)
	if method == 6 {
		params := new(miner.PreCommitSectorParams)
		b := new(bytes.Buffer)
		_, err := b.Write(msg.Message.Params)
		if err != nil {
			msgLog.Errorf("record preCommit msg:%+v write byte err:%+v", msg.Cid, err)
			return err
		}
		err = params.UnmarshalCBOR(b)
		if err != nil {
			msgLog.Errorf("record preCommit msg:%+v unmarshal err:%+v", msg.Cid, err)
			return err
		}
		sectorNum, err := strconv.ParseInt(params.SectorNumber.String(), 10, 64)
		if err != nil {
			msgLog.Errorf("record preCommit msg:%+v parse sector number err:%+v", msg.Cid, err)
			return err
		}
		m.SectorNumber = sectorNum
		m.Method = 6
	}
	if method == 7 {
		params := new(miner.ProveCommitSectorParams)
		b := new(bytes.Buffer)
		_, err := b.Write(msg.Message.Params)
		if err != nil {
			msgLog.Errorf("record  proveCommit msg:%+v write byte err:%+v", msg.Cid, err)
			return err
		}
		err = params.UnmarshalCBOR(b)
		if err != nil {
			msgLog.Errorf("record  proveCommit msg:%+v unmarshal err:%+v", msg.Cid, err)
			return err
		}
		sectorNum, err := strconv.ParseInt(params.SectorNumber.String(), 10, 64)
		if err != nil {
			msgLog.Errorf("record proveCommit msg:%+v parse sector number err:%+v", msg.Cid, err)
			return err
		}
		m.SectorNumber = sectorNum
		m.Method = 7
	}

	m.MessageId = msg.Cid.String()
	m.From = msg.Message.From.String()
	m.To = msg.Message.To.String()
	m.Epoch = int64(epoch)
	m.CreateTime = time.Unix(int64(timeStamp), 0)
	return m.Insert()
}

func withdrawMsgValue(msg api.Message) (abi.TokenAmount, error) {
	params := new(miner.WithdrawBalanceParams)
	value := abi.NewTokenAmount(0)
	b := new(bytes.Buffer)
	_, err := b.Write(msg.Message.Params)
	if err != nil {
		//msgLog.Errorf("record preCommit msg:%+v write byte err:%+v",msg.Cid, err)
		return value, err
	}
	err = params.UnmarshalCBOR(b)
	if err != nil {
		//msgLog.Errorf("record preCommit msg:%+v unmarshal err:%+v", msg.Cid,err)
		return value, err
	}

	return params.AmountRequested, nil
}

func getMinerByWallte(walletId string) (string, error) {
	o := orm.NewOrm()
	minerAndWalletRelations := new(models.MinerAndWalletRelation)
	num, err := o.QueryTable("fly_miner_and_wallet_relation").Filter("wallet_id", walletId).All(minerAndWalletRelations)
	if err != nil {
		msgLog.Errorf("get miner by wallte :%+v err:%+v", walletId, err)
		return "", err
	}
	if num == 0 {
		msgLog.Errorf("can not found miner by wallte :%+v ", walletId)
		return "", fmt.Errorf("can not found miner by wallte :%+v", walletId)
	}
	return minerAndWalletRelations.MinerId, nil
}
