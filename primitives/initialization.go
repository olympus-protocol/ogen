package primitives

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/utils/bech32"
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
	PayeeAddress string               `json:"withdraw_address"`
}

// InitializationParameters are used in conjunction with ChainParams to generate
// the new genesis state.
type InitializationParameters struct {
	InitialValidators []ValidatorInitialization
	GenesisTime       time.Time
}

// GetGenesisStateWithInitializationParameters gets the genesis state with certain parameters.
func GetGenesisStateWithInitializationParameters(genesisHash chainhash.Hash, ip *InitializationParameters, p *params.ChainParams) (*State, error) {
	initialValidators := make([]Worker, len(ip.InitialValidators))

	for i, v := range ip.InitialValidators {
		_, pkh, err := bech32.Decode(v.PayeeAddress)
		if err != nil {
			return nil, err
		}

		if len(pkh) != 20 {
			return nil, fmt.Errorf("expected payee address to be length 20, got %d", len(pkh))
		}

		var pkhBytes [20]byte
		copy(pkhBytes[:], pkh)

		initialValidators[i] = Worker{
			Balance:      p.DepositAmount * p.UnitsPerCoin,
			PubKey:       v.PubKey,
			PayeeAddress: pkhBytes,
			Status:       StatusActive,
		}
	}

	premineAddr, _ := hex.DecodeString("d6a0ebece6b6f9027236157a519ccffe14084f0d")

	var premineAddrArr [20]byte
	copy(premineAddrArr[:], premineAddr)

	s := &State{
		UtxoState: UtxoState{
			Balances: map[[20]byte]uint64{
				premineAddrArr: 1000 * 1000000, // 1 million coins
			},
			Nonces: make(map[[20]byte]uint64),
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
	s.CurrentEpochVoteAssignments = Shuffle(chainhash.Hash{}, s.GetActiveValidatorIndices())
	s.PreviousEpochVoteAssignments = Shuffle(chainhash.Hash{}, s.GetActiveValidatorIndices())

	return s, nil
}
