package p2p

import (
	"fmt"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
	"io"
)

type MsgProposal struct {
	Height        uint64
	Round         uint64
	BlockProposal primitives.Block
	ValidatorID   chainhash.Hash
	ValidRound    int64
	Signature     [96]byte
}

func (m *MsgProposal) Decode(r io.Reader) error {
	err := serializer.ReadElements(r, &m.Height, &m.Round, &m.ValidRound, &m.Signature, &m.ValidatorID)
	if err != nil {
		return err
	}

	return m.BlockProposal.Decode(r)
}

func (m *MsgProposal) String() string {
	return fmt.Sprintf("proposal(height=%d, round=%d, vr=%d, hash=%s) from %s", m.Height, m.Round, m.ValidRound, m.BlockProposal.Hash(), m.ValidatorID)
}

func (m *MsgProposal) Encode(w io.Writer) error {
	err := serializer.WriteElements(w, m.Height, m.Round, m.ValidRound, m.Signature, m.ValidatorID)
	if err != nil {
		return err
	}

	return m.BlockProposal.Encode(w)
}

func (m *MsgProposal) Command() string {
	return "proposal"
}

func (m *MsgProposal) Hash() chainhash.Hash {
	m2 := *m
	m2.Signature = [96]byte{}
	h := serializer.Hash(&m2)
	return *h
}

const maxMsgProposalSize = maxBlockSize + 8 + 8 + 8 + 96

func (m *MsgProposal) MaxPayloadLength() uint32 {
	return maxMsgProposalSize
}

var _ Message = &MsgProposal{}
