package block

import (
	"bytes"
	"context"
	"github.com/beego/beego/v2/client/orm"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/actors/builtin/reward"
	"github.com/filecoin-project/specs-actors/actors/util/smoothing"
	"github.com/filecoin-project/specs-actors/v2/actors/builtin"
	cbg "github.com/whyrusleeping/cbor-gen"
	"io"
	"math"
	"profit-allocation/lotus/client"
	"profit-allocation/models"
	"profit-allocation/util/bit"
	"strconv"
	"time"
)

//14天
const backtrack = abi.ChainEpoch(2 * 2880)

func RecordAllBlocks() {
Retry:
	ctx := context.Background()
	o := orm.NewOrm()
	allMined := make([]models.AllMinersMined, 0)
	num, err := o.QueryTable("fly_all_miners_mined").OrderBy("-epoch").All(&allMined)
	if err != nil {
		blockLog.Errorf("find fly_all_miners_mined error:%+v", err)
		sleep()
		goto Retry
	}
	begin := abi.ChainEpoch(0)
	if num == 0 {
		head, err := client.Client.ChainHead(ctx)
		if err != nil {
			blockLog.Errorf("get chain head error:%+v", err)
			sleep()
			goto Retry
		}
		begin = head.Height() - backtrack
	} else {
		begin = abi.ChainEpoch(allMined[0].Epoch)
	}

	for {
		head, err := client.Client.ChainHead(ctx)
		if err != nil {
			blockLog.Errorf("get chain head error:%+v", err)
			sleep()
			continue
		}
		//高度小于最新高度--记录
		if begin < head.Height()-10 {
			for begin < head.Height()-10 {
				tipset, err := client.Client.ChainGetTipSetByHeight(ctx, begin+1, types.EmptyTSK)
				if err != nil {
					blockLog.Errorf("get chain tipset by height error:%+v", err)
					sleep()
					goto Retry
				}
				blockLog.Infof("begin record height:%+v", begin+1)
				for _, block := range tipset.Blocks() {
					mineBlock := new(models.AllMinersMined)
					num, err := o.QueryTable("fly_all_miners_mined").Filter("epoch", int64(block.Height)).Filter("miner_id", block.Miner.String()).All(mineBlock)
					if err != nil {
						blockLog.Errorf("find fly_all_miners_mined error:%+v", err)
						sleep()
						goto Retry
					}
					//已有数据continue
					if num != 0 {
						blockLog.Warnf("miner:%+v height:%+v already been recorded", block.Miner, block.Height)
						continue
					}
					//获取奖励和算力
					rewardFloat, power, totalPower, err := calculateReward(ctx, tipset.Key(), block.ElectionProof.WinCount, block.Miner)
					if err != nil {
						blockLog.Errorf("calculate reward error miner:%+v height:%+v err:%+v", block.Miner, block.Height, err)
						sleep()
						goto Retry
					}
					//插入数据
					mineBlock.Epoch = int64(block.Height)
					mineBlock.MinerId = block.Miner.String()
					mineBlock.Reward = rewardFloat
					mineBlock.Power = power
					mineBlock.TotalPower = totalPower
					mineBlock.Time = time.Unix(int64(block.Timestamp), 0)
					_, err = o.Insert(mineBlock)
					if err != nil {
						blockLog.Errorf("insert into table error miner:%+v height:%+v err:%+v", block.Miner, block.Height, err)
						sleep()
						goto Retry
					}
				}
				//更新高度
				begin++
			}
		} else {
			sleep()
		}

	}
}

func sleep() {
	time.Sleep(time.Second * 60)
}

func calculateReward(ctx context.Context, tipsetKey types.TipSetKey, winCount int64, miner address.Address) (float64, float64, int64, error) {
	rewardActor, err := client.Client.StateGetActor(ctx, builtin.RewardActorAddr, tipsetKey)
	if err != nil {
		blockLog.Errorf("StateGetActor err:%+v", err)
		return 0, 0, 0, err
	}

	rewardStateRaw, err := client.Client.ChainReadObj(ctx, rewardActor.Head)
	if err != nil {
		blockLog.Errorf("ChainReadObj err:%+v", err)
		return 0, 0, 0, err
	}

	r := bytes.NewReader(rewardStateRaw)
	rewardActorState := unmarshalState(r)

	mineReward := big.Div(rewardActorState.ThisEpochReward, abi.NewTokenAmount(5))
	mineReward = big.Mul(mineReward, abi.NewTokenAmount(winCount))
	rewardStr := bit.TransFilToFIL(mineReward.String())
	rewardFloat, err := strconv.ParseFloat(rewardStr, 64)
	if err != nil {
		blockLog.Errorf("parse value to float err:%+v", err)
		return 0, 0, 0, err
	}

	power, err := client.Client.StateMinerPower(ctx, miner, tipsetKey)
	if err != nil {
		blockLog.Errorf("StateMinerPower err:%+v", err)
		return 0, 0, 0, err
	}
	var f float64 = 1024
	minerPower := float64(power.MinerPower.QualityAdjPower.Int64()) / f / f / f / f
	p := math.Pow(1024, 4)
	totalPower := power.TotalPower.QualityAdjPower.Int.Div(power.TotalPower.QualityAdjPower.Int, big.NewInt(int64(p)).Int)
	return rewardFloat, minerPower, totalPower.Int64(), nil
}

