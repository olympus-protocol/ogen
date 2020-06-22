package primitives

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

// IsGovernanceVoteValid checks if a governance vote is valid.
func (s *State) IsGovernanceVoteValid(vote *GovernanceVote, p *params.ChainParams) error {
	if vote.VoteEpoch != s.VoteEpoch {
		return fmt.Errorf("vote not valid with vote epoch: %d (expected: %d)", vote.VoteEpoch, s.VoteEpoch)
	}
	switch vote.Type {
	case EnterVotingPeriod:
		// must be during active period
		// must have >100 POLIS
		// must have not already voted
		// must be signed by the public key
		if s.VotingState != GovernanceStateActive {
			return fmt.Errorf("cannot vote for community vote during community vote period")
		}
		if len(vote.Data) != 0 {
			return fmt.Errorf("expected EnterVotingPeriod vote to have no data")
		}
		pkh := vote.Signature.GetPublicKey().Hash()
		if s.CoinsState.Balances[pkh] < p.MinVotingBalance*p.UnitsPerCoin {
			return fmt.Errorf("minimum balance is %d, but got %d", p.MinVotingBalance, s.CoinsState.Balances[pkh]/p.UnitsPerCoin)
		}
		if !vote.Valid() {
			return fmt.Errorf("vote signature did not validate")
		}
	case VoteFor:
		// must be during voting period
		// must have >100 POLIS
		// must have not already voted
		// must be signed by the public key
		if s.VotingState != GovernanceStateVoting {
			return fmt.Errorf("cannot vote for community vote during community vote period")
		}
		if len(vote.Data) != len(p.GovernancePercentages)*20 {
			return fmt.Errorf("expected VoteFor vote to have %d bytes of data got %d", len(p.GovernancePercentages)*32, len(vote.Data))
		}
		pkh := vote.Signature.GetPublicKey().Hash()
		if s.CoinsState.Balances[pkh] < p.MinVotingBalance*p.UnitsPerCoin {
			return fmt.Errorf("minimum balance is %d, but got %d", p.MinVotingBalance, s.CoinsState.Balances[pkh]/p.UnitsPerCoin)
		}
		if _, ok := s.ReplaceVotes[pkh]; ok {
			return fmt.Errorf("found existing vote for same public key hash")
		}
		if !vote.Valid() {
			return fmt.Errorf("vote signature did not validate")
		}
	case UpdateManagersInstantly:
		// must be during active period
		// must be signed by all managers
		if s.VotingState != GovernanceStateActive {
			return fmt.Errorf("cannot vote for community vote during community vote period")
		}
		if len(vote.Data) != len(p.GovernancePercentages)*20 {
			return fmt.Errorf("expected UpdateManagersInstantly vote to have %d bytes data but got %d", len(vote.Data), len(p.GovernancePercentages)*32)
		}
		pub := vote.Signature.GetPublicKey()
		mp, ok := pub.(*bls.Multipub)
		if !ok {
			return fmt.Errorf("expected UpdateManagersInstantly to have a multisig")
		}
		if mp.NumNeeded != 5 {
			return fmt.Errorf("expected 5 signatures needed")
		}
		for i := range mp.PublicKeys {
			ih := mp.PublicKeys[i].Hash()
			if !bytes.Equal(ih[:], s.CurrentManagers[i][:]) {
				return fmt.Errorf("expected public keys to match managers")
			}
		}

		if !vote.Valid() {
			return fmt.Errorf("vote signature is not valid")
		}
	case UpdateManagersVote:
		// must be during active period
		// must be signed by 3/5 managers
		if len(vote.Data) != (len(s.CurrentManagers)+7)/8 {
			return fmt.Errorf("expected UpdateManagersVote vote to have no data")
		}
		if s.VotingState != GovernanceStateActive {
			return fmt.Errorf("cannot vote for community vote during community vote period")
		}
		// must include multisig signed by 5/5 managers
		pub := vote.Signature.GetPublicKey()
		mp, ok := pub.(*bls.Multipub)
		if !ok {
			return fmt.Errorf("expected UpdateManagersInstantly to have a multisig")
		}
		if mp.NumNeeded >= 3 {
			return fmt.Errorf("expected 3 signatures needed")
		}
		for i := range mp.PublicKeys {
			ih := mp.PublicKeys[i].Hash()
			if !bytes.Equal(ih[:], s.CurrentManagers[i][:]) {
				return fmt.Errorf("expected public keys to match managers")
			}
		}

		if !vote.Valid() {
			return fmt.Errorf("vote signature is not valid")
		}
	default:
		return fmt.Errorf("unknown vote type")
	}

	return nil
}

