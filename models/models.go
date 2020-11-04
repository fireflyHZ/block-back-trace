package models

type RewardInfo struct {
	Id      int `orm:"pk;auto"`
	MinerId string
	//	WalletId string
	Epoch      int
	Reward     string
	Gas        string
	Penalty    string
	Value      string
	Power      float64
	Time       string
	UpdateTime int64
}
type MessageRewardInfo struct {
	Id      int `orm:"pk;auto"`
	MinerId string
	//	WalletId string
	Epoch string

	Value      string
	Time       string
	UpdateTime int64
}
type ExpendInfo struct {
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

type WalletHistoryData struct {
	Id       int `orm:"pk;auto"`
	WalletId string
	//NodeId         string
	BalanceFil     string
	BalanceAttofil string
	CreateTime     int64
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

type MinerInfo struct {
	Id           int `orm:"pk;auto"`
	MinerId      string
	QualityPower float64
	Pleage    float64
	CreateTime   int64
	UpdateTime   int64
}

type NetRunDataPro struct {
	Id                 int `orm:"pk;auto"`
	ReceiveBlockHeight int
	TotalShare         int
	AllShare           int
	CreateTime         int64
	UpdateTime         int64
}
type ExpendMessages struct {
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
type RewardMessages struct {
	Id        int `orm:"pk;auto"`
	MessageId string
	MinerId   string
	From      string
	Epoch     string

	Value string
	//RewardValue        string
	Method uint64

	Time       string
	CreateTime uint64
}

//出块记录
type MineBlocks struct {
	Id      int `orm:"pk;auto"`
	MinerId string
	//	WalletId string
	Epoch      int
	Reward     string
	Gas        string
	Penalty    string
	Value      string
	Power      float64
	Time       string
	CreateTime uint64
}

//打包的message记录
type MineMessages struct {
	Id         int `orm:"pk;auto"`
	MinerId    string
	MessageId  string
	Epoch      string
	Gas        string
	Penalty    string
	Time       string
	CreateTime uint64
}

type UserInfo struct {
	Id     int `orm:"pk;auto"`
	UserId int
	Share  float64
	Power  float64
	Reward float64
}

type UserBlockRewardInfo struct {
	Id         int `orm:"pk;auto"`
	UserId     int
	MinerId    string
	Reward     float64
	Power      float64
	Epoch      string
	CreateTime uint64
}

type UserDailyRewardInfo struct {
	Id     int `orm:"pk;auto"`
	UserId int
	//MinerId    string
	Reward     float64
	Power      float64
	Epoch      string
	Time       string
	UpdateTime uint64
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

type OrderBlockRewardInfo struct {
	Id         int `orm:"pk;auto"`
	OrderId    int
	MinerId    string
	Reward     float64
	Power      float64
	Epoch      int
	CreateTime uint64
}

type OrderDailyRewardInfo struct {
	Id      int `orm:"pk;auto"`
	OrderId int
	//MinerId    string
	Reward     float64
	Pleage     float64
	Power      float64
	Epoch      int
	Time       string
	UpdateTime uint64
}

type SettlePlan struct {
	Id        int
	TreatyId  int
	ArticleId int
	Quantity  float64
	AddTime   string
}

type OrderGoods struct {
	Id         int
	ChannelId  int
	ArticleId  int
	OrderId    int
	GoodsId    int
	GoodsNo    string
	GoodsTitle string
	ImgUrl     string
	SpecText   string
	GoodsPrice float64
	RealPrice  float64
	Quantity   int
	Point      int
}

type Orders struct {
	Id               int
	SiteId           int
	OrderNo          string
	TradeNo          string
	UserId           int
	UserName         string
	PaymentId        int
	PaymentFee       float64
	PaymentStatus    int
	PaymentTime      string
	IdcFee           float64
	TreatyId         int
	TreatyStart_date int64
	TreatyEnd_date   int64
	TreatyQuantity   float64
	TreatyStatus     int
	TreatyTime       string
	ExpressId        int
	ExpressNo        string
	ExpressFee       float64
	ExpressStatus    int
	ExpressTime      string
	AcceptName       string
	PostCode         string
	Telphone         string
	Mobile           string
	Email            string
	Area             string
	Address          string
	Message          string
	Remark           string
	IsEsign          int
	IsIdc            int
	IsConfirm        int
	IsInvoice        int
	InvoiceTitle     string
	InvoiceTaxes     float64
	PayableAmount    float64
	RealAmount       float64
	OrderAmount      float64
	Point            int
	Status           int
	AddTime          string
	ConfirmTime      string
	CompleteTime     string
	PaymentVoucher   string
}

type OrderDailyCostInfo struct {
	Id          int `orm:"pk;auto"`
	OrderId     int
	UserId      int
	Expend      float64
	ValueReward float64
	Time        string
}
