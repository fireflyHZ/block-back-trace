package reward

import (
	"context"
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	"github.com/fatih/color"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	lotusClient "github.com/filecoin-project/lotus/api/client"
	"github.com/filecoin-project/lotus/blockstore"
	"github.com/filecoin-project/lotus/chain/actors/adt"
	"github.com/filecoin-project/lotus/chain/actors/builtin/miner"
	"github.com/filecoin-project/lotus/chain/gen"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/actors/builtin/reward"
	cbor "github.com/ipfs/go-ipld-cbor"
	logging "github.com/ipfs/go-log/v2"
	"github.com/prometheus/common/log"
	"io"
	"io/ioutil"
	"net/http"
	"profit-allocation/models"
	"time"

	smoothing "github.com/filecoin-project/specs-actors/actors/util/smoothing"
	cbg "github.com/whyrusleeping/cbor-gen"
)

var rewardForLog = logging.Logger("reward-former-log")

type gasAndPenalty struct {
	gas     abi.TokenAmount
	penalty abi.TokenAmount
}

func unmarshalState(r io.Reader) *reward.State {
	rewardActorState := new(reward.State)

	br := cbg.GetPeeker(r)
	scratch := make([]byte, 8)

	maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		rewardForLog.Error("CborReadHeaderBuf err:%+v", err)
	}
	if maj != cbg.MajArray {
		rewardForLog.Debug("maj : %+v", maj)
	}

	if extra != 9 {
		rewardForLog.Debug("extra != 9 extra :%+v", extra)
	}

	// t.CumsumBaseline (big.Int) (struct)

	{

		if err := rewardActorState.CumsumBaseline.UnmarshalCBOR(br); err != nil {
			rewardForLog.Error("CumsumBaseline err : %+v", err)
		}

	}
	// t.CumsumRealized (big.Int) (struct)

	{

		if err := rewardActorState.CumsumRealized.UnmarshalCBOR(br); err != nil {
			rewardForLog.Error("CumsumRealized err : %+v", err)
		}

	}
	// t.EffectiveNetworkTime (abi.ChainEpoch) (int64)
	{
		maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
		var extraI int64
		if err != nil {
			rewardForLog.Error("CborReadHeaderBuf err : %+v", err)
		}
		switch maj {
		case cbg.MajUnsignedInt:
			extraI = int64(extra)
			if extraI < 0 {
				rewardForLog.Debug("int64 positive overflow")
			}
		case cbg.MajNegativeInt:
			extraI = int64(extra)
			if extraI < 0 {
				rewardForLog.Debug("int64 negative oveflow")
			}
			extraI = -1 - extraI
		default:
			rewardForLog.Debug("wrong type for int64 field: %d ", maj)
		}

		rewardActorState.EffectiveNetworkTime = abi.ChainEpoch(extraI)
	}
	// t.EffectiveBaselinePower (big.Int) (struct)

	{

		if err := rewardActorState.EffectiveBaselinePower.UnmarshalCBOR(br); err != nil {
			rewardForLog.Error("unmarshaling t.EffectiveBaselinePower: %+v", err)
		}

	}
	// t.ThisEpochReward (big.Int) (struct)

	{

		if err := rewardActorState.ThisEpochReward.UnmarshalCBOR(br); err != nil {
			rewardForLog.Error("unmarshaling t.ThisEpochReward: %+v", err)
		}

	}
	// t.ThisEpochRewardSmoothed (smoothing.FilterEstimate) (struct)

	{

		b, err := br.ReadByte()
		if err != nil {
			rewardForLog.Error("unmarshaling t.ReadByte: %+v", err)
		}
		if b != cbg.CborNull[0] {
			if err := br.UnreadByte(); err != nil {
				rewardForLog.Error("unmarshaling t.UnreadByte: %+v", err)
			}
			rewardActorState.ThisEpochRewardSmoothed = new(smoothing.FilterEstimate)
			if err := rewardActorState.ThisEpochRewardSmoothed.UnmarshalCBOR(br); err != nil {
				rewardForLog.Error("ThisEpochRewardSmoothed: %+v", err)
			}
		}

	}
	// t.ThisEpochBaselinePower (big.Int) (struct)

	{

		if err := rewardActorState.ThisEpochBaselinePower.UnmarshalCBOR(br); err != nil {
			rewardForLog.Error("ThisEpochBaselinePower: %+v", err)
		}

	}
	// t.Epoch (abi.ChainEpoch) (int64)
	{
		maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
		var extraI int64
		if err != nil {
			rewardForLog.Error("CborReadHeaderBuf: %+v", err)
		}
		switch maj {
		case cbg.MajUnsignedInt:
			extraI = int64(extra)
			if extraI < 0 {
				rewardForLog.Debug("int64 positive overflow")
			}
		case cbg.MajNegativeInt:
			extraI = int64(extra)
			if extraI < 0 {
				rewardForLog.Debug("int64 negative oveflow")
			}
			extraI = -1 - extraI
		default:
			rewardForLog.Debug("wrong type for int64 field: %+v", maj)
		}

		rewardActorState.Epoch = abi.ChainEpoch(extraI)
	}
	// t.TotalMined (big.Int) (struct)

	{

		if err := rewardActorState.TotalMined.UnmarshalCBOR(br); err != nil {
			rewardForLog.Error("unmarshaling t.TotalMined: %+v", err)
		}

	}
	return rewardActorState
}

