package near

import (
	"github.com/near/borsh-go"
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
	SuccessValue     interface{} `json:"SuccessValue,omitempty"`
	SuccessReceiptId interface{} `json:"SuccessReceiptId,omitempty"`
	Failure          interface{} `json:"Failure,omitempty"`
	Unknown          interface{} `json:"Unknown,omitempty"`
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

// Action simulates an enum for Borsh encoding.
type Action struct {
	Enum         borsh.Enum `borsh_enum:"true"` // treat struct as complex enum when serializing/deserializing
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

// A Transaction encodes a NEAR transaction.
type RawTransaction struct {
	SignerID   string
	PublicKey  PublicKey
	Nonce      uint64
	ReceiverID string
	BlockHash  [32]byte
	Actions    []Action
}

// PublicKey encoding for NEAR.
type PublicKey struct {
	KeyType uint8
	Data    [32]byte
}

type SignedTransaction struct {
	Transaction RawTransaction
	Signature   Signature
}

// A Signature used for signing transaction.
type Signature struct {
	KeyType uint8
	Data    [64]byte
}
