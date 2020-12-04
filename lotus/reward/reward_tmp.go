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
	"profit-allocation/tool"
	"profit-allocation/tool/bit"
	"profit-allocation/tool/log"
	"profit-allocation/tool/sync"
	"time"
)

var DealMessageBlockHeightTmp = 221040

func CollectTotalRerwardAndPledge() {
	defer sync.Wg.Done()
	if height, err := queryNetRunDataTmp(); err != nil {
		log.Logger.Error("ERROR: collectLotusChainBlockRunData(), err=%v", err)
		return
	} else {
		//log.Logger.Debug("debug: collectLotusChainBlockRunData(), height=%v", height)
		DealMessageBlockHeightTmp = height
	}
	log.Logger.Debug("DEBUG: collectLotusChainBlockRunData(),  DealMessageBlockHeightTmp:%v ", DealMessageBlockHeightTmp)

	chainHeadResp, err := collectLotusChainHeadBlock()
	if err != nil {
		log.Logger.Error("ERROR: collectLotusChainBlockRunData()  collectLotusChainHeadBlock err=%v", err)
		return
	}

	blockHeight := int(chainHeadResp.Height())
	if blockHeight <= DealMessageBlockHeightTmp+11 {
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
		log.Logger.Error("updateNetRunData height:%+v err :%+v\n", DealMessageBlockHeight, err)
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

		blockMessageResp, err := getParentsBlockMessage(chainHeightAfter.Cids()[0])
		if err != nil {
			log.Logger.Error("ERROR: handleRequestInfo() getParentsBlockMessage cid %s  err=%v", chainHeightAfter.Cids()[0].String(), err)
			return end, err
		}

		blocks := chainHeightHandle.Blocks()
		bFlag:=true
		for index, block := range blocks {

			if inMiners(block.Miner.String()) {
				err = calculateRewardAndPledge(index, blocks, chainHeightHandle.Cids(), chainHeightHandle.Key(), chainHeightAfter.Cids()[0], blockMessageResp)
				if err != nil {
					log.Logger.Error("ERROR: handleRequestInfo() calculateMineReward height:%+v err=%+v", end, err)
					return end, err
				}
				bFlag=false
			}

		}

		//pledge && power
		//没出块，且当前高度有块
		if bFlag&&len(blocks)>0{
			t := time.Unix(int64(blocks[0].Timestamp), 0)
			if isExecutingPoint(t) {
				tStr := time.Unix(int64(blocks[0].Timestamp), 0).Format("2006-01-02")

				err = calculatePowerAndPledge(blocks[0].Height, chainHeightHandle.Key(), tStr)
				if err != nil {
					log.Logger.Error("ERROR: handleRequestInfo() calculatePowerAndPledge height:%+v err=%+v", end, err)
					return end, err
				}
			}
		}


		//parentTipsetKey := chainHeightHandle.Parents()
		chainHeightHandle = chainHeightAfter
		bFlag=true
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

func calculatePowerAndPledge(height abi.ChainEpoch, tipsetKey types.TipSetKey, t string) error {
	epoch := int(height)
	//获取质押
	o := orm.NewOrm()
	err := o.Begin()
	if err != nil {
		log.Logger.Debug("DEBUG: collectWalletData orm transation begin error: %+v", err)
		return err
	}
	for _, miner := range tool.Miners {

		pleage, err := GetMienrPleage(miner, height)
		if err != nil {
			log.Logger.Error("ERROR GetMienrPleage ParseFloat miner:%+v height:%+v err:%+v", miner, height, err)
			err := o.Rollback()
			if err != nil {
				log.Logger.Debug("DEBUG: collectWalletData orm transation rollback error: %+v", err)
			}
			return err
		}
		power, err := getMinerPower(miner, tipsetKey)
		if err != nil {
			log.Logger.Error("ERROR GetMienrPleage getMinerPower miner:%+v height:%+v err:%+v", miner, height, err)
			err := o.Rollback()
			if err != nil {
				log.Logger.Debug("DEBUG: collectWalletData orm transation rollback error: %+v", err)
			}
			return err
		}
		//	log.Logger.Debug("------gas:%+v,mine:%+v,penalty:%+v,value:%+v", gas, mine, penalty, value)
		err=putMinerPowerStatus(o,miner,power,t)
		if err != nil {
			log.Logger.Error("ERROR GetMienrPleage putMinerPowerStatus miner:%+v height:%+v err:%+v", miner, height, err)
			err := o.Rollback()
			if err != nil {
				log.Logger.Debug("DEBUG: collectWalletData orm transation rollback error: %+v", err)
			}
			return err
		}
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
		oldPower := minerInfo.QualityPower
		//log.Logger.Debug("-=-=-=-=-=-power:%+v old:%+v",power,oldPower)

		if n == 0 {
			return errors.New("get miner power  error")
		} else {
			//更新miner info
			minerInfo.Pleage = pleage
			minerInfo.QualityPower = power

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
				log.Logger.Debug("DEBUG: calculatePowerAndPledge orm transation rollback error: %+v", err)
			}
			return err
		}
		if n == 0 {
			//记录块收益 todo

			rewardInfo.Time = t
			rewardInfo.MinerId = miner
			rewardInfo.Pledge = pleage - oldPleage
			rewardInfo.Power = power - oldPower
			rewardInfo.Value = "0.0"
			rewardInfo.Epoch = epoch
			rewardInfo.UpdateTime = time.Now().Unix()

			_, err = o.Insert(rewardInfo)
			if err != nil {
				log.Logger.Error("Error  Insert miner:%+v time:%+v err:%+v ", miner, t, err)
				err := o.Rollback()
				if err != nil {
					log.Logger.Error("Error: calculatePowerAndPledge orm transation rollback error: %+v", err)
				}
				return err
			}
		} else {
			//记录块收益 todo
			//更新walletinfo
			if rewardInfo.Epoch < epoch {

				rewardInfo.Pledge += pleage - oldPleage
				rewardInfo.Power += power - oldPower
				//rewardInfo.Value = bit.CalculateReward(rewardInfo.Value, value)
				rewardInfo.Epoch = epoch
				rewardInfo.UpdateTime = time.Now().Unix()
				_, err := o.Update(rewardInfo)
				if err != nil {
					log.Logger.Error("Error  Update miner:%+v time:%+v err:%+v ", miner, t, err)
					err := o.Rollback()
					if err != nil {
						log.Logger.Debug("DEBUG: calculatePowerAndPledge orm transation rollback error: %+v", err)
					}
					return err
				}
			}

		}
		log.Logger.Debug("Debug miner:%+v epoch:%+v  power:%+v  pledge:%+v", miner, epoch, rewardInfo.Power, rewardInfo.Pledge)

	}
	err = updateNetRunDataTmp(epoch + 1)
	if err != nil {
		log.Logger.Error("Error  Update net run data tmp  err:%+v height:%+v ", err, epoch)
		err := o.Rollback()
		if err != nil {
			log.Logger.Debug("DEBUG: calculatePowerAndPledge orm transation rollback error: %+v", err)
		}
		return err
	}
	err = o.Commit()
	if err != nil {
		log.Logger.Debug("DEBUG: calculatePowerAndPledge orm transation commit error: %+v", err)
		return err
	}
	log.Logger.Debug("Debug  calculate complete epoch:%+v ", epoch)
	return nil
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
	value, power, err := calculateReward(index, blocks[index].Miner, blockCid, tipsetKey, blocks[index].ParentBaseFee, winCount, blocks[index], blockAfter, messages)
	//	log.Logger.Debug("------gas:%+v,mine:%+v,penalty:%+v,value:%+v", gas, mine, penalty, value)
	if err != nil {
		err := o.Rollback()
		if err != nil {
			log.Logger.Debug("DEBUG: collectWalletData orm transation rollback error: %+v", err)
		}
		return err
	}
	//新增存储当天miner power状态、MinerPowerStatus
	err=putMinerPowerStatus(o,miner,power,t)
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
	oldPower := minerInfo.QualityPower
	//log.Logger.Debug("-=-=-=-=-=-power:%+v old:%+v",power,oldPower)

	if n == 0 {
		return errors.New("get miner power  error")
	} else {
		//更新miner info
		minerInfo.Pleage = pleage
		minerInfo.QualityPower = power

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
	//记录出块
	mineBlock := models.MineBlocks{
		//Id:         0,
		MinerId:    miner,
		Epoch:      epoch,
		Reward:     value,
		Gas:        "",
		Penalty:    "",
		Value:      "",
		Power:      0,
		Time:       t,
		CreateTime: blocks[0].Timestamp,
	}
	_, err = o.Insert(&mineBlock)
	if err != nil {
		log.Logger.Error("Error  Update minerInfo miner:%+v  err:%+v ", miner, err)
		err := o.Rollback()
		if err != nil {
			log.Logger.Debug("DEBUG: collectMinerData orm transation rollback error: %+v", err)
		}
		return err
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
		rewardInfo.Power = power - oldPower
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
		if rewardInfo.Epoch < epoch {

			rewardInfo.Pledge += pleage - oldPleage
			rewardInfo.Power += power - oldPower
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
	log.Logger.Debug("Debug miner:%+v epoch:%+v reward:%+v total:%+v", miner, epoch, value, rewardInfo.Value)
	return nil
}

func calculateReward(index int, miner address.Address, blockCid []cid.Cid, tipsetKey types.TipSetKey, basefee abi.TokenAmount, winCount int64, header *types.BlockHeader, blockAfter cid.Cid, msgs []api.Message) (string, float64, error) {
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
		return "0.0", 0, err
	}
	defer closer()
	for i := index; i >= 0; i-- {
		if i == index {
			messages, err := nodeApi.ChainGetBlockMessages(context.Background(), blockCid[i])
			if err != nil {
				log.Logger.Error("Error getRewardInfo ChainGetBlockMessages err:%+v", err)
				return "0.0", 0, err
			}
			for _, message := range messages.BlsMessages {
				rewardMap[message.Cid().String()] = base
			}
		} else {
			messages, err := nodeApi.ChainGetBlockMessages(context.Background(), blockCid[i])
			if err != nil {
				log.Logger.Error("Error getRewardInfo ChainGetBlockMessages err:%+v", err)
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
		//log.Logger.Debug("======i:%+v msgID:%+v len:%+v", i, message.Cid.String(), len(msgs))
		gasout, err := getGasout(blockAfter, message.Message, basefee, i)
		if err != nil {
			return "0.0", 0, err
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

		totalGas = big.Add(msgGas, totalGas)
		totalPenalty = big.Add(msgPenalty, totalPenalty)
	}

	power, err := nodeApi.StateMinerPower(ctx, miner, tipsetKey)
	if err != nil {
		log.Logger.Error("StateMinerPower err:%+v", err)
		return "0.0", 0, err
	}
	var f float64 = 1024
	minerPower := float64(power.MinerPower.QualityAdjPower.Int64()) / f / f / f / f

	rewardActor, err := nodeApi.StateGetActor(ctx, builtin.RewardActorAddr, tipsetKey)
	if err != nil {
		log.Logger.Error("StateGetActor err:%+v", err)
		return "0.0", 0, err
	}

	rewardStateRaw, err := nodeApi.ChainReadObj(ctx, rewardActor.Head)
	if err != nil {
		log.Logger.Error("ChainReadObj err:%+v", err)
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
	lotusHost := beego.AppConfig.String("lotusHost")
	requestHeader := http.Header{}
	ctx := context.Background()
	miner, err := address.NewFromString(minerStr)
	if err != nil {
		log.Logger.Error("getMinerPower NewFromString err:%+v", err)
		return 0, err
	}
	nodeApi, closer, err := lotusClient.NewFullNodeRPC(context.Background(), lotusHost, requestHeader)
	if err != nil {
		log.Logger.Error("getMinerPower NewFullNodeRPC err:%+v", err)
		return 0, err
	}
	defer closer()
	power, err := nodeApi.StateMinerPower(ctx, miner, tipsetKey)
	if err != nil {
		log.Logger.Error("StateMinerPower err:%+v", err)
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
		log.Logger.Error("ERROR: handleRequestInfo() getChainHeadByHeight err=%+v", err)
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
			log.Logger.Error("ERROR: handleRequestInfo() getChainHeadByHeight height:%+v err=%+v", i, err)
			return
		}
		//chainHeightHandle, err := getChainHeadByHeight(i)
		//if err != nil {
		//	log.Logger.Error("ERROR: handleRequestInfo() getChainHeadByHeight height:%+v err=%+v", dealBlcokHeight-1, err)
		//	return end, err
		//}
		blockMessageResp, err := getParentsBlockMessage(chainHeightAfter.Cids()[0])
		if err != nil {
			log.Logger.Error("ERROR: handleRequestInfo() getParentsBlockMessage cid %s  err=%v", chainHeightAfter.Cids()[0].String(), err)
			return
		}

		blocks := chainHeightHandle.Blocks()
		for index, block := range blocks {
			if inMiners(block.Miner.String()) {
				n++
				v, err := calculateRewardAndPledgeTest(index, blocks, chainHeightHandle.Cids(), chainHeightHandle.Key(), chainHeightAfter.Cids()[0], blockMessageResp)
				if err != nil {
					log.Logger.Error("ERROR: handleRequestInfo() calculateMineReward height:%+v err=%+v", i, err)
					return
				}
				re = bit.CalculateReward(v, re)
				log.Logger.Debug("Debug height:%+v total reward:%+v,count:%+v", i, re, n)
			}
		}
		chainHeightHandle = chainHeightAfter
		i++

	}
	log.Logger.Debug("Debug total ok")
}

func calculateRewardAndPledgeTest(index int, blocks []*types.BlockHeader, blockCid []cid.Cid, tipsetKey types.TipSetKey, blockAfter cid.Cid, messages []api.Message) (string, error) {

	epoch := int(blocks[0].Height)
	//log.Logger.Debug("Debug collectMinertData height:%+v", epoch)
	winCount := blocks[index].ElectionProof.WinCount
	value, _, err := calculateReward(index, blocks[index].Miner, blockCid, tipsetKey, blocks[index].ParentBaseFee, winCount, blocks[index], blockAfter, messages)
	if err != nil {
		return "0.0", err
	}
	log.Logger.Debug("------miner:%+v,epoch:%+v,value:%+v,wincount:%+v", blocks[index].Miner, epoch, value, winCount)

	return value, nil
}

func putMinerPowerStatus(o orm.Ormer,miner string,power float64,t string) error {
	minerPowerStatus:=new(models.MinerPowerStatus)
	num,err:=o.QueryTable("fly_miner_power_status").Filter("miner_id",miner).Filter("time",t).All(minerPowerStatus)
	if err != nil {
		log.Logger.Error("Error  QueryTable miner power status :%+v err:%+v num:%+v ", miner, err, num)
		err := o.Rollback()
		if err != nil {
			log.Logger.Debug("DEBUG: collectMinerData orm transation rollback error: %+v", err)
		}
		return err
	}
	if num==0{
		minerPowerStatus.Power=power
		minerPowerStatus.Time=t
		minerPowerStatus.MinerId=miner
		_,err=o.Insert(minerPowerStatus)
		if err != nil {
			log.Logger.Error("Error  InsertTable miner power status :%+v err:%+v ", miner, err)
			err := o.Rollback()
			if err != nil {
				log.Logger.Debug("DEBUG: collectMinerData orm transation rollback error: %+v", err)
			}
			return err
		}
	}else {
		minerPowerStatus.Power=power
		_,err=o.Update(minerPowerStatus)
		if err != nil {
			log.Logger.Error("Error  UpdateTable miner power status :%+v err:%+v ", miner, err)
			err := o.Rollback()
			if err != nil {
				log.Logger.Debug("DEBUG: collectMinerData orm transation rollback error: %+v", err)
			}
			return err
		}
	}
	return nil
}