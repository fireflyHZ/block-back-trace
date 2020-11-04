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
	Msg string
}

type OrderInfoRequest struct {
	OrderId    int
	UserId     int
	Share      int
	Power      float64
}

type OrderDailyRewardResp struct {
	Code string
	Msg string
	Data []models.OrderDailyRewardInfo
}