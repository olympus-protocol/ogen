package primitives

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

// GetVoteCommittee gets the committee for a certain block.
func (s *State) GetVoteCommittee(slot uint64, p *params.ChainParams) (min uint32, max uint32) {
	numValidatorsAtSlot := uint32(uint64(len(s.ValidatorRegistry))/p.EpochLength) + 1
	slotIndex := uint32(slot % p.EpochLength)

	min = slotIndex * numValidatorsAtSlot
	max = min + numValidatorsAtSlot

	if max >= uint32(len(s.ValidatorRegistry)) {
		max = uint32(len(s.ValidatorRegistry) - 1)
	}

	return
}

// IsDepositValid validates signatures and ensures that a deposit is valid.
func (s *State) IsDepositValid(deposit *Deposit, params *params.ChainParams) error {
	pkh := deposit.PublicKey.Hash()
	if s.UtxoState.Balances[pkh] < params.DepositAmount*params.UnitsPerCoin {
		return fmt.Errorf("balance is too low for deposit (got: %d, expected at least: %d)", s.UtxoState.Balances[pkh], params.DepositAmount*params.UnitsPerCoin)
	}

	// first validate signature
	buf := bytes.NewBuffer([]byte{})

	if err := deposit.Data.Encode(buf); err != nil {
		return err
	}

	depositHash := chainhash.HashH(buf.Bytes())

	valid, err := bls.VerifySig(&deposit.PublicKey, depositHash[:], &deposit.Signature)
	if err != nil {
		return err
	}
	if !valid {
		return errors.New("deposit signature is not valid")
	}

	validatorPubkey := deposit.Data.PublicKey.Serialize()

	// now, ensure we don't already have this validator
	for _, v := range s.ValidatorRegistry {
		if bytes.Equal(v.PubKey[:], validatorPubkey[:]) {
			return fmt.Errorf("validator already registered")
		}
	}

	pubkeyHash := chainhash.HashH(validatorPubkey[:])
	valid, err = bls.VerifySig(&deposit.Data.PublicKey, pubkeyHash[:], &deposit.Data.ProofOfPossession)
	if err != nil {
		return err
	}
	if !valid {
		return errors.New("proof-of-possession is not valid")
	}

	return nil
}

// ApplyDeposit applies a deposit to the state.
func (s *State) ApplyDeposit(deposit *Deposit, p *params.ChainParams) error {
	if err := s.IsDepositValid(deposit, p); err != nil {
		return err
	}

	pkh := deposit.PublicKey.Hash()

	s.UtxoState.Balances[pkh] -= p.DepositAmount * p.UnitsPerCoin

	s.ValidatorRegistry = append(s.ValidatorRegistry, Worker{
		Balance:      p.DepositAmount * p.UnitsPerCoin,
		PubKey:       deposit.PublicKey.Serialize(),
		PayeeAddress: deposit.Data.WithdrawalAddress,
		Status:       StatusStarting,
	})

	return nil
}

