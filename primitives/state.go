package primitives

import (
	"bytes"
	"io"

	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
)

// LastBlockHashesSize is the size of the last block hashes.
const LastBlockHashesSize = 8

// State is the state of consensus in the blockchain.
type State struct {
	CoinsState CoinsState
	UserState  UserState

	// ValidatorRegistry keeps track of validators in the state.
	ValidatorRegistry []Validator

	// LatestValidatorRegistryChange keeps track of the last time the validator
	// registry was changed. We only want to update the registry if a block was
	// finalized since the last time it was changed, so we keep track of that
	// here.
	LatestValidatorRegistryChange uint64

	// RANDAO for figuring out the proposer queue. We don't want any one validator
	// to have influence over the RANDAO, so we have each proposer contribute.
	RANDAO chainhash.Hash

	// NextRANDAO is the RANDAO currently being created. Every time a block is
	// created, we XOR the 32 least-significant bytes of the RandaoReveal with this
	// value to update it.
	NextRANDAO chainhash.Hash

	// Slot is the last slot ProcessSlot was called for.
	Slot uint64

	// EpochIndex is the last epoch ProcessEpoch was called for.
	EpochIndex uint64

	// ProposerQueue is the queue of validators scheduled to create a block.
	ProposerQueue []uint32

	PreviousEpochVoteAssignments []uint32
	CurrentEpochVoteAssignments  []uint32

	// NextProposerQueue is the queue of validators scheduled to create a block
	// in the next epoch.
	NextProposerQueue []uint32

	// JustifiedBitfield is a bitfield where the nth least significant bit
	// represents whether the nth last epoch was justified.
	JustificationBitfield uint64

	// FinalizedEpoch is the epoch that was finalized.
	FinalizedEpoch uint64

	// LastBlockHashes is the last LastBlockHashesSize block hashes.
	LatestBlockHashes []chainhash.Hash

	// JustifiedEpoch is the last epoch that >2/3 of validators voted for.
	JustifiedEpoch uint64

	// JustifiedEpochHash is the block hash of the last epoch that >2/3 of validators voted for.
	JustifiedEpochHash chainhash.Hash

	// CurrentEpochVotes are votes that are being submitted where
	// the source epoch matches justified epoch.
	CurrentEpochVotes []AcceptedVoteInfo

	// PreviousJustifiedEpoch is the second-to-last epoch that >2/3 of validators
	// voted for.
	PreviousJustifiedEpoch uint64

	// PreviousJustifiedEpochHash is the block hash of the last epoch that >2/3 of validators voted for.
	PreviousJustifiedEpochHash chainhash.Hash

	// PreviousEpochVotes are votes where the FromEpoch matches PreviousJustifiedEpoch.
	PreviousEpochVotes []AcceptedVoteInfo
}

// GetValidatorIndicesActiveAt gets validator indices where the validator is active at a certain slot.
func (s *State) GetValidatorIndicesActiveAt(epoch int64) []uint32 {
	vals := make([]uint32, 0, len(s.ValidatorRegistry))
	for i, v := range s.ValidatorRegistry {
		if v.IsActiveAtEpoch(epoch) {
			vals = append(vals, uint32(i))
		}
	}

	return vals
}

