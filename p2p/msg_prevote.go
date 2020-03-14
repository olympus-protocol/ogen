package p2p

import (
	"fmt"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
	"io"
)

type MsgPrevote struct {
	Height      uint64
	Round       uint64
	BlockHash   chainhash.Hash
	ValidatorID chainhash.Hash
	Signature   [96]byte
}

func (m *MsgPrevote) Decode(r io.Reader) error {
	return serializer.ReadElements(r, &m.Height, &m.Round, &m.BlockHash, &m.Signature, &m.ValidatorID)
}

func (m *MsgPrevote) Encode(w io.Writer) error {
	return serializer.WriteElements(w, m.Height, m.Round, m.BlockHash, m.Signature, m.ValidatorID)
}

func (m *MsgPrevote) String() string {
	return fmt.Sprintf("prevote(height=%d, round=%d, hash=%s) from %s", m.Height, m.Round, m.BlockHash, m.ValidatorID)
}

func (m *MsgPrevote) Command() string {
	return "prevote"
}

func (m *MsgPrevote) Hash() chainhash.Hash {
	m2 := *m
	m2.Signature = [96]byte{}
	h := serializer.Hash(&m2)
	return *h
}

// MaxMsgPrevoteSize is height (8), round (8), block hash (32), signature (96).
const MaxMsgPrevoteSize = 8 + 8 + 32 + 96

func (m *MsgPrevote) MaxPayloadLength() uint32 {
	return MaxMsgPrevoteSize
}

var _ Message = &MsgPrevote{}
