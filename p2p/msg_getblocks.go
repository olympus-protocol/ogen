package p2p

import (
	"bytes"
	"fmt"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
	"io"
)

type MsgGetBlocks struct {
	LastBlockHash chainhash.Hash
}

func (m *MsgGetBlocks) Encode(w io.Writer) error {
	err := serializer.WriteElement(w, m.LastBlockHash)
	if err != nil {
		return err
	}
	return nil
}

func (m *MsgGetBlocks) Decode(r io.Reader) error {
	buf, ok := r.(*bytes.Buffer)
	if !ok {
		return fmt.Errorf("MsgVersion.Decode reader is not a " +
			"*bytes.Buffer")
	}
	err := serializer.ReadElement(buf, &m.LastBlockHash)
	if err != nil {
		return err
	}
	return nil
}

func (m *MsgGetBlocks) Command() string {
	return MsgGetBlocksCmd
}

func (m *MsgGetBlocks) MaxPayloadLength() uint32 {
	return chainhash.HashSize
}

func NewMsgGetBlock(hash chainhash.Hash) *MsgGetBlocks {
	m := &MsgGetBlocks{
		LastBlockHash: hash,
	}
	return m
}
