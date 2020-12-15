package reward

import (
	"context"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api/apibstore"
	lotusClient "github.com/filecoin-project/lotus/api/client"
	"github.com/filecoin-project/lotus/chain/actors/adt"
	"github.com/filecoin-project/lotus/chain/actors/builtin/miner"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/lib/blockstore"
	"github.com/filecoin-project/lotus/lib/bufbstore"
	cbor "github.com/ipfs/go-ipld-cbor"
	"net/http"
	"profit-allocation/tool/bit"
	"strconv"
)

func GetMienrPleage(minerAddr string, epoch abi.ChainEpoch) (float64,float64,float64,float64, error) {
	maddr, err := address.NewFromString(minerAddr)
	if err != nil {
		return 0,0,0,0, err
	}
	//totalGas := abi.NewTokenAmount(0)
	//mineReward := abi.NewTokenAmount(0)
	//totalPenalty := abi.NewTokenAmount(0)
	lotusHost := beego.AppConfig.String("lotusHost")
	requestHeader := http.Header{}
	ctx := context.Background()

	api, closer, err := lotusClient.NewFullNodeRPC(context.Background(), lotusHost, requestHeader)
	if err != nil {
		fmt.Println(err)
		return 0,0,0,0, err
	}
	defer closer()
	tipset := types.NewTipSetKey()
	//epoch:=abi.ChainEpoch(148888)
	t, _ := api.ChainGetTipSetByHeight(ctx, epoch, tipset)
	tipsetKey := t.Key()

	mact, err := api.StateGetActor(ctx, maddr, tipsetKey)
	if err != nil {
		return 0,0,0,0, err
	}

	tbs := bufbstore.NewTieredBstore(apibstore.NewAPIBlockstore(api), blockstore.NewTemporary())
	mas, err := miner.Load(adt.WrapStore(ctx, cbor.NewCborStore(tbs)), mact)
	if err != nil {
		return 0,0,0,0, err
	}
	// NOTE: there's no need to unlock anything here. Funds only
	// vest on deadline boundaries, and they're unlocked by cron.
	lockedFunds, err := mas.LockedFunds()
	if err != nil {
		return 0,0,0,0, err
	}
	availBalance, err := mas.AvailableBalance(mact.Balance)
	if err != nil {
		return 0,0,0,0, err
	}
	pleageStr := bit.TransFilToFIL(lockedFunds.InitialPledgeRequirement.String())
	pleage, err := strconv.ParseFloat(pleageStr, 64)

	preCommitStr := bit.TransFilToFIL(lockedFunds.PreCommitDeposits.String())
	preCommit, err := strconv.ParseFloat(preCommitStr, 64)

	vestingStr := bit.TransFilToFIL(lockedFunds.VestingFunds.String())
	vesting, err := strconv.ParseFloat(vestingStr, 64)

	availStr := bit.TransFilToFIL(availBalance.String())
	available, err := strconv.ParseFloat(availStr, 64)
	if err != nil {
		return 0,0,0,0, err
	}
	return available,preCommit,vesting,pleage, nil
}
