package primitives

import (
	"errors"
	"fmt"
	"io"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/params"
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

	// PreviousJustifiedEpoch is the second-to-last epoch that >2/3 of validators
	// voted for.
	PreviousJustifiedEpoch uint64

	// JustifiedEpoch is the last epoch that >2/3 of validators voted for.
	JustifiedEpoch uint64

	// JustifiedBitfield is a bitfield where the nth least significant bit
	// represents whether the nth last epoch was justified.
	JustificationBitfield uint64

	// FinalizedEpoch is the epoch that was finalized.
	FinalizedEpoch uint64

	// LastBlockHashes is the last LastBlockHashesSize block hashes.
	LatestBlockHashes []chainhash.Hash

	// CurrentEpochVotes are votes that are being submitted where
	// the source epoch matches justified epoch.
	CurrentEpochVotes []AcceptedVoteInfo

	// PreviousEpochVotes are votes that are being submitted where the
	// source epoch matches the previous justified epoch.
	PreviousEpochVotes []AcceptedVoteInfo
}

// Serialize serializes the state to the writer.
func (s *State) Serialize(w io.Writer) error {
	if err := serializer.WriteElements(w, s.Slot, s.EpochIndex, s.PreviousJustifiedEpoch, s.JustifiedEpoch, s.FinalizedEpoch, s.JustificationBitfield); err != nil {
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

func (s *State) Deserialize(r io.Reader) error {
	if err := serializer.ReadElements(r, &s.Slot, &s.EpochIndex, &s.PreviousJustifiedEpoch, &s.JustifiedEpoch, &s.FinalizedEpoch, &s.JustificationBitfield); err != nil {
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

// ProcessBlock processes a block in state.
func (s *State) ProcessBlock(p *params.ChainParams, b *Block) error {
	if b.Header.Slot != s.Slot {
		return fmt.Errorf("state is not updated to slot %d, instead got %d", b.Header.Slot, s.Slot)
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
	proposer := s.WorkerState.Get(proposerIndex)

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

	return nil
}
