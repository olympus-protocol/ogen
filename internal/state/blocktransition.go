package state

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/pkg/bitfield"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/bls/common"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

// ApplyMultiTransactionSingle applies multiple single Tx to the state
func (s *state) ApplyMultiTransactionSingle(txs []*primitives.Tx, blockWithdrawalAddress [20]byte) error {

	u := s.CoinsState

	txsAmount := len(txs)

	txsSigs := make([]common.Signature, txsAmount)
	txsMsgs := make([][32]byte, txsAmount)
	txsPubs := make([]common.PublicKey, txsAmount)

	for i, tx := range txs {
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

		sig, err := tx.GetSignature()
		if err != nil {
			return err
		}

		pub, err := tx.GetPublic()
		if err != nil {
			return err
		}

		txsSigs[i] = sig
		txsMsgs[i] = tx.SignatureMessage()
		txsPubs[i] = pub
	}

	sig := bls.AggregateSignatures(txsSigs)

	valid := sig.AggregateVerify(txsPubs, txsMsgs)
	if !valid {
		return errors.New("invalid txs signatures")
	}

	for _, tx := range txs {
		pkh, err := tx.FromPubkeyHash()
		if err != nil {
			return err
		}
		u.Balances[pkh] -= tx.Amount + tx.Fee
		u.Balances[tx.To] += tx.Amount
		u.Balances[blockWithdrawalAddress] += tx.Fee
		u.Nonces[pkh] = tx.Nonce
	}
	return nil
}

