package stellar

import (
	"time"

	"github.com/anyswap/CrossChain-Router/v3/log"
	"github.com/anyswap/CrossChain-Router/v3/tokens"
	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/txnbuild"
)

// SendTransaction send signed tx
func (b *Bridge) SendTransaction(signedTx interface{}) (txHash string, err error) {
	tx, ok := signedTx.(*txnbuild.Transaction)
	if !ok {
		return "", tokens.ErrWrongRawTx
	}
	var success bool
	var resp hProtocol.Transaction
	for i := 0; i < rpcRetryTimes; i++ {
		for _, r := range b.Remotes {
			resp, err = r.SubmitTransaction(tx)
			if err != nil {
				log.Warn("Try sending transaction failed", "error", err)
				continue
			}
			if !resp.Successful {
				log.Warn("send tx with error result", "result", resp.Successful, "message")
			}
			txHash = resp.Hash
			success = true
		}
		if !success {
			time.Sleep(rpcRetryInterval)
		}
	}
	if success {
		// if !params.IsParallelSwapEnabled() {
		// 	b.SetNonce(tx.Account.String(), uint64(tx.Sequence)+1)
		// }
		return txHash, nil
	}
	return "", err
}
