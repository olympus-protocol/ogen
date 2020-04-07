package primitives

import (
	"errors"
	"fmt"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

func (s *State) getValidatorAtVoteSlot(validator uint64, slot uint64, p *params.ChainParams) Worker {
	numValidatorsAtSlot := (uint64(len(s.ProposerQueue)) / p.EpochLength) + 1
	slotIndex := slot % p.EpochLength

	valIdx := s.ProposerQueue[slotIndex*numValidatorsAtSlot+validator]

	return s.ValidatorRegistry[valIdx]
}

func (s *State) isVoteValid(v *MultiValidatorVote, p *params.ChainParams) error {
	if v.Data.ToEpoch == s.EpochIndex {
		if v.Data.FromEpoch != s.JustifiedEpoch {
			return fmt.Errorf("expected from epoch to match justified epoch (expected: %d, got: %d)", s.JustifiedEpoch, v.Data.FromEpoch)
		}

		if !s.JustifiedEpochHash.IsEqual(&v.Data.FromHash) {
			return fmt.Errorf("justified block hash is wrong (expected: %s, got: %s)", s.JustifiedEpochHash, v.Data.FromHash)
		}
	} else if v.Data.ToEpoch == s.EpochIndex-1 {
		if v.Data.FromEpoch != s.PreviousJustifiedEpoch {
			return fmt.Errorf("expected from epoch to match previous justified epoch (expected: %d, got: %d)", s.PreviousJustifiedEpoch, v.Data.FromEpoch)
		}

		if !s.PreviousJustifiedEpochHash.IsEqual(&v.Data.FromHash) {
			return fmt.Errorf("previous justified block hash is wrong (expected: %s, got: %s)", s.PreviousJustifiedEpochHash, v.Data.FromHash)
		}
	} else {
		return fmt.Errorf("vote should have target epoch of either the current epoch (%d) or the previous epoch (%d) but got %d", s.EpochIndex, s.EpochIndex-1, v.Data.FromEpoch)
	}

	aggPubs := bls.NewAggregatePublicKey()
	for i := range v.ParticipationBitfield {
		for j := 0; j < 8; j++ {
			validator := (i * 8) + j

			worker := s.getValidatorAtVoteSlot(uint64(validator), v.Data.Slot, p)
			pub, err := bls.DeserializePublicKey(worker.PubKey)
			if err != nil {
				return err
			}
			aggPubs.AggregatePubKey(pub)
		}
	}

	h := v.Data.Hash()
	valid, err := bls.VerifySig(aggPubs, h[:], &v.Signature)
	if err != nil {
		return err
	}

	if !valid {
		return fmt.Errorf("aggregate signature did not validate")
	}

	return nil
}

func (s *State) processVote(v *MultiValidatorVote, p *params.ChainParams, proposerIndex uint32) error {
	if v.Data.Slot+p.MinAttestationInclusionDelay > s.Slot {
		return fmt.Errorf("vote included too soon (expected s.Slot > %d, got %d)", v.Data.Slot+p.MinAttestationInclusionDelay, s.Slot)
	}

	if v.Data.Slot+p.EpochLength <= s.Slot {
		return fmt.Errorf("vote not included within 1 epoch (latest: %d, got: %d)", v.Data.Slot+p.EpochLength-1, s.Slot)
	}

	if (v.Data.Slot-1)/p.EpochLength != v.Data.ToEpoch {
		return errors.New("vote slot did not match target epoch")
	}

	err := s.isVoteValid(v, p)
	if err != nil {
		return err
	}

	if v.Data.ToEpoch == s.EpochIndex {
		s.CurrentEpochVotes = append(s.CurrentEpochVotes, AcceptedVoteInfo{
			Data:                  v.Data,
			ParticipationBitfield: v.ParticipationBitfield,
			Proposer:              proposerIndex,
			InclusionDelay:        s.Slot - v.Data.Slot,
		})
	} else {
		s.PreviousEpochVotes = append(s.PreviousEpochVotes, AcceptedVoteInfo{
			Data:                  v.Data,
			ParticipationBitfield: v.ParticipationBitfield,
			Proposer:              proposerIndex,
			InclusionDelay:        s.Slot - v.Data.Slot,
		})
	}

	return nil
}

// ProcessBlock runs a block transition on the state and mutates state.
func (s *State) ProcessBlock(b *Block, p *params.ChainParams) error {
	if b.Header.Slot != s.Slot {
		return fmt.Errorf("state is not updated to slot %d, instead got %d", b.Header.Slot, s.Slot)
	}

	if uint64(len(b.Votes)) > p.MaxVotesPerBlock {
		return fmt.Errorf("block has too many votes (max: %d, got: %d)", p.MaxVotesPerBlock, len(b.Votes))
	}

	blockHash := b.Hash()
	blockSig, err := bls.DeserializeSignature(b.Signature)
	if err != nil {
		return err
	}

	randaoSig, err := bls.DeserializeSignature(b.RandaoSignature)
	if err != nil {
		return err
	}

	slotIndex := b.Header.Slot % p.EpochLength

	proposerIndex := s.ProposerQueue[slotIndex]
	proposer := s.ValidatorRegistry[proposerIndex]

	workerPub, err := bls.DeserializePublicKey(proposer.PubKey)
	if err != nil {
		return err
	}

	slotHash := chainhash.HashH([]byte(fmt.Sprintf("%d", b.Header.Slot)))

	valid, err := bls.VerifySig(workerPub, blockHash[:], blockSig)
	if err != nil {
		return err
	}
	if !valid {
		return errors.New("error validating signature for block")
	}

	valid, err = bls.VerifySig(workerPub, slotHash[:], randaoSig)
	if err != nil {
		return err
	}
	if !valid {
		return errors.New("error validating RANDAO signature for block")
	}

	for _, v := range b.Votes {
		if err := s.processVote(&v, p, proposerIndex); err != nil {
			return err
		}
	}

	h := s.Hash()
	if !h.IsEqual(&b.Header.StateRoot) {
		return fmt.Errorf("block has incorrect state root (got: %s, expected: %s)", b.Header.StateRoot, h)
	}

	return nil
}
