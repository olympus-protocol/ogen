package primitives

import (
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

// ValidatorInitialization is the parameters needed to initialize validators.
type ValidatorInitialization struct {
	PubKey       [48]byte
	PayeeAddress string
}

// InitializationParameters are used in conjunction with ChainParams to generate
// the new genesis state.
type InitializationParameters struct {
	InitialValidators []ValidatorInitialization
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
