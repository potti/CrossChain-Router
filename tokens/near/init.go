package near

import (
	"github.com/anyswap/CrossChain-Router/v3/router"
)

// InitAfterConfig init variables (ie. extra members) after loading config
func (b *Bridge) InitAfterConfig() {
	router.SetMPCPublicKey("userdemo.testnet", "ed25519:MTjQVM8fgKSgfq8Uuer2nGRXL9dHLGYQDBkdxwmrdDB")
}