func unmarshalState(r io.Reader) *reward.State {
	rewardActorState := new(reward.State)

	br := cbg.GetPeeker(r)
	scratch := make([]byte, 8)

	maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		blockLog.Error("CborReadHeaderBuf err:%+v", err)
	}
	if maj != cbg.MajArray {
		blockLog.Debug("maj : %+v", maj)
	}

	if extra != 9 {
		blockLog.Debug("extra != 9 extra :%+v", extra)
	}

	// t.CumsumBaseline (big.Int) (struct)

	{

		if err := rewardActorState.CumsumBaseline.UnmarshalCBOR(br); err != nil {
			blockLog.Error("CumsumBaseline err : %+v", err)
		}

	}
	// t.CumsumRealized (big.Int) (struct)

	{

		if err := rewardActorState.CumsumRealized.UnmarshalCBOR(br); err != nil {
			blockLog.Error("CumsumRealized err : %+v", err)
		}

	}
	// t.EffectiveNetworkTime (abi.ChainEpoch) (int64)
	{
		maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
		var extraI int64
		if err != nil {
			blockLog.Error("CborReadHeaderBuf err : %+v", err)
		}
		switch maj {
		case cbg.MajUnsignedInt:
			extraI = int64(extra)
			if extraI < 0 {
				blockLog.Debug("int64 positive overflow")
			}
		case cbg.MajNegativeInt:
			extraI = int64(extra)
			if extraI < 0 {
				blockLog.Debug("int64 negative oveflow")
			}
			extraI = -1 - extraI
		default:
			blockLog.Debug("wrong type for int64 field: %d ", maj)
		}

		rewardActorState.EffectiveNetworkTime = abi.ChainEpoch(extraI)
	}
	// t.EffectiveBaselinePower (big.Int) (struct)

	{

		if err := rewardActorState.EffectiveBaselinePower.UnmarshalCBOR(br); err != nil {
			blockLog.Error("unmarshaling t.EffectiveBaselinePower: %+v", err)
		}

	}
	// t.ThisEpochReward (big.Int) (struct)

	{

		if err := rewardActorState.ThisEpochReward.UnmarshalCBOR(br); err != nil {
			blockLog.Error("unmarshaling t.ThisEpochReward: %+v", err)
		}

	}
	// t.ThisEpochRewardSmoothed (smoothing.FilterEstimate) (struct)

	{

		b, err := br.ReadByte()
		if err != nil {
			blockLog.Error("unmarshaling t.ReadByte: %+v", err)
		}
		if b != cbg.CborNull[0] {
			if err := br.UnreadByte(); err != nil {
				blockLog.Error("unmarshaling t.UnreadByte: %+v", err)
			}
			rewardActorState.ThisEpochRewardSmoothed = new(smoothing.FilterEstimate)
			if err := rewardActorState.ThisEpochRewardSmoothed.UnmarshalCBOR(br); err != nil {
				blockLog.Error("ThisEpochRewardSmoothed: %+v", err)
			}
		}

	}
	// t.ThisEpochBaselinePower (big.Int) (struct)

	{

		if err := rewardActorState.ThisEpochBaselinePower.UnmarshalCBOR(br); err != nil {
			blockLog.Error("ThisEpochBaselinePower: %+v", err)
		}

	}
	// t.Epoch (abi.ChainEpoch) (int64)
	{
		maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
		var extraI int64
		if err != nil {
			blockLog.Error("CborReadHeaderBuf: %+v", err)
		}
		switch maj {
		case cbg.MajUnsignedInt:
			extraI = int64(extra)
			if extraI < 0 {
				blockLog.Debug("int64 positive overflow")
			}
		case cbg.MajNegativeInt:
			extraI = int64(extra)
			if extraI < 0 {
				blockLog.Debug("int64 negative oveflow")
			}
			extraI = -1 - extraI
		default:
			blockLog.Debug("wrong type for int64 field: %+v", maj)
		}

		rewardActorState.Epoch = abi.ChainEpoch(extraI)
	}
	// t.TotalMined (big.Int) (struct)

	{

		if err := rewardActorState.TotalMined.UnmarshalCBOR(br); err != nil {
			blockLog.Error("unmarshaling t.TotalMined: %+v", err)
		}

	}
	return rewardActorState
}
