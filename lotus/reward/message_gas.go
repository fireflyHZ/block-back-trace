package reward

import (
	"bytes"
	"context"
	"github.com/beego/beego/v2/client/orm"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/vm"
	"github.com/filecoin-project/specs-actors/v2/actors/builtin"
	"github.com/filecoin-project/specs-actors/v2/actors/builtin/miner"
	cid "github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"
	"math/big"
	"profit-allocation/models"
	"profit-allocation/tool/bit"
	"profit-allocation/tool/sync"
	"strconv"
	"time"
)

var CostFromHeight = 343201
var msgLog = logging.Logger("message-log")

func CalculateMsgGasData() {
	defer sync.Wg.Done()
	if height, err := queryMsgGasNetStatus(); err != nil {
		msgLog.Error("ERROR: collectLotusChainBlockRunData(), err=%v", err)
		return
	} else {
		CostFromHeight = height
	}
	msgLog.Debug("DEBUG: CalculateMsgGasData(),  CostFromHeight:%v ", CostFromHeight)

	chainHeadResp, err := collectLotusChainHeadBlock()
	if err != nil {
		msgLog.Error("ERROR: CalculateMsgGasData()  collectLotusChainHeadBlock err=%v", err)
		return
	}
	blockHeight := int(chainHeadResp.Height())
	if blockHeight <= CostFromHeight+11 {
		return
	}

	if blockHeight-CostFromHeight > 50 {
		h, err := handleMsgGasInfo(CostFromHeight+50, CostFromHeight)
		if err != nil {
			msgLog.Error("ERROR: CalculateMsgGasData() handleRequestInfo >50 err:%+v", err)
			return
		}
		CostFromHeight = h
	} else {
		h, err := handleMsgGasInfo(blockHeight, CostFromHeight)
		if err != nil {
			msgLog.Error("ERROR: CalculateMsgGasData() handleRequestInfo <=50 err:%+v", err)
			return
		}
		CostFromHeight = h
	}
	err = updateMsgGasNetStatus(CostFromHeight)
	if err != nil {
		msgLog.Error("updateMsgGasNetRunDataTmp height:%+v err :%+v", CostFromHeight, err)
	}

}

func queryMsgGasNetStatus() (height int, err error) {
	o := orm.NewOrm()
	netRunData := new(models.ListenMsgGasNetStatus)
	n, err := o.QueryTable("fly_listen_msg_gas_net_status").All(netRunData)
	if err != nil {
		return
	}
	if n == 0 {
		height = 343201
		return
	} else {
		height = netRunData.ReceiveBlockHeight
	}
	return
}

