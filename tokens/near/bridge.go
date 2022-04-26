package near

import (
	"math/big"

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

//GetLatestBlockNumberOf gets latest block number from single api
// For ripple, GetLatestBlockNumberOf returns current ledger version
func (b *Bridge) GetLatestBlockNumberOf(apiAddress string) (uint64, error) {
	return GetLatestBlockNumber(apiAddress)
}

func (b *Bridge) GetLatestBlockNumberByHash(txhash string) (uint64, error) {
	urls := b.GatewayConfig.APIAddress
	for _, url := range urls {
		result, err := GetLatestBlockNumberByHash(url, txhash)
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
	for _, url := range urls {
		result, err = GetTransactionByHash(url, txHash)
		if err == nil {
			return result, nil
		}
	}
	return nil, tokens.ErrTxNotFound
}

// GetTransactionStatus impl
func (b *Bridge) GetTransactionStatus(txHash string) (status *tokens.TxStatus, err error) {
	return
}

// GetBalance gets balance
func (b *Bridge) GetBalance(accountAddress string) (*big.Int, error) {
	return big.NewInt(0), nil
}

// GetTokenBalance not supported
func (b *Bridge) GetTokenBalance(tokenType, tokenAddress, accountAddress string) (*big.Int, error) {
	return nil, nil
}

// GetTokenSupply not supported
func (b *Bridge) GetTokenSupply(tokenType, tokenAddress string) (*big.Int, error) {
	return nil, nil
}

// GetAccount returns account
func (b *Bridge) GetAccount(address string) (acct string, err error) {
	return
}
