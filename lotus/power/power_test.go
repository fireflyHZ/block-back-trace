package power

import (
	"profit-allocation/models"
	"testing"
)

func TestPartitionCheck(t *testing.T) {
	models.Miners = map[string]int{"f0144528": 1, "f02420": 2}
	models.LotusHost = "http://172.16.10.245:1234/rpc/v0"
	PartitionCheck()
}
