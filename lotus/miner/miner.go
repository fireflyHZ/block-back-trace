package miner

import (
	logging "github.com/ipfs/go-log/v2"
	"profit-allocation/models"

	"time"
)

var minerLog = logging.Logger("miner-log")

func collectMinerData() {

	//查询数据
	o, err := models.O.Begin()
	if err != nil {
		minerLog.Debug("DEBUG: collectMinerData orm transation begin error: %+v", err)
		return
	}

	for _, minerId := range models.Miners {
		qualityPower, _, _ := stateMinerPowerInfo(minerId)
		minerInfo := new(models.MinerInfo)
		n, err := o.QueryTable("fly_reward_info").Filter("miner_id", minerId).All(minerInfo)
		if err != nil {
			minerLog.Error("Error  QueryTable minerInfo:%+v err:%+v num:%+v ", minerId, err, n)
			err := o.Rollback()
			if err != nil {
				minerLog.Debug("DEBUG: collectMinerData orm transation rollback error: %+v", err)
			}
			return
		}
		if n == 0 {
			minerInfo.QualityPower = qualityPower
			//minerInfo.RawPower = rawPower
			//minerInfo.Available = available
			minerInfo.CreateTime = time.Now().Unix()
			minerInfo.UpdateTime = time.Now().Unix()

			_, err := o.Insert(minerInfo)
			if err != nil {
				minerLog.Error("Error  Insert minerInfo miner:%+v  err:%+v ", minerId, err)
				err := o.Rollback()
				if err != nil {
					minerLog.Debug("DEBUG: collectMinerData orm transation rollback error: %+v", err)
				}
				return
			}
		} else {
			//更新miner info
			minerInfo.QualityPower = qualityPower
			//minerInfo.RawPower = rawPower
			//minerInfo.Available = available
			minerInfo.UpdateTime = time.Now().Unix()

			_, err := o.Update(minerInfo)
			if err != nil {
				minerLog.Error("Error  Update minerInfo miner:%+v  err:%+v ", minerId, err)
				err := o.Rollback()
				if err != nil {
					minerLog.Debug("DEBUG: collectMinerData orm transation rollback error: %+v", err)
				}
				return
			}
		}

	}

}

func stateMinerPowerInfo(minerId string) (qualityPower float64, rawPower, available string) {
	/*lotusHost := beego.AppConfig.String("lotusHost")
	requestHeader := http.Header{}
	nodeApi, closer, err := lotusClient.NewFullNodeRPC(context.Background(), lotusHost, requestHeader)
	if err != nil {
		return
	}
	defer closer()

	addr, err := address.NewFromString(minerId)
	if err != nil {
		return
	}

	ctx := context.Background()
	ts, err := cli.LoadTipSet(ctx, cctx, api)
	if err != nil {
		return err
	}
	tipset := types.NewTipSetKey()

	power, err := nodeApi.StateMinerPower(ctx, addr, tipset)
	if err != nil {
		return
	}
	availableInt, err := nodeApi.StateMinerAvailableBalance(ctx, addr, types.EmptyTSK)

	available = bit.TransFilToFIL(availableInt.String())


	qualityPowerInt := power.MinerPower.QualityAdjPower.Int64()
	rawPowerInt := power.MinerPower.RawBytePower.Int64()
	count := 0
	for {
		if qualityPowerInt > 1024 {
			qualityPowerInt /= 1024
			rawPowerInt /= 1024
			count++
		} else {
			break
		}
	}
	//qualityPower = strconv.Itoa(int(qualityPowerInt))
	rawPower = strconv.Itoa(int(rawPowerInt))
	switch count {
	case 0:
		qualityPower += "byte"
		rawPower += "byte"
	case 1:
		qualityPower += "KiB"
		rawPower += "KiB"
	case 2:
		qualityPower += "MiB"
		rawPower += "MiB"
	case 3:
		qualityPower += "TiB"
		rawPower += "TiB"
	case 4:
		qualityPower += "PiB"
		rawPower += "PiB"
	case 5:
		qualityPower += "EiB"
		rawPower += "EiB"
	}*/
	return
}
