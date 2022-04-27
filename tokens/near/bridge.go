package near

import (
	"github.com/anyswap/CrossChain-Router/v3/log"
	"github.com/anyswap/CrossChain-Router/v3/tokens"
	"github.com/anyswap/CrossChain-Router/v3/tokens/base"
)

var (
	// ensure Bridge impl tokens.CrossChainBridge
	_ tokens.IBridge = &Bridge{}
	// ensure Bridge impl tokens.NonceSetter
	_ tokens.NonceSetter = &Bridge{}
)

// Bridge block bridge inherit from btc bridge
type Bridge struct {
	*base.NonceSetterBase
}

// NewCrossChainBridge new bridge
func NewCrossChainBridge() *Bridge {
	return &Bridge{
		NonceSetterBase: base.NewNonceSetterBase(),
	}
}

// VerifyTokenConfig verify token config
func (b *Bridge) VerifyTokenConfig(tokenCfg *tokens.TokenConfig) error {
	return nil
}

// GetLatestBlockNumber gets latest block number
// For ripple, GetLatestBlockNumber returns current ledger version
func (b *Bridge) GetLatestBlockNumber() (uint64, error) {
	urls := b.GatewayConfig.APIAddress
	for _, url := range urls {
		result, err := GetLatestBlockNumber(url)
		if err == nil {
			return result, nil
		}
	}
	return 0, tokens.ErrRPCQueryError
}

func (b *Bridge) GetLatestBlockHash() (string, error) {
	urls := b.GatewayConfig.APIAddress
	for _, url := range urls {
		result, err := GetLatestBlockHash(url)
		if err == nil {
			return result, nil
		}
	}
	return "", tokens.ErrRPCQueryError
}

//GetLatestBlockNumberOf gets latest block number from single api
// For ripple, GetLatestBlockNumberOf returns current ledger version
func (b *Bridge) GetLatestBlockNumberOf(apiAddress string) (uint64, error) {
	return GetLatestBlockNumber(apiAddress)
}

func (b *Bridge) GetBlockNumberByHash(txhash string) (uint64, error) {
	urls := b.GatewayConfig.APIAddress
	for _, url := range urls {
		result, err := GetBlockNumberByHash(url, txhash)
		if err == nil {
			return result, nil
		}
	}
	return 0, tokens.ErrRPCQueryError
}

// GetTransaction impl
func (b *Bridge) GetTransaction(txHash string) (tx interface{}, err error) {
	return b.GetTransactionByHash(txHash)
}

// GetTransactionByHash get tx response by hash
func (b *Bridge) GetTransactionByHash(txHash string) (result *TransactionResult, err error) {
	urls := b.GatewayConfig.APIAddress
	router := b.ChainConfig.RouterContract
	for _, url := range urls {
		result, err = GetTransactionByHash(url, txHash, router)
		if err == nil {
			return result, nil
		}
	}
	return nil, tokens.ErrTxNotFound
}

// GetTransactionStatus impl
func (b *Bridge) GetTransactionStatus(txHash string) (status *tokens.TxStatus, err error) {
	status = new(tokens.TxStatus)
	tx, err := b.GetTransaction(txHash)
	if err != nil {
		return nil, err
	}

	txres, ok := tx.(*TransactionResult)
	if !ok {
		// unexpected
		log.Warn("GetTransactionStatus", "error", errTxResultType)
		return nil, errTxResultType
	}

	// Check tx status
	if txres.Status.Failure != "" {
		log.Warn("Near tx status is not success", "result", txres.Status.Failure)
		return nil, tokens.ErrTxWithWrongStatus
	}

	status.Receipt = nil
	blockHeight, blockErr := b.GetBlockNumberByHash(txHash)
	if blockErr != nil {
		log.Warn("GetBlockNumberByHash", "error", blockErr)
		return nil, errTxResultType
	}
	status.BlockHeight = blockHeight

	if latest, err := b.GetLatestBlockNumber(); err == nil && latest > blockHeight {
		status.Confirmations = latest - blockHeight
	}
	return
}
