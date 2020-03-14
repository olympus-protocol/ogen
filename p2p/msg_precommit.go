package p2p

import (
	"fmt"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
	"io"
)

type MsgPrecommit struct {
	Height      uint64
	Round       uint64
	BlockHash   chainhash.Hash
	ValidatorID chainhash.Hash
	Signature   [96]byte
}

func (m *MsgPrecommit) Decode(r io.Reader) error {
	return serializer.ReadElements(r, &m.Height, &m.Round, &m.BlockHash, &m.Signature, &m.ValidatorID)
}

func (m *MsgPrecommit) Encode(w io.Writer) error {
	return serializer.WriteElements(w, m.Height, m.Round, m.BlockHash, m.Signature, m.ValidatorID)
}

func (m *MsgPrecommit) String() string {
	return fmt.Sprintf("precommit(height=%d, round=%d, hash=%s) from %s", m.Height, m.Round, m.BlockHash, m.ValidatorID)
}

func (*MsgPrecommit) Command() string {
	return "precommit"
}

// MaxMsgPrecommitSize is height (8), round (8), block hash (32), signature (96).
const MaxMsgPrecommitSize = 8 + 8 + 32 + 96

func (*MsgPrecommit) MaxPayloadLength() uint32 {
	return MaxMsgPrecommitSize
}

func (m *MsgPrecommit) Hash() chainhash.Hash {
	m2 := *m
	m2.Signature = [96]byte{}
	h := serializer.Hash(&m2)
	return *h
}

var _ Message = &MsgPrecommit{}
