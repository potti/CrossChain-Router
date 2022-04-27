package near

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/anyswap/CrossChain-Router/v3/common"
	"github.com/anyswap/CrossChain-Router/v3/log"
	"github.com/anyswap/CrossChain-Router/v3/params"
	"github.com/anyswap/CrossChain-Router/v3/router"
	"github.com/anyswap/CrossChain-Router/v3/tokens"
	"github.com/anyswap/CrossChain-Router/v3/tokens/near/serialize"
	"github.com/btcsuite/btcutil/base58"
)

// BuildRawTransaction build raw tx
func (b *Bridge) BuildRawTransaction(args *tokens.BuildTxArgs) (rawTx interface{}, err error) {
	if !params.IsTestMode && args.ToChainID.String() != b.ChainConfig.ChainID {
		return nil, tokens.ErrToChainIDMismatch
	}
	if args.Input != nil {
		return nil, fmt.Errorf("forbid build raw swap tx with input data")
	}
	if args.From == "" {
		return nil, fmt.Errorf("forbid empty sender")
	}
	routerMPC, getMpcErr := router.GetRouterMPC(args.GetTokenID(), b.ChainConfig.ChainID)
	if getMpcErr != nil {
		return nil, getMpcErr
	}
	if !common.IsEqualIgnoreCase(args.From, routerMPC) {
		log.Error("build tx mpc mismatch", "have", args.From, "want", routerMPC)
		return nil, tokens.ErrSenderMismatch
	}

	switch args.SwapType {
	case tokens.ERC20SwapType:
	default:
		return nil, tokens.ErrSwapTypeNotSupported
	}

	mpcPubkey := router.GetMPCPublicKey(args.From)
	if mpcPubkey == "" {
		return nil, tokens.ErrMissMPCPublicKey
	}

	erc20SwapInfo := args.ERC20SwapInfo
	multichainToken := router.GetCachedMultichainToken(erc20SwapInfo.TokenID, args.ToChainID.String())
	if multichainToken == "" {
		log.Warn("get multichain token failed", "tokenID", erc20SwapInfo.TokenID, "chainID", args.ToChainID)
		return nil, tokens.ErrMissTokenConfig
	}

	token := b.GetTokenConfig(multichainToken)
	if token == nil {
		return nil, tokens.ErrMissTokenConfig
	}

	nonce, getNonceErr := b.GetAccountNonce(args.From, mpcPubkey)
	if getNonceErr != nil {
		return nil, getNonceErr
	}

	blockHash, getBlockHashErr := b.GetLatestBlockHash()
	if getBlockHashErr != nil {
		return nil, tokens.ErrRPCQueryError
	}
	rawTx, err = CreateTransaction(args.From, args.Bind, mpcPubkey, blockHash, nonce, args.Value)
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

func CreateTransaction(from, to, publicKey, blockHash string, nonce uint64, value *big.Int) (*RawTransaction, error) {
	var err error
	tx := new(RawTransaction)
	tx.SignerId = serialize.String{Value: from}
	tx.Nonce = serialize.U64{Value: uint64(nonce)}
	tx.ReceiverId = serialize.String{Value: to}
	bh := base58.Decode(blockHash)
	if len(bh) == 0 {
		return nil, fmt.Errorf("base58  decode blockhash error ,BlockHash=%s", blockHash)
	}
	tx.BlockHash = serialize.BlockHash{
		Value: bh,
	}
	publicKey = strings.TrimPrefix(publicKey, "ed25519:")
	var pk []byte
	if len(publicKey) == 64 { //is hex
		pk, err = hex.DecodeString(publicKey)
		if err != nil {
			return nil, fmt.Errorf("decode public key error,Err=%v", err)
		}
	} else {
		pk = base58.Decode(publicKey)
		if len(pk) == 0 {
			return nil, fmt.Errorf("base58 decode public key error,Public Key=%s", publicKey)
		}
	}
	tx.PublicKey = serialize.PublicKey{
		KeyType: 0,
		Value:   pk,
	}
	action, createErr := serialize.CreateFuncCall("any_swap_in", []byte{}, 0, "")
	if createErr != nil {
		return nil, fmt.Errorf("CreateFuncCall error=%s", createErr)
	}
	tx.SetAction(action)
	return tx, nil
}

func (tx *RawTransaction) SetAction(action ...serialize.IAction) {
	tx.Actions = append(tx.Actions, action...)
}

func (tx *RawTransaction) Serialize() ([]byte, error) {
	var (
		data []byte
	)
	ss, err := tx.SignerId.Serialize()
	if err != nil {
		return nil, fmt.Errorf("tx serialize: signerId error,Err=%v", err)
	}
	data = append(data, ss...)
	ps, err := tx.PublicKey.Serialize()
	if err != nil {
		return nil, fmt.Errorf("tx serialize: publickey error,Err=%v", err)
	}
	data = append(data, ps...)
	ns, err := tx.Nonce.Serialize()
	if err != nil {
		return nil, fmt.Errorf("tx serialize: nonce error,Err=%v", err)
	}
	data = append(data, ns...)
	rs, err := tx.ReceiverId.Serialize()
	if err != nil {
		return nil, fmt.Errorf("tx serialize: ReceiverId error,Err=%v", err)
	}
	data = append(data, rs...)
	bs, err := tx.BlockHash.Serialize()
	if err != nil {
		return nil, fmt.Errorf("tx serialize: blockhash error,Err=%v", err)
	}
	data = append(data, bs...)
	//序列化action
	al := len(tx.Actions)
	uAL := serialize.U32{
		Value: uint32(al),
	}
	uALData, err := uAL.Serialize()
	if err != nil {
		return nil, fmt.Errorf("tx serialize: action length error,Err=%v", err)
	}
	data = append(data, uALData...)
	for _, action := range tx.Actions {
		as, err := action.Serialize()
		if err != nil {
			return nil, fmt.Errorf("tx serialize: action error,Err=%v", err)
		}
		data = append(data, as...)
	}
	return data, nil
}

func SignTransaction(tx_hex string, privateKey string) (string, error) {
	priv, err := hex.DecodeString(privateKey)
	if err != nil {
		return "", fmt.Errorf("decode private key error,Err=%v", err)
	}
	if len(priv) != 32 {
		return "", fmt.Errorf("private key kength is not equal 32,Len=%d", len(priv))
	}
	data, err := hex.DecodeString(tx_hex)
	if err != nil {
		return "", fmt.Errorf("decode tx hex error,Err=%v", err)
	}
	preSigData := sha256.Sum256(data)
	p := ed25519.NewKeyFromSeed(priv)
	sig := ed25519.Sign(p, preSigData[:])
	if len(sig) != 64 {
		return "", fmt.Errorf("sign error,length is not equal 64,length=%d", len(sig))
	}
	return hex.EncodeToString(sig), nil
}

func CreateSignatureTransaction(tx *RawTransaction, sig string) (*SignatureTransaction, error) {
	var signature []byte
	var err error
	if len(sig) == 128 {
		signature, err = hex.DecodeString(sig)
		if err != nil {
			return nil, err
		}
	} else {
		signature = base58.Decode(sig)
		if len(signature) == 0 {
			return nil, fmt.Errorf("b58 decode sig error,sig=%s", sig)
		}
	}
	stx := new(SignatureTransaction)
	stx.Tx = tx
	stx.Sig = serialize.Signature{
		KeyType: tx.PublicKey.KeyType,
		Value:   signature,
	}
	return stx, nil
}

func (stx *SignatureTransaction) Serialize() ([]byte, error) {
	data, err := stx.Tx.Serialize()
	if err != nil {
		return nil, fmt.Errorf("sign serialize: tx serialize error,Err=%v", err)
	}
	ss, err := stx.Sig.Serialize()
	if err != nil {
		return nil, fmt.Errorf("sign serialize: sig serialize error,Err=%v", err)
	}
	data = append(data, ss...)
	return data, nil
}
