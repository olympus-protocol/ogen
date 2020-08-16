package state

import (
	"encoding/hex"
	"fmt"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"time"

	"github.com/olympus-protocol/ogen/pkg/bech32"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/params"
)

// ChainFile represents the on-disk chain file used to initialize the chain.
type ChainFile struct {
	Validators         []ValidatorInitialization `json:"validators"`
	GenesisTime        uint64                    `json:"genesis_time"`
	InitialConnections []string                  `json:"initial_connections"`
	PremineAddress     string                    `json:"premine_address"`
}

// ToInitializationParameters converts the chain configuration file to initialization
// parameters.
func (cf *ChainFile) ToInitializationParameters() InitializationParameters {
	ip := InitializationParameters{
		InitialValidators: cf.Validators,
		GenesisTime:       time.Unix(int64(cf.GenesisTime), 0),
		PremineAddress:    cf.PremineAddress,
	}

	if cf.GenesisTime == 0 {
		ip.GenesisTime = time.Unix(time.Now().Add(5*time.Second).Unix(), 0)
	}

	return ip
}

// ValidatorInitialization is the parameters needed to initialize validators.
type ValidatorInitialization struct {
	PubKey       string `json:"pubkey"`
	PayeeAddress string `json:"withdraw_address"`
}

// InitializationParameters are used in conjunction with ChainParams to generate
// the new genesis state.
type InitializationParameters struct {
	InitialValidators []ValidatorInitialization
	PremineAddress    string
	GenesisTime       time.Time
}

// GetGenesisStateWithInitializationParameters gets the genesis state with certain parameters.
func GetGenesisStateWithInitializationParameters(genesisHash chainhash.Hash, ip *InitializationParameters, p *params.ChainParams) (State, error) {
	initialValidators := make([]*primitives.Validator, len(ip.InitialValidators))

	for i, v := range ip.InitialValidators {
		_, pkh, err := bech32.Decode(v.PayeeAddress)
		if err != nil {
			return nil, err
		}

		if len(pkh) != 20 {
			return nil, fmt.Errorf("expected payee address to be length 20, got %d", len(pkh))
		}

		var pkhBytes [20]byte
		var pubKey [48]byte
		copy(pkhBytes[:], pkh)
		pubKeyBytes, err := hex.DecodeString(v.PubKey)
		if err != nil {
			return nil, fmt.Errorf("unable to decode pubkey to bytes")
		}
		copy(pubKey[:], pubKeyBytes)
		initialValidators[i] = &primitives.Validator{
			Balance:          p.DepositAmount * p.UnitsPerCoin,
			PubKey:           pubKey,
			PayeeAddress:     pkhBytes,
			Status:           primitives.StatusActive,
			FirstActiveEpoch: 0,
			LastActiveEpoch:  0,
		}
	}

	_, premineAddr, err := bech32.Decode(ip.PremineAddress)
	if err != nil {
		return nil, err
	}

	var premineAddrArr [20]byte
	copy(premineAddrArr[:], premineAddr)

	cs := primitives.CoinsState{
		Balances: map[[20]byte]uint64{
			premineAddrArr: 400000 * p.UnitsPerCoin,
		},
		Nonces: make(map[[20]byte]uint64),
	}

	s := NewState(cs, initialValidators, genesisHash, p)

	return s, nil
}
