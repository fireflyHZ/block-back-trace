package block

import (
	"context"
	"fmt"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/chain/gen"
	"github.com/filecoin-project/lotus/chain/types"
	logging "github.com/ipfs/go-log/v2"
	"profit-allocation/lotus/client"
	"profit-allocation/models"
	"sort"
	"sync"
	"time"
)

type mineBlockNum struct {
	mineNum float64
	winNum  float64
	lock    sync.Mutex
}

var groupNum abi.ChainEpoch = 5

var wait sync.WaitGroup
var blockLog = logging.Logger("block-log")

func GetMinerMineBlockPercentage(start, end, miner string) (float64, []models.BlockInfo, []models.BlockInfo, error) {
	//totalBlockNum:=halfBlockNum*t

	from, to, err := calculateBlock(start, end)
	if err != nil {
		blockLog.Errorf("calculateBlock error:%+v", err)
		return 0, nil, nil, err
	}
	//blockLog.Infof("calculateMineBlockPercentage from:%+v to:%+v", from, to)
	var errCalc error
	counter, missed, mined := calculateMineBlockPercentage(from, to, miner, &errCalc)
	//if err != nil {
	//	blockLog.Errorf("calculateMineBlockPercentage error:%+v", err)
	//	return 0, err
	//}
	//var percentage float64
	blockLog.Infof("calculateMineBlockPercentage counter:%+v", counter)
	sort.Slice(missed, func(i, j int) bool {
		return missed[i].Epoch < missed[j].Epoch
	})
	sort.Slice(mined, func(i, j int) bool {
		return mined[i].Epoch < mined[j].Epoch
	})
	if counter.winNum == 0 {
		return 1, missed, mined, errCalc
	} else {
		return counter.mineNum / counter.winNum, missed, mined, errCalc
	}

}

func calculateBlock(start, end string) (abi.ChainEpoch, abi.ChainEpoch, error) {
	tipset, err := client.Client.ChainHead(context.Background())
	if err != nil {
		blockLog.Errorf("calculateBlock get chain head error:%+v", err)
		return 0, 0, err
	}
	t := time.Unix(int64(tipset.MinTimestamp()), 0)
	//h,m,_:=t.Clock()
	//totalNum:=h*120+m*2
	var startTime time.Time
	var endTime time.Time
	startTime, err = time.ParseInLocation("2006-01-02 15:04:05", start, time.Local)
	if err != nil {
		blockLog.Errorf("calculateBlock parse start time error:%+v", err)
		return 0, 0, err
	}
	//startTime = startTime.AddDate(0, 0, 1)

	endTime, err = time.ParseInLocation("2006-01-02 15:04:05", end, time.Local)
	if err != nil {
		blockLog.Errorf("calculateBlock parse end time error:%+v", err)
		return 0, 0, err
	}

	blockLog.Infof("calculateBlock parse start time :%+v end time :%+v", startTime.String(), endTime.String())

	from := tipset.Height() - abi.ChainEpoch(int64(t.Sub(startTime).Minutes())*2)
	to := tipset.Height() - abi.ChainEpoch(int64(t.Sub(endTime).Minutes())*2)
	return from, to, nil
}

func calculateMineBlockPercentage(begin, end abi.ChainEpoch, miner string, errCalc *error) (*mineBlockNum, []models.BlockInfo, []models.BlockInfo) {
	counter := new(mineBlockNum)
	minerAddr, err := address.NewFromString(miner)
	if err != nil {
		blockLog.Errorf("NewFromString err:", err)
		*errCalc = err
		return counter, nil, nil
	}
	missed := make([]models.BlockInfo, 0)
	mined := make([]models.BlockInfo, 0)
	wins := make([]models.BlockInfo, 0)
	//将begin end分组
	for {
		if begin-end <= groupNum {
			wait.Add(1)
			go do(begin, end, counter, minerAddr, &missed, &mined, &wins, errCalc)
			break
		}
		wait.Add(1)
		go do(begin, begin-groupNum, counter, minerAddr, &missed, &mined, &wins, errCalc)
		//是否sleep
		begin -= groupNum
	}
	wait.Wait()
	blockLog.Infof("mined:%+v    missed:%+v   wins:%+v", mined, missed, wins)
	return counter, missed, mined
}

func do(begin, end abi.ChainEpoch, counter *mineBlockNum, miner address.Address, missed, mined, wins *[]models.BlockInfo, errCalc *error) {

	defer wait.Done()
	for i := begin; i > end; i-- {
		start := time.Now()
		tipset, err := client.Client.ChainGetTipSetByHeight(context.Background(), i, types.NewTipSetKey())
		if err != nil {
			blockLog.Errorf("calculateBlock get chain head error:%+v", err)
			*errCalc = fmt.Errorf("calculateBlock get chain head error:%+v", err)
			return
		}
		if tipset.Height() != i {
			continue
		}
		//time
		tStr := time.Unix(int64(tipset.MinTimestamp()), 0).String()
		//flag
		flag := true
		//计算mined
		for _, b := range tipset.Blocks() {
			if b.Miner.String() == miner.String() {
				counter.lock.Lock()
				counter.mineNum++
				counter.lock.Unlock()
				flag = false
				*mined = append(*mined, models.BlockInfo{i, tStr})
			}
		}
		blockLog.Infof("mined:%+v", time.Now().Sub(start))
		end := time.Now()
		//计算winner
		if calculateWiner(i-1, miner, errCalc) {
			counter.lock.Lock()
			counter.winNum++
			counter.lock.Unlock()
			if flag {
				*missed = append(*missed, models.BlockInfo{i, tStr})
			}
			*wins = append(*wins, models.BlockInfo{i, tStr})
		}
		blockLog.Infof("winer:%+v", time.Now().Sub(end))
	}

}

func calculateWiner(h abi.ChainEpoch, miner address.Address, errCalc *error) bool {
	ctx := context.Background()
	round := h + 1
	tp, err := client.Client.ChainGetTipSetByHeight(ctx, h, types.NewTipSetKey())
	if err != nil {
		blockLog.Errorf("ChainGetTipSetByHeight err:%+v", err)
		*errCalc = fmt.Errorf("ChainGetTipSetByHeight err:%+v", err)
		return false
	}

	mbi, err := client.Client.MinerGetBaseInfo(ctx, miner, round, tp.Key())
	if err != nil {
		blockLog.Errorf("MinerGetBaseInfo err:%+v", err)
		*errCalc = fmt.Errorf("MinerGetBaseInfo err:%+v", err)
		return false
	}

	if mbi == nil {

		return false
	}
	if !mbi.EligibleForMining {
		// slashed or just have no power yet
		return false
	}

	beaconPrev := mbi.PrevBeaconEntry
	bvals := mbi.BeaconEntries

	rbase := beaconPrev
	if len(bvals) > 0 {
		rbase = bvals[len(bvals)-1]
	}

	p, err := gen.IsRoundWinner(ctx, tp, round, miner, rbase, mbi, client.SignClient)
	if err != nil {
		blockLog.Errorf("IsRoundWinner err:%+v", err)
		*errCalc = fmt.Errorf("IsRoundWinner err:%+v", err)
		return false
	}

	if p == nil {
		return false
	}
	return true
}
