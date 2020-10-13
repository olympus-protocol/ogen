package hdwallet

import (
	"github.com/olympus-protocol/ogen/pkg/bls"
	util "github.com/wealdtech/go-eth2-util"
)

// CreateHDWallet will create a single secret key from a seed and a path
func CreateHDWallet(seed []byte, path string) (*bls.SecretKey, error) {
	key, err := util.PrivateKeyFromSeedAndPath(seed, path)
	if err != nil {
		return nil, err
	}
	return bls.SecretKeyFromBytes(key.Marshal())
}
