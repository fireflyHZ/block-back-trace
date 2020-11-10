package reward

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/lotus/api"
	lotusClient "github.com/filecoin-project/lotus/api/client"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/v2/actors/builtin"
	cid "github.com/ipfs/go-cid"
	"net/http"
	"profit-allocation/models"
	"profit-allocation/tool/bit"
	"profit-allocation/tool/log"
	"profit-allocation/tool/sync"
	"time"
)

var DealMessageBlockHeightTmp = 148888

func CollectTotalRerwardAndPledge() {

	defer sync.Wg.Done()
	log.Logger.Debug("DEBUG: CollectTotalRerwardAndPledge()")

	if height, err := queryNetRunDataTmp(); err != nil {
		log.Logger.Error("ERROR: collectLotusChainBlockRunData(), err=%v", err)
		return
	} else {
		DealMessageBlockHeightTmp = height
	}

	log.Logger.Debug("DEBUG: collectLotusChainBlockRunData(),  DealMessageBlockHeight:%v ", DealMessageBlockHeight)

	chainHeadResp, err := collectLotusChainHeadBlock()
	if err != nil {
		log.Logger.Error("ERROR: collectLotusChainBlockRunData()  collectLotusChainHeadBlock err=%v", err)
		return
	}

	blockHeight := int(chainHeadResp.Height())
	if blockHeight <= DealMessageBlockHeightTmp+1 {
		return
	}

	if blockHeight-DealMessageBlockHeightTmp > 200 {
		//	log.Logger.Debug("DEBUG: collectLotusChainBlockRunData()  >200")

		h, err := getRewardAndPledge(DealMessageBlockHeightTmp+200, DealMessageBlockHeightTmp)
		if err != nil {
			log.Logger.Error("ERROR: collectLotusChainBlockRunData() handleRequestInfo >500 err:%+v", err)
			return
		}
		DealMessageBlockHeightTmp = h
		//log.Logger.Debug("======== >500 ok")
	} else {
		//log.Logger.Debug("DEBUG: collectLotusChainBlockRunData()  <200")

		h, err := getRewardAndPledge(blockHeight, DealMessageBlockHeightTmp)
		if err != nil {
			log.Logger.Error("ERROR: collectLotusChainBlockRunData() handleRequestInfo <=500 err:%+v", err)
			return
		}
		DealMessageBlockHeightTmp = h
		//log.Logger.Debug("======== <500 ok")
	}

	err = updateNetRunDataTmp(DealMessageBlockHeightTmp)
	if err != nil {
		fmt.Printf("updateNetRunData height:%+v err :%+v\n", DealMessageBlockHeight, err)
	}
}

func queryNetRunDataTmp() (height int, err error) {
	o := orm.NewOrm()
	netRunData := new(models.NetRunDataProTmp)
	n, err := o.QueryTable("fly_net_run_data_pro_tmp").All(netRunData)
	if err != nil {
		return
	}
	if n == 0 {
		height = -1
		return
	} else {
		height = netRunData.ReceiveBlockHeight
	}
	return
}

