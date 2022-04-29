package near

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/anyswap/CrossChain-Router/v3/common"
	"github.com/anyswap/CrossChain-Router/v3/log"
	"github.com/anyswap/CrossChain-Router/v3/params"
	"github.com/anyswap/CrossChain-Router/v3/router"
	"github.com/anyswap/CrossChain-Router/v3/tokens"
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
	nonce++
	if getNonceErr != nil {
		return nil, getNonceErr
	}

	blockHash, getBlockHashErr := b.GetLatestBlockHash()
	if getBlockHashErr != nil {
		return nil, getBlockHashErr
	}

	actions := createFunctionCall(args.SwapID, multichainToken, args.Bind, args.OriginValue.String(), args.FromChainID.String())
	rawTx = createTransaction(args.From, PublicKeyFromEd25519(StringToPublicKey(mpcPubkey)), b.ChainConfig.RouterContract, nonce, base58.Decode(blockHash), actions)
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

func createTransaction(
	signerID string,
	publicKey PublicKey,
	receiverID string,
	nonce uint64,
	blockHash []byte,
	actions []Action,
) *RawTransaction {
	var tx RawTransaction
	tx.SignerID = signerID
	tx.PublicKey = publicKey
	tx.ReceiverID = receiverID
	tx.Nonce = nonce
	copy(tx.BlockHash[:], blockHash)
	tx.Actions = actions
	return &tx
}

func createFunctionCall(txHash, token, to, amount, from_chain_id string) []Action {
	log.Info("createFunctionCall", "txHash", txHash, "token", token, "to", to, "amount", amount, "from_chain_id", from_chain_id)
	callArgs := &AnySwapIn{
		Tx:            txHash,
		Token:         token,
		To:            to,
		Amount:        amount,
		From_chain_id: from_chain_id,
	}
	argsBytes, _ := json.Marshal(callArgs)
	return []Action{{
		Enum: 2,
		FunctionCall: FunctionCall{
			MethodName: "any_swap_in",
			Args:       argsBytes,
			Gas:        300_000_000_000_000,
			Deposit:    *big.NewInt(0),
		},
	}}
}
