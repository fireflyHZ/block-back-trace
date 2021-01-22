package models

type MinerInfoTmp struct {
	Id           int `orm:"pk;auto"`
	MinerId      string
	QualityPower float64
	Pleage       float64
	CreateTime   int64
	UpdateTime   int64
}

type NetRunDataProTmp struct {
	Id                 int `orm:"pk;auto"`
	ReceiveBlockHeight int
	CreateTime         int64
	UpdateTime         int64
}

type MsgGasNetRunDataProTmp struct {
	Id                 int `orm:"pk;auto"`
	ReceiveBlockHeight int
	CreateTime         int64
	UpdateTime         int64
}

type RewardInfoTmp struct {
	Id      int `orm:"pk;auto"`
	MinerId string
	//	WalletId string
	Epoch      int
	Value      string
	Pledge     float64
	Power      float64
	BlockNum   int
	WinCounts  int64
	Time       string
	UpdateTime int64
}

type ExpendInfoTmp struct {
	Id int `orm:"pk;auto"`
	//MinerId            string
	WalletId           string
	Epoch              string
	Gas                string
	BaseBurnFee        string
	OverEstimationBurn string
	Value              string
	//RewardValue        string
	Penalty    string
	Time       string
	UpdateTime int64
}

type ExpendMessagesTmp struct {
	Id                 int `orm:"pk;auto"`
	MessageId          string
	WalletId           string
	To                 string
	Epoch              string
	Gas                string
	BaseBurnFee        string
	OverEstimationBurn string
	Value              string
	//RewardValue        string
	Penalty    string
	Method     uint64
	Time       string
	CreateTime uint64
}

type MsgGasNetRunDataProTmpTmp struct {
	Id                 int `orm:"pk;auto"`
	ReceiveBlockHeight int
	CreateTime         int64
	UpdateTime         int64
}
