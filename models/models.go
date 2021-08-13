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
	MinerId    string `orm:"index"`
	Epoch      int64  `orm:"index"`
	Reward     float64
	Power      float64 `orm:"index"`
	TotalPower float64
	Time       time.Time
}

type WalletBaseinfo struct {
	Id       int `orm:"pk;auto"`
	WalletId string
	//NodeId         string
	BalanceFil     string
	BalanceAttofil string
	CreateTime     int64
	UpdateTime     int64
	Status         string
}

type WalletProfitInfo struct {
	Id             int `orm:"pk;auto"`
	SettlementType string
	SettlementDate string
	WalletId       string
	//MinerId        string
	StartAmount string
	EndAmount   string
	Amount      string
	CreateTime  int64
	Status      string
}

type NetRunDataPro struct {
	Id                 int `orm:"pk;auto"`
	ReceiveBlockHeight int
	TotalShare         int
	AllShare           int
	CreateTime         int64
	UpdateTime         int64
}

//分配至order和user版本使用 以下
type UserInfo struct {
	Id            int `orm:"pk;auto"`
	UserId        int
	Share         int
	Power         float64 //算力
	Available     float64 //可用余额
	TotalPleage   float64 //总质押
	AdvancePleage float64 //垫付质押
	Vesting       float64 //锁定金额
	Release       float64 //已释放
	Reward        float64 //总奖励
	Fee           float64 //总奖励
	UpdateTime    string
}

type UserDailyRewardInfo struct {
	Id               int `orm:"pk;auto"`
	UserId           int
	Reward           float64
	Power            float64
	Pledge           float64
	Fee              float64
	ImmediateRelease float64
	LinearRelease    float64
	Time             string
}

type OrderInfo struct {
	Id         int `orm:"pk;auto"`
	OrderId    int
	UserId     int
	Share      int
	Power      float64
	Reward     float64
	Epoch      int
	Time       string
	UpdateTime uint64
}

type OrderDailyRewardInfo struct {
	Id      int `orm:"pk;auto"`
	OrderId int
	//MinerId    string
	Reward     float64
	Pleage     float64
	Power      float64
	Fee        float64
	Epoch      int
	Time       string
	UpdateTime uint64
}

type VestingInfo struct {
	Id        int `orm:"pk;auto"`
	UserId    int
	Vesting   float64
	Release   float64
	Times     int32
	StartTime string
}

type MinerAndWalletRelation struct {
	Id       int `orm:"pk;auto"`
	MinerId  string
	WalletId string
}