func updateNetRunDataTmp(height int) (err error) {
	o := orm.NewOrm()
	netRunData := new(models.NetRunDataProTmp)
	n, err := o.QueryTable("fly_net_run_data_pro_tmp").All(netRunData)
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

func getRewardAndPledge(dealBlcokHeight int, end int) (int, error) {

	//chainHeightLatest, err := getChainHeadByHeight(dealBlcokHeight)
	//if err != nil {
	//	log.Logger.Error("ERROR: handleRequestInfo() getChainHeadByHeight height:%+v err=%+v", dealBlcokHeight, err)
	//	return end, err
	//}
	chainHeightHandle, err := getChainHeadByHeight(end)
	if err != nil {
		log.Logger.Error("ERROR: handleRequestInfo() getChainHeadByHeight height:%+v err=%+v", dealBlcokHeight-1, err)
		return end, err
	}

	dh := dealBlcokHeight

	for i := end; i < dealBlcokHeight; i++ {
		chainHeightAfter, err := getChainHeadByHeight(i + 1)
		if err != nil {
			log.Logger.Error("ERROR: handleRequestInfo() getChainHeadByHeight height:%+v err=%+v", dealBlcokHeight, err)
			return end, err
		}
		//chainHeightHandle, err := getChainHeadByHeight(i)
		//if err != nil {
		//	log.Logger.Error("ERROR: handleRequestInfo() getChainHeadByHeight height:%+v err=%+v", dealBlcokHeight-1, err)
		//	return end, err
		//}
		blockMessageResp, err := getParentsBlockMessage(chainHeightAfter.Cids()[0])
		if err != nil {
			log.Logger.Error("ERROR: handleRequestInfo() getParentsBlockMessage cid %s  err=%v", chainHeightAfter.Cids()[0].String(), err)
			return end, err
		}

		blocks := chainHeightHandle.Blocks()
		for index, block := range blocks {
			if inMiners(block.Miner.String()) {
				err = calculateRewardAndPledge(index, blocks, chainHeightHandle.Cids(), chainHeightHandle.Key(), chainHeightAfter.Cids()[0], blockMessageResp)
				if err != nil {
					log.Logger.Error("ERROR: handleRequestInfo() calculateMineReward height:%+v err=%+v", dealBlcokHeight-1, err)
					return end, err
				}
			}
		}

		//parentTipsetKey := chainHeightHandle.Parents()
		chainHeightHandle = chainHeightAfter

	}
	return dh, nil
}

func calculateRewardAndPledge(index int, blocks []*types.BlockHeader, blockCid []cid.Cid, tipsetKey types.TipSetKey, blockAfter cid.Cid, messages []api.Message) error {
	//获取minerid
	miner := blocks[index].Miner.String()
	//log.Logger.Debug("Debug collectMinertData miner:%+v", miner)
	//查询数据
	o := orm.NewOrm()
	err := o.Begin()
	if err != nil {
		log.Logger.Debug("DEBUG: collectWalletData orm transation begin error: %+v", err)
		return err
	}
	t := time.Unix(int64(blocks[0].Timestamp), 0).Format("2006-01-02")
	//epoch := blocks[0].Height.String()
	epoch := int(blocks[0].Height)
	//log.Logger.Debug("Debug collectMinertData height:%+v", epoch)
	winCount := blocks[index].ElectionProof.WinCount
	value, err := calculateReward(index, blocks[index].Miner, blockCid, tipsetKey, blocks[index].ParentBaseFee, winCount, blocks[index], blockAfter, messages)
	//	log.Logger.Debug("------gas:%+v,mine:%+v,penalty:%+v,value:%+v", gas, mine, penalty, value)

	if err != nil {
		err := o.Rollback()
		if err != nil {
			log.Logger.Debug("DEBUG: collectWalletData orm transation rollback error: %+v", err)
		}
		return err
	}
	//获取质押
	pleage, err := GetMienrPleage(miner, blocks[0].Height)
	if err != nil {
		log.Logger.Error("ERROR GetMienrPleage ParseFloat err:%+v", err)
		err := o.Rollback()
		if err != nil {
			log.Logger.Debug("DEBUG: collectWalletData orm transation rollback error: %+v", err)
		}
		return err
	}
	//	log.Logger.Debug("------gas:%+v,mine:%+v,penalty:%+v,value:%+v", gas, mine, penalty, value)

	//收益分配
	minerInfo := new(models.MinerInfoTmp)
	n, err := o.QueryTable("fly_miner_info_tmp").Filter("miner_id", miner).All(minerInfo)

	if err != nil {
		log.Logger.Error("Error  QueryTable minerInfo:%+v err:%+v num:%+v ", miner, err, n)
		err := o.Rollback()
		if err != nil {
			log.Logger.Debug("DEBUG: collectMinerData orm transation rollback error: %+v", err)
		}
		return err
	}

	oldPleage := minerInfo.Pleage
	//log.Logger.Debug("-=-=-=-=-=-power:%+v old:%+v",power,oldPower)

	if n == 0 {
		return errors.New("get miner power  error")
	} else {
		//更新miner info
		minerInfo.Pleage = pleage

		minerInfo.UpdateTime = time.Now().Unix()

		_, err := o.Update(minerInfo)
		if err != nil {
			log.Logger.Error("Error  Update minerInfo miner:%+v  err:%+v ", miner, err)
			err := o.Rollback()
			if err != nil {
				log.Logger.Debug("DEBUG: collectMinerData orm transation rollback error: %+v", err)
			}
			return err
		}
	}

	rewardInfo := new(models.RewardInfoTmp)
	//入库
	n, err = o.QueryTable("fly_reward_info_tmp").Filter("miner_id", miner).Filter("time", t).All(rewardInfo)
	if err != nil {
		log.Logger.Error("Error  QueryTable rewardInfo:%+v err:%+v num:%+v time:%+v", miner, err, n, t)
		err := o.Rollback()
		if err != nil {
			log.Logger.Debug("DEBUG: collectWalletData orm transation rollback error: %+v", err)
		}
		return err
	}
	if n == 0 {
		//记录块收益 todo

		rewardInfo.Time = t
		rewardInfo.MinerId = miner
		rewardInfo.Pledge = pleage - oldPleage
		rewardInfo.Value = value
		rewardInfo.Epoch = epoch
		rewardInfo.UpdateTime = time.Now().Unix()

		_, err = o.Insert(rewardInfo)
		if err != nil {
			log.Logger.Error("Error  Insert miner:%+v time:%+v err:%+v ", miner, t, err)
			err := o.Rollback()
			if err != nil {
				log.Logger.Error("Error: collectWalletData orm transation rollback error: %+v", err)
			}
			return err
		}
	} else {
		//记录块收益 todo
		//更新walletinfo
		if rewardInfo.Epoch != epoch {

			rewardInfo.Pledge += pleage - oldPleage
			rewardInfo.Value = bit.CalculateReward(rewardInfo.Value, value)
			rewardInfo.Epoch = epoch
			rewardInfo.UpdateTime = time.Now().Unix()
			_, err := o.Update(rewardInfo)
			if err != nil {
				log.Logger.Error("Error  Update miner:%+v time:%+v err:%+v ", miner, t, err)
				err := o.Rollback()
				if err != nil {
					log.Logger.Debug("DEBUG: collectWalletData orm transation rollback error: %+v", err)
				}
				return err
			}
		}

	}
	err = updateNetRunDataTmp(epoch + 1)
	if err != nil {
		log.Logger.Error("Error  Update net run data tmp  err:%+v height:%+v ", err, epoch)
		err := o.Rollback()
		if err != nil {
			log.Logger.Debug("DEBUG: collectWalletData orm transation rollback error: %+v", err)
		}
		return err
	}
	err = o.Commit()
	if err != nil {
		log.Logger.Debug("DEBUG: collectWalletData orm transation commit error: %+v", err)
		return err
	}
	return nil
}

func calculateReward(index int, miner address.Address, blockCid []cid.Cid, tipsetKey types.TipSetKey, basefee abi.TokenAmount, winCount int64, header *types.BlockHeader, blockAfter cid.Cid, msgs []api.Message) (string, error) {
	totalGas := abi.NewTokenAmount(0)
	mineReward := abi.NewTokenAmount(0)
	totalPenalty := abi.NewTokenAmount(0)
	lotusHost := beego.AppConfig.String("lotusHost")
	requestHeader := http.Header{}
	ctx := context.Background()
	//o := orm.NewOrm()
	rewardMap := make(map[string]gasAndPenalty)
	allRewardMap := make(map[string]gasAndPenalty)
	base := gasAndPenalty{
		gas:     abi.NewTokenAmount(0),
		penalty: abi.NewTokenAmount(0),
	}
	//var totalGas string
	//var totalValue string
	//var mineReward string
	nodeApi, closer, err := lotusClient.NewFullNodeRPC(context.Background(), lotusHost, requestHeader)
	if err != nil {
		fmt.Println(err)
		return "0.0", err
	}
	defer closer()
	for i := index; i >= 0; i-- {
		if i == index {
			messages, err := nodeApi.ChainGetBlockMessages(context.Background(), blockCid[i])
			if err != nil {
				log.Logger.Error("Error getRewardInfo ChainGetBlockMessages err:%+v", err)
				return "0.0", err
			}
			for _, message := range messages.BlsMessages {
				rewardMap[message.Cid().String()] = base
			}
		} else {
			messages, err := nodeApi.ChainGetBlockMessages(context.Background(), blockCid[i])
			if err != nil {
				log.Logger.Error("Error getRewardInfo ChainGetBlockMessages err:%+v", err)
				return "0.0", err
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
		//log.Logger.Debug("======i:%+v msgID:%+v len:%+v", i, message.Cid.String(), len(msgs))
		gasout, err := getGasout(blockAfter, message.Message, basefee, i)
		if err != nil {
			return "0.0", err
		}
		//	log.Logger.Debug("7777777 gas:%+v",gasout)
		gasPenalty := gasAndPenalty{
			gas:     gasout.MinerTip,
			penalty: gasout.MinerPenalty,
		}
		allRewardMap[message.Cid.String()] = gasPenalty
	}

	for msgId, _ := range rewardMap {
		//记录收益的message todo
		var msgGas abi.TokenAmount
		var msgPenalty abi.TokenAmount
		if gas, ok := allRewardMap[msgId]; ok {
			msgGas = gas.gas
			msgPenalty = gas.penalty
		} else {
			msgGas = rewardMap[msgId].gas
			msgPenalty = rewardMap[msgId].penalty
		}
		/*mineMsg := new(models.MineMessages)
		mineMsg.MinerId = miner.String()
		mineMsg.MessageId = msgId
		mineMsg.Gas = msgGas.String()
		mineMsg.Penalty = msgPenalty.String()
		//	log.Logger.Debug("2333333 gas:%+v ,msg:%+v",msgGas.String(),msgId)
		mineMsg.Epoch = header.Height.String()
		mineMsg.Time = time.Unix(int64(header.Timestamp), 0).Format("2006-01-02")
		mineMsg.CreateTime = header.Timestamp

		_, err := o.Insert(mineMsg)
		if err != nil {
			log.Logger.Error("Error inert msg:%+v err:%+v", msgId, err)
			return "0.0", "0.0", "0.0", "0.0", 0, err
		}*/
		totalGas = big.Add(msgGas, totalGas)
		totalPenalty = big.Add(msgPenalty, totalPenalty)
	}

	rewardActor, err := nodeApi.StateGetActor(ctx, builtin.RewardActorAddr, tipsetKey)
	if err != nil {
		log.Logger.Error("StateGetActor err:%+v", err)
		return "0.0", err
	}

	rewardStateRaw, err := nodeApi.ChainReadObj(ctx, rewardActor.Head)
	if err != nil {
		log.Logger.Error("ChainReadObj err:%+v", err)
		return "0.0", err
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

	return totalValue, nil
}
