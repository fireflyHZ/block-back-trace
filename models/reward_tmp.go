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

type RewardInfoTmp struct {
	Id      int `orm:"pk;auto"`
	MinerId string
	//	WalletId string
	Epoch      int
	Value      string
	Pledge     float64
	Time       string
	UpdateTime int64
}
