package near

import (
	"github.com/anyswap/CrossChain-Router/v3/tokens"
	"github.com/near/borsh-go"
)

// SendTransaction send signed tx
func (b *Bridge) SendTransaction(signedTx interface{}) (txHash string, err error) {
	signTx := signedTx.(*SignedTransaction)
	buf, err := borsh.Serialize(*signTx)
	if err != nil {
		return "", err
	}
	return b.BroadcastTxCommit(buf)
}

func (b *Bridge) BroadcastTxCommit(signedTx []byte) (string, error) {
	urls := b.GatewayConfig.APIAddress
	for _, url := range urls {
		result, err := BroadcastTxCommit(url, signedTx)
		if err == nil {
			return result, nil
		}
	}
	return "", tokens.ErrRPCQueryError
}
