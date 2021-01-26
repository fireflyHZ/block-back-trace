package models

type MinerInfoTmp struct {
	Id           int `orm:"pk;auto"`
	MinerId      string
	QualityPower float64
	Pleage       float64
	CreateTime   int64
	UpdateTime   int64
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
