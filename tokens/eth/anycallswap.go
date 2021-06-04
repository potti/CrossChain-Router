package eth

import (
	"bytes"
	"errors"

	"github.com/anyswap/CrossChain-Router/v3/common"
	"github.com/anyswap/CrossChain-Router/v3/common/hexutil"
	"github.com/anyswap/CrossChain-Router/v3/log"
	"github.com/anyswap/CrossChain-Router/v3/tokens"
	"github.com/anyswap/CrossChain-Router/v3/tokens/eth/abicoder"
	"github.com/anyswap/CrossChain-Router/v3/types"
)

// anycall lot topics and func hashes
var (
	LogAnyCallTopic = common.FromHex("0x3d1b3d059223895589208a5541dce543eab6d5942b3b1129231a942d1c47bc45")

	AnyCallFuncHash = common.FromHex("0x32f29022")
)

// nolint:dupl // ok
func (b *Bridge) registerAnyCallSwapTx(txHash string, logIndex int) ([]*tokens.SwapTxInfo, []error) {
	commonInfo := &tokens.SwapTxInfo{SwapInfo: tokens.SwapInfo{AnyCallSwapInfo: &tokens.AnyCallSwapInfo{}}}
	commonInfo.SwapType = tokens.AnyCallSwapType // SwapType
	commonInfo.Hash = txHash                     // Hash
	commonInfo.LogIndex = logIndex               // LogIndex

	receipt, err := b.verifySwapTxReceipt(commonInfo, b.ChainConfig.RouterContract, true)
	if err != nil {
		return []*tokens.SwapTxInfo{commonInfo}, []error{err}
	}

	swapInfos := make([]*tokens.SwapTxInfo, 0)
	errs := make([]error, 0)
	startIndex, endIndex := 1, len(receipt.Logs)

	if logIndex != 0 {
		if logIndex >= endIndex || logIndex < 0 {
			return []*tokens.SwapTxInfo{commonInfo}, []error{tokens.ErrLogIndexOutOfRange}
		}
		startIndex = logIndex
		endIndex = logIndex + 1
	}

	for i := startIndex; i < endIndex; i++ {
		swapInfo := &tokens.SwapTxInfo{}
		*swapInfo = *commonInfo
		swapInfo.AnyCallSwapInfo = &tokens.AnyCallSwapInfo{}
		swapInfo.LogIndex = i // LogIndex
		err := b.verifyAnyCallSwapTxLog(swapInfo, receipt.Logs[i])
		switch {
		case errors.Is(err, tokens.ErrSwapoutLogNotFound):
			continue
		case err == nil:
			err = b.checkAnyCallSwapInfo(swapInfo)
		default:
			log.Debug(b.ChainConfig.BlockChain+" register anycall swap error", "txHash", txHash, "logIndex", swapInfo.LogIndex, "err", err)
		}
		swapInfos = append(swapInfos, swapInfo)
		errs = append(errs, err)
	}

	if len(swapInfos) == 0 {
		return []*tokens.SwapTxInfo{commonInfo}, []error{tokens.ErrSwapoutLogNotFound}
	}

	return swapInfos, errs
}

func (b *Bridge) verifyAnyCallSwapTx(txHash string, logIndex int, allowUnstable bool) (*tokens.SwapTxInfo, error) {
	swapInfo := &tokens.SwapTxInfo{SwapInfo: tokens.SwapInfo{AnyCallSwapInfo: &tokens.AnyCallSwapInfo{}}}
	swapInfo.SwapType = tokens.AnyCallSwapType // SwapType
	swapInfo.Hash = txHash                     // Hash
	swapInfo.LogIndex = logIndex               // LogIndex

	receipt, err := b.verifySwapTxReceipt(swapInfo, b.ChainConfig.RouterContract, allowUnstable)
	if err != nil {
		return swapInfo, err
	}

	if logIndex >= len(receipt.Logs) {
		return swapInfo, tokens.ErrLogIndexOutOfRange
	}

	err = b.verifyAnyCallSwapTxLog(swapInfo, receipt.Logs[logIndex])
	if err != nil {
		return swapInfo, err
	}

	err = b.checkAnyCallSwapInfo(swapInfo)
	if err != nil {
		return swapInfo, err
	}

	if !allowUnstable {
		log.Debug("verify anycall swap tx stable pass",
			"from", swapInfo.From, "to", swapInfo.To, "txid", txHash, "logIndex", logIndex,
			"height", swapInfo.Height, "timestamp", swapInfo.Timestamp,
			"fromChainID", swapInfo.FromChainID, "toChainID", swapInfo.ToChainID)
	}

	return swapInfo, nil
}

func (b *Bridge) verifyAnyCallSwapTxLog(swapInfo *tokens.SwapTxInfo, rlog *types.RPCLog) (err error) {
	logTopic := rlog.Topics[0].Bytes()
	if !bytes.Equal(logTopic, LogAnyCallTopic) {
		return tokens.ErrSwapoutLogNotFound
	}

	err = b.parseAnyCallSwapTxLog(swapInfo, rlog)
	if err != nil {
		log.Info(b.ChainConfig.BlockChain+" b.verifyAnyCallSwapTxLog fail", "tx", swapInfo.Hash, "logIndex", rlog.Index, "err", err)
		return err
	}

	if rlog.Removed != nil && *rlog.Removed {
		return tokens.ErrTxWithRemovedLog
	}
	return nil
}

func (b *Bridge) parseAnyCallSwapTxLog(swapInfo *tokens.SwapTxInfo, rlog *types.RPCLog) (err error) {
	logTopics := rlog.Topics
	if len(logTopics) != 2 {
		return tokens.ErrTxWithWrongTopics
	}
	logData := *rlog.Data
	if len(logData) < 320 {
		return abicoder.ErrParseDataError
	}

	swapInfo.CallFrom = common.BytesToAddress(logTopics[1].Bytes()).LowerHex()
	swapInfo.CallTo, err = abicoder.ParseAddressSliceInData(logData, 0)
	if err != nil {
		return err
	}
	swapInfo.CallData, err = abicoder.ParseBytesSliceInData(logData, 32)
	if err != nil {
		return err
	}
	swapInfo.Callbacks, err = abicoder.ParseAddressSliceInData(logData, 64)
	if err != nil {
		return err
	}
	swapInfo.CallNonces, err = abicoder.ParseNumberSliceAsBigIntsInData(logData, 96)
	if err != nil {
		return err
	}
	swapInfo.FromChainID = common.GetBigInt(logData, 128, 32)
	swapInfo.ToChainID = common.GetBigInt(logData, 160, 32)
	return nil
}

func (b *Bridge) checkAnyCallSwapInfo(swapInfo *tokens.SwapTxInfo) error {
	return nil
}

func (b *Bridge) buildAnyCallSwapTxInput(args *tokens.BuildTxArgs) (err error) {
	if args.AnyCallSwapInfo == nil {
		return errors.New("build anycall swaptx without swapinfo")
	}
	funcHash := AnyCallFuncHash

	if b.ChainConfig.ChainID != args.ToChainID.String() {
		return errors.New("anycall to chainId mismatch")
	}

	input := abicoder.PackDataWithFuncHash(funcHash,
		common.HexToAddress(args.CallFrom),
		toAddresses(args.CallTo),
		args.CallData,
		toAddresses(args.Callbacks),
		args.CallNonces,
		args.FromChainID,
	)
	args.Input = (*hexutil.Bytes)(&input)  // input
	args.To = b.ChainConfig.RouterContract // to

	return nil
}
