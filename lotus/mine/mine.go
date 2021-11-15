package mine

import (
	"context"
	"errors"
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api"
	lotusClient "github.com/filecoin-project/lotus/api/client"
	"github.com/filecoin-project/lotus/api/v0api"
	"github.com/filecoin-project/lotus/chain/gen"
	"github.com/filecoin-project/lotus/chain/types"
	logging "github.com/ipfs/go-log/v2"
	"net/http"
	"os"
	"profit-allocation/models"
	"strconv"
	"strings"
	"sync"
	"time"
)

var log = logging.Logger("mine-log")

func CalculateMineRight() {
	walletNodeApi, walletClose, err := initWalletClient()
	if err != nil {
		log.Errorf("init wallet client err=%+v", err)
		return
	}
	defer walletClose()

	dataNodeApi, dataClose, err := initDataClient()
	if err != nil {
		log.Errorf("init data client err=%+v", err)
		return
	}
	defer dataClose()

	ctx := context.Background()
	var round int64
	if height, err := getCalculateMineRightStatus(); err != nil {
		log.Errorf("get calculate mine right status error: %+v", err)
		return
	} else if height == 0 {
		head, err := dataNodeApi.ChainHead(ctx)
		if err != nil {
			log.Errorf("get chain head error: %v", err)
			return
		}
		round = int64(head.Height()) - 120
	} else {
		round = height
	}
	for {
		head, err := dataNodeApi.ChainHead(ctx)
		if err != nil {
			log.Errorf("get chain head error: %v", err)
			niceSleep()
			continue
		}
		if round > int64(head.Height())-10 {
			niceSleep()
			continue
		}
		log.Infof("calculate mine right round: %v ", round)
		err = calculate(round, walletNodeApi, dataNodeApi)
		if err != nil {
			log.Errorf("calculate error:%+v ", err)
			niceSleep()
			continue
		}
		log.Debug("calculate complete")
		err = updateCalculateMineRightStatus(round)
		if err != nil {
			log.Errorf("update calculate mine right status error:%+v ", err)
			continue
		}
		round++
		log.Debugf("complete round:%+v", round)
	}
}

func initWalletClient() (v0api.FullNode, jsonrpc.ClientCloser, error) {
	walletLotusHost := os.Getenv("WALLTE_LOTUS")
	walletToken := os.Getenv("WALLET_LOTUS_TOKEN")
	if walletLotusHost == "" || walletToken == "" {
		log.Errorf("WALLTE_LOTUS:%+v or WALLET_LOTUS_TOKEN:%+v is not set", walletLotusHost, walletToken)
		return nil, nil, errors.New("lotus info not set")
	}
	//ctx := context.Background()
	walletRequestHeader := http.Header{}
	walletRequestHeader.Add("Content-Type", "application/json")
	walletTokenHeader := fmt.Sprintf("Bearer %s", walletToken)
	walletRequestHeader.Set("Authorization", walletTokenHeader)
	return lotusClient.NewFullNodeRPCV0(context.Background(), walletLotusHost, walletRequestHeader)
}

func initDataClient() (v0api.FullNode, jsonrpc.ClientCloser, error) {
	dataLotusHost := os.Getenv("DATA_LOTUS")
	dataToken := os.Getenv("DATA_LOTUS_TOKEN")
	if dataLotusHost == "" || dataToken == "" {
		log.Errorf("DATA_LOTUS:%+v or DATA_LOTUS_TOKEN:%+v is not set", dataLotusHost, dataToken)
		return nil, nil, errors.New("lotus info not set")
	}
	dataRequestHeader := http.Header{}
	dataRequestHeader.Add("Content-Type", "application/json")
	dataTokenHeader := fmt.Sprintf("Bearer %s", dataToken)
	dataRequestHeader.Set("Authorization", dataTokenHeader)
	return lotusClient.NewFullNodeRPCV0(context.Background(), dataLotusHost, dataRequestHeader)
}

func getCalculateMineRightStatus() (height int64, err error) {
	o := orm.NewOrm()
	status := new(models.CalculateMineRightStatus)
	n, err := o.QueryTable("fly_calculate_mine_right_status").All(status)
	if err != nil {
		return
	}
	if n == 0 {
		height = 1180080
		return
	} else {
		height = status.ReceiveBlockHeight
	}
	return
}
func updateCalculateMineRightStatus(height int64) (err error) {
	o := orm.NewOrm()
	status := new(models.CalculateMineRightStatus)
	n, err := o.QueryTable("fly_calculate_mine_right_status").All(status)
	if err != nil {
		return
	}
	if n == 0 {
		status.ReceiveBlockHeight = height
		status.CreateTime = time.Now()
		status.UpdateTime = time.Now()
		_, err = o.Insert(status)
		if err != nil {
			return err
		}
		return
	} else {
		status.ReceiveBlockHeight = height
		status.UpdateTime = time.Now()
		_, err = o.Update(status)
		if err != nil {
			return err
		}
	}
	return
}

