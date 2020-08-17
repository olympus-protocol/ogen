package state

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/olympus-protocol/ogen/pkg/bitfield"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/params"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"reflect"
)

// IsGovernanceVoteValid checks if a governance vote is valid.
func (s *state) IsGovernanceVoteValid(vote *primitives.GovernanceVote, p *params.ChainParams) error {
	if vote.VoteEpoch != s.VoteEpoch {
		return fmt.Errorf("vote not valid with vote epoch: %d (expected: %d)", vote.VoteEpoch, s.VoteEpoch)
	}
	switch vote.Type {
	case primitives.EnterVotingPeriod:
		// must be during active period
		// must have >100 POLIS
		// must have not already voted
		// must be signed by the public key
		if s.VotingState != GovernanceStateActive {
			return fmt.Errorf("cannot vote for community vote during community vote period")
		}
		sig := vote.CombinedSig
		pubKey, err := sig.Pub()
		if err != nil {
			return err
		}
		pkh, err := pubKey.Hash()
		if err != nil {
			return err
		}
		if s.CoinsState.Balances[pkh] < p.MinVotingBalance*p.UnitsPerCoin {
			return fmt.Errorf("minimum balance is %d, but got %d", p.MinVotingBalance, s.CoinsState.Balances[pkh]/p.UnitsPerCoin)
		}
		if !vote.ValidCombined() {
			return fmt.Errorf("vote signature did not validate")
		}
	case primitives.VoteFor:
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
		sig := vote.CombinedSig
		pubKey, err := sig.Pub()
		if err != nil {
			return err
		}
		pkh, err := pubKey.Hash()
		if err != nil {
			return err
		}
		if s.CoinsState.Balances[pkh] < p.MinVotingBalance*p.UnitsPerCoin {
			return fmt.Errorf("minimum balance is %d, but got %d", p.MinVotingBalance, s.CoinsState.Balances[pkh]/p.UnitsPerCoin)
		}
		if _, ok := s.Governance.ReplaceVotes[pkh]; ok {
			return fmt.Errorf("found existing vote for same public key hash")
		}
		if !vote.ValidCombined() {
			return fmt.Errorf("vote signature did not validate")
		}
	case primitives.UpdateManagersInstantly:
		// must be during active period
		// must be signed by all managers
		if s.VotingState != GovernanceStateActive {
			return fmt.Errorf("cannot vote for community vote during community vote period")
		}
		if len(vote.Data) != len(p.GovernancePercentages)*20 {
			return fmt.Errorf("expected UpdateManagersInstantly vote to have %d bytes data but got %d", len(vote.Data), len(p.GovernancePercentages)*32)
		}
		sig := vote.Multisig
		pub, err := sig.GetPublicKey()
		if err != nil {
			return err
		}
		if pub.NumNeeded != 5 {
			return fmt.Errorf("expected 5 signatures needed")
		}
		for i := range pub.PublicKeys {
			pub, err := bls.PublicKeyFromBytes(pub.PublicKeys[i])
			if err != nil {
				return err
			}
			ih, err := pub.Hash()
			if err != nil {
				return err
			}
			if !bytes.Equal(ih[:], s.CurrentManagers[i][:]) {
				return fmt.Errorf("expected public keys to match managers")
			}
		}

		if !vote.ValidMultisig() {
			return fmt.Errorf("vote signature is not valid")
		}
	case primitives.UpdateManagersVote:
		// must be during active period
		// must be signed by 3/5 managers
		if len(vote.Data) != (len(s.CurrentManagers)+7)/8 {
			return fmt.Errorf("expected UpdateManagersVote vote to have no data")
		}
		if s.VotingState != GovernanceStateActive {
			return fmt.Errorf("cannot vote for community vote during community vote period")
		}
		sig := vote.Multisig
		// must include multisig signed by 5/5 managers
		pub, err := sig.GetPublicKey()
		if err != nil {
			return err
		}
		if pub.NumNeeded >= 3 {
			return fmt.Errorf("expected 3 signatures needed")
		}
		for i := range pub.PublicKeys {
			pub, err := bls.PublicKeyFromBytes(pub.PublicKeys[i])
			if err != nil {
				return err
			}
			ih, err := pub.Hash()
			if err != nil {
				return err
			}
			if !bytes.Equal(ih[:], s.CurrentManagers[i][:]) {
				return fmt.Errorf("expected public keys to match managers")
			}
		}

		if !vote.ValidMultisig() {
			return fmt.Errorf("vote signature is not valid")
		}
	default:
		return fmt.Errorf("unknown vote type")
	}

	return nil
}

