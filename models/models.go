package models

import "time"

type ListenMsgGasNetStatus struct {
	Id                 int `orm:"pk;auto"`
	ReceiveBlockHeight int64
	CreateTime         time.Time
	UpdateTime         time.Time
}

type ExpendMessages struct {
	Id                 int `orm:"pk;auto"`
	MinerId            string
	MessageId          string
	WalletId           string
	To                 string
	Epoch              int64
	Gas                float64
	BaseBurnFee        float64
	OverEstimationBurn float64
	Value              float64
	//RewardValue        string
	Penalty float64
	Method  uint32
	//Time       string
	CreateTime time.Time
}

type ExpendInfo struct {
	Id                 int `orm:"pk;auto"`
	MinerId            string
	WalletId           string
	Epoch              int64
	Gas                float64
	BaseBurnFee        float64
	OverEstimationBurn float64
	Value              float64
	//RewardValue        string
	Penalty float64
	//Time       string
	UpdateTime time.Time
}

//reeward
type ListenRewardNetStatus struct {
	Id                 int `orm:"pk;auto"`
	ReceiveBlockHeight int64
	CreateTime         time.Time
	UpdateTime         time.Time
}

type MinerInfo struct {
	Id           int `orm:"pk;auto"`
	MinerId      string
	Name         string
	QualityPower float64
	Pleage       float64
	Vesting      float64
	CreateTime   time.Time
	UpdateTime   time.Time
	Type         int
	Location     string
}

//miner power status
type MinerStatusAndDailyChange struct {
	Id      int `orm:"pk;auto"`
	MinerId string
	//最新状态
	TotalPower          float64
	TotalAvailable      float64
	TotalPreCommit      float64
	TotalVesting        float64
	TotalPledge         float64
	LiveSectorsNumber   uint64
	ActiveSectorsNumber uint64
	FaultySectorsNumber uint64
	PowerPercentage     float64
	MinedPercentage     float64
	TotalReward         float64
	TotalGas            float64
	TotalBlockNum       int
	TotalWinCounts      int64
	//当天变化
	Epoch      int64
	Reward     float64
	Pledge     float64
	Power      float64
	Vesting    float64
	Gas        float64
	BlockNum   int
	WinCounts  int64
	Time       time.Time `orm:"type(date)"`
	UpdateTime time.Time
}

type PreAndProveMessages struct {
	Id           int `orm:"pk;auto"`
	MessageId    string
	From         string
	To           string
	Epoch        int64
	Method       uint64
	SectorNumber int64
	Status       int
	Params       string
	CreateTime   time.Time
}

//打包的message记录
type MineMessages struct {
	Id         int `orm:"pk;auto"`
	MinerId    string
	MessageId  string
	Epoch      int64
	Gas        float64
	Penalty    float64
	CreateTime time.Time
}

//miner出块权记录表
type MineBlockRight struct {
	Id         int    `orm:"pk;auto"`
	MinerId    string `orm:"index"`
	Epoch      int64  `orm:"index"`
	Missed     bool
	Reward     float64
	WinCount   int64
	Time       time.Time `orm:"type(date)"`
	UpdateTime time.Time
}

//全部矿工出块记录
type AllMinersMined struct {
	Id         int    `orm:"pk;auto"`
	MinerId    string `orm:"index"`
	Epoch      int64  `orm:"index"`
	Reward     float64
	Power      float64 `orm:"index"`
	TotalPower int64
	Time       time.Time
}

type AllMinersPower struct {
	Id         int     `orm:"pk;auto"`
	MinerId    string  `orm:"index"`
	Power      float64 `orm:"index"`
	UpdateTime time.Time
}

type WalletInfo struct {
	Id         int `orm:"pk;auto"`
	WalletId   string
	Balance    float64
	CreateTime time.Time
}

type ReceiveMessages struct {
	Id         int `orm:"pk;auto"`
	MessageId  string
	From       string
	To         string
	Epoch      int64
	Value      float64
	Method     uint32
	CreateTime time.Time
}

type MinerAndWalletRelation struct {
	Id       int `orm:"pk;auto"`
	MinerId  string
	WalletId string
}
