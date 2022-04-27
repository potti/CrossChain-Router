package near

import (
	"crypto/ecdsa"

	"github.com/anyswap/CrossChain-Router/v3/tokens"
)

// MPCSignTransaction mpc sign raw tx
func (b *Bridge) MPCSignTransaction(rawTx interface{}, args *tokens.BuildTxArgs) (signedTx interface{}, txHash string, err error) {
	rawTxStruct := rawTx.(*RawTransaction)
	signedTx, errf := rawTxStruct.Serialize()
	if errf != nil {
		err = errf
		return
	}
	return
}

// SignTransactionWithPrivateKey sign tx with ECDSA private key
func (b *Bridge) SignTransactionWithPrivateKey(rawTx interface{}, privKey *ecdsa.PrivateKey) (signTx interface{}, txHash string, err error) {
	return
}
