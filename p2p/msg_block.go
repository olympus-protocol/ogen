package p2p

import (
	"bytes"
	"fmt"
	"github.com/olympus-protocol/ogen/utils/serializer"
	"io"
)

const maxBlockSize = 1024 * 512 // 512 KB

type MsgBlock struct {
	Header    BlockHeader
	Txs       []*MsgTx
	PubKey    [48]byte
	Signature [96]byte
}

func (m *MsgBlock) AddTx(tx *MsgTx) {
	m.Txs = append(m.Txs, tx)
}

func (m *MsgBlock) Encode(w io.Writer) error {
	err := m.Header.Serialize(w)
	if err != nil {
		return err
	}
	err = serializer.WriteVarInt(w, uint64(len(m.Txs)))
	if err != nil {
		return err
	}
	for _, tx := range m.Txs {
		err := tx.Encode(w)
		if err != nil {
			return err
		}
	}
	err = serializer.WriteElements(w, m.PubKey, m.Signature)
	if err != nil {
		return err
	}
	return nil
}

func (m *MsgBlock) Decode(r io.Reader) error {
	buf, ok := r.(*bytes.Buffer)
	if !ok {
		return fmt.Errorf("MsgBlock.Decode reader is not a " +
			"*bytes.Buffer")
	}
	err := m.Header.Deserialize(r)
	if err != nil {
		return err
	}
	txCount, err := serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	for i := uint64(0); i < txCount; i++ {
		var tx MsgTx
		err := tx.Decode(r)
		if err != nil {
			return err
		}
		m.AddTx(&tx)
	}
	err = serializer.ReadElements(buf, &m.PubKey, &m.Signature)
	if err != nil {
		return err
	}
	return nil
}

func (m *MsgBlock) Command() string {
	return MsgBlockCmd
}

func (m *MsgBlock) MaxPayloadLength() uint32 {
	return maxBlockSize
}

func NewMsgBlock(header BlockHeader, blockSign [96]byte) *MsgBlock {
	m := &MsgBlock{
		Header:    header,
		Signature: blockSign,
	}
	return m
}
