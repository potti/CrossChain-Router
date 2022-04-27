package near

import (
	"github.com/anyswap/CrossChain-Router/v3/tokens/near/serialize"
)

type TransactionResult struct {
	Status             Status             `json:"status"`
	Transaction        Transaction        `json:"transaction"`
	TransactionOutcome TransactionOutcome `json:"transaction_outcome"`
	ReceiptsOutcome    []ReceiptsOutcome  `json:"receipts_outcome"`
}

type BlockDetail struct {
	Header BlockHeader `json:"header"`
}

type BlockHeader struct {
	Hash   string `json:"hash"`
	Height uint64 `json:"height"`
}

type Status struct {
	SuccessValue     string `json:"SuccessValue,omitempty"`
	SuccessReceiptId string `json:"SuccessReceiptId,omitempty"`
	Failure          string `json:"Failure,omitempty"`
	Unknown          string `json:"Unknown,omitempty"`
}

type Transaction struct {
	Actions    []Action `json:"actions"`
	Hash       string   `json:"hash"`
	Nonce      uint64   `json:"nonce"`
	PublicKey  string   `json:"public_key"`
	ReceiverID string   `json:"receiver_id"`
	Signature  string   `json:"signature"`
	SignerID   string   `json:"signer_id"`
}

type TransactionOutcome struct {
	BlockHash string  `json:"block_hash"`
	ID        string  `json:"id"`
	Outcome   Outcome `json:"outcome"`
	Proof     []Proof `json:"proof"`
}

type ReceiptsOutcome struct {
	BlockHash string  `json:"block_hash"`
	ID        string  `json:"id"`
	Outcome   Outcome `json:"outcome"`
	Proof     []Proof `json:"proof"`
}

type Outcome struct {
	ExecutorID  string   `json:"executor_id"`
	GasBurnt    int64    `json:"gas_burnt"`
	Logs        []string `json:"logs"`
	ReceiptIds  []string `json:"receipt_ids"`
	Status      Status   `json:"status"`
	TokensBurnt string   `json:"tokens_burnt"`
}

type Proof struct {
	Direction string `json:"direction"`
	Hash      string `json:"hash"`
}

type Action struct {
	FunctionCall FunctionCall
	Transfer     Transfer
}

type Transfer struct {
	Deposit string
}

type FunctionCall struct {
	MethodName string
	Args       []byte
	Gas        uint64
	Deposit    string
}

type RawTransaction struct {
	SignerId   serialize.String
	PublicKey  serialize.PublicKey
	Nonce      serialize.U64
	ReceiverId serialize.String
	BlockHash  serialize.BlockHash
	Actions    []serialize.IAction
}

type SignatureTransaction struct {
	Sig serialize.Signature
	Tx  *RawTransaction
}