func inWallets(walletId string) bool {

	for wallet, _ := range models.Wallets {
		if wallet == walletId {
			return true
		}
	}
	return false
}

func inMiners(minerId string) bool {
	for mid, _ := range models.Miners {
		if mid == minerId {
			return true
		}
	}
	return false
}

func TetsGetInfo() {
	requestHeader := http.Header{}
	requestHeader.Add("Content-Type", "application/json")
	LotusHost, err := web.AppConfig.String("lotusHost")
	if err != nil {
		log.Errorf("get lotusHost  err:%+v\n", err)
		return
	}
	nodeApi, closer, err := lotusClient.NewFullNodeRPC(context.Background(), LotusHost, requestHeader)
	if err != nil {
		fmt.Println("NewFullNodeRPC err:", err)
		return
	}
	defer closer()
	minerAddr, _ := address.NewFromString("f0515389")
	data := []byte{}
	for i := 706000; i < 710400; i++ {
		var epoch = abi.ChainEpoch(i)
		//tipset, _ := nodeApi.ChainHead(context.Background())
		//fmt.Printf("444444%+v \n ", time.Unix(int64(tipset.Blocks()[0].Timestamp), 0).Format("2006-01-02 15:04:05"))
		t := types.NewTipSetKey()
		//ver, _ := nodeApi.StateNetworkVersion(context.Background(), tipset.Key())
		//fmt.Printf("version:%+v\n", ver)
		blocks, err := nodeApi.ChainGetTipSetByHeight(context.Background(), epoch, t)
		if err != nil {
			//	rewardForLog.Error("Error get chain head err:%+v",err)
			fmt.Printf("Error get chain head err:%+v\n", err)
			return
		}

		p, _ := nodeApi.StateMinerPower(context.Background(), minerAddr, blocks.Key())
		fmt.Printf("==========%+v\n", p)

		//---------------------
		ctx := context.Background()
		mact, err := nodeApi.StateGetActor(ctx, minerAddr, blocks.Key())
		if err != nil {
			fmt.Println(err)
		}

		tbs := blockstore.NewTieredBstore(blockstore.NewAPIBlockstore(nodeApi), blockstore.NewMemory())
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
		ep := fmt.Sprintf("epoch: %d\n", i)

		mb := fmt.Sprintf("Miner Balance: %s\n", color.YellowString("%s", types.FIL(mact.Balance)))
		pre := fmt.Sprintf("\tPreCommit:   %s\n", types.FIL(lockedFunds.PreCommitDeposits))
		pl := fmt.Sprintf("\tPledge:      %s\n", types.FIL(lockedFunds.InitialPledgeRequirement))
		v := fmt.Sprintf("\tVesting:     %s\n", types.FIL(lockedFunds.VestingFunds))
		a := fmt.Sprintf("\tAvailable:   %s\n", types.FIL(availBalance))
		data = append(data, []byte(ep)...)
		data = append(data, []byte(mb)...)
		data = append(data, []byte(pre)...)
		data = append(data, []byte(pl)...)
		data = append(data, []byte(v)...)
		data = append(data, []byte(a)...)
		data = append(data, []byte("-------------------------\n")...)

	}
	ioutil.WriteFile("vesting", data, 0644)
	//block,err:=nodeApi.ChainHead(context.Background())

	//color.Green("\tAvailable:   %s", types.FIL(availBalance))

	//pr, err := crypto.GenerateKey()
	//if err != nil {
	//	fmt.Printf("err", err)
	//}
	//
	//fmt.Printf("priv:%+v\n", pr)
	//fmt.Printf("priv:%+v\n", len(pr))
	//
	//secCounts, err := nodeApi.StateMinerSectorCount(ctx, minerAddr, types.EmptyTSK)
	//if err != nil {
	//	return
	//}
	//fmt.Printf("sector counts:%+v\n", secCounts)

}

