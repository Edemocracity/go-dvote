package commands

import (
	"go.vocdoni.io/dvote/crypto/ethereum"
)

var signKey *ethereum.SignKeys

func genRandSignKey() error {
	signKey = ethereum.NewSignKeys()
	if privKey != "" {
		if err := signKey.AddHexKey(privKey); err != nil {
			return err
		}
	} else {
		signKey.Generate()
	}
	return nil
}
