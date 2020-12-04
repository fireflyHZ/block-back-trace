package reward

import (
	"github.com/astaxie/beego/orm"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/vm"
	cid "github.com/ipfs/go-cid"
	"profit-allocation/models"
	"profit-allocation/tool/bit"
	"profit-allocation/tool/log"
	"profit-allocation/tool/sync"
	"time"
)

var CostFromHeight = 221040

func CalculateMsgGasData() {
	defer sync.Wg.Done()
	if height, err := queryMsgGasNetRunDataTmp(); err != nil {
		log.Logger.Error("ERROR: collectLotusChainBlockRunData(), err=%v", err)
		return
	} else {
		//log.Logger.Debug("debug: collectLotusChainBlockRunData(), height=%v", height)
		CostFromHeight = height
	}
	log.Logger.Debug("DEBUG: CalculateMsgGasData(),  CostFromHeight:%v ", CostFromHeight)


	chainHeadResp, err := collectLotusChainHeadBlock()
	if err != nil {
		log.Logger.Error("ERROR: CalculateMsgGasData()  collectLotusChainHeadBlock err=%v", err)
		return
	}
	blockHeight := int(chainHeadResp.Height())
	if blockHeight <= CostFromHeight+11 {
		return
	}

	if blockHeight-CostFromHeight > 50 {
		//	log.Logger.Debug("DEBUG: collectLotusChainBlockRunData()  >200")

		h, err := handleMsgGasInfo(CostFromHeight+50, CostFromHeight)
		if err != nil {
			log.Logger.Error("ERROR: CalculateMsgGasData() handleRequestInfo >500 err:%+v", err)
			return
		}
		CostFromHeight = h
		//log.Logger.Debug("======== >500 ok")
	} else {
		//log.Logger.Debug("DEBUG: collectLotusChainBlockRunData()  <200")

		h, err := handleMsgGasInfo(blockHeight, CostFromHeight)
		if err != nil {
			log.Logger.Error("ERROR: CalculateMsgGasData() handleRequestInfo <=500 err:%+v", err)
			return
		}
		CostFromHeight = h
		//log.Logger.Debug("======== <500 ok")
	}
	err=updateMsgGasNetRunDataTmp(CostFromHeight)
	if err != nil {
		log.Logger.Error("updateMsgGasNetRunDataTmp height:%+v err :%+v\n", CostFromHeight, err)
	}

}

func queryMsgGasNetRunDataTmp() (height int, err error) {
	o := orm.NewOrm()
	netRunData := new(models.MsgGasNetRunDataProTmp)
	n, err := o.QueryTable("fly_msg_gas_net_run_data_pro_tmp").All(netRunData)
	if err != nil {
		return
	}
	if n == 0 {
		height = 221040
		return
	} else {
		height = netRunData.ReceiveBlockHeight
	}
	return
}

func updateMsgGasNetRunDataTmp(height int) (err error) {
	o := orm.NewOrm()
	netRunData := new(models.MsgGasNetRunDataProTmp)
	n, err := o.QueryTable("fly_msg_gas_net_run_data_pro_tmp").All(netRunData)
	if err != nil {
		return
	}
	if n == 0 {
		netRunData.ReceiveBlockHeight = height
		netRunData.CreateTime = time.Now().Unix()
		netRunData.UpdateTime = time.Now().Unix()
		_, err = o.Insert(netRunData)
		if err != nil {
			return err
		}

	} else {
		netRunData.ReceiveBlockHeight = height
		//netRunData.CreateTime=time.Now().Unix()
		netRunData.UpdateTime = time.Now().Unix()
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
		log.Logger.Error("ERROR: handleMsgGasInfo() getChainHeadByHeight height:%+v err=%+v", dealBlcokHeight-1, err)
		return end, err
	}

	dh := dealBlcokHeight

	for i := end; i < dealBlcokHeight; i++ {
		chainHeightAfter, err := getChainHeadByHeight(i + 1)
		if err != nil {
			log.Logger.Error("ERROR: handleMsgGasInfo() getChainHeadByHeight height:%+v err=%+v", dealBlcokHeight, err)
			return end, err
		}
		//chainHeightHandle, err := getChainHeadByHeight(i)
		//if err != nil {
		//	log.Logger.Error("ERROR: handleRequestInfo() getChainHeadByHeight height:%+v err=%+v", dealBlcokHeight-1, err)
		//	return end, err
		//}
		blockMessageResp, err := getParentsBlockMessage(chainHeightAfter.Cids()[0])
		if err != nil {
			log.Logger.Error("ERROR: handleMsgGasInfo() getParentsBlockMessage cid %s  err=%v", chainHeightAfter.Cids()[0].String(), err)
			return end, err
		}

		//
		blocks := chainHeightHandle.Blocks()
		//timeStamp:=blocks[0].Timestamp

		//计算支出
		err = calculateWalletCost(*blocks[0], blockMessageResp, blocks[0].ParentBaseFee, chainHeightAfter.Cids()[0])
		if err != nil {
			return end, err
		}

		chainHeightHandle = chainHeightAfter

	}
	return dh, nil
}