// Serialize serializes the state to the writer.
func (s *State) Serialize(w io.Writer) error {
	if err := serializer.WriteElements(w,
		s.Slot,
		s.EpochIndex,
		s.JustifiedEpoch,
		s.FinalizedEpoch,
		s.JustificationBitfield,
		s.PreviousJustifiedEpoch,
		s.JustifiedEpochHash,
		s.PreviousJustifiedEpochHash,
		s.LatestValidatorRegistryChange,
		s.RANDAO,
		s.NextRANDAO); err != nil {
		return err
	}
	if err := s.CoinsState.Serialize(w); err != nil {
		return err
	}
	if err := s.UserState.Serialize(w); err != nil {
		return err
	}
	if err := serializer.WriteVarInt(w, uint64(len(s.ValidatorRegistry))); err != nil {
		return err
	}
	for _, v := range s.ValidatorRegistry {
		if err := v.Serialize(w); err != nil {
			return err
		}
	}
	if err := serializer.WriteVarInt(w, uint64(len(s.ProposerQueue))); err != nil {
		return err
	}
	for _, p := range s.ProposerQueue {
		if err := serializer.WriteElement(w, p); err != nil {
			return err
		}
	}
	if err := serializer.WriteVarInt(w, uint64(len(s.NextProposerQueue))); err != nil {
		return err
	}
	for _, p := range s.NextProposerQueue {
		if err := serializer.WriteElement(w, p); err != nil {
			return err
		}
	}
	if err := serializer.WriteVarInt(w, uint64(len(s.CurrentEpochVoteAssignments))); err != nil {
		return err
	}
	for _, p := range s.CurrentEpochVoteAssignments {
		if err := serializer.WriteElement(w, p); err != nil {
			return err
		}
	}
	if err := serializer.WriteVarInt(w, uint64(len(s.PreviousEpochVoteAssignments))); err != nil {
		return err
	}
	for _, p := range s.PreviousEpochVoteAssignments {
		if err := serializer.WriteElement(w, p); err != nil {
			return err
		}
	}
	if err := serializer.WriteVarInt(w, uint64(len(s.LatestBlockHashes))); err != nil {
		return err
	}
	for _, p := range s.LatestBlockHashes {
		if err := serializer.WriteElement(w, p); err != nil {
			return err
		}
	}
	if err := serializer.WriteVarInt(w, uint64(len(s.CurrentEpochVotes))); err != nil {
		return err
	}
	for _, p := range s.CurrentEpochVotes {
		if err := p.Serialize(w); err != nil {
			return err
		}
	}
	if err := serializer.WriteVarInt(w, uint64(len(s.PreviousEpochVotes))); err != nil {
		return err
	}
	for _, p := range s.PreviousEpochVotes {
		if err := p.Serialize(w); err != nil {
			return err
		}
	}
	return nil
}

// Deserialize deserializes state from the reader.
func (s *State) Deserialize(r io.Reader) error {
	if err := serializer.ReadElements(r,
		&s.Slot,
		&s.EpochIndex,
		&s.JustifiedEpoch,
		&s.FinalizedEpoch,
		&s.JustificationBitfield,
		&s.PreviousJustifiedEpoch,
		&s.JustifiedEpochHash,
		&s.PreviousJustifiedEpochHash,
		&s.LatestValidatorRegistryChange,
		&s.RANDAO,
		&s.NextRANDAO); err != nil {
		return err
	}
	if err := s.CoinsState.Deserialize(r); err != nil {
		return err
	}
	if err := s.UserState.Deserialize(r); err != nil {
		return err
	}
	num, err := serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	s.ValidatorRegistry = make([]Validator, num)
	for i := range s.ValidatorRegistry {
		if err := s.ValidatorRegistry[i].Deserialize(r); err != nil {
			return err
		}
	}
	num, err = serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	s.ProposerQueue = make([]uint32, num)
	for i := range s.ProposerQueue {
		if err := serializer.ReadElement(r, &s.ProposerQueue[i]); err != nil {
			return err
		}
	}
	num, err = serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	s.NextProposerQueue = make([]uint32, num)
	for i := range s.NextProposerQueue {
		if err := serializer.ReadElement(r, &s.NextProposerQueue[i]); err != nil {
			return err
		}
	}
	num, err = serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	s.CurrentEpochVoteAssignments = make([]uint32, num)
	for i := range s.CurrentEpochVoteAssignments {
		if err := serializer.ReadElement(r, &s.CurrentEpochVoteAssignments[i]); err != nil {
			return err
		}
	}
	num, err = serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	s.PreviousEpochVoteAssignments = make([]uint32, num)
	for i := range s.PreviousEpochVoteAssignments {
		if err := serializer.ReadElement(r, &s.PreviousEpochVoteAssignments[i]); err != nil {
			return err
		}
	}
	num, err = serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	s.LatestBlockHashes = make([]chainhash.Hash, num)
	for i := range s.LatestBlockHashes {
		if err := serializer.ReadElement(r, &s.LatestBlockHashes[i]); err != nil {
			return err
		}
	}
	num, err = serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	s.CurrentEpochVotes = make([]AcceptedVoteInfo, num)
	for i := range s.CurrentEpochVotes {
		if err := s.CurrentEpochVotes[i].Deserialize(r); err != nil {
			return err
		}
	}
	num, err = serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	s.PreviousEpochVotes = make([]AcceptedVoteInfo, num)
	for i := range s.PreviousEpochVotes {
		if err := s.PreviousEpochVotes[i].Deserialize(r); err != nil {
			return err
		}
	}
	return nil
}

