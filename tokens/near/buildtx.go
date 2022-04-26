package near

import (
	"github.com/anyswap/CrossChain-Router/v3/tokens"
)

var ()

// BuildRawTransaction build raw tx
func (b *Bridge) BuildRawTransaction(args *tokens.BuildTxArgs) (rawTx interface{}, err error) {
	return
}

// GetTxBlockInfo impl NonceSetter interface
func (b *Bridge) GetTxBlockInfo(txHash string) (blockHeight, blockTime uint64) {
	txStatus, err := b.GetTransactionStatus(txHash)
	if err != nil {
		return 0, 0
	}
	return txStatus.BlockHeight, txStatus.BlockTime
}

// GetPoolNonce impl NonceSetter interface
func (b *Bridge) GetPoolNonce(address, _height string) (uint64, error) {
	return uint64(0), nil
}

// GetSeq returns account tx sequence
func (b *Bridge) GetSeq(args *tokens.BuildTxArgs) (nonceptr *uint32, err error) {
	nonceVal, err := b.GetPoolNonce(args.From, "")
	if err != nil {
		return nil, err
	}
	if args == nil {
		nonce := uint32(nonceVal)
		return &nonce, nil
	}
	nonceVal = b.AdjustNonce(args.From, nonceVal)
	nonce := uint32(nonceVal)
	return &nonce, nil
}