//sleep 5s
func niceSleep() {
	time.Sleep(time.Second * 10)
}

func getMiners() ([]models.MinerInfo, error) {
	o := orm.NewOrm()
	miners := make([]models.MinerInfo, 0)
	_, err := o.QueryTable("fly_miner_info").All(&miners)
	if err != nil {
		return nil, err
	}
	return miners, nil
}
func calculate(round int64, walletNodeApi, dataNodeApi v0api.FullNode) error {
	ctx := context.Background()
	log.Debug("get miners")
	miners, err := getMiners()
	if err != nil {
		log.Errorf("get miners error:%+v", err)
		return err
	}
	log.Infof("miners number:%+v", len(miners))
	ws, err := walletNodeApi.WalletList(ctx)
	if err != nil {
		log.Errorf("wallet list error:%+v", err)
		return err
	}
	log.Infof("wallets number:%+v", len(ws))
	wait := new(sync.WaitGroup)
	success := true
	for _, m := range miners {
		wait.Add(1)
		go calculateMiner(ctx, wait, round, walletNodeApi, dataNodeApi, m, ws, &success)
	}
	wait.Wait()
	if !success {
		return errors.New("calculate not success!!!!")
	}
	return nil
}

func calculateMiner(ctx context.Context, waitMiner *sync.WaitGroup, round int64, walletNodeApi, dataNodeApi v0api.FullNode, m models.MinerInfo, ws []address.Address, success *bool) {
	defer waitMiner.Done()
	minerAddr, err := address.NewFromString(m.MinerId)
	if err != nil {
		log.Errorf("NewFromString err:%+v", err)
		*success = false
		return
	}

	tp, err := dataNodeApi.ChainGetTipSetByHeight(ctx, abi.ChainEpoch(round-1), types.NewTipSetKey())
	if err != nil {
		log.Errorf("ChainGetTipSetByHeight err:%+v", err)
		*success = false
		return
	}

	mbi, err := dataNodeApi.MinerGetBaseInfo(ctx, minerAddr, abi.ChainEpoch(round), tp.Key())
	if err != nil {
		if strings.Contains(err.Error(), "actor not found") {
			return
		}
		*success = false
		log.Errorf("MinerGetBaseInfo err:%+v", err)
		return
	}

	if mbi == nil {
		log.Warnf("miner: %+v epoch: %+v mbi is nil", m.MinerId, round)
		return
	}
	if !mbi.EligibleForMining {
		// slashed or just have no power yet
		log.Warnf("eligible!!!!!!!!!!!")
		return
	}

	beaconPrev := mbi.PrevBeaconEntry
	bvals := mbi.BeaconEntries

	rbase := beaconPrev
	if len(bvals) > 0 {
		rbase = bvals[len(bvals)-1]
	}
	//wait := new(sync.WaitGroup)
	for _, w := range ws {
		//wait.Add(1)
		//go calculateWallet(ctx, wait, round, walletNodeApi, dataNodeApi, minerAddr, tp, w, rbase, success)
		calculateWallet(ctx, round, walletNodeApi, mbi, minerAddr, tp, w, rbase, success)
	}
	//wait.Wait()
}

func calculateWallet(ctx context.Context, round int64, walletNodeApi v0api.FullNode, mbi *api.MiningBaseInfo, minerAddr address.Address, tp *types.TipSet, w address.Address, rbase types.BeaconEntry, success *bool) {
	//defer waitWallet.Done()
	//mbi, err := dataNodeApi.MinerGetBaseInfo(ctx, minerAddr, abi.ChainEpoch(round), tp.Key())
	//if err != nil {
	//	if strings.Contains(err.Error(), "actor not found") {
	//		return
	//	}
	//	*success = false
	//	log.Errorf("MinerGetBaseInfo err:%+v", err)
	//	return
	//}
	mbi.WorkerKey = w
	p, err := gen.IsRoundWinner(ctx, tp, abi.ChainEpoch(round), minerAddr, rbase, mbi, walletNodeApi)
	if err != nil {
		log.Errorf("IsRoundWinner err:%+v", err)
		*success = false
		return
	}

	if p != nil {
		t := time.Unix(int64(tp.MinTimestamp()+30), 0)
		mr := &models.MineRight{
			MinerId:    minerAddr.String(),
			Wallet:     w.String(),
			Epoch:      round,
			WinCount:   p.WinCount,
			Time:       t,
			UpdateTime: time.Now(),
		}
		err := mr.Insert()
		if err != nil {
			*success = false
			log.Errorf("miner: %+v wallet: %+v epoch: %+v record error:%+v", minerAddr, w, round, err)
			return
		}
	}

}

