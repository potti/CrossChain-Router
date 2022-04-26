package near

import (
	"fmt"
	"strings"
	"time"

	"github.com/anyswap/CrossChain-Router/v3/common"
	"github.com/anyswap/CrossChain-Router/v3/rpc/client"
)

var (
	rpcTimeout = 60
)

// SetRPCTimeout set rpc timeout
func SetRPCTimeout(timeout int) {
	rpcTimeout = timeout
}

// GetLatestBlock get latest block
func GetLatestBlock(url string) (string, error) {
	request := &client.Request{}
	request.Method = "status"
	request.Params = []string{}
	request.ID = int(time.Now().UnixNano())
	request.Timeout = rpcTimeout
	var result NetworkStatus
	err := client.RPCPostRequest(url, request, &result)
	if err != nil {
		return "0", err
	}
	return result.SyncInfo.LatestBlockHeight, nil
}

func GetBlockByHash(url, hash string) (string, error) {
	request := &client.Request{}
	request.Method = "block"
	request.Params = map[string]string{"block_id": hash}
	request.ID = int(time.Now().UnixNano())
	request.Timeout = rpcTimeout
	var result BlockDetail
	err := client.RPCPostRequest(url, request, &result)
	if err != nil {
		return "0", err
	}
	return result.Header.Height, nil
}

// GetLatestBlockNumber get latest block height
func GetLatestBlockNumber(url string) (height uint64, err error) {
	block, err := GetLatestBlock(url)
	if err != nil {
		return 0, err
	}
	return common.GetUint64FromStr(block)
}

// GetTransactionByHash get tx by hash
func GetTransactionByHash(url, txHash string) (*TransactionResult, error) {
	request := &client.Request{}
	request.Method = "tx"
	request.Params = []string{txHash, "userdemo.testnet"}
	request.ID = int(time.Now().UnixNano())
	request.Timeout = rpcTimeout
	var result TransactionResult
	err := client.RPCPostRequest(url, request, &result)
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(result.Transaction.Hash, txHash) {
		return nil, fmt.Errorf("get tx hash mismatch, have %v want %v", result.Transaction.Hash, txHash)
	}
	return &result, nil
}