func calculateWalletCost(block types.BlockHeader, messages []api.Message, basefee abi.TokenAmount, blockAfter cid.Cid) error {

	messagesCostMap := make(map[string]bool)

	for i, message := range messages {
		//?????
		/*if v==5{
			break
		}*/
		//计算支出
		if inWallets(message.Message.From.String())||inMiners(message.Message.From.String()) {
			if messagesCostMap[message.Cid.String()] {
				continue
			}
			//		log.Logger.Debug("======i:%+v msgID:%+v len:%+v", i, message.Cid.String(), len(messages))

			gasout, err := getGasout(blockAfter, message.Message, basefee, i)
			if err != nil {
				return err
			}
			err = recordCostMessage(gasout, message, block)
			if err != nil {
				return err
			}
			messagesCostMap[message.Cid.String()] = true
			//fmt.Println("----------------------------------------")
		}

		/*if message.Message.Method==15&&inMiners(message.Message.To.String()){
			gasout, err := getGasout(blockAfter, message.Message, basefee, i)
			if err != nil {
				return err
			}
			err = recordCostMessage(gasout, message, block)
			if err != nil {
				return err
			}
			messagesCostMap[message.Cid.String()] = true
		}*/

	}

	return nil
}

func recordCostMessage(gasout vm.GasOutputs, message api.Message, block types.BlockHeader) error {
	//获取minerid
	walletId := message.Message.From.String()
	to := message.Message.To.String()
	value := message.Message.Value
	msgId := message.Cid.String()

	//	log.Logger.Debug("Debug recordCostMessageInfo wallets:%+v", walletId)
	//查询数据
	o := orm.NewOrm()
	err := o.Begin()
	if err != nil {
		log.Logger.Debug("DEBUG: recordCostMessageInfo orm transation begin error: %+v", err)
		return err
	}

	t := time.Unix(int64(block.Timestamp), 0).Format("2006-01-02")
	epoch := block.Height.String()

	expendMsg := models.ExpendMessages{
		MessageId:          msgId,
		WalletId:           walletId,
		To:                 to,
		Epoch:              epoch,
		Gas:                gasout.MinerTip.String(),
		BaseBurnFee:        gasout.BaseFeeBurn.String(),
		OverEstimationBurn: gasout.OverEstimationBurn.String(),
		Value:              value.String(),
		Penalty:            gasout.MinerPenalty.String(),
		Method:             uint64(message.Message.Method),
		Time:               t,
		CreateTime:         block.Timestamp,
	}
	_, err = o.Insert(&expendMsg)
	if err != nil {
		log.Logger.Error("Error  Insert miner:%+v time:%+v err:%+v ", walletId, t, err)
		err := o.Rollback()
		if err != nil {
			log.Logger.Debug("DEBUG: collectWalletData orm transation rollback error: %+v", err)
		}
		return err
	}
	expendInfo := new(models.ExpendInfo)
	//入库
	n, err := o.QueryTable("fly_expend_info").Filter("wallet_id", walletId).Filter("time", t).All(expendInfo)
	if err != nil {
		log.Logger.Error("Error  QueryTable rewardInfo:%+v err:%+v num:%+v time:%+v", walletId, err, n, t)
		err := o.Rollback()
		if err != nil {
			log.Logger.Debug("DEBUG: collectWalletData orm transation rollback error: %+v", err)
		}
		return err
	}
	if n == 0 {
		//记录块收益 todo
		expendInfo.WalletId = walletId
		expendInfo.Epoch = epoch
		expendInfo.Gas = bit.TransFilToFIL(gasout.MinerTip.String())
		expendInfo.BaseBurnFee = bit.TransFilToFIL(gasout.BaseFeeBurn.String())
		expendInfo.OverEstimationBurn = bit.TransFilToFIL(gasout.OverEstimationBurn.String())
		expendInfo.Value = bit.TransFilToFIL(value.String())
		expendInfo.Time = t
		expendInfo.UpdateTime = time.Now().Unix()

		_, err = o.Insert(expendInfo)
		if err != nil {
			log.Logger.Error("Error  Insert miner:%+v time:%+v err:%+v ", walletId, t, err)
			err := o.Rollback()
			if err != nil {
				log.Logger.Debug("DEBUG: collectWalletData orm transation rollback error: %+v", err)
			}
			return err
		}
	} else {
		//记录块收益 todo
		expendInfo.Epoch = epoch
		expendInfo.Gas = bit.CalculateReward(expendInfo.Gas, bit.TransFilToFIL(gasout.MinerTip.String()))
		expendInfo.BaseBurnFee = bit.CalculateReward(expendInfo.BaseBurnFee, bit.TransFilToFIL(gasout.BaseFeeBurn.String()))
		expendInfo.OverEstimationBurn = bit.CalculateReward(expendInfo.OverEstimationBurn, bit.TransFilToFIL(gasout.OverEstimationBurn.String()))
		expendInfo.Value = bit.CalculateReward(expendInfo.Value, bit.TransFilToFIL(value.String()))
		expendInfo.UpdateTime = time.Now().Unix()
		_, err := o.Update(expendInfo)
		if err != nil {
			log.Logger.Error("Error  Update miner:%+v time:%+v err:%+v ", walletId, t, err)
			err := o.Rollback()
			if err != nil {
				log.Logger.Debug("DEBUG: collectWalletData orm transation rollback error: %+v", err)
			}
			return err
		}
	}

	err = o.Commit()
	if err != nil {
		log.Logger.Debug("DEBUG: recordCostMessageInfo orm transation Commit error: %+v", err)
		return err
	}
	return nil
}

