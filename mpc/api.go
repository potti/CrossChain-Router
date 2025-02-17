package mpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/anyswap/CrossChain-Router/v3/common"
	"github.com/anyswap/CrossChain-Router/v3/log"
	"github.com/anyswap/CrossChain-Router/v3/rpc/client"
)

// get mpc sign status error
var (
	ErrGetSignStatusTimeout     = errors.New("getSignStatus timeout")
	ErrGetSignStatusFailed      = errors.New("getSignStatus failure")
	ErrGetSignStatusHasDisagree = errors.New("getSignStatus has disagree")
)

const (
	successStatus = "Success"
)

func newWrongStatusError(subject, status, errInfo string) error {
	return fmt.Errorf("[%v] Wrong status \"%v\", err=\"%v\"", subject, status, errInfo)
}

func (c *Config) wrapPostError(method string, err error) error {
	return fmt.Errorf("[post] %v error, %w", c.mpcAPIPrefix+method, err)
}

func (c *Config) httpPost(result interface{}, method string, params ...interface{}) error {
	return client.RPCPostWithTimeout(c.mpcRPCTimeout, &result, c.defaultMPCNode.mpcRPCAddress, c.mpcAPIPrefix+method, params...)
}

func (c *Config) httpPostTo(result interface{}, rpcAddress, method string, params ...interface{}) error {
	return client.RPCPostWithTimeout(c.mpcRPCTimeout, &result, rpcAddress, c.mpcAPIPrefix+method, params...)
}

// GetEnode call getEnode
func (c *Config) GetEnode(rpcAddr string) (string, error) {
	var result GetEnodeResp
	err := c.httpPostTo(&result, rpcAddr, "getEnode")
	if err != nil {
		return "", c.wrapPostError("getEnode", err)
	}
	if result.Status != successStatus {
		return "", newWrongStatusError("getEnode", result.Status, result.Error)
	}
	return result.Data.Enode, nil
}

// GetSignNonce call getSignNonce
func (c *Config) GetSignNonce(mpcUser, rpcAddr string) (uint64, error) {
	var result DataResultResp
	err := c.httpPostTo(&result, rpcAddr, "getSignNonce", mpcUser)
	if err != nil {
		return 0, c.wrapPostError("getSignNonce", err)
	}
	if result.Status != successStatus {
		return 0, newWrongStatusError("getSignNonce", result.Status, result.Error)
	}
	bi, err := common.GetBigIntFromStr(result.Data.Result)
	if err != nil {
		return 0, fmt.Errorf("getSignNonce can't parse result as big int, %w", err)
	}
	return bi.Uint64(), nil
}

// GetSignStatus call getSignStatus
func (c *Config) GetSignStatus(key, rpcAddr string) (*SignStatus, error) {
	var result DataResultResp
	err := c.httpPostTo(&result, rpcAddr, "getSignStatus", key)
	if err != nil {
		return nil, c.wrapPostError("getSignStatus", err)
	}
	if result.Status != successStatus {
		return nil, newWrongStatusError("getSignStatus", result.Status, "response error "+result.Error)
	}
	data := result.Data.Result
	var signStatus SignStatus
	err = json.Unmarshal([]byte(data), &signStatus)
	if err != nil {
		return nil, c.wrapPostError("getSignStatus", err)
	}
	switch signStatus.Status {
	case "Failure":
		log.Info("getSignStatus Failure", "keyID", key, "status", data)
		if signStatus.HasDisagree() {
			return nil, ErrGetSignStatusHasDisagree
		}
		return nil, ErrGetSignStatusFailed
	case "Timeout":
		log.Info("getSignStatus Timeout", "keyID", key, "status", data)
		return nil, ErrGetSignStatusTimeout
	case successStatus:
		return &signStatus, nil
	default:
		return nil, newWrongStatusError("getSignStatus", signStatus.Status, "sign status error "+signStatus.Error)
	}
}

// GetCurNodeSignInfo call getCurNodeSignInfo
// filter out invalid sign info and
// filter out expired sign info if `expiredInterval` is greater than 0
func (c *Config) GetCurNodeSignInfo(expiredInterval int64) ([]*SignInfoData, error) {
	var result SignInfoResp
	err := c.httpPost(&result, "getCurNodeSignInfo", c.defaultMPCNode.keyWrapper.Address.String())
	if err != nil {
		return nil, c.wrapPostError("getCurNodeSignInfo", err)
	}
	if result.Status != successStatus {
		return nil, newWrongStatusError("getCurNodeSignInfo", result.Status, result.Error)
	}
	signInfoSortedSlice := make(SignInfoSortedSlice, 0, len(result.Data))
	for _, signInfo := range result.Data {
		if !signInfo.IsValid() {
			log.Trace("filter out invalid sign info", "signInfo", signInfo)
			continue
		}
		signInfo.timestamp, _ = common.GetUint64FromStr(signInfo.TimeStamp)
		if expiredInterval > 0 && int64(signInfo.timestamp/1000)+expiredInterval < time.Now().Unix() {
			log.Trace("filter out expired sign info", "signInfo", signInfo)
			continue
		}
		signInfoSortedSlice = append(signInfoSortedSlice, signInfo)
	}
	sort.Stable(signInfoSortedSlice)
	return signInfoSortedSlice, nil
}

// Sign call sign
func (c *Config) Sign(raw, rpcAddr string) (string, error) {
	var result DataResultResp
	err := c.httpPostTo(&result, rpcAddr, "sign", raw)
	if err != nil {
		return "", c.wrapPostError("sign", err)
	}
	if result.Status != successStatus {
		return "", newWrongStatusError("sign", result.Status, result.Error)
	}
	return result.Data.Result, nil
}

// AcceptSign call acceptSign
func (c *Config) AcceptSign(raw string) (string, error) {
	var result DataResultResp
	err := c.httpPost(&result, "acceptSign", raw)
	if err != nil {
		return "", c.wrapPostError("acceptSign", err)
	}
	if result.Status != successStatus {
		return "", newWrongStatusError("acceptSign", result.Status, result.Error)
	}
	return result.Data.Result, nil
}

// GetGroupByID call getGroupByID
func (c *Config) GetGroupByID(groupID, rpcAddr string) (*GroupInfo, error) {
	var result GetGroupByIDResp
	err := c.httpPostTo(&result, rpcAddr, "getGroupByID", groupID)
	if err != nil {
		return nil, c.wrapPostError("getGroupByID", err)
	}
	if result.Status != successStatus {
		return nil, newWrongStatusError("getGroupByID", result.Status, result.Error)
	}
	return result.Data, nil
}
