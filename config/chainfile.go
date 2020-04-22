package config

import "github.com/olympus-protocol/ogen/primitives"

type ChainFile struct {
	Validators         []primitives.ValidatorInitialization `json:"validators"`
	GenesisTime        int64                             `json:"genesis_time"`
	InitialConnections []string                           `json:"initial_connections"`
}
