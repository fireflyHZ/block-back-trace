package models

import (
	"github.com/filecoin-project/go-state-types/abi"
	"time"
)

type WalletInfoResp struct {
	Code          string
	Msg           string
	WalletBalance string
}

type WalletProfitResp struct {
	Code   string
	Msg    string
	Amount string
}

type RewardRespFormer struct {
	Code   string
	Msg    string
	Reward string
}

type PostResp struct {
	Code string
	Msg  string
}

type OrderInfoRequest struct {
	OrderId int
	UserId  int
	Share   int
	Power   float64
}

type OrderDailyRewardResp struct {
	Code string
	Msg  string
	Data []OrderDailyRewardInfo
}

type UserInfoResp struct {
	Code string
	Msg  string
	Data UserInfo
}

type UserDailyInfoResp struct {
	Code       string
	Msg        string
	TotalCount int
	TotalPage  int
	Data       []UserDailyRewardInfo
}

//--------------------------------
type RewardResp struct {
	Code           string
	Msg            string
	Reward         string
	Pledge         string
	Power          string
	Gas            string
	BlockNumber    int
	SectorsNumber  int64
	WinCount       int64
	TotalPower     string
	TotalAvailable string
	TotalPreCommit string
	TotalPleage    string
	TotalVesting   string
	WindowPostGas  string
	Penalty        string
	Update         time.Time
}

type MessageGasTmp struct {
	Code string
	Msg  string
	Gas  float64
}

//-----------------------------
type GetBlockPercentageResp struct {
	Code            string
	Msg             string
	MinedPercentage string
	Mined           []BlockInfo
	Missed          []BlockInfo
}

type GetMinersLuckResp struct {
	Code       string
	Msg        string
	MinersLuck []MinerLuck
}

type MinerLuck struct {
	Miner       string
	Luck        string
	BlockNumber int
	TotalValue  float64
}

type BlockInfo struct {
	Epoch abi.ChainEpoch
	Time  string
}
