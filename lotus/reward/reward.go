package reward

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/fatih/color"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/apibstore"
	lotusClient "github.com/filecoin-project/lotus/api/client"
	"github.com/filecoin-project/lotus/chain/actors/adt"
	"github.com/filecoin-project/lotus/chain/actors/builtin/miner"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/vm"
	"github.com/filecoin-project/lotus/lib/blockstore"
	"github.com/filecoin-project/lotus/lib/bufbstore"
	"github.com/filecoin-project/specs-actors/actors/builtin/reward"
	"github.com/filecoin-project/specs-actors/v2/actors/builtin"
	cid "github.com/ipfs/go-cid"
	cbor "github.com/ipfs/go-ipld-cbor"
	"io"
	"net/http"
	"profit-allocation/models"
	"profit-allocation/tool"
	"profit-allocation/tool/bit"
	"profit-allocation/tool/log"
	"profit-allocation/tool/sync"
	"strconv"
	"time"

	smoothing "github.com/filecoin-project/specs-actors/actors/util/smoothing"
	cbg "github.com/whyrusleeping/cbor-gen"
)

var DealMessageBlockHeight = 148888
var UserInfoFundData = "2020-10-15"

func CollectLotusChainBlockRunData() {
	defer sync.Wg.Done()
	log.Logger.Debug("DEBUG: collectLotusChainBlockRunData()")

	if height, err := queryNetRunData(); err != nil {
		log.Logger.Error("ERROR: collectLotusChainBlockRunData(), err=%v", err)
		return
	} else {
		DealMessageBlockHeight = height
	}

	if data, err := queryUserInfoFundDate(); err != nil {
		log.Logger.Error("ERROR: queryUserInfoFundDate(), err=%v", err)
		return
	} else {
		UserInfoFundData = data
	}

	//全网总算力
	//if powerStr, err := collectLotusPower(""); err != nil {
	//	log.Logger.Error("ERROR: collectLotusChainBlockRunData() collectLotusPower err=%v", err)
	//	return
	//} else {
	//	NetRD.Storage = powerStr
	//}
	//log.Logger.Debug("DEBUG: collectLotusChainBlockRunData(), ReceiveBlockHeight:%v, DealMessageBlockHeight:%v , power:%+v", NetRD.ReceiveBlockHeight, NetRD.DealMessageBlockHeight, NetRD.Storage)
	log.Logger.Debug("DEBUG: collectLotusChainBlockRunData(),  DealMessageBlockHeight:%v ", DealMessageBlockHeight)

	chainHeadResp, err := collectLotusChainHeadBlock()
	if err != nil {
		log.Logger.Error("ERROR: collectLotusChainBlockRunData()  collectLotusChainHeadBlock err=%v", err)
		return
	}

	blockHeight := int(chainHeadResp.Height())
	if blockHeight <= DealMessageBlockHeight+1 {
		return
	}

	if blockHeight-DealMessageBlockHeight > 50 {
		//	log.Logger.Debug("DEBUG: collectLotusChainBlockRunData()  >200")

		h, err := handleRequestInfo(DealMessageBlockHeight+50, DealMessageBlockHeight)
		if err != nil {
			log.Logger.Error("ERROR: collectLotusChainBlockRunData() handleRequestInfo >500 err:%+v", err)
			return
		}
		DealMessageBlockHeight = h
		//log.Logger.Debug("======== >500 ok")
	} else {
		//log.Logger.Debug("DEBUG: collectLotusChainBlockRunData()  <200")

		h, err := handleRequestInfo(blockHeight, DealMessageBlockHeight)
		if err != nil {
			log.Logger.Error("ERROR: collectLotusChainBlockRunData() handleRequestInfo <=500 err:%+v", err)
			return
		}
		DealMessageBlockHeight = h
		//log.Logger.Debug("======== <500 ok")
	}

	err = updateNetRunData(DealMessageBlockHeight)
	if err != nil {
		fmt.Printf("updateNetRunData height:%+v err :%+v\n", DealMessageBlockHeight, err)
	}
}

