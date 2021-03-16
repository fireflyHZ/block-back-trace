package block

import (
	"context"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/chain/gen"
	"github.com/filecoin-project/lotus/chain/types"
	logging "github.com/ipfs/go-log/v2"
	"profit-allocation/lotus/reward"
	"sync"
	"time"
)

type mineBlockNum struct {
	mineNum float64
	winNum  float64
	lock    sync.Mutex
}

var wait sync.WaitGroup
var blockLog = logging.Logger("block-log")

func GetMinerMineBlockPercentage(start, end, miner string) (float64, error) {
	//totalBlockNum:=halfBlockNum*t
	from, to, err := calculateBlock(start, end)
	if err != nil {
		blockLog.Errorf("calculateBlock error:%+v", err)
		return 0, err
	}
	blockLog.Infof("calculateMineBlockPercentage from:%+v to:%+v", from, to)

	counter := calculateMineBlockPercentage(from, to, miner)
	//if err != nil {
	//	blockLog.Errorf("calculateMineBlockPercentage error:%+v", err)
	//	return 0, err
	//}
	//var percentage float64
	blockLog.Infof("calculateMineBlockPercentage counter:%+v", counter)
	if counter.winNum == 0 {
		return 0, nil
	} else {
		return counter.mineNum / counter.winNum, nil
	}
}

func calculateBlock(start, end string) (abi.ChainEpoch, abi.ChainEpoch, error) {
	tipset, err := reward.Client.ChainHead(context.Background())
	if err != nil {
		blockLog.Errorf("calculateBlock get chain head error:%+v", err)
		return 0, 0, err
	}
	t := time.Unix(int64(tipset.MinTimestamp()), 0)
	//h,m,_:=t.Clock()
	//totalNum:=h*120+m*2
	var startTime time.Time
	var endTime time.Time
	if t.Format("2006-01-02") == start {
		startTime = t
	} else {
		startTime, err = time.ParseInLocation("2006-01-02", start, time.Local)
		if err != nil {
			blockLog.Errorf("calculateBlock parse start time error:%+v", err)
			return 0, 0, err
		}
		startTime = startTime.AddDate(0, 0, 1)
	}

	endTime, err = time.ParseInLocation("2006-01-02", end, time.Local)
	if err != nil {
		blockLog.Errorf("calculateBlock parse end time error:%+v", err)
		return 0, 0, err
	}

	blockLog.Infof("calculateBlock parse start time :%+v end time :%+v", startTime.String(), endTime.String())

	from := tipset.Height() - abi.ChainEpoch(int64(t.Sub(startTime).Minutes())*2)
	to := tipset.Height() - abi.ChainEpoch(int64(t.Sub(endTime).Minutes())*2)
	return from, to, nil
}

func calculateMineBlockPercentage(begin, end abi.ChainEpoch, miner string) *mineBlockNum {
	counter := new(mineBlockNum)
	minerAddr, err := address.NewFromString(miner)
	if err != nil {
		blockLog.Errorf("NewFromString err:", err)
		return counter
	}
	//将begin end分组
	for {
		if begin-end <= 50 {
			wait.Add(1)
			go do(begin, end, counter, minerAddr)
			break
		}
		wait.Add(1)
		go do(begin, begin-50, counter, minerAddr)
		//是否sleep
		begin -= 50
	}
	wait.Wait()
	return counter
}

func do(begin, end abi.ChainEpoch, counter *mineBlockNum, miner address.Address) {

	defer wait.Done()
	for i := begin; i > end; i-- {
		tipset, err := reward.Client.ChainGetTipSetByHeight(context.Background(), i, types.NewTipSetKey())
		if err != nil {
			blockLog.Errorf("calculateBlock get chain head error:%+v", err)
			return
		}
		for _, b := range tipset.Blocks() {
			if b.Miner.String() == miner.String() {
				counter.lock.Lock()
				counter.mineNum++
				counter.lock.Unlock()
			}
		}
		if calculateWincount(i-1, miner) {
			counter.lock.Lock()
			counter.winNum++
			counter.lock.Unlock()
		}
	}

}

func calculateWincount(h abi.ChainEpoch, miner address.Address) bool {
	ctx := context.Background()
	round := h + 1
	tp, err := reward.Client.ChainGetTipSetByHeight(ctx, h, types.NewTipSetKey())
	if err != nil {
		blockLog.Errorf("ChainGetTipSetByHeight err:%+v", err)
		return false
	}

	mbi, err := reward.Client.MinerGetBaseInfo(ctx, miner, round, tp.Key())
	if err != nil {
		blockLog.Errorf("MinerGetBaseInfo err:%+v", err)
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

	p, err := gen.IsRoundWinner(ctx, tp, round, miner, rbase, mbi, reward.Client)
	if err != nil {
		blockLog.Errorf("IsRoundWinner err:%+v", err)
		return false
	}

	if p == nil {
		return false
	}
	return true
}