// Hash calculates the hash of the state.
func (s *State) Hash() chainhash.Hash {
	buf := bytes.NewBuffer([]byte{})
	_ = s.Serialize(buf)
	return chainhash.HashH(buf.Bytes())
}

// Copy returns a copy of the state.
func (s *State) Copy() State {
	s2 := *s

	s2.CoinsState = s.CoinsState.Copy()
	s2.UserState = s.UserState.Copy()
	s2.ValidatorRegistry = make([]Validator, len(s.ValidatorRegistry))

	for i, c := range s.ValidatorRegistry {
		s2.ValidatorRegistry[i] = c.Copy()
	}

	s2.ProposerQueue = make([]uint32, len(s.ProposerQueue))
	for i, c := range s.ProposerQueue {
		s2.ProposerQueue[i] = c
	}

	s2.NextProposerQueue = make([]uint32, len(s.NextProposerQueue))
	for i, c := range s.NextProposerQueue {
		s2.NextProposerQueue[i] = c
	}

	s2.CurrentEpochVoteAssignments = make([]uint32, len(s.CurrentEpochVoteAssignments))
	for i, c := range s.CurrentEpochVoteAssignments {
		s2.CurrentEpochVoteAssignments[i] = c
	}

	s2.PreviousEpochVoteAssignments = make([]uint32, len(s.PreviousEpochVoteAssignments))
	for i, c := range s.PreviousEpochVoteAssignments {
		s2.PreviousEpochVoteAssignments[i] = c
	}

	s2.LatestBlockHashes = make([]chainhash.Hash, len(s.LatestBlockHashes))
	for i, c := range s.LatestBlockHashes {
		s2.LatestBlockHashes[i] = c
	}

	s2.CurrentEpochVotes = make([]AcceptedVoteInfo, len(s.CurrentEpochVotes))
	for i, c := range s.CurrentEpochVotes {
		s2.CurrentEpochVotes[i] = c.Copy()
	}

	s2.PreviousEpochVotes = make([]AcceptedVoteInfo, len(s.PreviousEpochVotes))
	for i, c := range s.PreviousEpochVotes {
		s2.PreviousEpochVotes[i] = c.Copy()
	}

	return s2
}

// AccountTxs is just a helper struct for database storage of account transactions.
type AccountTxs struct {
	TxsAmount uint64
	Txs       []chainhash.Hash
}

// Encode serializes the data into a writer.
func (atxs *AccountTxs) Encode(w io.Writer) error {
	err := serializer.WriteVarInt(w, atxs.TxsAmount)
	if err != nil {
		return err
	}
	for _, tx := range atxs.Txs {
		err := serializer.WriteElement(w, tx)
		if err != nil {
			return err
		}
	}
	return nil
}

// Decode deserializes the data into a reader.
func (atxs *AccountTxs) Decode(r io.Reader) error {
	txs, err := serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	atxs.TxsAmount = txs
	atxs.Txs = make([]chainhash.Hash, txs)
	for i := range atxs.Txs {
		err := serializer.ReadElement(r, &atxs.Txs[i])
		if err != nil {
			return err
		}
	}
	return nil
}
