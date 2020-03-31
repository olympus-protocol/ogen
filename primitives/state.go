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
	UtxoState       UtxoState
	GovernanceState GovernanceState
	UserState       UserState
	WorkerState     WorkerState

	// Slot is the last slot ProcessSlot was called for.
	Slot uint64

	// EpochIndex is the last epoch ProcessEpoch was called for.
	EpochIndex uint64

	// ProposerQueue is the queue of validators scheduled to create a block.
	ProposerQueue []chainhash.Hash

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

// Serialize serializes the state to the writer.
func (s *State) Serialize(w io.Writer) error {
	if err := serializer.WriteElements(w, s.Slot, s.EpochIndex, s.JustifiedEpoch, s.FinalizedEpoch, s.JustificationBitfield, s.PreviousJustifiedEpoch, s.JustifiedEpochHash, s.PreviousJustifiedEpochHash); err != nil {
		return err
	}
	if err := s.UtxoState.Serialize(w); err != nil {
		return err
	}
	if err := s.GovernanceState.Serialize(w); err != nil {
		return err
	}
	if err := s.UserState.Serialize(w); err != nil {
		return err
	}
	if err := s.WorkerState.Serialize(w); err != nil {
		return err
	}
	if err := serializer.WriteVarInt(w, uint64(len(s.ProposerQueue))); err != nil {
		return err
	}
	for _, p := range s.ProposerQueue {
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
	if err := serializer.ReadElements(r, &s.Slot, &s.EpochIndex, &s.JustifiedEpoch, &s.FinalizedEpoch, &s.JustificationBitfield, &s.PreviousJustifiedEpoch, &s.JustifiedEpochHash, &s.PreviousJustifiedEpochHash); err != nil {
		return err
	}
	if err := s.UtxoState.Deserialize(r); err != nil {
		return err
	}
	if err := s.GovernanceState.Deserialize(r); err != nil {
		return err
	}
	if err := s.UserState.Deserialize(r); err != nil {
		return err
	}
	if err := s.WorkerState.Deserialize(r); err != nil {
		return err
	}
	num, err := serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	s.ProposerQueue = make([]chainhash.Hash, num)
	for i := range s.ProposerQueue {
		if err := serializer.ReadElement(r, &s.ProposerQueue[i]); err != nil {
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