func queryNetRunData() (height int, err error) {
	o := orm.NewOrm()
	netRunData := new(models.NetRunDataPro)
	n, err := o.QueryTable("fly_net_run_data_pro").All(netRunData)
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

func queryUserInfoFundDate() (string, error) {
	var date string
	o := orm.NewOrm()
	userInfo := new(models.UserInfo)
	n, err := o.QueryTable("fly_user_info").All(userInfo)
	if err != nil {
		return date, err
	}
	if n == 0 {
		date = UserInfoFundData
		return date, err
	} else {
		t, err := time.Parse("2006-01-02", userInfo.UpdateTime)
		if err != nil {
			return date, err
		}
		handleTime := t.AddDate(0, 0, 1)
		date = handleTime.Format("2006-01-02")
	}
	return date, nil
}

func updateNetRunData(height int) (err error) {
	o := orm.NewOrm()
	netRunData := new(models.NetRunDataPro)
	n, err := o.QueryTable("fly_net_run_data_pro").All(netRunData)
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

func collectLotusChainHeadBlock() (tipset *types.TipSet, err error) {
	lotusHost := beego.AppConfig.String("lotusHost")
	requestHeader := http.Header{}
	nodeApi, closer, err := lotusClient.NewFullNodeRPC(context.Background(), lotusHost, requestHeader)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer closer()

	//	tipset := "bafy2bzacedhn5csx747j5j7bb7qqnadzjbtau4x3izm434vnb22g5xmvtjdb6bafy2bzacec22yebi7pyefs4uzsjxellrwcb4pc7bpwa5kbqwmrixw56bevbds"
	//minfo, err := nodeApi.StateMinerInfo(context.Background(), maddr, types.EmptyTSK)

	tipset, err = nodeApi.ChainHead(context.Background())
	//log.Logger.Debug("collectLotusChainHeadBlock tipset:%+v", tipset)
	return
}

func handleRequestInfo(dealBlcokHeight int, end int) (int, error) {

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

		//
		blocks := chainHeightHandle.Blocks()
		//timeStamp:=blocks[0].Timestamp
		go userInfoFund(int64(blocks[0].Timestamp))

		for index, block := range blocks {
			if inMiners(block.Miner.String()) {
				err = calculateMineReward(index, blocks, chainHeightHandle.Cids(), chainHeightHandle.Key(), chainHeightAfter.Cids()[0], blockMessageResp)
				if err != nil {
					log.Logger.Error("ERROR: handleRequestInfo() calculateMineReward height:%+v err=%+v", dealBlcokHeight-1, err)
					return end, err
				}
			}
		}

		//计算支出
		err = calculateWalletCostAndMinerReward(*blocks[0], blockMessageResp, blocks[0].ParentBaseFee, chainHeightAfter.Cids()[0])
		if err != nil {
			return end, err
		}

		//入库

		//转换
		//parentTipsetKey := chainHeightHandle.Parents()
		chainHeightHandle = chainHeightAfter
		//chainHeightHandle, err = getBlockByTipsetKey(parentTipsetKey)
		//if err != nil {
		//	log.Logger.Error("ERROR: handleRequestInfo() getBlockByTipsetKey  err=%v", err)
		//	return 0, err
		//}
		////判断跳出
		////if height <= end+1 {
		////	return dh, nil
		////}
	}
	return dh, nil
}

func userInfoFund(t int64) {
	blockTimeStr := time.Unix(t, 0).Format("2006-01-02")
	blockTime, err := time.Parse("2006-01-02", blockTimeStr)
	if err != nil {
		return
	}
	userInfoTime, err := time.Parse("2006-01-02", UserInfoFundData)
	if err != nil {
		return
	}
	if blockTime.Unix() > userInfoTime.Unix() {
		CalculateUserFund(blockTime.Unix())
	}
}

func getChainHeadByHeight(height int) (tipset *types.TipSet, err error) {
	lotusHost := beego.AppConfig.String("lotusHost")
	requestHeader := http.Header{}
	nodeApi, closer, err := lotusClient.NewFullNodeRPC(context.Background(), lotusHost, requestHeader)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer closer()

	//	tipset := "bafy2bzacedhn5csx747j5j7bb7qqnadzjbtau4x3izm434vnb22g5xmvtjdb6bafy2bzacec22yebi7pyefs4uzsjxellrwcb4pc7bpwa5kbqwmrixw56bevbds"
	//minfo, err := nodeApi.StateMinerInfo(context.Background(), maddr, types.EmptyTSK)
	epoch := abi.ChainEpoch(height)
	tipsetKey := types.NewTipSetKey()
	tipset, err = nodeApi.ChainGetTipSetByHeight(context.Background(), epoch, tipsetKey)
	//	log.Logger.Debug("collectLotusChainHeadBlock tipset:%+v", tipset)
	return
}

/*func getBlockById(cid string) (*blockDataResp, error) {
	//cmdContent := fmt.Sprintf(`{ "jsonrpc": "2.0", "method": "Filecoin.ChainGetBlock", "params": [{"/":"%s"}], "id": 3 }`, cid)
	cmdContent := request.GetChainGetBlock(cid)
	//log.Logger.Debug("getBlockById cmd %+v", cmdContent)

	resp, err := execute(cmdContent)
	if err != nil {
		log.Logger.Error("ERROR: getBlockById() execute  err:%v", err)
		return nil, err
	}
	//log.Logger.Debug("-------getBlockById resp  %+v", string(resp))

	blockDatas := new(blockDataResp)
	err = json.Unmarshal(resp, blockDatas)
	if err != nil {
		log.Logger.Error("ERROR: getBlockById() unmarshal blockDatas resp err:%v", err)
		return nil, err
	}
	//log.Logger.Debug("DEBUG: ---------getBlockById(), blockDatas=%+v", blockDatas)
	return blockDatas, nil
}*/

func getParentsBlockMessage(cid cid.Cid) (messages []api.Message, err error) {
	lotusHost := beego.AppConfig.String("lotusHost")
	requestHeader := http.Header{}
	nodeApi, closer, err := lotusClient.NewFullNodeRPC(context.Background(), lotusHost, requestHeader)
	if err != nil {
		//fmt.Println(err)
		return
	}
	defer closer()

	messages, err = nodeApi.ChainGetParentMessages(context.Background(), cid)
	//log.Logger.Debug("collectLotusChainHeadBlock tipset:%+v", tipset)
	return
}

func getBlockByTipsetKey(tipsetKey types.TipSetKey) (tipset *types.TipSet, err error) {
	lotusHost := beego.AppConfig.String("lotusHost")
	requestHeader := http.Header{}
	nodeApi, closer, err := lotusClient.NewFullNodeRPC(context.Background(), lotusHost, requestHeader)
	if err != nil {
		//fmt.Println(err)
		return
	}
	defer closer()

	tipset, err = nodeApi.ChainGetTipSet(context.Background(), tipsetKey)
	//log.Logger.Debug("collectLotusChainHeadBlock tipset:%+v", tipset)
	return
}

func calculateWalletCostAndMinerReward(block types.BlockHeader, messages []api.Message, basefee abi.TokenAmount, blockAfter cid.Cid) error {

	messagesCostMap := make(map[string]bool)
	messagesRewardMap := make(map[string]bool)
	// var baseFeeBurn abi.TokenAmount
	// var overBaseFeeBurn abi.TokenAmount
	// var minerTip abi.TokenAmount
	// var costValue abi.TokenAmount
	// var rewardValue abi.TokenAmount
	for i, message := range messages {
		//?????
		/*if v==5{
			break
		}*/
		//计算支出
		if inWallets(message.Message.From.String()) {
			if messagesCostMap[message.Cid.String()] {
				continue
			}
			//		log.Logger.Debug("======i:%+v msgID:%+v len:%+v", i, message.Cid.String(), len(messages))

			gasout, err := getGasout(blockAfter, message.Message, basefee, i)
			if err != nil {
				return err
			}
			err = recordCostMessageInfo(gasout, message, block)
			if err != nil {
				return err
			}
			messagesCostMap[message.Cid.String()] = true
			//fmt.Println("----------------------------------------")
		}
		if inMiners(message.Message.To.String()) {
			if messagesRewardMap[message.Cid.String()] {
				continue
			}
			err := recordRewardMessageInfo(message, block)
			if err != nil {
				return err
			}
			messagesRewardMap[message.Cid.String()] = true
		}

	}

	return nil
}

func recordCostMessageInfo(gasout vm.GasOutputs, message api.Message, block types.BlockHeader) error {
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
		//记录块收益
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
		//记录块收益
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
	feeStr := bit.CalculateReward(expendInfo.Gas, expendInfo.BaseBurnFee)
	feeStr = bit.CalculateReward(feeStr, expendInfo.OverEstimationBurn)
	fee, _ := strconv.ParseFloat(feeStr, 64)
	ordersInfo := make([]models.OrderInfo, 0)
	_, err = o.QueryTable("fly_order_info").All(&ordersInfo)
	if err != nil {
		log.Logger.Error("Error  query order info time:%+v err:%+v ", walletId, t, err)
		err := o.Rollback()
		if err != nil {
			log.Logger.Debug("DEBUG: query order info transation rollback error: %+v", err)
		}
		return err
	}
	netData := new(models.NetRunDataPro)
	_, err = o.QueryTable("fly_net_run_data_pro").All(netData)
	if err != nil {
		log.Logger.Error("Error query net run data table err:%+v", err)
		err = o.Rollback()
		if err != nil {
			log.Logger.Error("Error query table rollback err:%+v", err)
		}
		return err
	}
	for _, orderInfo := range ordersInfo {
		orderFee := fee * float64(orderInfo.Share) / float64(netData.AllShare)
		orderDaliyReward := new(models.OrderDailyRewardInfo)
		n, err := o.QueryTable("fly_order_daily_reward_info").Filter("order_id", orderInfo.OrderId).Filter("time", t).All(orderDaliyReward)
		if err != nil {
			log.Logger.Error("Error  QueryTable OrderDailyRewardInfo user:%+v err:%+v", orderInfo.OrderId, err)
			err = o.Rollback()
			if err != nil {
				log.Logger.Error("Error  QueryTable rollback err:%+v", err)
			}
			return err
		}
		if n == 0 {
			orderDaliyReward.OrderId = orderInfo.OrderId
			orderDaliyReward.Fee = orderFee
			orderDaliyReward.UpdateTime = block.Timestamp
			orderDaliyReward.Time = t
			_, err = o.Insert(orderDaliyReward)
			if err != nil {
				log.Logger.Error("Error  Insert OrderDailyRewardInfo user:%+v err:%+v", orderInfo.OrderId, err)
				err = o.Rollback()
				if err != nil {
					log.Logger.Error("Error  Insert rollback err:%+v", err)
				}
				return err
			}
		} else {
			orderDaliyReward.Fee = orderFee
			orderDaliyReward.UpdateTime = block.Timestamp
			_, err = o.Update(orderDaliyReward)
			if err != nil {
				log.Logger.Error("Error  Update OrderDailyRewardInfo user:%+v err:%+v", orderInfo.OrderId, err)
				err = o.Rollback()
				if err != nil {
					log.Logger.Error("Error  Update rollback err:%+v", err)
				}
				return err
			}
		}

	}

	err = o.Commit()
	if err != nil {
		log.Logger.Debug("DEBUG: recordCostMessageInfo orm transation Commit error: %+v", err)
		return err
	}
	return nil
}

func recordRewardMessageInfo(message api.Message, block types.BlockHeader) error {
	//获取minerid
	minerId := message.Message.To.String()
	from := message.Message.From.String()
	value := message.Message.Value
	msgId := message.Cid.String()
	//log.Logger.Debug("Debug collectWalletData minerId:%+v", minerId)
	//查询数据
	o := orm.NewOrm()
	err := o.Begin()
	if err != nil {
		log.Logger.Debug("DEBUG: collectWalletData orm transation begin error: %+v", err)
		return err
	}

	t := time.Unix(int64(block.Timestamp), 0).Format("2006-01-02")
	epoch := block.Height.String()

	rewardMsg := models.RewardMessages{
		MessageId:  msgId,
		MinerId:    minerId,
		From:       from,
		Epoch:      epoch,
		Value:      value.String(),
		Method:     uint64(message.Message.Method),
		Time:       t,
		CreateTime: block.Timestamp,
	}
	_, err = o.Insert(&rewardMsg)
	if err != nil {
		log.Logger.Error("Error  Insert miner:%+v time:%+v err:%+v ", minerId, t, err)
		err := o.Rollback()
		if err != nil {
			log.Logger.Debug("DEBUG: collectWalletData orm transation rollback error: %+v", err)
		}
		return err
	}
	msgRewardInfo := new(models.MessageRewardInfo)
	//入库
	n, err := o.QueryTable("fly_message_reward_info").Filter("miner_id", minerId).Filter("time", t).All(msgRewardInfo)
	if err != nil {
		log.Logger.Error("Error  QueryTable rewardInfo:%+v err:%+v num:%+v time:%+v", minerId, err, n, t)
		err := o.Rollback()
		if err != nil {
			log.Logger.Debug("DEBUG: collectWalletData orm transation rollback error: %+v", err)
		}
		return err
	}
	if n == 0 {
		//记录块收益
		msgRewardInfo.MinerId = minerId
		msgRewardInfo.Time = t
		msgRewardInfo.Value = bit.TransFilToFIL(value.String())
		//rewardInfo.Value = value
		msgRewardInfo.Epoch = epoch
		msgRewardInfo.UpdateTime = time.Now().Unix()

		_, err = o.Insert(msgRewardInfo)
		if err != nil {
			log.Logger.Error("Error  Insert miner:%+v time:%+v err:%+v ", minerId, t, err)
			err := o.Rollback()
			if err != nil {
				log.Logger.Debug("DEBUG: collectWalletData orm transation rollback error: %+v", err)
			}
			return err
		}
	} else {
		//记录块收益

		msgRewardInfo.Value = bit.CalculateReward(msgRewardInfo.Value, bit.TransFilToFIL(value.String()))
		//rewardInfo.Value = bit.StringAdd(value, rewardInfo.Value)
		msgRewardInfo.Epoch = epoch
		msgRewardInfo.UpdateTime = time.Now().Unix()
		_, err := o.Update(msgRewardInfo)
		if err != nil {
			log.Logger.Error("Error  Update miner:%+v time:%+v err:%+v ", minerId, t, err)
			err := o.Rollback()
			if err != nil {
				log.Logger.Debug("DEBUG: collectWalletData orm transation rollback error: %+v", err)
			}
			return err
		}
	}
	err = o.Commit()
	if err != nil {
		log.Logger.Debug("DEBUG: collectWalletData orm transation Commit error: %+v", err)
		return err
	}
	return nil
}

func getGasout(blockCid cid.Cid, messages *types.Message, basefee abi.TokenAmount, i int) (gasout vm.GasOutputs, err error) {
	lotusHost := beego.AppConfig.String("lotusHost")
	requestHeader := http.Header{}
	ctx := context.Background()

	nodeApi, closer, err := lotusClient.NewFullNodeRPC(ctx, lotusHost, requestHeader)
	if err != nil {
		log.Logger.Error("getGasout  NewFullNodeRPC err:%+v", err)
		return
	}
	defer closer()
	//fmt.Println("3333333")

	resp, err := nodeApi.ChainGetParentReceipts(ctx, blockCid)
	//fmt.Printf("resp :%+v", resp[0])
	if err != nil {
		log.Logger.Error("getGasout  ChainGetParentReceipts err:%+v", err)
		return
	}
	/*	for in, r := range resp {
		log.Logger.Debug("11111111 i:%+v  in:%+v used:%+v\n ", i, in, r.GasUsed)
	}*/

	gasout = vm.ComputeGasOutputs(resp[i].GasUsed, messages.GasLimit, basefee, messages.GasFeeCap, messages.GasPremium)
	//fmt.Printf("gasout:%+v\n", gasout)
	return
}

func getBlockMessage(blockCid cid.Cid) (blockMsg *api.BlockMessages) {
	lotusHost := beego.AppConfig.String("lotusHost")
	requestHeader := http.Header{}
	ctx := context.Background()

	nodeApi, closer, err := lotusClient.NewFullNodeRPC(ctx, lotusHost, requestHeader)
	if err != nil {
		log.Logger.Error("getBlockMessage  NewFullNodeRPC err:%+v", err)
		return
	}
	defer closer()
	blockMsg, err = nodeApi.ChainGetBlockMessages(ctx, blockCid)
	if err != nil {
		log.Logger.Error("getBlockMessage  ChainGetBlockMessages err:%+v", err)
		return
	}
	return
}

func calculateMineReward(index int, blocks []*types.BlockHeader, blockCid []cid.Cid, tipsetKey types.TipSetKey, blockAfter cid.Cid, messages []api.Message) error {
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
	gas, mine, penalty, value, power, err := getRewardInfo(index, blocks[index].Miner, blockCid, tipsetKey, blocks[index].ParentBaseFee, winCount, blocks[index], blockAfter, messages)
	//	log.Logger.Debug("------gas:%+v,mine:%+v,penalty:%+v,value:%+v", gas, mine, penalty, value)

	if err != nil {
		err := o.Rollback()
		if err != nil {
			log.Logger.Debug("DEBUG: collectWalletData orm transation rollback error: %+v", err)
		}
		return err
	}
	//获取质押
	_,_,_,pleage, err := GetMienrPleage(miner, blocks[0].Height)
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
	minerInfo := new(models.MinerInfo)
	n, err := o.QueryTable("fly_miner_info").Filter("miner_id", miner).All(minerInfo)

	if err != nil {
		log.Logger.Error("Error  QueryTable minerInfo:%+v err:%+v num:%+v ", miner, err, n)
		err := o.Rollback()
		if err != nil {
			log.Logger.Debug("DEBUG: collectMinerData orm transation rollback error: %+v", err)
		}
		return err
	}
	oldPower := minerInfo.QualityPower
	oldPleage := minerInfo.Pleage
	//log.Logger.Debug("-=-=-=-=-=-power:%+v old:%+v",power,oldPower)

	if n == 0 {
		return errors.New("get miner power  error")
	} else {
		//更新miner info
		minerInfo.QualityPower = power
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
	err = allocation(o, value, power-oldPower, pleage-oldPleage, epoch, miner, blocks[index].Timestamp)
	if err != nil {
		err := o.Rollback()
		if err != nil {
			log.Logger.Debug("DEBUG: collectWalletData orm transation rollback error: %+v", err)
		}
		return err
	}

	//-----------------------------------------
	mineBlock := models.MineBlocks{
		MinerId:    miner,
		Epoch:      epoch,
		Reward:     mine,
		Gas:        gas,
		Penalty:    penalty,
		Value:      value,
		Power:      power - oldPower,
		Time:       time.Unix(int64(blocks[0].Timestamp), 0).Format("2006-01-02"),
		CreateTime: blocks[0].Timestamp,
	}
	_, err = o.Insert(&mineBlock)
	if err != nil {
		log.Logger.Error("Error  Insert mineBlock:%+v err:%+v ", blocks[index].Cid(), err)
		err := o.Rollback()
		if err != nil {
			log.Logger.Debug("DEBUG: collectWalletData orm transation rollback error: %+v", err)
		}
		return err
	}

	rewardInfo := new(models.RewardInfo)
	//入库
	n, err = o.QueryTable("fly_reward_info").Filter("miner_id", miner).Filter("time", t).All(rewardInfo)
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
		rewardInfo.Reward = mine
		rewardInfo.Time = t
		rewardInfo.MinerId = miner
		rewardInfo.Gas = gas
		rewardInfo.Penalty = penalty
		rewardInfo.Value = value
		rewardInfo.Epoch = epoch
		rewardInfo.Power = power - oldPower

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
			rewardInfo.Reward = bit.CalculateReward(rewardInfo.Reward, mine)
			//rewardInfo.Time=t
			//rewardInfo.MinerId=minerId
			rewardInfo.Gas = bit.CalculateReward(rewardInfo.Gas, gas)
			rewardInfo.Penalty = bit.CalculateReward(rewardInfo.Penalty, penalty)
			rewardInfo.Power = power - oldPower
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

	err = updateNetRunData(epoch + 1)
	if err != nil {
		log.Logger.Error("Error  Update net run data  err:%+v height:%+v ", err, epoch)
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

type gasAndPenalty struct {
	gas     abi.TokenAmount
	penalty abi.TokenAmount
}

func getRewardInfo(index int, miner address.Address, blockCid []cid.Cid, tipsetKey types.TipSetKey, basefee abi.TokenAmount, winCount int64, header *types.BlockHeader, blockAfter cid.Cid, msgs []api.Message) (string, string, string, string, float64, error) {
	totalGas := abi.NewTokenAmount(0)
	mineReward := abi.NewTokenAmount(0)
	totalPenalty := abi.NewTokenAmount(0)
	lotusHost := beego.AppConfig.String("lotusHost")
	requestHeader := http.Header{}
	ctx := context.Background()
	o := orm.NewOrm()
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
		log.Logger.Error("getRewardInfo NewFullNodeRPC err:%+v", err)
		return "0.0", "0.0", "0.0", "0.0", 0, err
	}
	defer closer()
	for i := index; i >= 0; i-- {
		//自己挖出块的msgs
		if i == index {
			messages, err := nodeApi.ChainGetBlockMessages(context.Background(), blockCid[i])
			if err != nil {
				log.Logger.Error("Error getRewardInfo ChainGetBlockMessages err:%+v", err)
				return "0.0", "0.0", "0.0", "0.0", 0, err
			}
			for _, message := range messages.BlsMessages {
				rewardMap[message.Cid().String()] = base
			}
		} else {
			//在自己出块之前的矿工打包的msgs
			messages, err := nodeApi.ChainGetBlockMessages(context.Background(), blockCid[i])
			if err != nil {
				log.Logger.Error("Error getRewardInfo ChainGetBlockMessages err:%+v", err)
				return "0.0", "0.0", "0.0", "0.0", 0, err
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
			return "0.0", "0.0", "0.0", "0.0", 0, err
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
		mineMsg := new(models.MineMessages)
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
		}
		totalGas = big.Add(msgGas, totalGas)
		totalPenalty = big.Add(msgPenalty, totalPenalty)
	}

	rewardActor, err := nodeApi.StateGetActor(ctx, builtin.RewardActorAddr, tipsetKey)
	if err != nil {
		log.Logger.Error("StateGetActor err:%+v", err)
		return "0.0", "0.0", "0.0", "0.0", 0, err
	}

	rewardStateRaw, err := nodeApi.ChainReadObj(ctx, rewardActor.Head)
	if err != nil {
		log.Logger.Error("ChainReadObj err:%+v", err)
		return "0.0", "0.0", "0.0", "0.0", 0, err
	}
	//fmt.Println("ChainReadObj resp", string(rewardStateRaw))

	//记录miner的算力
	power, err := nodeApi.StateMinerPower(ctx, miner, tipsetKey)
	if err != nil {
		log.Logger.Error("StateMinerPower err:%+v", err)
		return "0.0", "0.0", "0.0", "0.0", 0, err
	}
	var f float64 = 1024
	minerPower := float64(power.MinerPower.QualityAdjPower.Int64()) / f / f / f / f
	//fmt.Printf("power %+v total:%+v \n", power, power.TotalPower.QualityAdjPower.Int64()/1024/1024/1024/1024/1024)

	//------------------------------------------------
	r := bytes.NewReader(rewardStateRaw)
	rewardActorState := unmarshalState(r)
	//mineReward = big.Add(rewardActorState.ThisEpochReward,rewardActorState.ThisEpochReward)
	mineReward = big.Div(rewardActorState.ThisEpochReward, abi.NewTokenAmount(5))
	mineReward = big.Mul(mineReward, abi.NewTokenAmount(winCount))
	//log.Logger.Debug("Debug thisepoch wincount :%+v",winCount)
	//log.Logger.Debug("Debug thisepoch reward:%+v  gas:%+v",mineReward,totalGas)
	//log.Logger.Debug("=====Debug thisepoch reward:%+v  gas:%+v", mineReward, totalGas)
	mine := bit.TransFilToFIL(mineReward.String())
	gas := bit.TransFilToFIL(totalGas.String())
	penalty := bit.TransFilToFIL(totalPenalty.String())
	value := big.Sub(big.Add(mineReward, totalGas), totalPenalty)
	if value.LessThan(abi.NewTokenAmount(0)) {
		value = abi.NewTokenAmount(0)
	}
	totalValue := bit.TransFilToFIL(value.String())
	//mine,err:=strconv.ParseFloat(mineStr, 64)
	//gas,err:=strconv.ParseFloat(gasStr, 64)
	//	log.Logger.Debug("_______%+v",bit.CalculateReward(mine,gas))
	//log.Logger.Debug("=====Debug thisepoch reward:%+v  gas:%+v", mine, gas)

	//-----------------------------------------------
	//fmt.Printf("Debug get info %+v  minereward:%+v\n", rewardActorState, mineReward)

	return gas, mine, penalty, totalValue, minerPower, nil
}

func allocation(o orm.Ormer, mine string, power float64, pleage float64, epoch int, miner string, timestamp uint64) error {
	//log.Logger.Debug("allocation  epoch:%+v", epoch)
	netData := new(models.NetRunDataPro)
	_, err := o.QueryTable("fly_net_run_data_pro").All(netData)
	if err != nil {
		log.Logger.Error("Error query net run data table err:%+v", err)
		err = o.Rollback()
		if err != nil {
			log.Logger.Error("Error query table rollback err:%+v", err)
		}
		return err
	}
	profitOrders := make([]models.OrderInfo, 0)
	commonOrders := make([]models.OrderInfo, 0)
	var allocatePower int
	_, err = o.QueryTable("fly_order_info").Filter("power__gt", 10).All(&profitOrders)
	if err != nil {
		log.Logger.Error("Error query table err:%+v", err)
		err = o.Rollback()
		if err != nil {
			log.Logger.Error("Error query table rollback err:%+v", err)
		}
		return err
	}

	_, err = o.QueryTable("fly_order_info").Filter("power__lte", 10).All(&commonOrders)
	if err != nil {
		log.Logger.Error("Error query table err:%+v", err)
		err = o.Rollback()
		if err != nil {
			log.Logger.Error("Error query table rollback err:%+v", err)
		}
		return err
	}
	//要分配的总算力
	for _, order := range profitOrders {
		allocatePower += order.Share
	}
	allocatePower += netData.AllShare - netData.TotalShare
	log.Logger.Debug("DEBUG allocatePower :%+v", allocatePower)
	mineFloat, err := strconv.ParseFloat(mine, 64)
	mineFloat *= 0.8
	if err != nil {
		log.Logger.Error("Error  ParseFloat err:%+v", err)
		err = o.Rollback()
		if err != nil {
			log.Logger.Error("Error  ParseFloat rollback err:%+v", err)
		}
		return err
	}
	for _, order := range profitOrders {
		//分配收益
		if order.Epoch < epoch {
			order.Reward += float64(order.Share) / float64(allocatePower) * mineFloat
			//分配算力

			order.Power += power * float64(order.Share) / float64(netData.AllShare)
			//log.Logger.Debug("------ share:%+v order:%+v total:%+v", float64(order.Share) /float64(netData.TotalShare), order.Share, netData.TotalShare)

			//log.Logger.Debug("------Update profitUsers reward:%+v increass %+v userPower:%+v", order.Reward, power*float64(int(order.Share)/netData.TotalShare), order.Power)
			//log.Logger.Debug("======Update profitUsers:%+v power %+v share:%+v", user.UserId, power,user.Share)
			_, err = o.Update(&order)
			if err != nil {
				log.Logger.Error("Error  Update profitUsers:%+v err:%+v", order.UserId, err)
				err = o.Rollback()
				if err != nil {
					log.Logger.Error("Error  Update rollback err:%+v", err)
				}
				return err
			}
			increaseReward := float64(order.Share) / float64(allocatePower) * mineFloat
			increasePower := power * float64(order.Share) / float64(netData.AllShare)
			increasePleage := pleage * float64(order.Share) / float64(netData.AllShare)
			err = recordUserBlockAndDailyReward(o, order.OrderId, increaseReward, increasePower, increasePleage, epoch, miner, timestamp)
			if err != nil {
				log.Logger.Error("Error  Update profitUsers:%+v err:%+v", order.UserId, err)
				err = o.Rollback()
				if err != nil {
					log.Logger.Error("Error  Update rollback err:%+v", err)
				}
				return err
			}
		}

	}
	for _, order := range commonOrders {
		//分配算力
		if order.Epoch < epoch {
			order.Power += power * float64(order.Share) / float64(netData.AllShare)
			//log.Logger.Debug("------Update profitUsers:%+v increass %+v userPower:%+v", user.UserId, power*user.Share, user.Power)

			_, err = o.Update(&order)
			if err != nil {
				log.Logger.Error("Error  Update commonUsers:%+v err:%+v", order.UserId, err)
				err = o.Rollback()
				if err != nil {
					log.Logger.Error("Error  Update rollback err:%+v", err)
				}
				return err
			}
			increasePower := power * float64(order.Share) / float64(netData.AllShare)
			increasePleage := pleage * float64(order.Share) / float64(netData.AllShare)
			err = recordUserBlockAndDailyReward(o, order.OrderId, 0, increasePower, increasePleage, epoch, miner, timestamp)
			if err != nil {
				log.Logger.Error("Error  Update profitUsers:%+v err:%+v", order.OrderId, err)
				err = o.Rollback()
				if err != nil {
					log.Logger.Error("Error  Update rollback err:%+v", err)
				}
				return err
			}
		}

	}

	return nil
}

func recordUserBlockAndDailyReward(o orm.Ormer, orderId int, reward, power, pleage float64, epoch int, miner string, timestamp uint64) error {
	//插入块收益
	if orderId == 2 {
		log.Logger.Debug("recordUserBlockAndDailyReward reward :%+v epoch :%+v", reward, epoch)
	}
	orderEpochReward := new(models.OrderBlockRewardInfo)
	orderEpochReward.OrderId = orderId
	orderEpochReward.Reward = reward
	orderEpochReward.Power = power
	orderEpochReward.Epoch = epoch
	orderEpochReward.MinerId = miner
	orderEpochReward.CreateTime = timestamp
	_, err := o.Insert(orderEpochReward)
	if err != nil {
		log.Logger.Error("Error  isnert Table OrderBlockRewardInfo user:%+v err:%+v", orderId, err)
		err = o.Rollback()
		if err != nil {
			log.Logger.Error("Error  insert Table rollback err:%+v", err)
		}
		return err
	}
	//更新日收益
	t := time.Unix(int64(timestamp), 0).Format("2006-01-02")
	orderDaliyReward := new(models.OrderDailyRewardInfo)

	n, err := o.QueryTable("fly_order_daily_reward_info").Filter("order_id", orderId).Filter("time", t).All(orderDaliyReward)
	if err != nil {
		log.Logger.Error("Error  QueryTable OrderDailyRewardInfo user:%+v err:%+v", orderId, err)
		err = o.Rollback()
		if err != nil {
			log.Logger.Error("Error  QueryTable rollback err:%+v", err)
		}
		return err
	}
	if n == 0 {
		orderDaliyReward.OrderId = orderId
		orderDaliyReward.Reward = reward
		orderDaliyReward.Pleage = pleage
		orderDaliyReward.Power = power
		orderDaliyReward.Epoch = epoch
		orderDaliyReward.UpdateTime = timestamp
		orderDaliyReward.Time = t
		_, err = o.Insert(orderDaliyReward)
		if err != nil {
			log.Logger.Error("Error  Insert OrderDailyRewardInfo user:%+v err:%+v", orderId, err)
			err = o.Rollback()
			if err != nil {
				log.Logger.Error("Error  Insert rollback err:%+v", err)
			}
			return err
		}
	} else {
		orderDaliyReward.Reward += reward
		orderDaliyReward.Pleage += pleage
		orderDaliyReward.Power += power
		orderDaliyReward.Epoch = epoch
		orderDaliyReward.UpdateTime = timestamp
		_, err = o.Update(orderDaliyReward)
		if err != nil {
			log.Logger.Error("Error  Update OrderDailyRewardInfo user:%+v err:%+v", orderId, err)
			err = o.Rollback()
			if err != nil {
				log.Logger.Error("Error  Update rollback err:%+v", err)
			}
			return err
		}
	}
	return nil
}

func unmarshalState(r io.Reader) *reward.State {
	rewardActorState := new(reward.State)

	br := cbg.GetPeeker(r)
	scratch := make([]byte, 8)

	maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		log.Logger.Error("CborReadHeaderBuf err:%+v", err)
	}
	if maj != cbg.MajArray {
		log.Logger.Debug("maj : %+v", maj)
	}

	if extra != 9 {
		log.Logger.Debug("extra != 9 extra :%+v", extra)
	}

	// t.CumsumBaseline (big.Int) (struct)

	{

		if err := rewardActorState.CumsumBaseline.UnmarshalCBOR(br); err != nil {
			log.Logger.Error("CumsumBaseline err : %+v", err)
		}

	}
	// t.CumsumRealized (big.Int) (struct)

	{

		if err := rewardActorState.CumsumRealized.UnmarshalCBOR(br); err != nil {
			log.Logger.Error("CumsumRealized err : %+v", err)
		}

	}
	// t.EffectiveNetworkTime (abi.ChainEpoch) (int64)
	{
		maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
		var extraI int64
		fmt.Println("maj", maj, "extar", extra)
		if err != nil {
			log.Logger.Error("CborReadHeaderBuf err : %+v", err)
		}
		switch maj {
		case cbg.MajUnsignedInt:
			extraI = int64(extra)
			if extraI < 0 {
				log.Logger.Debug("int64 positive overflow")
			}
		case cbg.MajNegativeInt:
			extraI = int64(extra)
			if extraI < 0 {
				log.Logger.Debug("int64 negative oveflow")
			}
			extraI = -1 - extraI
		default:
			log.Logger.Debug("wrong type for int64 field: %d ", maj)
		}

		rewardActorState.EffectiveNetworkTime = abi.ChainEpoch(extraI)
	}
	// t.EffectiveBaselinePower (big.Int) (struct)

	{

		if err := rewardActorState.EffectiveBaselinePower.UnmarshalCBOR(br); err != nil {
			log.Logger.Error("unmarshaling t.EffectiveBaselinePower: %+v", err)
		}

	}
	// t.ThisEpochReward (big.Int) (struct)

	{

		if err := rewardActorState.ThisEpochReward.UnmarshalCBOR(br); err != nil {
			log.Logger.Error("unmarshaling t.ThisEpochReward: %+v", err)
		}

	}
	// t.ThisEpochRewardSmoothed (smoothing.FilterEstimate) (struct)

	{

		b, err := br.ReadByte()
		if err != nil {
			log.Logger.Error("unmarshaling t.ReadByte: %+v", err)
		}
		if b != cbg.CborNull[0] {
			if err := br.UnreadByte(); err != nil {
				log.Logger.Error("unmarshaling t.UnreadByte: %+v", err)
			}
			rewardActorState.ThisEpochRewardSmoothed = new(smoothing.FilterEstimate)
			if err := rewardActorState.ThisEpochRewardSmoothed.UnmarshalCBOR(br); err != nil {
				log.Logger.Error("ThisEpochRewardSmoothed: %+v", err)
			}
		}

	}
	// t.ThisEpochBaselinePower (big.Int) (struct)

	{

		if err := rewardActorState.ThisEpochBaselinePower.UnmarshalCBOR(br); err != nil {
			log.Logger.Error("ThisEpochBaselinePower: %+v", err)
		}

	}
	// t.Epoch (abi.ChainEpoch) (int64)
	{
		maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
		var extraI int64
		if err != nil {
			log.Logger.Error("CborReadHeaderBuf: %+v", err)
		}
		switch maj {
		case cbg.MajUnsignedInt:
			extraI = int64(extra)
			if extraI < 0 {
				log.Logger.Debug("int64 positive overflow")
			}
		case cbg.MajNegativeInt:
			extraI = int64(extra)
			if extraI < 0 {
				log.Logger.Debug("int64 negative oveflow")
			}
			extraI = -1 - extraI
		default:
			log.Logger.Debug("wrong type for int64 field: %+v", maj)
		}

		rewardActorState.Epoch = abi.ChainEpoch(extraI)
	}
	// t.TotalMined (big.Int) (struct)

	{

		if err := rewardActorState.TotalMined.UnmarshalCBOR(br); err != nil {
			log.Logger.Error("unmarshaling t.TotalMined: %+v", err)
		}

	}
	return rewardActorState
}

func inWallets(walletId string) bool {

	for _, wallet := range tool.Wallets {
		if wallet == walletId {
			return true
		}
	}
	return false
}

func inMiners(minerId string) bool {

	for _, miner := range tool.Miners {
		if miner == minerId {
			return true
		}
	}
	return false
}

func TetsGetInfo() {
	lotusHost := beego.AppConfig.String("lotusHost")
	requestHeader := http.Header{}
	requestHeader.Add("Content-Type", "application/json")
	nodeApi, closer, err := lotusClient.NewFullNodeRPC(context.Background(), lotusHost, requestHeader)
	if err != nil {
		fmt.Println("NewFullNodeRPC err:", err)
		return
	}
	defer closer()
	//block,err:=nodeApi.ChainHead(context.Background())
	var epoch = abi.ChainEpoch(326420)
	tipset, _ := nodeApi.ChainHead(context.Background())
	fmt.Printf("444444%+v \n ", tipset.Height())
	t := types.NewTipSetKey()
	blocks, err := nodeApi.ChainGetTipSetByHeight(context.Background(), epoch, t)
	if err != nil {
		//	log.Logger.Error("Error get chain head err:%+v",err)
		fmt.Printf("Error get chain head err:%+v\n", err)
		return
	}
	minerAddr, _ := address.NewFromString("f088290")
	p, _ := nodeApi.StateMinerPower(context.Background(), minerAddr, blocks.Key())
	fmt.Printf("==========%+v\n", p)

	//---------------------
	ctx := context.Background()
	mact, err := nodeApi.StateGetActor(ctx, minerAddr, blocks.Key())
	if err != nil {
		fmt.Println(err)
	}

	tbs := bufbstore.NewTieredBstore(apibstore.NewAPIBlockstore(nodeApi), blockstore.NewTemporary())
	mas, err := miner.Load(adt.WrapStore(ctx, cbor.NewCborStore(tbs)), mact)
	if err != nil {
		fmt.Println(err)
	}
	lockedFunds, err := mas.LockedFunds()
	if err != nil {
		fmt.Println(err)
	}
	availBalance, err := mas.AvailableBalance(mact.Balance)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Miner Balance: %s\n", color.YellowString("%s", types.FIL(mact.Balance)))
	fmt.Printf("\tPreCommit:   %s\n", types.FIL(lockedFunds.PreCommitDeposits))
	fmt.Printf("\tPledge:      %s\n", types.FIL(lockedFunds.InitialPledgeRequirement))
	fmt.Printf("\tVesting:     %s\n", types.FIL(lockedFunds.VestingFunds))
	color.Green("\tAvailable:   %s", types.FIL(availBalance))

}
