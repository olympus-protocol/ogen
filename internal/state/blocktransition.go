package state

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/pkg/bitfield"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/burnproof"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

// IsGovernanceVoteValid checks if a governance vote is valid.
func (s *state) IsGovernanceVoteValid(vote *primitives.GovernanceVote) error {
	netParams := config.GlobalParams.NetParams

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
		// TODO check multisig as single signatures
		pub, err := vote.Multisig.GetPublicKey()
		if err != nil {
			return err
		}
		pkh, err := pub.Hash()
		if err != nil {
			return err
		}
		if s.CoinsState.Balances[pkh] < netParams.MinVotingBalance*netParams.UnitsPerCoin {
			return fmt.Errorf("minimum balance is %d, but got %d", netParams.MinVotingBalance, s.CoinsState.Balances[pkh]/netParams.UnitsPerCoin)
		}
		if !vote.Valid() {
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
		if len(vote.Data) != len(netParams.GovernancePercentages)*20 {
			return fmt.Errorf("expected VoteFor vote to have %d bytes of data got %d", len(netParams.GovernancePercentages)*32, len(vote.Data))
		}
		// TODO check multisig as single signatures
		pub, err := vote.Multisig.GetPublicKey()
		if err != nil {
			return err
		}
		pkh, err := pub.Hash()
		if err != nil {
			return err
		}
		if s.CoinsState.Balances[pkh] < netParams.MinVotingBalance*netParams.UnitsPerCoin {
			return fmt.Errorf("minimum balance is %d, but got %d", netParams.MinVotingBalance, s.CoinsState.Balances[pkh]/netParams.UnitsPerCoin)
		}
		if _, ok := s.Governance.ReplaceVotes[pkh]; ok {
			return fmt.Errorf("found existing vote for same public key hash")
		}
		if !vote.Valid() {
			return fmt.Errorf("vote signature did not validate")
		}
	case primitives.UpdateManagersInstantly:
		// must be during active period
		// must be signed by all managers
		if s.VotingState != GovernanceStateActive {
			return fmt.Errorf("cannot vote for community vote during community vote period")
		}
		if len(vote.Data) != len(netParams.GovernancePercentages)*20 {
			return fmt.Errorf("expected UpdateManagersInstantly vote to have %d bytes data but got %d", len(vote.Data), len(netParams.GovernancePercentages)*32)
		}
		pub, err := vote.Multisig.GetPublicKey()
		if err != nil {
			return err
		}
		if pub.NumNeeded != 5 {
			return fmt.Errorf("expected 5 signatures needed")
		}
		for i := range pub.PublicKeys {
			pub, err := bls.PublicKeyFromBytes(pub.PublicKeys[i][:])
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

		if !vote.Valid() {
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
		// must include multisig signed by 5/5 managers
		pub, err := vote.Multisig.GetPublicKey()
		if err != nil {
			return err
		}
		if pub.NumNeeded >= 3 {
			return fmt.Errorf("expected 3 signatures needed")
		}
		for i := range pub.PublicKeys {
			pub, err := bls.PublicKeyFromBytes(pub.PublicKeys[i][:])
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

		if !vote.Valid() {
			return fmt.Errorf("vote signature is not valid")
		}
	default:
		return fmt.Errorf("unknown vote type")
	}

	return nil
}

// ProcessGovernanceVote processes governance votes.
func (s *state) ProcessGovernanceVote(vote *primitives.GovernanceVote) error {
	if err := s.IsGovernanceVoteValid(vote); err != nil {
		return err
	}
	pub, err := vote.Multisig.GetPublicKey()
	if err != nil {
		return err
	}
	hash, err := pub.Hash()
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

func (s *state) ApplyTransactionsSingle(txs []*primitives.Tx, blockWithdrawalAddress [20]byte) error {
	netParams := config.GlobalParams.NetParams

	u := s.CoinsState

	txsAmount := len(txs)

	txsSigs := make([]*bls.Signature, txsAmount)
	txsMsgs := make([][32]byte, txsAmount)
	txsPubs := make([]*bls.PublicKey, txsAmount)

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

		if _, ok := s.Governance.ReplaceVotes[pkh]; u.Balances[pkh] < netParams.UnitsPerCoin*netParams.MinVotingBalance && ok {
			delete(s.Governance.ReplaceVotes, pkh)
		}
	}
	return nil
}

// ApplyTransactionSingle applies a transaction to the coin state.
func (s *state) ApplyTransactionSingle(tx *primitives.Tx, blockWithdrawalAddress [20]byte) error {
	netParams := config.GlobalParams.NetParams

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

	if _, ok := s.Governance.ReplaceVotes[pkh]; u.Balances[pkh] < netParams.UnitsPerCoin*netParams.MinVotingBalance && ok {
		delete(s.Governance.ReplaceVotes, pkh)
	}

	return nil
}

// ApplyTransactionMulti applies a multisig transaction to the coin state.
func (s *state) ApplyTransactionMulti(tx *primitives.TxMulti, blockWithdrawalAddress [20]byte) error {
	netParams := config.GlobalParams.NetParams

	u := s.GetCoinsState()
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

	if _, ok := s.Governance.ReplaceVotes[pkh]; u.Balances[pkh] < netParams.UnitsPerCoin*netParams.MinVotingBalance && ok {
		delete(s.Governance.ReplaceVotes, pkh)
	}

	return nil
}

// IsCoinProofValid checks if an coin proof is valid.
func (s *state) IsCoinProofValid(p *burnproof.CoinsProofSerializable) error {
	proof, err := p.ToCoinProof()
	if err != nil {
		return err
	}

	err = burnproof.VerifyBurnProof(proof, p.RedeemAccount)
	if err != nil {
		return err
	}

	if _, ok := s.CoinsState.ProofsVerified[p.Hash()]; ok {
		return errors.New("proof already verified")
	}

	return nil
}

// ApplyCoinProof applies a migration proof to the coin state.
func (s *state) ApplyCoinProof(p *burnproof.CoinsProofSerializable) error {
	err := s.IsCoinProofValid(p)
	if err != nil {
		return err
	}

	proof, err := p.ToCoinProof()
	if err != nil {
		return err
	}

	u := s.CoinsState

	sumBalance := uint64(0)
	for _, out := range proof.Transaction.TxOut {
		sumBalance += uint64(out.Value)
	}

	pkh, err := p.RedeemAccountHash()
	if err != nil {
		return err
	}

	u.Balances[pkh] += sumBalance

	u.ProofsVerified[p.Hash()] = struct{}{}

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

	common := make([]uint64, 0)
	voteCommittee1 := make(map[uint64]struct{})

	validators1, err := s.GetVoteCommittee(vs.Vote1.Data.Slot)
	if err != nil {
		return nil, err
	}

	validators2, err := s.GetVoteCommittee(vs.Vote2.Data.Slot)
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
			common = append(common, idx)
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
func (s *state) ApplyVoteSlashing(vs *primitives.VoteSlashing) error {
	common, err := s.IsVoteSlashingValid(vs)
	if err != nil {
		return err
	}

	for _, v := range common {
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

// AreDepositsValid validates multiple deposits
func (s *state) AreDepositsValid(deposits []*primitives.Deposit) error {
	netParams := config.GlobalParams.NetParams

	depNum := len(deposits)

	depSigs := make([]*bls.Signature, depNum)
	depPubs := make([]*bls.PublicKey, depNum)
	depMsgs := make([][32]byte, depNum)
	proofsSigs := make([]*bls.Signature, depNum)
	proofsPubs := make([]*bls.PublicKey, depNum)
	proofsMsgs := make([][32]byte, depNum)

	balances := make(map[[20]byte]uint64)

	for i, d := range deposits {
		pub, err := d.GetPublicKey()
		if err != nil {
			return err
		}

		depPubs[i] = pub

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

		depMsgs[i] = chainhash.HashH(buf)
		proofsMsgs[i] = chainhash.HashH(d.Data.PublicKey[:])

		proofPub, err := d.Data.GetPublicKey()
		if err != nil {
			return err
		}

		proofsPubs[i] = proofPub

		depositSig, err := d.GetSignature()
		if err != nil {
			return err
		}
		proofSig, err := d.Data.GetSignature()

		if err != nil {
			return err
		}

		depSigs[i] = depositSig
		proofsSigs[i] = proofSig

	}

	for pkh, balance := range balances {
		if s.CoinsState.Balances[pkh] < balance {
			return fmt.Errorf("balance is too low for deposit (got: %d, expected at least: %d)", s.CoinsState.Balances[pkh], balance)
		}
	}

	depositsSig := bls.AggregateSignatures(depSigs)
	proofOfPossessionSig := bls.AggregateSignatures(proofsSigs)

	valid1 := depositsSig.AggregateVerify(depPubs, depMsgs)
	if !valid1 {
		return errors.New("deposit signatures don't verify")
	}

	valid2 := proofOfPossessionSig.AggregateVerify(proofsPubs, proofsMsgs)
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

// ApplyDeposits applies multiple deposits to the state
func (s *state) ApplyDeposits(deposits []*primitives.Deposit) error {
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

	aggPubs := make([]*bls.PublicKey, 0)
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
func (s *state) GetProposerPublicKey(b *primitives.Block) (*bls.PublicKey, error) {
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
	transactionMerkleRoot := b.TransactionMerkleRoot()
	transactionMultiMerkleRoot := b.TransactionMultiMerkleRoot()
	depositMerkleRoot := b.DepositMerkleRoot()
	exitMerkleRoot := b.ExitMerkleRoot()
	voteSlashingMerkleRoot := b.VoteSlashingRoot()
	proposerSlashingMerkleRoot := b.ProposerSlashingsRoot()
	randaoSlashingMerkleRoot := b.RANDAOSlashingsRoot()
	governanceVoteMerkleRoot := b.GovernanceVoteMerkleRoot()
	coinProofsMerkleRoot := b.CoinProofsMerkleRoot()

	if !bytes.Equal(transactionMerkleRoot[:], b.Header.TxMerkleRoot[:]) {
		return fmt.Errorf("expected transaction merkle root to be %s but got %s", hex.EncodeToString(transactionMerkleRoot[:]), hex.EncodeToString(b.Header.TxMerkleRoot[:]))
	}

	if !bytes.Equal(transactionMultiMerkleRoot[:], b.Header.TxMultiMerkleRoot[:]) {
		return fmt.Errorf("expected transaction multi merkle root to be %s but got %s", hex.EncodeToString(transactionMultiMerkleRoot[:]), hex.EncodeToString(b.Header.TxMultiMerkleRoot[:]))
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
		return fmt.Errorf("expected vote slashing merkle root to be %s but got %s", hex.EncodeToString(voteSlashingMerkleRoot[:]), hex.EncodeToString(b.Header.VoteSlashingMerkleRoot[:]))
	}

	if !bytes.Equal(proposerSlashingMerkleRoot[:], b.Header.ProposerSlashingMerkleRoot[:]) {
		return fmt.Errorf("expected proposer slashing merkle root to be %s but got %s", hex.EncodeToString(proposerSlashingMerkleRoot[:]), hex.EncodeToString(b.Header.ProposerSlashingMerkleRoot[:]))
	}

	if !bytes.Equal(randaoSlashingMerkleRoot[:], b.Header.RANDAOSlashingMerkleRoot[:]) {
		return fmt.Errorf("expected randao slashing merkle root to be %s but got %s", hex.EncodeToString(randaoSlashingMerkleRoot[:]), hex.EncodeToString(b.Header.RANDAOSlashingMerkleRoot[:]))
	}

	if !bytes.Equal(governanceVoteMerkleRoot[:], b.Header.GovernanceVotesMerkleRoot[:]) {
		return fmt.Errorf("expected governance votes merkle root to be %s but got %s", hex.EncodeToString(governanceVoteMerkleRoot[:]), hex.EncodeToString(b.Header.GovernanceVotesMerkleRoot[:]))
	}

	if !bytes.Equal(coinProofsMerkleRoot[:], b.Header.CoinProofsMerkleRoot[:]) {
		return fmt.Errorf("expected coin proofs merkle root to be %s but got %s", hex.EncodeToString(governanceVoteMerkleRoot[:]), hex.EncodeToString(b.Header.GovernanceVotesMerkleRoot[:]))
	}

	if uint64(len(b.Votes)) > netParams.MaxVotesPerBlock {
		return fmt.Errorf("block has too many votes (max: %d, got: %d)", netParams.MaxVotesPerBlock, len(b.Votes))
	}

	if uint64(len(b.Txs)) > netParams.MaxTxsPerBlock {
		return fmt.Errorf("block has too many txs (max: %d, got: %d)", netParams.MaxTxsPerBlock, len(b.Txs))
	}

	if uint64(len(b.Deposits)) > netParams.MaxDepositsPerBlock {
		return fmt.Errorf("block has too many deposits (max: %d, got: %d)", netParams.MaxDepositsPerBlock, len(b.Deposits))
	}

	if uint64(len(b.Exits)) > netParams.MaxExitsPerBlock {
		return fmt.Errorf("block has too many exits (max: %d, got: %d)", netParams.MaxExitsPerBlock, len(b.Exits))
	}

	if uint64(len(b.RANDAOSlashings)) > netParams.MaxRANDAOSlashingsPerBlock {
		return fmt.Errorf("block has too many RANDAO slashings (max: %d, got: %d)", netParams.MaxRANDAOSlashingsPerBlock, len(b.RANDAOSlashings))
	}

	if uint64(len(b.VoteSlashings)) > netParams.MaxVoteSlashingsPerBlock {
		return fmt.Errorf("block has too many vote slashings (max: %d, got: %d)", netParams.MaxVoteSlashingsPerBlock, len(b.VoteSlashings))
	}

	if uint64(len(b.ProposerSlashings)) > netParams.MaxProposerSlashingsPerBlock {
		return fmt.Errorf("block has too many proposer slashings (max: %d, got: %d)", netParams.MaxProposerSlashingsPerBlock, len(b.ProposerSlashings))
	}

	if uint64(len(b.CoinProofs)) > netParams.MaxCoinProofsPerBlock {
		return fmt.Errorf("block has too many migration proofs (max: %d, got: %d)", netParams.MaxCoinProofsPerBlock, len(b.CoinProofs))
	}

	if len(b.Deposits) > 0 {
		if err := s.ApplyDeposits(b.Deposits); err != nil {
			return err
		}
	}

	if len(b.Txs) > 0 {
		if err := s.ApplyTransactionsSingle(b.Txs, b.Header.FeeAddress); err != nil {
			return err
		}
	}

	for _, vote := range b.GovernanceVotes {
		if err := s.ProcessGovernanceVote(vote); err != nil {
			return err
		}
	}

	for _, p := range b.CoinProofs {
		if err := s.ApplyCoinProof(p); err != nil {
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

	for _, e := range b.Exits {
		if err := s.ApplyExit(e); err != nil {
			return err
		}
	}

	for _, rs := range b.RANDAOSlashings {
		if err := s.ApplyRANDAOSlashing(rs); err != nil {
			return err
		}
	}

	for _, vs := range b.VoteSlashings {
		if err := s.ApplyVoteSlashing(vs); err != nil {
			return err
		}
	}

	for _, ps := range b.ProposerSlashings {
		if err := s.ApplyProposerSlashing(ps); err != nil {
			return err
		}
	}

	for i := range s.NextRANDAO {
		s.NextRANDAO[i] ^= b.RandaoSignature[i]
	}

	return nil
}