// ProcessGovernanceVote processes governance votes.
func (s *state) ProcessGovernanceVote(vote *primitives.GovernanceVote, p *params.ChainParams) error {
	if err := s.IsGovernanceVoteValid(vote, p); err != nil {
		return err
	}
	sig := vote.CombinedSig
	votePub, err := sig.Pub()
	if err != nil {
		return err
	}
	hash, err := votePub.Hash()
	if err != nil {
		return err
	}
	switch vote.Type {
	case primitives.EnterVotingPeriod:
		s.Governance.ReplaceVotes[hash] = chainhash.Hash{}
		// we check if it's above the threshold every few epochs, but not here
	case primitives.VoteFor:
		voteData := primitives.CommunityVoteData{
			ReplacementCandidates: [][20]byte{},
		}

		for i := range voteData.ReplacementCandidates {
			copy(voteData.ReplacementCandidates[i][:], vote.Data[i*20:(i+1)*20])
		}

		voteHash := voteData.Hash()
		s.Governance.CommunityVotes[voteHash] = voteData
		s.Governance.ReplaceVotes[hash] = voteHash
	case primitives.UpdateManagersInstantly:
		for i := range s.CurrentManagers {
			copy(s.CurrentManagers[i][:], vote.Data[i*20:(i+1)*20])
		}
		s.NextVoteEpoch(GovernanceStateActive)
	case primitives.UpdateManagersVote:
		s.NextVoteEpoch(GovernanceStateVoting)
	default:
		return fmt.Errorf("unknown vote type")
	}

	return nil
}

// ApplyTransactionSingle applies a transaction to the coin state.
func (s *state) ApplyTransactionSingle(tx *primitives.Tx, blockWithdrawalAddress [20]byte, p *params.ChainParams) error {
	u := s.CoinsState
	pkh, err := tx.FromPubkeyHash()
	if err != nil {
		return err
	}
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

	if _, ok := s.Governance.ReplaceVotes[pkh]; u.Balances[pkh] < p.UnitsPerCoin*p.MinVotingBalance && ok {
		delete(s.Governance.ReplaceVotes, pkh)
	}

	return nil
}

// ApplyTransactionMulti applies a multisig transaction to the coin state.
// func (s *State) ApplyTransactionMulti(tx *TransferMultiPayload, blockWithdrawalAddress [20]byte, p *params.ChainParams) error {
// 	u := s.CoinsState
// 	pkh, err := tx.FromPubkeyHash()
// 	if err != nil {
// 		return err
// 	}
// 	if u.Balances[pkh] < tx.Amount+tx.Fee {
// 		return fmt.Errorf("insufficient balance of %d for %d transaction", u.Balances[pkh], tx.Amount)
// 	}

// 	if u.Nonces[pkh] >= tx.Nonce {
// 		return fmt.Errorf("nonce is too small (already processed: %d, trying: %d)", u.Nonces[pkh], tx.Nonce)
// 	}

// 	if err := tx.VerifySig(); err != nil {
// 		return err
// 	}

// 	u.Balances[pkh] -= tx.Amount + tx.Fee
// 	u.Balances[tx.To] += tx.Amount
// 	u.Balances[blockWithdrawalAddress] += tx.Fee
// 	u.Nonces[pkh] = tx.Nonce

// 	if _, ok := s.Governance.ReplaceVotes[pkh]; u.Balances[pkh] < p.UnitsPerCoin*p.MinVotingBalance && ok {
// 		delete(s.Governance.ReplaceVotes, pkh)
// 	}

// 	return nil
// }