func TetsGetInfo1() {
	ctx := context.Background()
	requestHeader := http.Header{}
	requestHeader.Add("Content-Type", "application/json")
	LotusHost, err := web.AppConfig.String("lotusHost")
	if err != nil {
		log.Errorf("get lotusHost  err:%+v\n", err)
		return
	}
	nodeApi, closer, err := lotusClient.NewFullNodeRPC(context.Background(), LotusHost, requestHeader)
	if err != nil {
		fmt.Println("NewFullNodeRPC err:", err)
		return
	}
	defer closer()
	var h = abi.ChainEpoch(716400)
	tp, err := nodeApi.ChainGetTipSetByHeight(ctx, h, types.NewTipSetKey())
	if err != nil {
		fmt.Println("sdfadf1 err:", err)
	}
	fmt.Println("tp:", tp.Cids())
	ms, err := nodeApi.ChainGetParentMessages(ctx, tp.Cids()[0])
	if err != nil {
		fmt.Println("sdfadf2 err:", err)
		return
	}
	resp, err := nodeApi.ChainGetParentReceipts(ctx, tp.Cids()[0])
	if err != nil {
		fmt.Println("sdfadf3 err:", err)
		return
	}
	for i, m := range ms {
		if m.Message.Method == 0 {
			fmt.Println("msg id:", m.Cid)
			fmt.Println("gas usde:", resp[i].GasUsed)
		}
	}

}

type Ttttime struct {
	Id    int `orm:"pk;auto"`
	Count int
	Time  time.Time
}

func TestTimefind() {
	o := orm.NewOrm()
	for {
		time.Sleep(5 * time.Second)
		tt := make([]models.MinerStatusAndDailyChange, 0)
		n, err := o.Raw("select update_time from fly_miner_status_and_daily_change where miner_id=? and update_time::date=to_date(?,'YYYY-MM-DD')", "f02420", time.Now().Format("2006-01-02")).QueryRows(&tt)

		//		n,err:=o.QueryTable("fly_ttttime").Filter("time",time.Now().Format("2006-01-02")).All(tt)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println("nnnnnn:", n)
		fmt.Printf("asdfsadfsdafsdfasdf:%+v\n", tt[0])
	}
}

func TestMine() {
	ctx := context.Background()
	requestHeader := http.Header{}
	requestHeader.Add("Content-Type", "application/json")
	tokenHeader := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIl19.pL24pbzfXE-ZdEdfYGJabnMORAHvGr7WmEmUnVeiuW4"
	requestHeader.Set("Authorization", tokenHeader)
	LotusHost, err := web.AppConfig.String("lotusHost")
	if err != nil {
		log.Errorf("get lotusHost  err:%+v\n", err)
		return
	}
	nodeApi, closer, err := lotusClient.NewFullNodeRPC(context.Background(), LotusHost, requestHeader)
	if err != nil {
		fmt.Println("NewFullNodeRPC err:", err)
		return
	}
	defer closer()
	minerAddr, err := address.NewFromString("f02420")
	if err != nil {
		fmt.Println("NewFromString err:", err)
		return
	}
	for i := 579560; i < 590800; i++ {
		var h = abi.ChainEpoch(i)
		round := h + 1
		tp, err := nodeApi.ChainGetTipSetByHeight(ctx, h, types.NewTipSetKey())
		if err != nil {
			fmt.Println("ChainGetTipSetByHeight err:", err)
			return
		}

		mbi, err := nodeApi.MinerGetBaseInfo(ctx, minerAddr, round, tp.Key())
		if err != nil {
			fmt.Println("MinerGetBaseInfo err:", err)
			return
		}

		if mbi == nil {

			return
		}
		if !mbi.EligibleForMining {
			// slashed or just have no power yet
			return
		}

		beaconPrev := mbi.PrevBeaconEntry
		bvals := mbi.BeaconEntries

		rbase := beaconPrev
		if len(bvals) > 0 {
			rbase = bvals[len(bvals)-1]
		}

		p, err := gen.IsRoundWinner(ctx, tp, round, minerAddr, rbase, mbi, nodeApi)
		if err != nil {
			fmt.Println("IsRoundWinner err:", err)
			return
		}

		if p != nil {
			fmt.Printf("height:%+v\n", round)
			fmt.Printf("ppp:%+v\n", p)
		}
	}

}
