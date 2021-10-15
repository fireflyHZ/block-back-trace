package models

type CheckPartitionsResp struct {
	Result []MinerPartitionInfo `json:"result"`
}

type MinerPartitionInfo struct {
	Maddr               string `json:"Maddr"`
	SystemTime          string `json:"SystemTime"`
	TipsetTime          string `json:"TipsetTime"`
	DeadlineElapsedTime string `json:"DeadlineElapsedTime"`
	DeadlineOpenTime    string `json:"DeadlineOpenTime"`
	DeadlineCloseTime   string `json:"DeadlineCloseTime"`
	DeadlineIndex       int    `json:"DeadlineIndex"`
	AllSectors          int    `json:"AllSectors"`
	FaultSectors        int    `json:"FaultSectors"`
	AllPartitions       int    `json:"AllPartitions"`
	ProvenPartitions    int    `json:"ProvenPartitions"`
}