// ProcessGovernanceVote processes governance votes.
func (s *State) ProcessGovernanceVote(vote *GovernanceVote, p *params.ChainParams) error {
	if err := s.IsGovernanceVoteValid(vote, p); err != nil {
		return err
	}

	votePub := vote.Signature.GetPublicKey()

	switch vote.Type {
	case EnterVotingPeriod:
		s.ReplaceVotes[votePub.Hash()] = chainhash.Hash{}
		// we check if it's above the threshold every few epochs, but not here
	case VoteFor:
		voteData := CommunityVoteData{
			ReplacementCandidates: make([][]byte, len(p.GovernancePercentages)),
		}

		for i := range voteData.ReplacementCandidates {
			copy(voteData.ReplacementCandidates[i][:], vote.Data[i*20:(i+1)*20])
		}

		voteHash, err := voteData.Hash()
		if err != nil {
			return err
		}
		s.CommunityVotes[voteHash] = voteData
		s.ReplaceVotes[votePub.Hash()] = voteHash
	case UpdateManagersInstantly:
		for i := range s.CurrentManagers {
			copy(s.CurrentManagers[i][:], vote.Data[i*20:(i+1)*20])
		}
		s.NextVoteEpoch(GovernanceStateActive)
	case UpdateManagersVote:
		s.NextVoteEpoch(GovernanceStateVoting)
	default:
		return fmt.Errorf("unknown vote type")
	}

	return nil
}

// ApplyTransactionSingle applies a transaction to the coin state.
func (s *State) ApplyTransactionSingle(tx *TransferSinglePayload, blockWithdrawalAddress [20]byte, p *params.ChainParams) error {
	u := s.CoinsState
	pkh := tx.FromPubkeyHash()
	if u.Balances[pkh] < tx.Amount+tx.Fee {
		return fmt.Errorf("insufficient balance of %d for %d transaction", u.Balances[pkh], tx.Amount)
	}

	if u.Nonces[pkh] >= tx.Nonce {
		return fmt.Errorf("nonce is too small (already processed: %d, trying: %d)", u.Nonces[pkh], tx.Nonce)
	}

	if err := tx.VerifySig(); err != nil {
		return err
	}

	u.Balances[pkh] -= tx.Amount + tx.Fee
	u.Balances[tx.To] += tx.Amount
	u.Balances[blockWithdrawalAddress] += tx.Fee
	u.Nonces[pkh] = tx.Nonce

	if _, ok := s.ReplaceVotes[pkh]; u.Balances[pkh] < p.UnitsPerCoin*p.MinVotingBalance && ok {
		delete(s.ReplaceVotes, pkh)
	}

	return nil
}

// ApplyTransactionMulti applies a multisig transaction to the coin state.
func (s *State) ApplyTransactionMulti(tx *TransferMultiPayload, blockWithdrawalAddress [20]byte, p *params.ChainParams) error {
	u := s.CoinsState
	pkh := tx.FromPubkeyHash()
	if u.Balances[pkh] < tx.Amount+tx.Fee {
		return fmt.Errorf("insufficient balance of %d for %d transaction", u.Balances[pkh], tx.Amount)
	}

	if u.Nonces[pkh] >= tx.Nonce {
		return fmt.Errorf("nonce is too small (already processed: %d, trying: %d)", u.Nonces[pkh], tx.Nonce)
	}

	if err := tx.VerifySig(); err != nil {
		return err
	}

	u.Balances[pkh] -= tx.Amount + tx.Fee
	u.Balances[tx.To] += tx.Amount
	u.Balances[blockWithdrawalAddress] += tx.Fee
	u.Nonces[pkh] = tx.Nonce

	if _, ok := s.ReplaceVotes[pkh]; u.Balances[pkh] < p.UnitsPerCoin*p.MinVotingBalance && ok {
		delete(s.ReplaceVotes, pkh)
	}

	return nil
}

