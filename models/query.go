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
	Reward         float64
	Pledge         float64
	Power          float64
	Gas            float64
	BlockNumber    int
	WinCount       int64
	TotalPower     float64
	TotalAvailable float64
	TotalPreCommit float64
	TotalPleage    float64
	TotalVesting   float64
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

type BlockInfo struct {
	Epoch abi.ChainEpoch
	Time  string
}
