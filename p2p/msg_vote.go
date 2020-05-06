package p2p

import (
	"bytes"
	"io"

	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
)

const (
	MaxVotesInMsgVotes = 5000
	maxVoteSize        = (100 + primitives.MaxVoteDataSize) * MaxVotesInMsgVotes
)

type MsgVotes struct {
	Votes []primitives.SingleValidatorVote
}

func (m *MsgVotes) Encode(w io.Writer) error {
	err := serializer.WriteVarInt(w, uint64(len(m.Votes)))
	if err != nil {
		return err
	}
	for _, v := range m.Votes {
		if err := v.Encode(w); err != nil {
			return err
		}
	}
	return nil
}

func (m *MsgVotes) Decode(r io.Reader) error {
	numVotes, err := serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	m.Votes = make([]primitives.SingleValidatorVote, numVotes)
	for i := range m.Votes {
		if err := m.Votes[i].Decode(r); err != nil {
			return err
		}
	}

	return nil
}

func (m *MsgVotes) Hash() (chainhash.Hash, error) {
	buf := bytes.NewBuffer([]byte{})
	err := m.Encode(buf)
	if err != nil {
		return chainhash.Hash{}, err
	}
	return chainhash.DoubleHashH(buf.Bytes()), nil
}

func (m *MsgVotes) Command() string {
	return MsgVoteCmd
}

func (m *MsgVotes) MaxPayloadLength() uint32 {
	return maxVoteSize
}
