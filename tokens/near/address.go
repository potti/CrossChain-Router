package near

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/anyswap/CrossChain-Router/v3/common"
	"github.com/anyswap/CrossChain-Router/v3/tokens"
	"github.com/btcsuite/btcutil/base58"
)

var (
	ed25519Prefix = "ed25519:"
)

// IsValidAddress check address
func (b *Bridge) IsValidAddress(address string) bool {
	return address != ""
}

func (b *Bridge) GetAccountNonce(account, publicKey string) (uint64, error) {
	urls := append(b.GatewayConfig.APIAddress, b.GatewayConfig.APIAddressExt...)
	for _, url := range urls {
		result, err := GetAccountNonce(url, account, publicKey)
		if err == nil {
			return result, nil
		}
	}
	return 0, tokens.ErrGetAccountNonce
}

// PublicKeyToAddress public key to address
func (b *Bridge) PublicKeyToAddress(nearPublicKey string) (string, error) {
	return nearPublicKeyTompcSignPublicKey(nearPublicKey), nil
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

func nearPublicKeyTompcSignPublicKey(nearPublicKey string) string {
	pubKey := StringToPublicKey(nearPublicKey)
	return PublicKeyToAddress(pubKey)
}

func (b *Bridge) VerifyPubKey(address, pubkey string) error {
	if common.IsHexHash(address) {
		pubAddr := nearPublicKeyTompcSignPublicKey(pubkey)
		if !strings.EqualFold(pubAddr, address) {
			return fmt.Errorf("address %v and public key address %v is not match", address, pubAddr)
		}
	}
	_, err := b.GetAccountNonce(address, pubkey)
	if err != nil {
		return fmt.Errorf("verify public key failed, %w", err)
	}
	return nil
}
