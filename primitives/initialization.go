package primitives

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/utils/bech32"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/prysmaticlabs/go-bitfield"
)

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
func GetGenesisStateWithInitializationParameters(genesisHash chainhash.Hash, ip *InitializationParameters, p *params.ChainParams) (*State, error) {
	initialValidators := make([]*Validator, len(ip.InitialValidators))

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
		initialValidators[i] = &Validator{
			Balance:          p.DepositAmount * p.UnitsPerCoin,
			PubKey:           pubKey,
			PayeeAddress:     pkhBytes,
			Status:           StatusActive,
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
	s := &State{
		CoinsState: CoinsState{
			Balances: map[[20]byte]uint64{
				premineAddrArr: 400 * 1000000, // 400k coins
			},
			Nonces: make(map[[20]byte]uint64),
		},
		ValidatorRegistry:             initialValidators,
		LatestValidatorRegistryChange: 0,
		RANDAO:                        chainhash.Hash{},
		NextRANDAO:                    chainhash.Hash{},
		Slot:                          0,
		EpochIndex:                    0,
		JustificationBitfield:         0,
		JustifiedEpoch:                0,
		FinalizedEpoch:                0,
		LatestBlockHashes:             make([][32]byte, p.LatestBlockRootsLength),
		JustifiedEpochHash:            genesisHash,
		CurrentEpochVotes:             make([]*AcceptedVoteInfo, 0),
		PreviousJustifiedEpoch:        0,
		PreviousJustifiedEpochHash:    genesisHash,
		PreviousEpochVotes:            make([]*AcceptedVoteInfo, 0),
		CurrentManagers:               p.InitialManagers,
		VoteEpoch:                     0,
		VoteEpochStartSlot:            0,
		VotingState:                   GovernanceStateActive,
		LastPaidSlot:                  0,
	}
	activeValidators := s.GetValidatorIndicesActiveAt(0)
	s.ProposerQueue = DetermineNextProposers(chainhash.Hash{}, activeValidators, p)
	s.NextProposerQueue = DetermineNextProposers(chainhash.Hash{}, activeValidators, p)
	s.CurrentEpochVoteAssignments = Shuffle(chainhash.Hash{}, activeValidators)
	s.PreviousEpochVoteAssignments = Shuffle(chainhash.Hash{}, activeValidators)
	s.ManagerReplacement = bitfield.NewBitlist(uint64(len(s.CurrentManagers)))

	return s, nil
}
