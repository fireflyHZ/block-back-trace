package data

import "profit-allocation/models"

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

type RewardResp struct {
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
	Data []models.OrderDailyRewardInfo
}

type UserInfoResp struct {
	Code string
	Msg  string
	Data models.UserInfo
}

type UserDailyInfoResp struct {
	Code       string
	Msg        string
	TotalCount int
	TotalPage  int
	Data       []models.UserDailyRewardInfo
}

//--------------------------------
type RewardRespTmp struct {
	Code   string
	Msg    string
	Reward float64
	Pledge float64
	Power float64
	Update string
}

type MessageGasTmp struct {
	Code   string
	Msg    string
	Gas    string
}
//-----------------------------
