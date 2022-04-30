package near

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/anyswap/CrossChain-Router/v3/tokens"
	"github.com/btcsuite/btcutil/base58"
)

var (
	ed25519Prefix = "ed25519:"
)

// IsValidAddress check address
func (b *Bridge) IsValidAddress(address string) bool {
	return true
}

func (b *Bridge) GetAccountNonce(account, publicKey string) (uint64, error) {
	urls := b.GatewayConfig.APIAddress
	for _, url := range urls {
		result, err := GetAccountNonce(url, account, publicKey)
		if err == nil {
			return result, nil
		}
	}
	return 0, tokens.ErrRPCQueryError
}

// PublicKeyToAddress public key to address
func (b *Bridge) PublicKeyToAddress(pubKey string) (string, error) {
	return "", tokens.ErrNotImplemented
}

func GenerateKey() (seed, pub []byte, err error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	return priv.Seed(), pub, nil
}

func GeneratePubKeyBySeed(seed []byte) ([]byte, error) {
	priv := ed25519.NewKeyFromSeed(seed)

	pub := priv[32:]
	if len(pub) != 32 {
		return nil, errors.New("public key length is not equal 32")
	}
	return pub, nil
}

func GeneratePubKeyByBase58(b58Key string) ([]byte, error) {
	seed := base58.Decode(b58Key)
	if len(seed) == 0 {
		return nil, errors.New("base 58 decode error")
	}
	if len(seed) != 32 {
		return nil, errors.New("seed length is not equal 32")
	}
	return GeneratePubKeyBySeed(seed)
}

func PublicKeyToString(pub []byte) string {
	publicKey := base58.Encode(pub)
	return "ed25519:" + publicKey
}

func PublicKeyToAddress(pub []byte) string {
	return hex.EncodeToString(pub)
}

func StringToPublicKey(pub string) ed25519.PublicKey {
	pubKey := base58.Decode(strings.TrimPrefix(pub, ed25519Prefix))
	return ed25519.PublicKey(pubKey)
}

func StringToPrivateKey(priv string) ed25519.PrivateKey {
	privateKey := base58.Decode(strings.TrimPrefix(priv, ed25519Prefix))
	return ed25519.PrivateKey(privateKey)
}

// VerifyMPCPubKey verify mpc address and public key is matching
func VerifyMPCPubKey(mpcAddress, mpcPubkey string) error {
	return tokens.ErrNotImplemented
}
