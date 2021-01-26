package models

import "time"

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
	Gas            string
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
	Gas  string
}

//-----------------------------