// IsProposerSlashingValid checks if a given proposer slashing is valid.
func (s *state) IsProposerSlashingValid(ps *primitives.ProposerSlashing) (uint64, error) {
	h1 := ps.BlockHeader1.Hash()
	h2 := ps.BlockHeader2.Hash()

	if h1.IsEqual(&h2) {
		return 0, fmt.Errorf("proposer-slashing: block headers are equal")
	}

	if ps.BlockHeader1.Slot != ps.BlockHeader2.Slot {
		return 0, fmt.Errorf("proposer-slashing: block headers do not have the same slot")
	}
	pub, err := ps.GetValidatorPubkey()
	if err != nil {
		return 0, err
	}
	s1, err := ps.GetSignature1()
	if err != nil {
		return 0, err
	}
	s2, err := ps.GetSignature2()
	if err != nil {
		return 0, err
	}
	if !s1.Verify(h1[:], pub) {
		return 0, fmt.Errorf("proposer-slashing: signature does not validate for block header 1")
	}

	if !s2.Verify(h2[:], pub) {
		return 0, fmt.Errorf("proposer-slashing: signature does not validate for block header 2")
	}

	pubkeyBytes := ps.ValidatorPublicKey

	proposerIndex := -1
	for i, v := range s.ValidatorRegistry {
		if bytes.Equal(v.PubKey[:], pubkeyBytes[:]) {
			proposerIndex = i
		}
	}

	if proposerIndex < 0 {
		return 0, fmt.Errorf("proposer-slashing: validator is already exited")
	}

	return uint64(proposerIndex), nil
}

// ApplyProposerSlashing applies the proposer slashing to the state.
func (s *state) ApplyProposerSlashing(ps *primitives.ProposerSlashing, p *params.ChainParams) error {
	proposerIndex, err := s.IsProposerSlashingValid(ps)
	if err != nil {
		return err
	}

	return s.UpdateValidatorStatus(proposerIndex, primitives.StatusExitedWithPenalty, p)
}

// IsVoteSlashingValid checks if the vote slashing is valid.
func (s *state) IsVoteSlashingValid(vs *primitives.VoteSlashing, p *params.ChainParams) ([]uint64, error) {
	if vs.Vote1.Data.Equals(vs.Vote2.Data) {
		return nil, fmt.Errorf("vote-slashing: votes are not distinct")
	}

	if !vs.Vote1.Data.IsDoubleVote(vs.Vote2.Data) && !vs.Vote1.Data.IsSurroundVote(vs.Vote2.Data) {
		return nil, fmt.Errorf("vote-slashing: votes do not violate slashing rule")
	}

	common := make([]uint64, 0)
	voteCommittee1 := make(map[uint64]struct{})

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

	for i, idx := range validators1 {

		if !vs.Vote1.ParticipationBitfield.Get(uint(i)) {
			continue
		}

		voteCommittee1[idx] = struct{}{}

		pub, err := bls.PublicKeyFromBytes(s.ValidatorRegistry[idx].PubKey)
		if err != nil {
			return nil, err
		}
		aggPubs1 = append(aggPubs1, pub)
	}

	for i, idx := range validators2 {
		if !vs.Vote2.ParticipationBitfield.Get(uint(i)) {
			continue
		}

		if _, ok := voteCommittee1[idx]; ok {
			common = append(common, idx)
		}

		pub, err := bls.PublicKeyFromBytes(s.ValidatorRegistry[idx].PubKey)
		if err != nil {
			return nil, err
		}
		aggPubs2 = append(aggPubs2, pub)
	}

	v1Sig, err := vs.Vote1.Signature()
	if err != nil {
		return nil, err
	}
	if !v1Sig.FastAggregateVerify(aggPubs1, vs.Vote1.Data.Hash()) {
		return nil, fmt.Errorf("vote-slashing: vote 1 does not validate")
	}

	if len(common) == 0 {
		return nil, fmt.Errorf("vote-slashing: votes do not contain any common validators")
	}

	v2Sig, err := vs.Vote2.Signature()
	if err != nil {
		return nil, err
	}
	if !v2Sig.FastAggregateVerify(aggPubs2, vs.Vote2.Data.Hash()) {
		return nil, fmt.Errorf("vote-slashing: vote 2 does not validate")
	}

	return common, nil
}