func updateMsgGasNetStatus(height int) (err error) {
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

func handleMsgGasInfo(dealBlcokHeight int, end int) (int, error) {

	chainHeightHandle, err := getChainHeadByHeight(end)
	if err != nil {
		msgLog.Error("ERROR: handleMsgGasInfo() getChainHeadByHeight height:%+v err=%+v", dealBlcokHeight-1, err)
		return end, err
	}

	dh := dealBlcokHeight

	for i := end; i < dealBlcokHeight; i++ {
		chainHeightAfter, err := getChainHeadByHeight(i + 1)
		if err != nil {
			msgLog.Error("ERROR: handleMsgGasInfo() getChainHeadByHeight height:%+v err=%+v", dealBlcokHeight, err)
			return i, err
		}

		blockMessageResp, err := getParentsBlockMessage(chainHeightAfter.Cids()[0])
		if err != nil {
			msgLog.Error("ERROR: handleMsgGasInfo() getParentsBlockMessage cid %s  err=%v", chainHeightAfter.Cids()[0].String(), err)
			return i, err
		}
		blocks := chainHeightHandle.Blocks()
		//计算支出
		err = calculateWalletCost(*blocks[0], blockMessageResp, blocks[0].ParentBaseFee, chainHeightAfter.Cids()[0], chainHeightHandle.Key())
		if err != nil {
			return i, err
		}
		chainHeightHandle = chainHeightAfter

	}
	return dh, nil
}

func calculateWalletCost(block types.BlockHeader, messages []api.Message, basefee abi.TokenAmount, blockAfter cid.Cid, tipsetKey types.TipSetKey) error {
	messagesCostMap := make(map[string]bool)
	consensusFaultMap := make(map[string]bool)
	for i, message := range messages {
		//计算支出
		if inWallets(message.Message.From.String()) || inMiners(message.Message.From.String()) {
			if messagesCostMap[message.Cid.String()] {
				continue
			}
			gasout, err := getGasout(blockAfter, message.Message, basefee, i)
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

	}
	h, err := strconv.ParseInt(block.Height.String(), 10, 64)
	if err != nil {
		msgLog.Errorf("parse hight:%+v err:%+v", block.Height.String(), err)
	}
	err = updateMsgGasNetStatus(int(h))
	if err != nil {
		msgLog.Errorf("update hight:%+v err:%+v", h, err)
	}
	return nil
}

func recordCostMessage(gasout vm.GasOutputs, message api.Message, block types.BlockHeader) error {
	//获取minerid
	walletId := message.Message.From.String()
	to := message.Message.To.String()
	value := message.Message.Value
	msgId := message.Cid.String()
	o := orm.NewOrm()
	txOmer, err := o.Begin()
	if err != nil {
		msgLog.Debug("DEBUG: recordCostMessageInfo orm transation begin error: %+v", err)
		return err
	}
	expendMsg := new(models.ExpendMessages)
	n, err := txOmer.QueryTable("fly_expend_messages").Filter("message_id", msgId).All(expendMsg)
	if err != nil {
		msgLog.Error("Error  QueryTable rewardInfo:%+v err:%+v num:%+v ", walletId, err, n)
		errTx := txOmer.Rollback()
		if errTx != nil {
			msgLog.Debug("DEBUG: collectWalletData orm transation rollback error: %+v", errTx)
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
			msgLog.Debug("DEBUG: recordCostMessageInfo orm transation Commit error: %+v", err)
			return err
		}
		return nil
	}
	t := time.Unix(int64(block.Timestamp), 0)
	epoch := block.Height.String()

	expendMsg.MessageId = msgId
	expendMsg.WalletId = walletId
	expendMsg.To = to
	expendMsg.Epoch = epoch
	expendMsg.Gas = gasout.MinerTip.String()
	expendMsg.BaseBurnFee = gasout.BaseFeeBurn.String()
	expendMsg.OverEstimationBurn = gasout.OverEstimationBurn.String()
	expendMsg.Value = value.String()
	expendMsg.Penalty = gasout.MinerPenalty.String()
	expendMsg.Method = uint64(message.Message.Method)
	expendMsg.CreateTime = t

	_, err = txOmer.Insert(expendMsg)
	if err != nil {
		msgLog.Error("Error  Insert miner:%+v time:%+v err:%+v ", walletId, t.Format("2006-01-02"), err)
		errTx := txOmer.Rollback()
		if errTx != nil {
			msgLog.Debug("DEBUG: collectWalletData orm transation rollback error: %+v", errTx)
		}
		return err
	}
	expendInfo := new(models.ExpendInfo)
	//入库
	n, err = txOmer.QueryTable("fly_expend_info").Filter("wallet_id", walletId).Filter("time", t.Format("2006-01-02")).All(expendInfo)
	if err != nil {
		msgLog.Error("Error  QueryTable rewardInfo:%+v err:%+v num:%+v time:%+v", walletId, err, n, t.Format("2006-01-02"))
		errTx := txOmer.Rollback()
		if errTx != nil {
			msgLog.Debug("DEBUG: collectWalletData orm transation rollback error: %+v", errTx)
		}
		return err
	}
	if n == 0 {
		//记录块收益
		expendInfo.WalletId = walletId
		expendInfo.Epoch = epoch
		expendInfo.Gas = bit.TransFilToFIL(gasout.MinerTip.String())
		expendInfo.BaseBurnFee = bit.TransFilToFIL(gasout.BaseFeeBurn.String())
		expendInfo.OverEstimationBurn = bit.TransFilToFIL(gasout.OverEstimationBurn.String())
		expendInfo.Value = bit.TransFilToFIL(value.String())
		expendInfo.Time = t.Format("2006-01-02")
		expendInfo.UpdateTime = t

		_, err = txOmer.Insert(expendInfo)
		if err != nil {
			msgLog.Error("Error  Insert miner:%+v time:%+v err:%+v ", walletId, t, err)
			errTx := txOmer.Rollback()
			if errTx != nil {
				msgLog.Debug("DEBUG: collectWalletData orm transation rollback error: %+v", errTx)
			}
			return err
		}
	} else {
		//记录块收益
		expendInfo.Epoch = epoch
		expendInfo.Gas = bit.CalculateReward(expendInfo.Gas, bit.TransFilToFIL(gasout.MinerTip.String()))
		expendInfo.BaseBurnFee = bit.CalculateReward(expendInfo.BaseBurnFee, bit.TransFilToFIL(gasout.BaseFeeBurn.String()))
		expendInfo.OverEstimationBurn = bit.CalculateReward(expendInfo.OverEstimationBurn, bit.TransFilToFIL(gasout.OverEstimationBurn.String()))
		expendInfo.Value = bit.CalculateReward(expendInfo.Value, bit.TransFilToFIL(value.String()))
		expendInfo.UpdateTime = t
		_, err := txOmer.Update(expendInfo)
		if err != nil {
			msgLog.Error("Error  Update miner:%+v time:%+v err:%+v ", walletId, t, err)
			errTx := txOmer.Rollback()
			if errTx != nil {
				msgLog.Debug("DEBUG: collectWalletData orm transation rollback error: %+v", errTx)
			}
			return err
		}
	}

	err = txOmer.Commit()
	if err != nil {
		msgLog.Debug("DEBUG: recordCostMessageInfo orm transation Commit error: %+v", err)
		return err
	}
	return nil
}

func TestMsg() {
	chainHeightHandle, err := getChainHeadByHeight(230056)
	if err != nil {
		msgLog.Error("ERROR: handleRequestInfo() getChainHeadByHeight err=%+v", err)
		return
	}

	chainHeightAfter, err := getChainHeadByHeight(230057)
	if err != nil {
		msgLog.Error("ERROR: handleRequestInfo() getChainHeadByHeight  err=%+v", err)
		return
	}
	//chainHeightHandle, err := getChainHeadByHeight(i)
	//if err != nil {
	//	msgLog.Error("ERROR: handleRequestInfo() getChainHeadByHeight height:%+v err=%+v", dealBlcokHeight-1, err)
	//	return end, err
	//}
	blockMessageResp, err := getParentsBlockMessage(chainHeightAfter.Cids()[0])
	if err != nil {
		msgLog.Error("ERROR: handleRequestInfo() getParentsBlockMessage   err=%v", err)
		return
	}

	//
	//blocks := chainHeightHandle.Blocks()
	//timeStamp:=blocks[0].Timestamp

	//计算支出
	err = calculateWalletCostTest(chainHeightHandle.Key(), blockMessageResp)
	if err != nil {
		msgLog.Error("ERROR: handleRequestInfo() calculateWalletCost   err=%v", err)
		return
	}

}

func calculateWalletCostTest(tipsetKey types.TipSetKey, messages []api.Message) error {

	for _, message := range messages {
		//?????
		/*if v==5{
			break
		}*/
		//计算支出
		/*if inWallets(message.Message.From.String())||inMiners(message.Message.From.String()) {

			gasout, err := getGasout(blockAfter, message.Message, basefee, i)
			if err != nil {
				return err
			}
			msgLog.Debug("Debug  msgid:%+v --gas:%+v",message.Cid,gasout)

			//fmt.Println("----------------------------------------")
		}
		*/
		if message.Message.Method == 15 && inMiners(message.Message.To.String()) {
			burn, value, err := reportConsensusFaultPenalty(tipsetKey, message)
			if err != nil {
				return err
			}
			//	err = recordCostMessage(gasout, message, block)
			msgLog.Debug("Debug method == 15  msgid:%+v --gas:%+v", burn, value)
		}

	}

	return nil
}

func reportConsensusFaultPenalty(tipsetKey types.TipSetKey, msg api.Message) (string, string, error) {
	ctx := context.Background()
	rewardActor, err := Client.StateGetActor(ctx, builtin.RewardActorAddr, tipsetKey)
	if err != nil {
		msgLog.Error("StateGetActor err:%+v", err)
	}

	rewardStateRaw, err := Client.ChainReadObj(ctx, rewardActor.Head)
	if err != nil {
		msgLog.Error("ChainReadObj err:%+v", err)
	}

	r := bytes.NewReader(rewardStateRaw)
	rewardActorState := unmarshalState(r)
	//fmt.Printf("%+v\n", rewardActorState.ThisEpochRewardSmoothed.Estimate())
	penaltyFee := miner.ConsensusFaultPenalty(rewardActorState.ThisEpochRewardSmoothed.Estimate())
	//fmt.Printf("%+v\n", aiya)
	rcfp := new(miner.ReportConsensusFaultParams)
	b := new(bytes.Buffer)
	_, err = b.Write(msg.Message.Params)
	//fmt.Printf("msg :%+v",msg.Message)
	//fmt.Println(n, err)
	if err != nil {
		msgLog.Error("reportConsensusFaultPenalty Write Message.Params err:%+v", err)
		return "0", "0", err
	}
	err = rcfp.UnmarshalCBOR(b)
	if err != nil {
		msgLog.Error("reportConsensusFaultPenalty rcfp UnmarshalCBOR err:%+v", err)
		return "0", "0", err
	}
	faultAge := abi.ChainEpoch(1000)
	slasherReward := miner.RewardForConsensusSlashReport(faultAge, penaltyFee)
	//fmt.Printf("%+v\n%+v\n", aiya.String(), slasherReward.String())
	burnFee := big.NewInt(0)
	burnFee.Sub(penaltyFee.Int, slasherReward.Int)
	burnFeeStr := burnFee.String()
	if len(burnFeeStr) > 19 {
		burnFeeStr = burnFeeStr[:19]
	}
	slasherRewardStr := slasherReward.String()
	if len(slasherRewardStr) > 19 {
		slasherRewardStr = slasherRewardStr[:19]
	}
	return burnFeeStr, slasherRewardStr, nil
}
