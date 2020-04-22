package primitives

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

type InitializationPubkey [48]byte

func (ip InitializationPubkey) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	outBuf := base64.NewEncoder(base64.StdEncoding, buf)
	_, err := outBuf.Write(ip[:])
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf("\"%s\"", string(buf.Bytes()))), nil
}

func (ip *InitializationPubkey) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	reader := base64.NewDecoder(base64.StdEncoding, bytes.NewBuffer([]byte(s)))
	out, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	copy(ip[:], out)
	return nil
}

// ValidatorInitialization is the parameters needed to initialize validators.
type ValidatorInitialization struct {
	PubKey       InitializationPubkey `json:"pubkey"`
	PayeeAddress string `json:"withdraw_address"`
}

// InitializationParameters are used in conjunction with ChainParams to generate
// the new genesis state.
type InitializationParameters struct {
	InitialValidators []ValidatorInitialization
	GenesisTime       time.Time
}

// GetGenesisStateWithInitializationParameters gets the genesis state with certain parameters.
func GetGenesisStateWithInitializationParameters(genesisHash chainhash.Hash, ip *InitializationParameters, p *params.ChainParams) *State {
	initialValidators := make([]Worker, len(ip.InitialValidators))

	for i, v := range ip.InitialValidators {
		initialValidators[i] = Worker{
			OutPoint: OutPoint{
				TxHash: [32]byte{},
				Index:  0,
			},
			Balance:      p.DepositAmount * p.UnitsPerCoin,
			PubKey:       v.PubKey,
			PayeeAddress: v.PayeeAddress,
			Status:       StatusActive,
		}
	}

	return &State{
		UtxoState: UtxoState{
			UTXOs: make(map[chainhash.Hash]Utxo),
		},
		GovernanceState: GovernanceState{
			Proposals: make(map[chainhash.Hash]GovernanceProposal),
		},
		UserState: UserState{
			Users: make(map[chainhash.Hash]User),
		},
		ValidatorRegistry:             initialValidators,
		LatestValidatorRegistryChange: 0,
		RANDAO:                        chainhash.Hash{},
		NextRANDAO:                    chainhash.Hash{},
		Slot:                          0,
		EpochIndex:                    0,
		ProposerQueue:                 DetermineNextProposers(chainhash.Hash{}, initialValidators, p),
		NextProposerQueue:             DetermineNextProposers(chainhash.Hash{}, initialValidators, p),
		JustificationBitfield:         0,
		JustifiedEpoch:                0,
		FinalizedEpoch:                0,
		LatestBlockHashes:             make([]chainhash.Hash, p.LatestBlockRootsLength),
		JustifiedEpochHash:            genesisHash,
		CurrentEpochVotes:             make([]AcceptedVoteInfo, 0),
		PreviousJustifiedEpoch:        0,
		PreviousJustifiedEpochHash:    genesisHash,
		PreviousEpochVotes:            make([]AcceptedVoteInfo, 0),
	}
}