// IsProposerSlashingValid checks if a given proposer slashing is valid.
func (s *State) IsProposerSlashingValid(ps *ProposerSlashing) (uint32, error) {
	h1, err := ps.BlockHeader1.Hash()
	if err != nil {
		return 0, err
	}
	h2, err := ps.BlockHeader2.Hash()
	if err != nil {
		return 0, err
	}
	if h1.IsEqual(&h2) {
		return 0, fmt.Errorf("proposer-slashing: block headers are equal")
	}

	if ps.BlockHeader1.Slot != ps.BlockHeader2.Slot {
		return 0, fmt.Errorf("proposer-slashing: block headers do not have the same slot")
	}

	if !ps.Signature1.Verify(h1[:], &ps.ValidatorPublicKey) {
		return 0, fmt.Errorf("proposer-slashing: signature does not validate for block header 1")
	}

	if !ps.Signature2.Verify(h2[:], &ps.ValidatorPublicKey) {
		return 0, fmt.Errorf("proposer-slashing: signature does not validate for block header 2")
	}

	pubkeyBytes, err := ps.ValidatorPublicKey.Marshal()
	if err != nil {
		return 0, err
	}
	proposerIndex := -1
	for i, v := range s.ValidatorRegistry {
		if bytes.Equal(v.PubKey, pubkeyBytes) {
			proposerIndex = i
		}
	}

	if proposerIndex < 0 {
		return 0, fmt.Errorf("proposer-slashing: validator is already exited")
	}

	return uint32(proposerIndex), nil
}

// ApplyProposerSlashing applies the proposer slashing to the state.
func (s *State) ApplyProposerSlashing(ps *ProposerSlashing, p *params.ChainParams) error {
	proposerIndex, err := s.IsProposerSlashingValid(ps)
	if err != nil {
		return err
	}

	return s.UpdateValidatorStatus(proposerIndex, StatusExitedWithPenalty, p)
}

// IsVoteSlashingValid checks if the vote slashing is valid.
func (s *State) IsVoteSlashingValid(vs *VoteSlashing, p *params.ChainParams) ([]uint32, error) {
	if vs.Vote1.Data.Equals(&vs.Vote2.Data) {
		return nil, fmt.Errorf("vote-slashing: votes are not distinct")
	}

	if !vs.Vote1.Data.IsDoubleVote(vs.Vote2.Data) && !vs.Vote1.Data.IsSurroundVote(vs.Vote2.Data) {
		return nil, fmt.Errorf("vote-slashing: votes do not violate slashing rule")
	}

	voteParticipation1 := vs.Vote1.ParticipationBitfield
	voteParticipation2 := vs.Vote2.ParticipationBitfield

	voteCommittee1 := make(map[uint32]struct{})
	common := make([]uint32, 0)

	validators1, err := s.GetVoteCommittee(vs.Vote1.Data.Slot, p)
	if err != nil {
		return nil, err
	}

	validators2, err := s.GetVoteCommittee(vs.Vote2.Data.Slot, p)
	if err != nil {
		return nil, err
	}

	aggPubs1 := make([]*bls.PublicKey, 0)
	aggPubs2 := make([]*bls.PublicKey, 0)

	for i := range voteParticipation1 {
		for j := 0; j < 8; j++ {
			validator := uint32((i * 8) + j)

			if validator >= uint32(len(validators1)) {
				break
			}

			if voteParticipation1[i]&(1<<uint(j)) == 0 {
				continue
			}

			validatorIdx := validators1[validator]

			voteCommittee1[validatorIdx] = struct{}{}

			pub, err := bls.PublicKeyFromBytes(s.ValidatorRegistry[validatorIdx].PubKey)
			if err != nil {
				return nil, err
			}
			aggPubs1 = append(aggPubs1, pub)
		}
	}

	if !vs.Vote1.Signature.FastAggregateVerify(aggPubs1, vs.Vote1.Data.Hash()) {
		return nil, fmt.Errorf("vote-slashing: vote 1 does not validate")
	}

	for i := range voteParticipation2 {
		for j := 0; j < 8; j++ {
			validator := uint32((i * 8) + j)

			if validator >= uint32(len(validators2)) {
				break
			}

			if voteParticipation2[i]&(1<<uint(j)) == 0 {
				continue
			}

			validatorIdx := validators2[validator]

			if _, ok := voteCommittee1[validatorIdx]; ok {
				common = append(common, validatorIdx)
			}

			pub, err := bls.PublicKeyFromBytes(s.ValidatorRegistry[validatorIdx].PubKey)
			if err != nil {
				return nil, err
			}
			aggPubs2 = append(aggPubs2, pub)
		}
	}

	if len(common) == 0 {
		return nil, fmt.Errorf("vote-slashing: votes do not contain any common validators")
	}

	if !vs.Vote2.Signature.FastAggregateVerify(aggPubs2, vs.Vote2.Data.Hash()) {
		return nil, fmt.Errorf("vote-slashing: vote 2 does not validate")
	}

	return common, nil
}

