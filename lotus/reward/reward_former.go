package reward

import (
	"bytes"
	"context"
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	"github.com/fatih/color"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/crypto"
	"github.com/filecoin-project/lotus/api"
	lotusClient "github.com/filecoin-project/lotus/api/client"
	"github.com/filecoin-project/lotus/api/v0api"
	"github.com/filecoin-project/lotus/blockstore"
	"github.com/filecoin-project/lotus/build"
	"github.com/filecoin-project/lotus/chain/actors/adt"
	"github.com/filecoin-project/lotus/chain/actors/builtin/miner"
	"github.com/filecoin-project/lotus/chain/gen"
	"github.com/filecoin-project/lotus/chain/store"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/actors/builtin/reward"
	"github.com/filecoin-project/specs-actors/v5/actors/builtin"
	miner2 "github.com/filecoin-project/specs-actors/v5/actors/builtin/miner"
	"github.com/ipfs/go-cid"
	cbor "github.com/ipfs/go-ipld-cbor"
	logging "github.com/ipfs/go-log/v2"
	"github.com/prometheus/common/log"
	"io"
	"io/ioutil"
	"math"

	//big2 "math/big"
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
	nodeApi, closer, err := lotusClient.NewFullNodeRPCV0(context.Background(), LotusHost, requestHeader)
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
	nodeApi, closer, err := lotusClient.NewFullNodeRPCV0(context.Background(), LotusHost, requestHeader)
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
	nodeApi, closer, err := lotusClient.NewFullNodeRPCV0(context.Background(), LotusHost, requestHeader)
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

func TestMinerInfo() {
	ctx := context.Background()
	requestHeader := http.Header{}
	requestHeader.Add("Content-Type", "application/json")
	LotusHost, err := web.AppConfig.String("lotusHost")
	if err != nil {
		log.Errorf("get lotusHost  err:%+v\n", err)
		return
	}
	nodeApi, closer, err := lotusClient.NewFullNodeRPCV0(context.Background(), LotusHost, requestHeader)
	if err != nil {
		fmt.Println("NewFullNodeRPC err:", err)
		return
	}
	defer closer()
	minerAddr, err := address.NewFromString("f0144528")
	if err != nil {
		fmt.Println("NewFromString err:", err)
		return
	}
	tip, err := nodeApi.ChainHead(ctx)
	if err != nil {
		fmt.Println("chain head err:", err)
		return
	}
	mi, err := nodeApi.StateMinerInfo(ctx, minerAddr, tip.Key())
	if err != nil {
		fmt.Println("info err:", err)
		return
	}
	fmt.Printf("miner info:%+v\n", mi)
	fmt.Println("protocol", mi.Worker.Protocol())
	addr, err := nodeApi.StateAccountKey(ctx, mi.Worker, types.EmptyTSK)
	if err != nil {
		fmt.Println("state account key err:", err)
		return
	}
	fmt.Println("work ====", addr)
	addr, err = nodeApi.StateAccountKey(ctx, mi.ControlAddresses[0], types.EmptyTSK)
	if err != nil {
		fmt.Println("state account key err:", err)
		return
	}
	fmt.Println("controller ====", addr)
	addr, err = nodeApi.StateAccountKey(ctx, mi.Owner, types.EmptyTSK)
	if err != nil {
		fmt.Println("state account key err:", err)
		return
	}
	fmt.Println("owner ====", addr)

}

func TestProveCommitAggregateParams() {
	ctx := context.Background()
	requestHeader := http.Header{}
	requestHeader.Add("Content-Type", "application/json")
	LotusHost, err := web.AppConfig.String("lotusHost")
	if err != nil {
		log.Errorf("get lotusHost  err:%+v\n", err)
		return
	}
	nodeApi, closer, err := lotusClient.NewFullNodeRPCV0(context.Background(), LotusHost, requestHeader)
	if err != nil {
		fmt.Println("NewFullNodeRPC err:", err)
		return
	}
	defer closer()
	tip, _ := nodeApi.ChainGetTipSetByHeight(ctx, abi.ChainEpoch(905185), types.EmptyTSK)
	msgs, _ := nodeApi.ChainGetParentMessages(ctx, tip.Cids()[0])
	var msg api.Message
	for _, m := range msgs {
		if m.Message.To.String() == "f0144528" && m.Message.Method == builtin.MethodsMiner.ProveCommitAggregate {
			msg = m
			break
		}
	}
	fmt.Println(msg)
	params := new(miner2.ProveCommitAggregateParams)
	b := new(bytes.Buffer)
	_, err = b.Write(msg.Message.Params)
	if err != nil {
		fmt.Printf("record  proveCommit msg:%+v write byte err:%+v", msg.Cid, err)
		return
	}
	err = params.UnmarshalCBOR(b)
	if err != nil {
		fmt.Printf("record  proveCommit msg:%+v unmarshal err:%+v", msg.Cid, err)
		return
	}
	c, _ := params.SectorNumbers.Count()
	fmt.Println(params.SectorNumbers.AllMap(c))

}

func TestStateReplay() {
	ctx := context.Background()
	requestHeader := http.Header{}
	requestHeader.Add("Content-Type", "application/json")
	LotusHost, err := web.AppConfig.String("lotusHost")
	if err != nil {
		log.Errorf("get lotusHost  err:%+v\n", err)
		return
	}
	nodeApi, closer, err := lotusClient.NewFullNodeRPCV0(ctx, LotusHost, requestHeader)
	if err != nil {
		fmt.Println("NewFullNodeRPC err:", err)
		return
	}
	defer closer()
	mcid, err := cid.Decode("bafy2bzacea6za6yl3poxxg7uxq4ciyiqpvkge66ehyfodeegfhabxqx2lmir4")
	if err != nil {
		fmt.Printf("message cid was invalid: %s\n", err)
		return
	}

	res, err := nodeApi.StateReplay(ctx, types.EmptyTSK, mcid)
	if err != nil {
		return
	}

	fmt.Println("Replay receipt:")
	fmt.Printf("Exit code: %d\n", res.MsgRct.ExitCode)
	fmt.Printf("Return: %x\n", res.MsgRct.Return)
	fmt.Printf("Gas Used: %d\n", res.MsgRct.GasUsed)

	fmt.Printf("Base Fee Burn: %d\n", res.GasCost.BaseFeeBurn)
	fmt.Printf("Overestimaton Burn: %d\n", res.GasCost.OverEstimationBurn)
	fmt.Printf("Miner Penalty: %d\n", res.GasCost.MinerPenalty)
	fmt.Printf("Miner Tip: %d\n", res.GasCost.MinerTip)
	fmt.Printf("Refund: %d\n", res.GasCost.Refund)

	fmt.Printf("Total Message Cost: %d\n", res.GasCost.TotalCost)

	if res.MsgRct.ExitCode != 0 {
		fmt.Printf("Error message: %q\n", res.Error)
	}

	fmt.Printf("%s\t%s\t%s\t%d\t%d\t\n", res.Msg.From, res.Msg.To, res.Msg.Value, res.Msg.Method, res.MsgRct.ExitCode)
	burn := big.NewInt(res.GasCost.BaseFeeBurn.Int64())
	printInternalExecutions("\t", res.ExecutionTrace.Subcalls, &burn)
	fmt.Println(burn.Int64())
}
func printInternalExecutions(prefix string, trace []types.ExecutionTrace, burn *big.Int) {
	for _, im := range trace {

		*burn = big.Add(big.NewInt(burn.Int64()), big.NewInt(im.Msg.Value.Int64()))
		fmt.Printf("%s%s\t%s\t%s\t%d\t%d\t\n", prefix, im.Msg.From, im.Msg.To, im.Msg.Value, im.Msg.Method, im.MsgRct.ExitCode)
		printInternalExecutions(prefix+"\t", im.Subcalls, burn)
	}
}

func TestMinerPower() {
	requestHeader := http.Header{}
	requestHeader.Add("Content-Type", "application/json")
	LotusHost, err := web.AppConfig.String("lotusHost")
	if err != nil {
		log.Errorf("get lotusHost  err:%+v\n", err)
		return
	}
	nodeApi, closer, err := lotusClient.NewFullNodeRPCV0(context.Background(), LotusHost, requestHeader)
	if err != nil {
		fmt.Println("NewFullNodeRPC err:", err)
		return
	}
	defer closer()
	minerAddr, _ := address.NewFromString("f0144528")
	var epoch = abi.ChainEpoch(926640)
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

	fmt.Printf("Miner Balance: %s\n", color.YellowString("%s", types.FIL(mact.Balance)))
	fmt.Printf("\tPreCommit:   %s\n", types.FIL(lockedFunds.PreCommitDeposits))
	fmt.Printf("\tPledge:      %s\n", types.FIL(lockedFunds.InitialPledgeRequirement))
	fmt.Printf("\tVesting:     %s\n", types.FIL(lockedFunds.VestingFunds))
	fmt.Printf("\tAvailable:   %s\n", types.FIL(availBalance))

}

func Testmine() {
	requestHeader := http.Header{}
	requestHeader.Add("Content-Type", "application/json")
	tokenHeader := fmt.Sprintf("Bearer %s", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIl19.9cmZqbtPKaJ5q1KFeb67ZTnE-17G61Es6Gckf2eUVXM")
	requestHeader.Set("Authorization", tokenHeader)
	SignClient, _, err := lotusClient.NewFullNodeRPCV0(context.Background(), "http://172.16.10.243:1235/rpc/v0", requestHeader)
	if err != nil {
		fmt.Println(err)
		return
	}

	m, err := address.NewFromString("f0601583")
	if err != nil {
		fmt.Println(err)
	}
	ctx := context.Background()
	round := abi.ChainEpoch(1060440)
	tp, err := SignClient.ChainGetTipSetByHeight(ctx, round, types.NewTipSetKey())
	if err != nil {
		fmt.Println(err)
		return
	}

	mbi, err := SignClient.MinerGetBaseInfo(ctx, m, round, tp.Key())
	if err != nil {
		fmt.Println(err)
		return
	}
	if mbi == nil {
		fmt.Println("mbi == nil")
		return
	}
	fmt.Printf("%+v", mbi.Sectors)
	if !mbi.EligibleForMining {
		// slashed or just have no power yet
		fmt.Println("EligibleForMining")
		return
	}

	beaconPrev := mbi.PrevBeaconEntry
	bvals := mbi.BeaconEntries

	rbase := beaconPrev
	if len(bvals) > 0 {
		rbase = bvals[len(bvals)-1]
	}

	buf := new(bytes.Buffer)
	if err := m.MarshalCBOR(buf); err != nil {
		return
	}

	input, err := store.DrawRandomness(rbase.Data, crypto.DomainSeparationTag_TicketProduction, round-build.TicketRandomnessLookback, buf.Bytes())
	if err != nil {
		return
	}
	fmt.Printf("input %x\n", input)

	vrfOut, err := gen.ComputeVRF(ctx, SignClient.WalletSign, mbi.WorkerKey, input)
	if err != nil {
		return
	}
	fmt.Printf("vrfOut %x \n", vrfOut)

	p, err := gen.IsRoundWinner(ctx, tp, round, m, rbase, mbi, SignClient)
	if err != nil {
		fmt.Println("IsRoundWinner err:%+v", err)
		return
	}

	if p == nil {
		fmt.Println("p==nil")
		return
	}
	fmt.Println(p)
	return
}

func TestPower() {
	ctx := context.Background()

	requestHeader := http.Header{}
	requestHeader.Add("Content-Type", "application/json")
	LotusHost, err := web.AppConfig.String("lotusHost")
	if err != nil {
		log.Errorf("get lotusHost  err:%+v\n", err)
		return
	}
	nodeApi, closer, err := lotusClient.NewFullNodeRPCV0(context.Background(), LotusHost, requestHeader)
	if err != nil {
		fmt.Println("NewFullNodeRPC err:", err)
		return
	}
	defer closer()
	m, err := address.NewFromString("f0393119")
	if err != nil {
		fmt.Println(err)
	}
	round := abi.ChainEpoch(1018800)
	tp, err := nodeApi.ChainGetTipSetByHeight(ctx, round, types.NewTipSetKey())
	if err != nil {
		fmt.Println("1", err)
		return
	}
	power, err := nodeApi.StateMinerPower(ctx, m, tp.Key())
	if err != nil {
		fmt.Println("2", err)
		return
	}
	var f float64 = 1024
	minerPower := float64(power.MinerPower.QualityAdjPower.Int64()) / f / f / f / f

	fmt.Println(math.Pow(1024, 4))
	p := math.Pow(1024, 4)
	ll := big.NewInt(0).Div(power.TotalPower.QualityAdjPower.Int, big.NewInt(int64(p)).Int)
	fmt.Println(minerPower, minerPower/float64(ll.Int64()))
	_, err = nodeApi.StateGetActor(ctx, builtin.RewardActorAddr, tp.Key())
	if err != nil {
		fmt.Println("3", err)
	}
}

func TestAllMiners() {
	ctx := context.Background()

	requestHeader := http.Header{}
	requestHeader.Add("Content-Type", "application/json")
	LotusHost, err := web.AppConfig.String("lotusHost")
	if err != nil {
		log.Errorf("get lotusHost  err:%+v\n", err)
		return
	}
	nodeApi, closer, err := lotusClient.NewFullNodeRPCV0(context.Background(), LotusHost, requestHeader)
	if err != nil {
		fmt.Println("NewFullNodeRPC err:", err)
		return
	}
	defer closer()

	round := abi.ChainEpoch(1032000)
	tp, err := nodeApi.ChainGetTipSetByHeight(ctx, round, types.NewTipSetKey())
	if err != nil {
		fmt.Println("1", err)
		return
	}
	ms, err := nodeApi.StateListMiners(ctx, tp.Key())
	if err != nil {
		fmt.Println("3", err)
		return
	}
	fmt.Println(len(ms))
	//fmt.Println(ms)

}

func TestSector() {
	ctx := context.Background()

	requestHeader := http.Header{}
	requestHeader.Add("Content-Type", "application/json")
	LotusHost, err := web.AppConfig.String("lotusHost")
	if err != nil {
		log.Errorf("get lotusHost  err:%+v\n", err)
		return
	}
	nodeApi, closer, err := lotusClient.NewFullNodeRPCV0(context.Background(), LotusHost, requestHeader)
	if err != nil {
		fmt.Println("NewFullNodeRPC err:", err)
		return
	}
	defer closer()

	round := abi.ChainEpoch(1060440)
	tp, err := nodeApi.ChainGetTipSetByHeight(ctx, round, types.NewTipSetKey())
	if err != nil {
		fmt.Println("1", err)
		return
	}

	minerAddr, _ := address.NewFromString("f0144528")

	secCounts, err := nodeApi.StateMinerSectorCount(ctx, minerAddr, tp.Key())
	if err != nil {
		fmt.Println("2", err)
		return
	}
	fmt.Printf("sector counts:%+v\n", secCounts)
	mbi, err := nodeApi.MinerGetBaseInfo(ctx, minerAddr, round, tp.Key())
	if err != nil {
		fmt.Println(err)
		return
	}
	if mbi == nil {
		fmt.Println("mbi == nil")
		return
	}
	fmt.Printf("%+v", mbi.Sectors)
}

func TestFaultsSectors() {
	ctx := context.Background()
	requestHeader := http.Header{}
	requestHeader.Add("Content-Type", "application/json")
	LotusHost, err := web.AppConfig.String("lotusHost")
	if err != nil {
		log.Errorf("get lotusHost  err:%+v\n", err)
		return
	}
	nodeApi, closer, err := lotusClient.NewFullNodeRPCV0(ctx, LotusHost, requestHeader)
	if err != nil {
		fmt.Println("NewFullNodeRPC err:", err)
		return
	}
	defer closer()

	minerAddr, _ := address.NewFromString("f0419945")

	round := abi.ChainEpoch(1060440)
	tp, err := nodeApi.ChainGetTipSetByHeight(ctx, round, types.NewTipSetKey())
	if err != nil {
		fmt.Println("1", err)
		return
	}

	fail, err := nodeApi.StateMinerFaults(ctx, minerAddr, tp.Key())
	if err != nil {
		fmt.Println("StateMinerFaults err:", err)
		return
	}
	fmt.Printf("fail:%+v\n", fail)
	fail.ForEach(func(num uint64) error {
		_, _ = fmt.Printf("%d\n", num)
		return nil
	})

	//ds,err:=nodeApi.StateMinerDeadlines(ctx,minerAddr,tp.Key())
	//if err != nil {
	//	fmt.Println("StateMinerFaults err:", err)
	//	return
	//}
	di, err := nodeApi.StateMinerProvingDeadline(ctx, minerAddr, tp.Key())
	if err != nil {
		fmt.Println("StateMinerProvingDeadline err:", err)
		return
	}

	fmt.Printf("deadline %+v\n", di)
	mact, err := nodeApi.StateGetActor(ctx, minerAddr, tp.Key())
	if err != nil {
		return
	}

	//tbs := bufbstore.NewTieredBstore(apibstore.NewAPIBlockstore(api), blockstore.NewTemporary())
	tbs := blockstore.NewTieredBstore(blockstore.NewAPIBlockstore(nodeApi), blockstore.NewMemory())
	mas, err := miner.Load(adt.WrapStore(ctx, cbor.NewCborStore(tbs)), mact)
	dl, err := mas.LoadDeadline(di.Index)
	if err != nil {
		fmt.Println("LoadDeadline err:", err)
		return
	}

	par, err := dl.LoadPartition(0)
	if err != nil {
		fmt.Println("LoadPartition err:", err)
		return
	}
	f, err := par.FaultySectors()
	if err != nil {
		fmt.Println("FaultySectors err:", err)
		return
	}
	f.ForEach(func(num uint64) error {
		_, _ = fmt.Printf("%d\n", num)
		return nil
	})
	//net version
	v, err := nodeApi.StateNetworkVersion(ctx, tp.Key())
	if err != nil {
		fmt.Println("StateNetworkVersion err:", err)
		return
	}
	fmt.Println("ver:", v)

	ast, err := nodeApi.StateReadState(ctx, builtin.StorageMarketActorAddr, tp.Key())
	if err != nil {
		fmt.Println("StateReadState err:", err)
		return
	}
	fmt.Printf("ast:%+v\n", ast)

	//msgs,err:=nodeApi.ChainGetParentMessages(ctx,tp.Blocks()[0].Cid())
	//if err != nil {
	//	fmt.Println("ChainGetParentMessages err:", err)
	//	return
	//}
	//round = abi.ChainEpoch(1060449)
	round = abi.ChainEpoch(1071724)
	for {
		fmt.Println(round)
		tp, err := nodeApi.ChainGetTipSetByHeight(ctx, round, types.NewTipSetKey())
		if err != nil {
			fmt.Println("1", err)
			return
		}
		coms, err := nodeApi.StateCompute(ctx, round, nil, tp.Key())
		if err != nil {
			fmt.Println("StateCompute err:", err)
			return
		}
		for _, ins := range coms.Trace {

			printSub(nodeApi, ins.Msg.Cid(), ins.ExecutionTrace.Subcalls)
		}
		round++
	}

}

func printSub(nodeApi v0api.FullNode, msg cid.Cid, subs []types.ExecutionTrace) {
	for _, sub := range subs {
		if sub.Subcalls != nil {
			printSub(nodeApi, msg, sub.Subcalls)
		}
		if sub.Msg.From.String() == "f0419945" && sub.Msg.To.String() == "f099" {
			fmt.Println("msg:", msg)
			fmt.Println(nodeApi.ChainGetMessage(context.Background(), msg))
			fmt.Printf("sub :%+v\n", sub.Msg)
		}
	}
}