// ApplyTransactionSingle applies a transaction to the coin state.
func (s *state) ApplyTransactionSingle(tx *primitives.Tx, blockWithdrawalAddress [20]byte) error {

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

	return nil
}

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
	if !s1.Verify(pub, h1[:]) {
		return 0, fmt.Errorf("proposer-slashing: signature does not validate for block header 1")
	}

	if !s2.Verify(pub, h2[:]) {
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
func (s *state) ApplyProposerSlashing(ps *primitives.ProposerSlashing) error {
	proposerIndex, err := s.IsProposerSlashingValid(ps)
	if err != nil {
		return err
	}

	return s.UpdateValidatorStatus(proposerIndex, primitives.StatusExitedWithPenalty)
}

// IsVoteSlashingValid checks if the vote slashing is valid.
func (s *state) IsVoteSlashingValid(vs *primitives.VoteSlashing) ([]uint64, error) {

	if vs.Vote1.Data.Equals(vs.Vote2.Data) {
		return nil, fmt.Errorf("vote-slashing: votes are not distinct")
	}

	if !vs.Vote1.Data.IsDoubleVote(vs.Vote2.Data) && !vs.Vote1.Data.IsSurroundVote(vs.Vote2.Data) {
		return nil, fmt.Errorf("vote-slashing: votes do not violate slashing rule")
	}

	c := make([]uint64, 0)
	voteCommittee1 := make(map[uint64]struct{})

	validators1, err := s.GetVoteCommittee(vs.Vote1.Data.Slot)
	if err != nil {
		return nil, err
	}

	validators2, err := s.GetVoteCommittee(vs.Vote2.Data.Slot)
	if err != nil {
		return nil, err
	}

	aggPubs1 := make([]common.PublicKey, 0)
	aggPubs2 := make([]common.PublicKey, 0)

	for i, idx := range validators1 {

		if !vs.Vote1.ParticipationBitfield.Get(uint(i)) {
			continue
		}

		voteCommittee1[idx] = struct{}{}

		pub, err := bls.PublicKeyFromBytes(s.ValidatorRegistry[idx].PubKey[:])
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
			c = append(c, idx)
		}

		pub, err := bls.PublicKeyFromBytes(s.ValidatorRegistry[idx].PubKey[:])
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

	if len(c) == 0 {
		return nil, fmt.Errorf("vote-slashing: votes do not contain any common validators")
	}

	v2Sig, err := vs.Vote2.Signature()
	if err != nil {
		return nil, err
	}
	if !v2Sig.FastAggregateVerify(aggPubs2, vs.Vote2.Data.Hash()) {
		return nil, fmt.Errorf("vote-slashing: vote 2 does not validate")
	}

	return c, nil
}

// ApplyVoteSlashing applies a vote slashing to the state.
func (s *state) ApplyVoteSlashing(vs *primitives.VoteSlashing) error {
	commonVotes, err := s.IsVoteSlashingValid(vs)
	if err != nil {
		return err
	}

	for _, v := range commonVotes {
		if err := s.UpdateValidatorStatus(v, primitives.StatusExitedWithPenalty); err != nil {
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
	if !sig.Verify(pub, slotHash[:]) {
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
func (s *state) ApplyRANDAOSlashing(rs *primitives.RANDAOSlashing) error {
	proposer, err := s.IsRANDAOSlashingValid(rs)
	if err != nil {
		return err
	}

	return s.UpdateValidatorStatus(proposer, primitives.StatusExitedWithPenalty)
}

// GetVoteCommittee gets the committee for a certain block.
func (s *state) GetVoteCommittee(slot uint64) ([]uint64, error) {
	netParams := config.GlobalParams.NetParams

	if (slot-1)/netParams.EpochLength == s.EpochIndex {
		assignments := s.CurrentEpochVoteAssignments
		slotIndex := slot % netParams.EpochLength
		min := (slotIndex * uint64(len(assignments))) / netParams.EpochLength
		max := ((slotIndex + 1) * uint64(len(assignments))) / netParams.EpochLength
		return assignments[min:max], nil

	} else if (slot-1)/netParams.EpochLength == s.EpochIndex-1 {
		assignments := s.PreviousEpochVoteAssignments
		slotIndex := slot % netParams.EpochLength
		min := (slotIndex * uint64(len(assignments))) / netParams.EpochLength
		max := ((slotIndex + 1) * uint64(len(assignments))) / netParams.EpochLength
		return assignments[min:max], nil
	}

	// TODO: better handling
	return nil, fmt.Errorf("tried to get vote committee out of range: %d", slot)
}

// IsExitValid checks if an exit is valid.
func (s *state) IsExitValid(exit *primitives.Exit) error {
	msgHash := chainhash.HashH(exit.ValidatorPubkey[:])
	wPubKey, err := exit.GetWithdrawPubKey()
	if err != nil {
		return err
	}
	sig, err := exit.GetSignature()
	if err != nil {
		return err
	}
	valid := sig.Verify(wPubKey, msgHash[:])
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

// IsPartialExitValid checks if an exit is valid.
func (s *state) IsPartialExitValid(p *primitives.PartialExit) error {
	params := config.GlobalParams.NetParams

	if p.Amount < 15*params.UnitsPerCoin {
		return errors.New("partial exit tries to unlock a very little amount of coins")
	}

	msgHash := chainhash.HashH(p.ValidatorPubkey[:])
	wPubKey, err := p.GetWithdrawPubKey()
	if err != nil {
		return err
	}

	sig, err := p.GetSignature()
	if err != nil {
		return err
	}

	valid := sig.Verify(wPubKey, msgHash[:])
	if !valid {
		return fmt.Errorf("exit signature is not valid")
	}

	pkh, err := wPubKey.Hash()
	if err != nil {
		return err
	}

	pubkeySerialized := p.ValidatorPubkey

	foundActiveValidator := false
	expectedValidatorBalance := (params.DepositAmount + p.Amount) * params.UnitsPerCoin
	for _, v := range s.ValidatorRegistry {
		if bytes.Equal(v.PubKey[:], pubkeySerialized[:]) && v.IsActive() {
			if !bytes.Equal(v.PayeeAddress[:], pkh[:]) {
				return fmt.Errorf("withdraw pubkey does not match withdraw address (expected: %x, got: %x)", pkh, v.PayeeAddress)
			}
			foundActiveValidator = true
			if v.Balance < expectedValidatorBalance {
				return errors.New("validator doesn't have enough balance to withdraw")
			}
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

	for i, v := range s.ValidatorRegistry {
		if bytes.Equal(v.PubKey[:], exit.ValidatorPubkey[:]) && v.IsActive() {
			s.ValidatorRegistry[i].Status = primitives.StatusActivePendingExit
			s.ValidatorRegistry[i].LastActiveEpoch = s.EpochIndex + 2
		}
	}

	return nil
}

// ApplyPartialExit processes an exit request.
func (s *state) ApplyPartialExit(p *primitives.PartialExit) error {
	if err := s.IsPartialExitValid(p); err != nil {
		return err
	}

	for i, v := range s.ValidatorRegistry {
		if bytes.Equal(v.PubKey[:], p.ValidatorPubkey[:]) && v.IsActive() {
			s.ValidatorRegistry[i].Balance -= p.Amount
			s.CoinsState.Balances[v.PayeeAddress] += p.Amount
		}
	}

	return nil
}

// AreDepositsValid validates multiple deposits
func (s *state) AreDepositsValid(deposits []*primitives.Deposit) error {

	netParams := config.GlobalParams.NetParams

	num := len(deposits)

	sigs := make([]common.Signature, num)
	pubs := make([]common.PublicKey, num)
	msgs := make([][32]byte, num)

	pSigs := make([]common.Signature, num)
	pPubs := make([]common.PublicKey, num)
	pMsgs := make([][32]byte, num)

	balances := make(map[[20]byte]uint64)

	for i, d := range deposits {

		pub, err := d.GetPublicKey()
		if err != nil {
			return err
		}

		pubs[i] = pub

		pkh, err := pub.Hash()
		if err != nil {
			return err
		}

		balances[pkh] += netParams.DepositAmount * netParams.UnitsPerCoin

		for _, v := range s.ValidatorRegistry {
			if bytes.Equal(v.PubKey[:], d.Data.PublicKey[:]) {
				return fmt.Errorf("validator already registered")
			}
		}

		buf, err := d.Data.Marshal()
		if err != nil {
			return err
		}

		msgs[i] = chainhash.HashH(buf)
		pMsgs[i] = chainhash.HashH(d.Data.PublicKey[:])

		proofPub, err := d.Data.GetPublicKey()
		if err != nil {
			return err
		}

		pPubs[i] = proofPub

		depositSig, err := d.GetSignature()
		if err != nil {
			return err
		}

		proofSig, err := d.Data.GetSignature()

		if err != nil {
			return err
		}

		sigs[i] = depositSig
		pSigs[i] = proofSig

	}

	for pkh, balance := range balances {
		if s.CoinsState.Balances[pkh] < balance {
			return fmt.Errorf("balance is too low for deposit (got: %d, expected at least: %d)", s.CoinsState.Balances[pkh], balance)
		}
	}

	sig := bls.AggregateSignatures(sigs)
	pSig := bls.AggregateSignatures(pSigs)

	valid1 := sig.AggregateVerify(pubs, msgs)
	if !valid1 {
		return errors.New("deposit signatures don't verify")
	}

	valid2 := pSig.AggregateVerify(pPubs, pMsgs)
	if !valid2 {
		return errors.New("proof-of-possession signatures don't verify")
	}

	return nil
}

// IsDepositValid validates signatures and ensures that a deposit is valid.
func (s *state) IsDepositValid(deposit *primitives.Deposit) error {
	netParams := config.GlobalParams.NetParams

	dPub, err := deposit.GetPublicKey()
	if err != nil {
		return err
	}

	pkh, err := dPub.Hash()
	if err != nil {
		return err
	}

	if s.CoinsState.Balances[pkh] < netParams.DepositAmount*netParams.UnitsPerCoin {
		return fmt.Errorf("balance is too low for deposit (got: %d, expected at least: %d)", s.CoinsState.Balances[pkh], netParams.DepositAmount*netParams.UnitsPerCoin)
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

	valid := dSig.Verify(dPub, depositHash[:])
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

	valid = dataSig.Verify(dataPub, pubkeyHash[:])
	if !valid {
		return errors.New("proof-of-possession is not valid")
	}

	return nil
}

// ApplyMultiDeposit applies multiple deposits to the state
func (s *state) ApplyMultiDeposit(deposits []*primitives.Deposit) error {
	netParams := config.GlobalParams.NetParams

	if err := s.AreDepositsValid(deposits); err != nil {
		return err
	}

	for _, d := range deposits {
		pub, err := d.GetPublicKey()
		if err != nil {
			return err
		}
		pkh, err := pub.Hash()
		if err != nil {
			return err
		}

		s.CoinsState.Balances[pkh] -= netParams.DepositAmount * netParams.UnitsPerCoin

		s.ValidatorRegistry = append(s.ValidatorRegistry, &primitives.Validator{
			Balance:          netParams.DepositAmount * netParams.UnitsPerCoin,
			PubKey:           d.Data.PublicKey,
			PayeeAddress:     d.Data.WithdrawalAddress,
			Status:           primitives.StatusStarting,
			FirstActiveEpoch: s.EpochIndex + 2,
		})
	}
	return nil
}

// ApplyDeposit applies a deposit to the state.
func (s *state) ApplyDeposit(deposit *primitives.Deposit) error {
	netParams := config.GlobalParams.NetParams

	if err := s.IsDepositValid(deposit); err != nil {
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

	s.CoinsState.Balances[pkh] -= netParams.DepositAmount * netParams.UnitsPerCoin

	s.ValidatorRegistry = append(s.ValidatorRegistry, &primitives.Validator{
		Balance:          netParams.DepositAmount * netParams.UnitsPerCoin,
		PubKey:           deposit.Data.PublicKey,
		PayeeAddress:     deposit.Data.WithdrawalAddress,
		Status:           primitives.StatusStarting,
		FirstActiveEpoch: s.EpochIndex + 2,
	})

	return nil
}

var (
	// ErrorVoteSlot returns when the vote data slot is out of range
	ErrorVoteSlot = errors.New("slot out of range")
	// ErrorFromEpoch returns when the vote From Epoch doesn't match a justified epoch
	ErrorFromEpoch = errors.New("expected from epoch to match justified epoch")
	// ErrorJustifiedHash returns when the vote justified hash doesn't match state justified hash
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
func (s *state) IsVoteValid(v *primitives.MultiValidatorVote) error {

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

	aggPubs := make([]common.PublicKey, 0)
	validators, err := s.GetVoteCommittee(v.Data.Slot)
	if err != nil {
		return err
	}

	for i, validatorIdx := range validators {
		if !v.ParticipationBitfield.Get(uint(i)) {
			continue
		}
		pub, err := bls.PublicKeyFromBytes(s.ValidatorRegistry[validatorIdx].PubKey[:])
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

func (s *state) ProcessVote(v *primitives.MultiValidatorVote, proposerIndex uint64) error {
	netParams := config.GlobalParams.NetParams

	if v.Data.Slot+netParams.MinAttestationInclusionDelay > s.Slot {
		return fmt.Errorf("vote included too soon (expected s.Slot > %d, got %d)", v.Data.Slot+netParams.MinAttestationInclusionDelay, s.Slot)
	}

	if v.Data.Slot+netParams.EpochLength <= s.Slot {
		return fmt.Errorf("vote not included within 1 epoch (latest: %d, got: %d)", v.Data.Slot+netParams.EpochLength-1, s.Slot)
	}

	if (v.Data.Slot-1)/netParams.EpochLength != v.Data.ToEpoch {
		return errors.New("vote slot did not match target epoch")
	}

	err := s.IsVoteValid(v)

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
func (s *state) GetProposerPublicKey(b *primitives.Block) (common.PublicKey, error) {
	netParams := config.GlobalParams.NetParams

	slotIndex := (b.Header.Slot + netParams.EpochLength - 1) % netParams.EpochLength

	proposerIndex := s.ProposerQueue[slotIndex]
	proposer := s.ValidatorRegistry[proposerIndex]

	return bls.PublicKeyFromBytes(proposer.PubKey[:])
}

// CheckBlockSignature checks the block signature.
func (s *state) CheckBlockSignature(b *primitives.Block) error {

	blockHash := b.Hash()
	blockSig, err := bls.SignatureFromBytes(b.Signature[:])
	if err != nil {
		return err
	}

	validatorPub, err := s.GetProposerPublicKey(b)
	if err != nil {
		return err
	}

	randaoSig, err := bls.SignatureFromBytes(b.RandaoSignature[:])
	if err != nil {
		return err
	}

	valid := blockSig.Verify(validatorPub, blockHash[:])
	if !valid {
		return errors.New("error validating signature for block")
	}

	slotHash := chainhash.HashH([]byte(fmt.Sprintf("%d", b.Header.Slot)))

	valid = randaoSig.Verify(validatorPub, slotHash[:])
	if !valid {
		return errors.New("error validating RANDAO signature for block")
	}

	return nil
}

// ProcessBlock runs a block transition on the state and mutates state.
func (s *state) ProcessBlock(b *primitives.Block) error {
	netParams := config.GlobalParams.NetParams

	if b.Header.Slot != s.Slot {
		return fmt.Errorf("state is not updated to slot %d, instead got %d", b.Header.Slot, s.Slot)
	}

	if err := s.CheckBlockSignature(b); err != nil {
		return err
	}

	voteMerkleRoot := b.VotesMerkleRoot()
	depositMerkleRoot := b.DepositMerkleRoot()
	exitMerkleRoot := b.ExitMerkleRoot()
	partialExitsMerkleRoot := b.PartialExitsMerkleRoot()
	txsMerkleRoot := b.TxsMerkleRoot()
	voteSlashingMerkleRoot := b.VoteSlashingRoot()
	proposerSlashingMerkleRoot := b.ProposerSlashingsRoot()
	randaoSlashingMerkleRoot := b.RANDAOSlashingsRoot()

	if !bytes.Equal(depositMerkleRoot[:], b.Header.DepositMerkleRoot[:]) {
		return fmt.Errorf("expected deposit merkle root to be %s but got %s", hex.EncodeToString(depositMerkleRoot[:]), hex.EncodeToString(b.Header.DepositMerkleRoot[:]))
	}

	if !bytes.Equal(exitMerkleRoot[:], b.Header.ExitMerkleRoot[:]) {
		return fmt.Errorf("expected exit merkle root to be %s but got %s", hex.EncodeToString(exitMerkleRoot[:]), hex.EncodeToString(b.Header.ExitMerkleRoot[:]))
	}

	if !bytes.Equal(voteMerkleRoot[:], b.Header.VoteMerkleRoot[:]) {
		return fmt.Errorf("expected vote merkle root to be %s but got %s", hex.EncodeToString(voteMerkleRoot[:]), hex.EncodeToString(b.Header.VoteMerkleRoot[:]))
	}

	if !bytes.Equal(partialExitsMerkleRoot[:], b.Header.PartialExitMerkleRoot[:]) {
		return fmt.Errorf("expected partial exits merkle root to be %s but got %s", hex.EncodeToString(partialExitsMerkleRoot[:]), hex.EncodeToString(b.Header.PartialExitMerkleRoot[:]))
	}

	if !bytes.Equal(txsMerkleRoot[:], b.Header.TxsMerkleRoot[:]) {
		return fmt.Errorf("expected transaction merkle root to be %s but got %s", hex.EncodeToString(txsMerkleRoot[:]), hex.EncodeToString(b.Header.TxsMerkleRoot[:]))
	}

	if !bytes.Equal(voteSlashingMerkleRoot[:], b.Header.VoteSlashingMerkleRoot[:]) {
		return fmt.Errorf("expected vote slashing merkle root to be %s but got %s", hex.EncodeToString(voteSlashingMerkleRoot[:]), hex.EncodeToString(b.Header.VoteSlashingMerkleRoot[:]))
	}

	if !bytes.Equal(proposerSlashingMerkleRoot[:], b.Header.ProposerSlashingMerkleRoot[:]) {
		return fmt.Errorf("expected proposer slashing merkle root to be %s but got %s", hex.EncodeToString(proposerSlashingMerkleRoot[:]), hex.EncodeToString(b.Header.ProposerSlashingMerkleRoot[:]))
	}

	if !bytes.Equal(randaoSlashingMerkleRoot[:], b.Header.RANDAOSlashingMerkleRoot[:]) {
		return fmt.Errorf("expected randao slashing merkle root to be %s but got %s", hex.EncodeToString(randaoSlashingMerkleRoot[:]), hex.EncodeToString(b.Header.RANDAOSlashingMerkleRoot[:]))
	}

	if uint64(len(b.Votes)) > primitives.MaxVotesPerBlock {
		return fmt.Errorf("block has too many votes (max: %d, got: %d)", primitives.MaxVotesPerBlock, len(b.Votes))
	}

	if uint64(len(b.Deposits)) > primitives.MaxDepositsPerBlock {
		return fmt.Errorf("block has too many deposits (max: %d, got: %d)", primitives.MaxDepositsPerBlock, len(b.Deposits))
	}

	if uint64(len(b.Exits)) > primitives.MaxExitsPerBlock {
		return fmt.Errorf("block has too many exits (max: %d, got: %d)", primitives.MaxExitsPerBlock, len(b.Exits))
	}

	if uint64(len(b.PartialExit)) > primitives.MaxPartialExitsPerBlock {
		return fmt.Errorf("block has too many partial exits (max: %d, got: %d)", primitives.MaxPartialExitsPerBlock, len(b.PartialExit))
	}

	if uint64(len(b.Txs)) > primitives.MaxTxsPerBlock {
		return fmt.Errorf("block has too many txs (max: %d, got: %d)", primitives.MaxTxsPerBlock, len(b.Txs))
	}

	if uint64(len(b.ProposerSlashings)) > primitives.MaxProposerSlashingsPerBlock {
		return fmt.Errorf("block has too many proposer slashings (max: %d, got: %d)", primitives.MaxProposerSlashingsPerBlock, len(b.ProposerSlashings))
	}

	if uint64(len(b.VoteSlashings)) > primitives.MaxVoteSlashingsPerBlock {
		return fmt.Errorf("block has too many vote slashings (max: %d, got: %d)", primitives.MaxVoteSlashingsPerBlock, len(b.VoteSlashings))
	}

	if uint64(len(b.RANDAOSlashings)) > primitives.MaxRANDAOSlashingsPerBlock {
		return fmt.Errorf("block has too many RANDAO slashings (max: %d, got: %d)", primitives.MaxRANDAOSlashingsPerBlock, len(b.RANDAOSlashings))
	}

	if len(b.Txs) > 0 {
		if err := s.ApplyMultiTransactionSingle(b.Txs, b.Header.FeeAddress); err != nil {
			return err
		}
	}

	slotIndex := (b.Header.Slot + netParams.EpochLength - 1) % netParams.EpochLength

	proposerIndex := s.ProposerQueue[slotIndex]

	for _, v := range b.Votes {
		if err := s.ProcessVote(v, proposerIndex); err != nil {
			return err
		}
	}

	if len(b.Deposits) > 0 {
		if err := s.ApplyMultiDeposit(b.Deposits); err != nil {
			return err
		}
	}

	for _, e := range b.Exits {
		if err := s.ApplyExit(e); err != nil {
			return err
		}
	}

	for _, p := range b.PartialExit {
		if err := s.ApplyPartialExit(p); err != nil {
			return err
		}
	}

	for _, ps := range b.ProposerSlashings {
		if err := s.ApplyProposerSlashing(ps); err != nil {
			return err
		}
	}

	for _, vs := range b.VoteSlashings {
		if err := s.ApplyVoteSlashing(vs); err != nil {
			return err
		}
	}

	for _, rs := range b.RANDAOSlashings {
		if err := s.ApplyRANDAOSlashing(rs); err != nil {
			return err
		}
	}

	for i := range s.NextRANDAO {
		s.NextRANDAO[i] ^= b.RandaoSignature[i]
	}

	return nil
}