// ApplyVoteSlashing applies a vote slashing to the state.
func (s *State) ApplyVoteSlashing(vs *VoteSlashing, p *params.ChainParams) error {
	common, err := s.IsVoteSlashingValid(vs, p)
	if err != nil {
		return err
	}

	for _, v := range common {
		if err := s.UpdateValidatorStatus(v, StatusExitedWithPenalty, p); err != nil {
			return err
		}
	}

	return nil
}

// IsRANDAOSlashingValid checks if the RANDAO slashing is valid.
func (s *State) IsRANDAOSlashingValid(rs *RANDAOSlashing) (uint32, error) {
	if rs.Slot >= s.Slot {
		return 0, fmt.Errorf("randao-slashing: RANDAO was already assumed to be revealed")
	}

	slotHash := chainhash.HashH([]byte(fmt.Sprintf("%d", rs.Slot)))

	if !rs.RandaoReveal.Verify(slotHash[:], &rs.ValidatorPubkey) {
		return 0, fmt.Errorf("randao-slashing: RANDAO reveal does not verify")
	}

	pubkeyBytes, err := rs.ValidatorPubkey.Marshal()
	if err != nil {
		return 0, err
	}
	proposerIndex := -1
	for i, v := range s.ValidatorRegistry {
		if bytes.Equal(v.PubKey, pubkeyBytes) {
			proposerIndex = i
		}
	}

	if proposerIndex < 0 {
		return 0, fmt.Errorf("proposer-slashing: validator is already exited")
	}

	return uint32(proposerIndex), nil
}

// ApplyRANDAOSlashing applies the RANDAO slashing to the state.
func (s *State) ApplyRANDAOSlashing(rs *RANDAOSlashing, p *params.ChainParams) error {
	proposer, err := s.IsRANDAOSlashingValid(rs)
	if err != nil {
		return err
	}

	return s.UpdateValidatorStatus(proposer, StatusExitedWithPenalty, p)
}

// GetVoteCommittee gets the committee for a certain block.
func (s *State) GetVoteCommittee(slot uint64, p *params.ChainParams) ([]uint32, error) {
	if (slot-1)/p.EpochLength == s.EpochIndex {
		assignments := s.CurrentEpochVoteAssignments
		slotIndex := uint64(slot % p.EpochLength)
		min := (slotIndex * uint64(len(assignments))) / p.EpochLength
		max := ((slotIndex + 1) * uint64(len(assignments))) / p.EpochLength

		return assignments[min:max], nil
	} else if (slot-1)/p.EpochLength == s.EpochIndex-1 {
		assignments := s.PreviousEpochVoteAssignments
		slotIndex := uint64(slot % p.EpochLength)
		min := (slotIndex * uint64(len(assignments))) / p.EpochLength
		max := ((slotIndex + 1) * uint64(len(assignments))) / p.EpochLength

		return assignments[min:max], nil
	}

	// TODO: better handling
	return nil, fmt.Errorf("tried to get vote committee out of range: %d", slot)
}