func TestMsg()  {
	chainHeightHandle, err := getChainHeadByHeight(230056)
	if err != nil {
		log.Logger.Error("ERROR: handleRequestInfo() getChainHeadByHeight err=%+v", err)
		return
	}


		chainHeightAfter, err := getChainHeadByHeight(230057)
		if err != nil {
			log.Logger.Error("ERROR: handleRequestInfo() getChainHeadByHeight  err=%+v", err)
			return
		}
		//chainHeightHandle, err := getChainHeadByHeight(i)
		//if err != nil {
		//	log.Logger.Error("ERROR: handleRequestInfo() getChainHeadByHeight height:%+v err=%+v", dealBlcokHeight-1, err)
		//	return end, err
		//}
		blockMessageResp, err := getParentsBlockMessage(chainHeightAfter.Cids()[0])
		if err != nil {
			log.Logger.Error("ERROR: handleRequestInfo() getParentsBlockMessage   err=%v",  err)
			return
		}

		//
		blocks := chainHeightHandle.Blocks()
		//timeStamp:=blocks[0].Timestamp

		//计算支出
		err = calculateWalletCostTest(*blocks[0], blockMessageResp, blocks[0].ParentBaseFee, chainHeightAfter.Cids()[0])
		if err != nil {
			log.Logger.Error("ERROR: handleRequestInfo() calculateWalletCost   err=%v",  err)
			return
		}


}

func calculateWalletCostTest(block types.BlockHeader, messages []api.Message, basefee abi.TokenAmount, blockAfter cid.Cid) error {


	for i, message := range messages {
		//?????
		/*if v==5{
			break
		}*/
		//计算支出
		if inWallets(message.Message.From.String())||inMiners(message.Message.From.String()) {

			gasout, err := getGasout(blockAfter, message.Message, basefee, i)
			if err != nil {
				return err
			}
			log.Logger.Debug("Debug  msgid:%+v --gas:%+v",message.Cid,gasout)

			//fmt.Println("----------------------------------------")
		}

		if message.Message.Method==15&&inMiners(message.Message.To.String()){
			gasout, err := getGasout(blockAfter, message.Message, basefee, i)
			if err != nil {
				return err
			}
		//	err = recordCostMessage(gasout, message, block)
			log.Logger.Debug("Debug method == 15  msgid:%+v --gas:%+v",message.Cid,gasout)
		}

	}

	return nil
}
