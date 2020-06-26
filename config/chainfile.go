package config

import (
	"time"

	"github.com/olympus-protocol/ogen/primitives"
)

// ChainFile represents the on-disk chain file used to initialize the chain.
type ChainFile struct {
	Validators         []primitives.ValidatorInitialization `json:"validators"`
	GenesisTime        uint64                               `json:"genesis_time"`
	InitialConnections []string                             `json:"initial_connections"`
	PremineAddress     string                               `json:"premine_address"`
}

// ToInitializationParameters converts the chain configuration file to initialization
// parameters.
func (cf *ChainFile) ToInitializationParameters() primitives.InitializationParameters {
	ip := primitives.InitializationParameters{
		InitialValidators: cf.Validators,
		GenesisTime:       time.Unix(int64(cf.GenesisTime), 0),
		PremineAddress:    cf.PremineAddress,
	}

	if cf.GenesisTime == 0 {
		ip.GenesisTime = time.Unix(time.Now().Add(5*time.Second).Unix(), 0)
	}

	return ip
}