// IsExitValid checks if an exit is valid.
func (s *State) IsExitValid(exit *Exit) error {
	exitPub, err := exit.ValidatorPubkey.Marshal()
	if err != nil {
		return err
	}
	msg := fmt.Sprintf("exit %x", exitPub)
	msgHash := chainhash.HashH([]byte(msg))

	valid := exit.Signature.Verify(msgHash[:], &exit.WithdrawPubkey)
	if !valid {
		return fmt.Errorf("exit signature is not valid")
	}

	pkh := exit.WithdrawPubkey.Hash()

	pubkeySerialized, err := exit.ValidatorPubkey.Marshal()
	if err != nil {
		return err
	}
	foundActiveValidator := false
	for _, v := range s.ValidatorRegistry {
		if bytes.Equal(v.PubKey[:], pubkeySerialized[:]) && v.IsActive() {
			if !bytes.Equal(v.PayeeAddress[:], pkh[:]) {
				return fmt.Errorf("withdraw pubkey does not match withdraw address (expected: %x, got: %x)", pkh, v.PayeeAddress)
			}

			foundActiveValidator = true
		}
	}

	if !foundActiveValidator {
		return fmt.Errorf("could not find active validator with pubkey: %x", pubkeySerialized[:])
	}

	return nil
}

// ApplyExit processes an exit request.
func (s *State) ApplyExit(exit *Exit) error {
	if err := s.IsExitValid(exit); err != nil {
		return err
	}

	pubkeySerialized, err := exit.ValidatorPubkey.Marshal()
	if err != nil {
		return err
	}
	for i, v := range s.ValidatorRegistry {
		if bytes.Equal(v.PubKey[:], pubkeySerialized[:]) && v.IsActive() {
			s.ValidatorRegistry[i].Status = StatusActivePendingExit
			s.ValidatorRegistry[i].LastActiveEpoch = s.EpochIndex + 2
		}
	}

	return nil
}

// IsDepositValid validates signatures and ensures that a deposit is valid.
func (s *State) IsDepositValid(deposit *Deposit, params *params.ChainParams) error {
	pkh := deposit.PublicKey.Hash()
	if s.CoinsState.Balances[pkh] < params.DepositAmount*params.UnitsPerCoin {
		return fmt.Errorf("balance is too low for deposit (got: %d, expected at least: %d)", s.CoinsState.Balances[pkh], params.DepositAmount*params.UnitsPerCoin)
	}

	depositHash, err := deposit.Hash()
	if err != nil {
		return err
	}
	valid := deposit.Signature.Verify(depositHash[:], &deposit.PublicKey)
	if !valid {
		return errors.New("deposit signature is not valid")
	}

	validatorPubkey, err := deposit.Data.PublicKey.Marshal()
	if err != nil {
		return err
	}
	// now, ensure we don't already have this validator
	for _, v := range s.ValidatorRegistry {
		if bytes.Equal(v.PubKey[:], validatorPubkey[:]) {
			return fmt.Errorf("validator already registered")
		}
	}

	pubkeyHash := chainhash.HashH(validatorPubkey[:])
	valid = deposit.Data.ProofOfPossession.Verify(pubkeyHash[:], &deposit.Data.PublicKey)
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

	s.CoinsState.Balances[pkh] -= p.DepositAmount * p.UnitsPerCoin
	pubKey, err := deposit.Data.PublicKey.Marshal()
	if err != nil {
		return err
	}
	s.ValidatorRegistry = append(s.ValidatorRegistry, Validator{
		Balance:          p.DepositAmount * p.UnitsPerCoin,
		PubKey:           pubKey,
		PayeeAddress:     deposit.Data.WithdrawalAddress,
		Status:           StatusStarting,
		FirstActiveEpoch: (s.EpochIndex + 2),
		LastActiveEpoch:  0,
	})

	return nil
}

