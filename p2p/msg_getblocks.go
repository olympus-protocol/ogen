package p2p

import (
	"bytes"
	"fmt"
	"io"

	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
)

type MsgGetBlocks struct {
	HashStop      chainhash.Hash
	LocatorHashes []chainhash.Hash
}

func (m *MsgGetBlocks) Encode(w io.Writer) error {
	err := serializer.WriteElement(w, m.HashStop)
	if err != nil {
		return err
	}

	err = serializer.WriteVarInt(w, uint64(len(m.LocatorHashes)))
	if err != nil {
		return err
	}

	for _, c := range m.LocatorHashes {
		err = serializer.WriteElement(w, c)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *MsgGetBlocks) Decode(r io.Reader) error {
	buf, ok := r.(*bytes.Buffer)
	if !ok {
		return fmt.Errorf("MsgVersion.Decode reader is not a " +
			"*bytes.Buffer")
	}
	err := serializer.ReadElement(buf, &m.HashStop)
	if err != nil {
		return err
	}
	numHashes, err := serializer.ReadVarInt(r)
	if err != nil {
		return err
	}

	m.LocatorHashes = make([]chainhash.Hash, numHashes)
	for i := range m.LocatorHashes {
		err = serializer.ReadElement(r, &m.LocatorHashes[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *MsgGetBlocks) Command() string {
	return MsgGetBlocksCmd
}

func (m *MsgGetBlocks) MaxPayloadLength() uint32 {
	return chainhash.HashSize + 40*chainhash.HashSize + 9
}

func NewMsgGetBlock(hashStop chainhash.Hash, locatorHashes []chainhash.Hash) *MsgGetBlocks {
	m := &MsgGetBlocks{
		HashStop:      hashStop,
		LocatorHashes: locatorHashes,
	}
	return m
}