// ApplyVoteSlashing applies a vote slashing to the state.
func (s *state) ApplyVoteSlashing(vs *primitives.VoteSlashing, p *params.ChainParams) error {
	common, err := s.IsVoteSlashingValid(vs, p)
	if err != nil {
		return err
	}

	for _, v := range common {
		if err := s.UpdateValidatorStatus(v, primitives.StatusExitedWithPenalty, p); err != nil {
			return err
		}
	}

	return nil
}

// IsRANDAOSlashingValid checks if the RANDAO slashing is valid.
func (s *state) IsRANDAOSlashingValid(rs *primitives.RANDAOSlashing) (uint64, error) {
	if rs.Slot >= s.Slot {
		return 0, fmt.Errorf("randao-slashing: RANDAO was already assumed to be revealed")
	}

	slotHash := chainhash.HashH([]byte(fmt.Sprintf("%d", rs.Slot)))
	pub, err := rs.GetValidatorPubkey()
	if err != nil {
		return 0, err
	}
	sig, err := rs.GetRandaoReveal()
	if err != nil {
		return 0, err
	}
	if !sig.Verify(slotHash[:], pub) {
		return 0, fmt.Errorf("randao-slashing: RANDAO reveal does not verify")
	}

	pubkeyBytes := rs.ValidatorPubkey

	proposerIndex := -1
	for i, v := range s.ValidatorRegistry {
		if bytes.Equal(v.PubKey[:], pubkeyBytes[:]) {
			proposerIndex = i
		}
	}

	if proposerIndex < 0 {
		return 0, fmt.Errorf("proposer-slashing: validator is already exited")
	}

	return uint64(proposerIndex), nil
}

// ApplyRANDAOSlashing applies the RANDAO slashing to the state.
func (s *state) ApplyRANDAOSlashing(rs *primitives.RANDAOSlashing, p *params.ChainParams) error {
	proposer, err := s.IsRANDAOSlashingValid(rs)
	if err != nil {
		return err
	}

	return s.UpdateValidatorStatus(proposer, primitives.StatusExitedWithPenalty, p)
}

