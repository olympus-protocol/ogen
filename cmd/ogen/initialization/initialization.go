package initialization

import (
	"encoding/json"
	"io/ioutil"
	"path"
)

type Validators struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

type NetworkInitialParams struct {
	Validators        []Validators `json:"validators"`
	PremineAddress    string       `json:"premine_address"`
	PreminePrivateKey string       `json:"premine_private_key"`
	GenesisTime       int64        `json:"genesis_time"`
}

// LoadParams returns the initialization params required for the network specified.
func LoadParams(network string) (NetworkInitialParams, error) {
	filename := network + "_params.json"
	b, err := ioutil.ReadFile(path.Join("./cmd/ogen/initialization", filename))
	if err != nil {
		return NetworkInitialParams{}, err
	}
	var params NetworkInitialParams
	err = json.Unmarshal(b, &params)
	if err != nil {
		return NetworkInitialParams{}, err
	}
	return params, err
}
