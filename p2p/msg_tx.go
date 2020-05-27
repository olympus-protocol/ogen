package p2p

import (
	"bytes"
	"io"

	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
)

const (
	maxTxSize = 1024 * 300 // 300 KB
)

type MsgTx struct {
	Txs []primitives.TransferSinglePayload
}

func (m *MsgTx) Encode(w io.Writer) error {
	if err := serializer.WriteVarInt(w, uint64(len(m.Txs))); err != nil {
		return err
	}

	for _, tx := range m.Txs {
		if err := tx.Encode(w); err != nil {
			return err
		}
	}

	return nil
}

func (m *MsgTx) Decode(r io.Reader) error {
	numTxs, err := serializer.ReadVarInt(r)
	if err != nil {
		return err
	}

	m.Txs = make([]primitives.TransferSinglePayload, numTxs)
	for i := range m.Txs {
		if err := m.Txs[i].Decode(r); err != nil {
			return err
		}
	}

	return nil
}

func (m *MsgTx) TxHash() (chainhash.Hash, error) {
	buf := bytes.NewBuffer([]byte{})
	err := m.Encode(buf)
	if err != nil {
		return chainhash.Hash{}, err
	}
	return chainhash.DoubleHashH(buf.Bytes()), nil
}

func (m *MsgTx) Command() string {
	return MsgTxCmd
}

func (m *MsgTx) MaxPayloadLength() uint32 {
	return maxTxSize
}