// IsVoteValid checks if a vote is valid.
func (s *State) IsVoteValid(v *MultiValidatorVote, p *params.ChainParams) error {
	if v.Data.Slot == 0 {
		return fmt.Errorf("slot out of range")
	}
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

	aggPubs := make([]*bls.PublicKey, 0)
	validators, err := s.GetVoteCommittee(v.Data.Slot, p)
	if err != nil {
		return err
	}
	for i := range v.ParticipationBitfield {
		for j := 0; j < 8; j++ {
			validator := uint32((i * 8) + j)

			if validator >= uint32(len(validators)) {
				break
			}

			if v.ParticipationBitfield[i]&(1<<uint(j)) == 0 {
				continue
			}

			validatorIdx := validators[validator]

			pub, err := bls.PublicKeyFromBytes(s.ValidatorRegistry[validatorIdx].PubKey)
			if err != nil {
				return err
			}
			aggPubs = append(aggPubs, pub)
		}
	}

	h := v.Data.Hash()
	valid := v.Signature.FastAggregateVerify(aggPubs, h)
	if !valid {
		return fmt.Errorf("aggregate signature did not validate")
	}

	return nil
}

func (s *State) ProcessVote(v *MultiValidatorVote, p *params.ChainParams, proposerIndex uint32) error {
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

// GetProposerPublicKey gets the public key for the proposer of a block.
func (s *State) GetProposerPublicKey(b *Block, p *params.ChainParams) (*bls.PublicKey, error) {
	slotIndex := (b.Header.Slot + p.EpochLength - 1) % p.EpochLength

	proposerIndex := s.ProposerQueue[slotIndex]
	proposer := s.ValidatorRegistry[proposerIndex]

	return bls.PublicKeyFromBytes(proposer.PubKey)
}

// CheckBlockSignature checks the block signature.
func (s *State) CheckBlockSignature(b *Block, p *params.ChainParams) error {
	blockHash, err := b.Hash()
	if err != nil {
		return err
	}
	blockSig, err := bls.SignatureFromBytes(b.Signature)
	if err != nil {
		return err
	}

	validatorPub, err := s.GetProposerPublicKey(b, p)
	if err != nil {
		return err
	}

	randaoSig, err := bls.SignatureFromBytes(b.RandaoSignature)
	if err != nil {
		return err
	}

	valid := blockSig.Verify(blockHash[:], validatorPub)
	if !valid {
		return errors.New("error validating signature for block")
	}

	slotHash := chainhash.HashH([]byte(fmt.Sprintf("%d", b.Header.Slot)))

	valid = randaoSig.Verify(slotHash[:], validatorPub)
	if !valid {
		return errors.New("error validating RANDAO signature for block")
	}

	return nil
}

// ProcessBlock runs a block transition on the state and mutates state.
func (s *State) ProcessBlock(b *Block, p *params.ChainParams) error {
	if b.Header.Slot != s.Slot {
		return fmt.Errorf("state is not updated to slot %d, instead got %d", b.Header.Slot, s.Slot)
	}

	if err := s.CheckBlockSignature(b, p); err != nil {
		return err
	}

	voteMerkleRoot := b.VotesMerkleRoot()
	transactionMerkleRoot := b.TransactionMerkleRoot()
	depositMerkleRoot, err := b.DepositMerkleRoot()
	if err != nil {
		return err
	}
	exitMerkleRoot, err := b.ExitMerkleRoot()
	if err != nil {
		return err
	}
	voteSlashingMerkleRoot, err := b.VoteSlashingRoot()
	if err != nil {
		return err
	}
	proposerSlashingMerkleRoot, err := b.ProposerSlashingsRoot()
	if err != nil {
		return err
	}
	randaoSlashingMerkleRoot, err := b.RANDAOSlashingsRoot()
	if err != nil {
		return err
	}
	governanceVoteMerkleRoot, err := b.GovernanceVoteMerkleRoot()
	if err != nil {
		return err
	}
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
		return fmt.Errorf("expected exit merkle root to be %s but got %s", exitMerkleRoot, b.Header.ExitMerkleRoot)
	}

	if !b.Header.VoteSlashingMerkleRoot.IsEqual(&voteSlashingMerkleRoot) {
		return fmt.Errorf("expected exit merkle root to be %s but got %s", voteSlashingMerkleRoot, b.Header.VoteSlashingMerkleRoot)
	}

	if !b.Header.ProposerSlashingMerkleRoot.IsEqual(&proposerSlashingMerkleRoot) {
		return fmt.Errorf("expected exit merkle root to be %s but got %s", proposerSlashingMerkleRoot, b.Header.ProposerSlashingMerkleRoot)
	}

	if !b.Header.RANDAOSlashingMerkleRoot.IsEqual(&randaoSlashingMerkleRoot) {
		return fmt.Errorf("expected exit merkle root to be %s but got %s", randaoSlashingMerkleRoot, b.Header.RANDAOSlashingMerkleRoot)
	}

	if !b.Header.GovernanceVotesMerkleRoot.IsEqual(&governanceVoteMerkleRoot) {
		return fmt.Errorf("expected exit merkle root to be %s but got %s", governanceVoteMerkleRoot, b.Header.GovernanceVotesMerkleRoot)
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

	if uint64(len(b.RANDAOSlashings)) > p.MaxRANDAOSlashingsPerBlock {
		return fmt.Errorf("block has too many RANDAO slashings (max: %d, got: %d)", p.MaxRANDAOSlashingsPerBlock, len(b.RANDAOSlashings))
	}

	if uint64(len(b.VoteSlashings)) > p.MaxVoteSlashingsPerBlock {
		return fmt.Errorf("block has too many vote slashings (max: %d, got: %d)", p.MaxVoteSlashingsPerBlock, len(b.VoteSlashings))
	}

	if uint64(len(b.ProposerSlashings)) > p.MaxProposerSlashingsPerBlock {
		return fmt.Errorf("block has too many proposer slashings (max: %d, got: %d)", p.MaxProposerSlashingsPerBlock, len(b.ProposerSlashings))
	}

	for _, d := range b.Deposits {
		if err := s.ApplyDeposit(&d, p); err != nil {
			return err
		}
	}

	for _, tx := range b.Txs {
		switch payload := tx.Payload.(type) {
		case *TransferSinglePayload:
			if err := s.ApplyTransactionSingle(payload, b.Header.FeeAddress, p); err != nil {
				return err
			}
		case *TransferMultiPayload:
			if err := s.ApplyTransactionMulti(payload, b.Header.FeeAddress, p); err != nil {
				return err
			}
		default:
			return fmt.Errorf("payload missing from transaction")
		}
	}

	for _, vote := range b.GovernanceVotes {
		if err := s.ProcessGovernanceVote(&vote, p); err != nil {
			return err
		}
	}

	slotIndex := (b.Header.Slot + p.EpochLength - 1) % p.EpochLength

	proposerIndex := s.ProposerQueue[slotIndex]

	for _, v := range b.Votes {
		if err := s.ProcessVote(&v, p, proposerIndex); err != nil {
			return err
		}
	}

	for _, e := range b.Exits {
		if err := s.ApplyExit(&e); err != nil {
			return err
		}
	}

	for _, rs := range b.RANDAOSlashings {
		if err := s.ApplyRANDAOSlashing(&rs, p); err != nil {
			return err
		}
	}

	for _, vs := range b.VoteSlashings {
		if err := s.ApplyVoteSlashing(&vs, p); err != nil {
			return err
		}
	}

	for _, ps := range b.ProposerSlashings {
		if err := s.ApplyProposerSlashing(&ps, p); err != nil {
			return err
		}
	}

	for i := range s.NextRANDAO {
		s.NextRANDAO[i] ^= b.RandaoSignature[i]
	}

	return nil
}
