package client

import (
	"context"
	"fmt"
	"github.com/filecoin-project/lotus/api"
	lotusClient "github.com/filecoin-project/lotus/api/client"
	logging "github.com/ipfs/go-log/v2"
	"net/http"
	"profit-allocation/models"
)

var Client api.FullNode
var SignClient api.FullNode
var clientLog = logging.Logger("client-log")

func CreateLotusClient() error {
	var err error
	requestHeader := http.Header{}
	requestHeader.Add("Content-Type", "application/json")
	Client, _, err = lotusClient.NewFullNodeRPC(context.Background(), models.LotusHost, requestHeader)
	if err != nil {
		clientLog.Errorf("create lotus client%+v,host:%+v", err, models.LotusHost)
		return err
	}
	return nil
}
func CreateLotusSignClient() error {
	var err error
	requestHeader := http.Header{}
	requestHeader.Add("Content-Type", "application/json")
	tokenHeader := fmt.Sprintf("Bearer %s", models.LotusSignToken)
	requestHeader.Set("Authorization", tokenHeader)
	SignClient, _, err = lotusClient.NewFullNodeRPC(context.Background(), models.LotusHost, requestHeader)
	if err != nil {
		clientLog.Errorf("create lotus client%+v,host:%+v", err, models.LotusHost)
		return err
	}
	return nil
}
