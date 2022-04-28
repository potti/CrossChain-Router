package near

import (
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"

	"github.com/anyswap/CrossChain-Router/v3/log"
	"github.com/anyswap/CrossChain-Router/v3/params"
	"github.com/anyswap/CrossChain-Router/v3/router"
	"github.com/anyswap/CrossChain-Router/v3/tokens"
	"github.com/near/borsh-go"
)

// MPCSignTransaction mpc sign raw tx
func (b *Bridge) MPCSignTransaction(rawTx interface{}, args *tokens.BuildTxArgs) (signedTx interface{}, txHash string, err error) {
	log.Debug("Near MPCSignTransaction")
	tx, err := b.verifyTransactionReceiver(rawTx, args.GetTokenID())
	if err != nil {
		return nil, "", err
	}
	fmt.Println("tx========", tx)
	if params.SignWithPrivateKey() {
		privKey := params.GetSignerPrivateKey(b.ChainConfig.ChainID)
		signedTx, txHash, err = b.SignTransactionWithPrivateKey(rawTx, ed25519.PrivateKey(privKey))
		return
	}
	return
}

// SignTransactionWithPrivateKey sign tx with ECDSA private key
func (b *Bridge) SignTransactionWithPrivateKey(rawTx interface{}, privKey ed25519.PrivateKey) (signedTx interface{}, txHash string, err error) {
	tx := rawTx.(*RawTransaction)
	signedTx, txHash, err = signTransaction(tx, ed25519.PrivateKey(privKey), b.ChainConfig.RouterContract)
	return
}

func (b *Bridge) verifyTransactionReceiver(rawTx interface{}, tokenID string) (*RawTransaction, error) {
	tx, ok := rawTx.(*RawTransaction)
	if !ok {
		return nil, errors.New("[sign] wrong raw tx param")
	}
	checkReceiver, err := router.GetTokenRouterContract(tokenID, b.ChainConfig.ChainID)
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(tx.ReceiverID, checkReceiver) {
		return nil, fmt.Errorf("[sign] tx receiver mismatch. have %v want %v", tx.ReceiverID, checkReceiver)
	}
	return tx, nil
}

func signTransaction(tx *RawTransaction, privKey ed25519.PrivateKey, accountId string) (signedTx *SignedTransaction, txHash string, err error) {
	buf, err := borsh.Serialize(*tx)
	if err != nil {
		return nil, "", err
	}

	hash := sha256.Sum256(buf)

	sig, err := privKey.Sign(rand.Reader, hash[:], crypto.Hash(0))
	if err != nil {
		return nil, "", err
	}

	var signature Signature
	signature.KeyType = ED25519
	copy(signature.Data[:], sig)

	var stx SignedTransaction
	stx.Transaction = *tx
	stx.Signature = signature

	return &stx, string(hash[:]), nil
}