// IsVoteValid checks if a vote is valid.
func (s *State) IsVoteValid(v *MultiValidatorVote, p *params.ChainParams) error {
	if v.Data.ToEpoch == s.EpochIndex {
		if v.Data.FromEpoch != s.JustifiedEpoch {
			return fmt.Errorf("expected from epoch to match justified epoch (expected: %d, got: %d)", s.JustifiedEpoch, v.Data.FromEpoch)
		}

		if !s.JustifiedEpochHash.IsEqual(&v.Data.FromHash) {
			return fmt.Errorf("justified block hash is wrong (expected: %s, got: %s)", s.JustifiedEpochHash, v.Data.FromHash)
		}
	} else if s.EpochIndex > 0 && v.Data.ToEpoch == s.EpochIndex-1 {
		if v.Data.FromEpoch != s.PreviousJustifiedEpoch {
			return fmt.Errorf("expected from epoch to match previous justified epoch (expected: %d, got: %d)", s.PreviousJustifiedEpoch, v.Data.FromEpoch)
		}

		if !s.PreviousJustifiedEpochHash.IsEqual(&v.Data.FromHash) {
			return fmt.Errorf("previous justified block hash is wrong (expected: %s, got: %s)", s.PreviousJustifiedEpochHash, v.Data.FromHash)
		}
	} else {
		return fmt.Errorf("vote should have target epoch of either the current epoch (%d) or the previous epoch (%d) but got %d", s.EpochIndex, s.EpochIndex-1, v.Data.ToEpoch)
	}

	aggPubs := bls.NewAggregatePublicKey()
	min, max := s.GetVoteCommittee(v.Data.Slot, p)
	for i := range v.ParticipationBitfield {
		for j := 0; j < 8; j++ {
			validator := min + uint32((i*8)+j)

			if validator > max {
				break
			}

			if v.ParticipationBitfield[i]&(1<<uint(j)) == 0 {
				continue
			}

			pub, err := bls.DeserializePublicKey(s.ValidatorRegistry[validator].PubKey)
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

	err := s.IsVoteValid(v, p)
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

	voteMerkleRoot := b.VotesMerkleRoot()
	transactionMerkleRoot := b.TransactionMerkleRoot()
	depositMerkleRoot := b.DepositMerkleRoot()
	exitMerkleRoot := b.ExitMerkleRoot()

	if !b.Header.TxMerkleRoot.IsEqual(&transactionMerkleRoot) {
		return fmt.Errorf("expected transaction merkle root to be %s but got %s", transactionMerkleRoot, b.Header.TxMerkleRoot)
	}

	if !b.Header.VoteMerkleRoot.IsEqual(&voteMerkleRoot) {
		return fmt.Errorf("expected vote merkle root to be %s but got %s", voteMerkleRoot, b.Header.VoteMerkleRoot)
	}

	if !b.Header.DepositMerkleRoot.IsEqual(&depositMerkleRoot) {
		return fmt.Errorf("expected deposit merkle root to be %s but got %s", depositMerkleRoot, b.Header.DepositMerkleRoot)
	}

	if !b.Header.ExitMerkleRoot.IsEqual(&exitMerkleRoot) {
		return fmt.Errorf("expected exit merkle root to be %s but got %s", depositMerkleRoot, b.Header.DepositMerkleRoot)
	}

	if uint64(len(b.Votes)) > p.MaxVotesPerBlock {
		return fmt.Errorf("block has too many votes (max: %d, got: %d)", p.MaxVotesPerBlock, len(b.Votes))
	}

	if uint64(len(b.Txs)) > p.MaxTxsPerBlock {
		return fmt.Errorf("block has too many txs (max: %d, got: %d)", p.MaxTxsPerBlock, len(b.Txs))
	}

	if uint64(len(b.Deposits)) > p.MaxDepositsPerBlock {
		return fmt.Errorf("block has too many deposits (max: %d, got: %d)", p.MaxDepositsPerBlock, len(b.Deposits))
	}

	if uint64(len(b.Exits)) > p.MaxExitsPerBlock {
		return fmt.Errorf("block has too many exits (max: %d, got: %d)", p.MaxExitsPerBlock, len(b.Exits))
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

	for _, d := range b.Deposits {
		if err := s.ApplyDeposit(&d, p); err != nil {
			return err
		}
	}

	for _, tx := range b.Txs {
		switch p := tx.Payload.(type) {
		case *CoinPayload:
			if err := s.UtxoState.ApplyTransaction(p, b.Header.FeeAddress); err != nil {
				return err
			}
		default:
			return fmt.Errorf("payload missing from transaction")
		}
	}

	slotIndex := (b.Header.Slot + p.EpochLength - 1) % p.EpochLength

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

	for i := range s.NextRANDAO {
		s.NextRANDAO[i] ^= b.RandaoSignature[i]
	}

	return nil
}