// GetVoteCommittee gets the committee for a certain block.
func (s *state) GetVoteCommittee(slot uint64, p *params.ChainParams) ([]uint64, error) {

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
func (s *state) IsExitValid(exit *primitives.Exit) error {
	msg := fmt.Sprintf("exit %x", exit.ValidatorPubkey)
	msgHash := chainhash.HashH([]byte(msg))
	wPubKey, err := exit.GetWithdrawPubKey()
	if err != nil {
		return err
	}
	sig, err := exit.GetSignature()
	if err != nil {
		return err
	}
	valid := sig.Verify(msgHash[:], wPubKey)
	if !valid {
		return fmt.Errorf("exit signature is not valid")
	}

	pkh, err := wPubKey.Hash()
	if err != nil {
		return err
	}
	pubkeySerialized := exit.ValidatorPubkey

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
func (s *state) ApplyExit(exit *primitives.Exit) error {
	if err := s.IsExitValid(exit); err != nil {
		return err
	}

	pubkeySerialized := exit.ValidatorPubkey

	for i, v := range s.ValidatorRegistry {
		if bytes.Equal(v.PubKey[:], pubkeySerialized[:]) && v.IsActive() {
			s.ValidatorRegistry[i].Status = primitives.StatusActivePendingExit
			s.ValidatorRegistry[i].LastActiveEpoch = s.EpochIndex + 2
		}
	}

	return nil
}

// IsDepositValid validates signatures and ensures that a deposit is valid.
func (s *state) IsDepositValid(deposit *primitives.Deposit, params *params.ChainParams) error {
	dPub, err := deposit.GetPublicKey()
	if err != nil {
		return err
	}
	pkh, err := dPub.Hash()
	if err != nil {
		return err
	}
	if s.CoinsState.Balances[pkh] < params.DepositAmount*params.UnitsPerCoin {
		return fmt.Errorf("balance is too low for deposit (got: %d, expected at least: %d)", s.CoinsState.Balances[pkh], params.DepositAmount*params.UnitsPerCoin)
	}

	buf, err := deposit.Data.Marshal()
	if err != nil {
		return err
	}

	depositHash := chainhash.HashH(buf)
	dSig, err := deposit.GetSignature()
	if err != nil {
		return err
	}
	valid := dSig.Verify(depositHash[:], dPub)
	if !valid {
		return errors.New("deposit signature is not valid")
	}

	validatorPubkey := deposit.Data.PublicKey

	// now, ensure we don't already have this validator
	for _, v := range s.ValidatorRegistry {
		if bytes.Equal(v.PubKey[:], validatorPubkey[:]) {
			return fmt.Errorf("validator already registered")
		}
	}

	pubkeyHash := chainhash.HashH(validatorPubkey[:])
	dataSig, err := deposit.Data.GetSignature()
	if err != nil {
		return err
	}
	dataPub, err := deposit.Data.GetPublicKey()
	if err != nil {
		return err
	}
	valid = dataSig.Verify(pubkeyHash[:], dataPub)
	if !valid {
		return errors.New("proof-of-possession is not valid")
	}

	return nil
}

// ApplyDeposit applies a deposit to the state.
func (s *state) ApplyDeposit(deposit *primitives.Deposit, p *params.ChainParams) error {
	if err := s.IsDepositValid(deposit, p); err != nil {
		return err
	}
	pub, err := deposit.GetPublicKey()
	if err != nil {
		return err
	}
	pkh, err := pub.Hash()
	if err != nil {
		return err
	}

	s.CoinsState.Balances[pkh] -= p.DepositAmount * p.UnitsPerCoin

	s.ValidatorRegistry = append(s.ValidatorRegistry, &primitives.Validator{
		Balance:          p.DepositAmount * p.UnitsPerCoin,
		PubKey:           deposit.Data.PublicKey,
		PayeeAddress:     deposit.Data.WithdrawalAddress,
		Status:           primitives.StatusStarting,
		FirstActiveEpoch: s.EpochIndex + 2,
		LastActiveEpoch:  0,
	})

	return nil
}

var (
	// ErrorVoteEmpty returns when some information of a MultiValidatorVote is missing
	ErrorVoteEmpty = errors.New("vote information is not complete")
	// ErrorVoteSlot returns when the vote data slot is out of range
	ErrorVoteSlot = errors.New("slot out of range")
	// ErrorFromEpoch returns when the vote From Epoch doesn't match a justified epoch
	ErrorFromEpoch = errors.New("expected from epoch to match justified epoch")
	// ErrorJustifiedHashWrong returns when the vote justified hash doesn't match state justified hash
	ErrorJustifiedHash = errors.New("justified block hash is wrong")
	// ErrorFromEpochPreviousJustified returns when the from epoch slot doesn't match the previous justified epoch
	ErrorFromEpochPreviousJustified = errors.New("expected from epoch to match previous justified epoch")
	// ErrorFromEpochPreviousJustifiedHash returns when the from epoch hash doesn't match the previous justified epoch hash
	ErrorFromEpochPreviousJustifiedHash = errors.New("expected from epoch hash to match previous justified epoch hash")
	// ErrorTargetEpoch returns when the to epoch hash a wrong target
	ErrorTargetEpoch = errors.New("vote should have target epoch of either the current epoch or the previous epoch")
	// ErrorVoteSignature returns when the vote aggregate signature doesn't validate
	ErrorVoteSignature = errors.New("vote aggregate signature did not validate")
)

// IsVoteValid checks if a vote is valid.
func (s *state) IsVoteValid(v *primitives.MultiValidatorVote, p *params.ChainParams) error {

	if v.Data == nil || v.ParticipationBitfield == nil || reflect.DeepEqual(v.Sig, [96]byte{}) {
		return ErrorVoteEmpty
	}

	if v.Data.Slot == 0 {
		return ErrorVoteSlot
	}

	if v.Data.ToEpoch == s.EpochIndex {
		if v.Data.FromEpoch != s.JustifiedEpoch {
			return ErrorFromEpoch
		}
		if !bytes.Equal(s.JustifiedEpochHash[:], v.Data.FromHash[:]) {
			return ErrorJustifiedHash
		}
	} else if s.EpochIndex > 0 && v.Data.ToEpoch == s.EpochIndex-1 {
		if v.Data.FromEpoch != s.PreviousJustifiedEpoch {
			return ErrorFromEpochPreviousJustified
		}
		if !bytes.Equal(s.PreviousJustifiedEpochHash[:], v.Data.FromHash[:]) {
			return ErrorFromEpochPreviousJustifiedHash
		}
	} else {
		return ErrorTargetEpoch
	}

	aggPubs := make([]*bls.PublicKey, 0)
	validators, err := s.GetVoteCommittee(v.Data.Slot, p)
	if err != nil {
		return err
	}

	for i, validatorIdx := range validators {
		if !v.ParticipationBitfield.Get(uint(i)) {
			continue
		}
		pub, err := bls.PublicKeyFromBytes(s.ValidatorRegistry[validatorIdx].PubKey)
		if err != nil {
			return err
		}
		aggPubs = append(aggPubs, pub)
	}

	h := v.Data.Hash()
	vSig, err := v.Signature()
	if err != nil {
		return err
	}

	valid := vSig.FastAggregateVerify(aggPubs, h)
	if !valid {
		return ErrorVoteSignature
	}

	return nil
}

func (s *state) ProcessVote(v *primitives.MultiValidatorVote, p *params.ChainParams, proposerIndex uint64) error {
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

	bl := bitfield.NewBitlist(v.ParticipationBitfield.Len())

	for i, p := range v.ParticipationBitfield {
		bl[i] = p
	}

	if v.Data.ToEpoch == s.EpochIndex {
		s.CurrentEpochVotes = append(s.CurrentEpochVotes, &primitives.AcceptedVoteInfo{
			Data:                  v.Data,
			ParticipationBitfield: bl,
			Proposer:              proposerIndex,
			InclusionDelay:        s.Slot - v.Data.Slot,
		})
	} else {
		s.PreviousEpochVotes = append(s.PreviousEpochVotes, &primitives.AcceptedVoteInfo{
			Data:                  v.Data,
			ParticipationBitfield: bl,
			Proposer:              proposerIndex,
			InclusionDelay:        s.Slot - v.Data.Slot,
		})
	}

	return nil
}

// GetProposerPublicKey gets the public key for the proposer of a block.
func (s *state) GetProposerPublicKey(b *primitives.Block, p *params.ChainParams) (*bls.PublicKey, error) {
	slotIndex := (b.Header.Slot + p.EpochLength - 1) % p.EpochLength

	proposerIndex := s.ProposerQueue[slotIndex]
	proposer := s.ValidatorRegistry[proposerIndex]

	return bls.PublicKeyFromBytes(proposer.PubKey)
}

// CheckBlockSignature checks the block signature.
func (s *state) CheckBlockSignature(b *primitives.Block, p *params.ChainParams) error {
	blockHash := b.Hash()
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
func (s *state) ProcessBlock(b *primitives.Block, p *params.ChainParams) error {
	if b.Header.Slot != s.Slot {
		return fmt.Errorf("state is not updated to slot %d, instead got %d", b.Header.Slot, s.Slot)
	}

	if err := s.CheckBlockSignature(b, p); err != nil {
		return err
	}

	voteMerkleRoot := b.VotesMerkleRoot()
	transactionMerkleRoot := b.TransactionMerkleRoot()
	depositMerkleRoot := b.DepositMerkleRoot()
	exitMerkleRoot := b.ExitMerkleRoot()
	voteSlashingMerkleRoot := b.VoteSlashingRoot()
	proposerSlashingMerkleRoot := b.ProposerSlashingsRoot()
	randaoSlashingMerkleRoot := b.RANDAOSlashingsRoot()
	governanceVoteMerkleRoot := b.GovernanceVoteMerkleRoot()

	if !bytes.Equal(transactionMerkleRoot[:], b.Header.TxMerkleRoot[:]) {
		return fmt.Errorf("expected transaction merkle root to be %s but got %s", hex.EncodeToString(transactionMerkleRoot[:]), hex.EncodeToString(b.Header.TxMerkleRoot[:]))
	}

	if !bytes.Equal(voteMerkleRoot[:], b.Header.VoteMerkleRoot[:]) {
		return fmt.Errorf("expected vote merkle root to be %s but got %s", hex.EncodeToString(voteMerkleRoot[:]), hex.EncodeToString(b.Header.VoteMerkleRoot[:]))
	}

	if !bytes.Equal(depositMerkleRoot[:], b.Header.DepositMerkleRoot[:]) {
		return fmt.Errorf("expected deposit merkle root to be %s but got %s", hex.EncodeToString(depositMerkleRoot[:]), hex.EncodeToString(b.Header.DepositMerkleRoot[:]))
	}
	if !bytes.Equal(exitMerkleRoot[:], b.Header.ExitMerkleRoot[:]) {
		return fmt.Errorf("expected exit merkle root to be %s but got %s", hex.EncodeToString(exitMerkleRoot[:]), hex.EncodeToString(b.Header.ExitMerkleRoot[:]))
	}

	if !bytes.Equal(voteSlashingMerkleRoot[:], b.Header.VoteSlashingMerkleRoot[:]) {
		return fmt.Errorf("expected exit merkle root to be %s but got %s", hex.EncodeToString(voteSlashingMerkleRoot[:]), hex.EncodeToString(b.Header.VoteSlashingMerkleRoot[:]))
	}

	if !bytes.Equal(proposerSlashingMerkleRoot[:], b.Header.ProposerSlashingMerkleRoot[:]) {
		return fmt.Errorf("expected exit merkle root to be %s but got %s", hex.EncodeToString(proposerSlashingMerkleRoot[:]), hex.EncodeToString(b.Header.ProposerSlashingMerkleRoot[:]))
	}

	if !bytes.Equal(randaoSlashingMerkleRoot[:], b.Header.RANDAOSlashingMerkleRoot[:]) {
		return fmt.Errorf("expected exit merkle root to be %s but got %s", hex.EncodeToString(randaoSlashingMerkleRoot[:]), hex.EncodeToString(b.Header.RANDAOSlashingMerkleRoot[:]))
	}

	if !bytes.Equal(governanceVoteMerkleRoot[:], b.Header.GovernanceVotesMerkleRoot[:]) {
		return fmt.Errorf("expected exit merkle root to be %s but got %s", hex.EncodeToString(governanceVoteMerkleRoot[:]), hex.EncodeToString(b.Header.GovernanceVotesMerkleRoot[:]))
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
		if err := s.ApplyDeposit(d, p); err != nil {
			return err
		}
	}

	for _, tx := range b.Txs {
		if err := s.ApplyTransactionSingle(tx, b.Header.FeeAddress, p); err != nil {
			return err
		}
	}

	for _, vote := range b.GovernanceVotes {
		if err := s.ProcessGovernanceVote(vote, p); err != nil {
			return err
		}
	}

	slotIndex := (b.Header.Slot + p.EpochLength - 1) % p.EpochLength

	proposerIndex := s.ProposerQueue[slotIndex]

	for _, v := range b.Votes {
		if err := s.ProcessVote(v, p, proposerIndex); err != nil {
			return err
		}
	}

	for _, e := range b.Exits {
		if err := s.ApplyExit(e); err != nil {
			return err
		}
	}

	for _, rs := range b.RANDAOSlashings {
		if err := s.ApplyRANDAOSlashing(rs, p); err != nil {
			return err
		}
	}

	for _, vs := range b.VoteSlashings {
		if err := s.ApplyVoteSlashing(vs, p); err != nil {
			return err
		}
	}

	for _, ps := range b.ProposerSlashings {
		if err := s.ApplyProposerSlashing(ps, p); err != nil {
			return err
		}
	}

	for i := range s.NextRANDAO {
		s.NextRANDAO[i] ^= b.RandaoSignature[i]
	}

	return nil
}