func CalculateOne() {
	walletNodeApi, walletClose, err := initWalletClient()
	if err != nil {
		log.Errorf("init wallet client err=%+v", err)
		return
	}
	defer walletClose()

	dataNodeApi, dataClose, err := initDataClient()
	if err != nil {
		log.Errorf("init data client err=%+v", err)
		return
	}
	defer dataClose()
	m := os.Getenv("MINER")
	if m == "" {
		log.Error("miner not set")
		return
	}
	ctx := context.Background()
	startStr := os.Getenv("LOTUS_START")
	start, err := strconv.Atoi(startStr)
	if err != nil {
		log.Errorf("get lotus start err:%+v", err)
		return
	}
	round := int64(start)
	for {
		head, err := dataNodeApi.ChainHead(ctx)
		if err != nil {
			log.Errorf("get chain head error: %v", err)
			niceSleep()
			continue
		}
		if round > int64(head.Height())-10 {
			niceSleep()
			continue
		}
		log.Infof("calculate mine right round: %v ", round)
		err = calculateOne(round, m, walletNodeApi, dataNodeApi)
		if err != nil {
			log.Errorf("calculate error:%+v ", err)
			niceSleep()
			continue
		}
		log.Debug("calculate complete")
		round++
		log.Debugf("complete round:%+v", round)
	}
}

func calculateOne(round int64, m string, walletNodeApi, dataNodeApi v0api.FullNode) error {
	ctx := context.Background()

	ws, err := walletNodeApi.WalletList(ctx)
	if err != nil {
		log.Errorf("wallet list error:%+v", err)
		return err
	}
	log.Infof("wallets number:%+v", len(ws))
	minerAddr, err := address.NewFromString(m)
	if err != nil {
		log.Errorf("NewFromString err:%+v", err)
		return err
	}

	tp, err := dataNodeApi.ChainGetTipSetByHeight(ctx, abi.ChainEpoch(round-1), types.NewTipSetKey())
	if err != nil {
		log.Errorf("ChainGetTipSetByHeight err:%+v", err)
		return err
	}

	mbi, err := dataNodeApi.MinerGetBaseInfo(ctx, minerAddr, abi.ChainEpoch(round), tp.Key())
	if err != nil {
		if strings.Contains(err.Error(), "actor not found") {
			log.Warn("MinerGetBaseInfo err:%+v", err)

		} else {
			log.Errorf("MinerGetBaseInfo err:%+v", err)
			return err
		}
	}

	if mbi == nil {
		log.Warnf("miner: %+v epoch: %+v mbi is nil", m, round)
		return nil
	}
	if !mbi.EligibleForMining {
		// slashed or just have no power yet
		log.Errorf("eligible!!!!!!!!!!!")
		return nil
	}

	beaconPrev := mbi.PrevBeaconEntry
	bvals := mbi.BeaconEntries

	rbase := beaconPrev
	if len(bvals) > 0 {
		rbase = bvals[len(bvals)-1]
	}
	for _, w := range ws {
		mbi.WorkerKey = w
		p, err := gen.IsRoundWinner(ctx, tp, abi.ChainEpoch(round), minerAddr, rbase, mbi, walletNodeApi)
		if err != nil {
			log.Errorf("IsRoundWinner err:%+v", err)
			return err
		}

		if p != nil {
			t := time.Unix(int64(tp.MinTimestamp()+30), 0)
			mr := &models.MineRight{
				MinerId:    m,
				Wallet:     w.String(),
				Epoch:      round,
				WinCount:   p.WinCount,
				Time:       t,
				UpdateTime: time.Now(),
			}
			err := mr.Insert()
			if err != nil {
				log.Errorf("miner: %+v wallet: %+v epoch: %+v record error:%+v", m, w, round, err)
				return err
			}
		}
	}

	return nil
}
