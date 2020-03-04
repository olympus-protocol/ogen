package p2p

import (
	"bytes"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
	"io"
)

const (
	maxTxSize = 1024 * 300 // 300 KB
)


type MsgTx struct {
	primitives.Tx
}

type OutPoint struct {
	TxHash chainhash.Hash
	Index  int64
}

func (o *OutPoint) IsNull() bool {
	zeroHash := chainhash.Hash{}
	if o.TxHash == zeroHash && o.Index == 0 {
		return true
	}
	return false
}

func (o *OutPoint) Serialize(w io.Writer) error {
	err := serializer.WriteElements(w, o.TxHash, o.Index)
	if err != nil {
		return err
	}
	return nil
}

func (o *OutPoint) Deserialize(r io.Reader) error {
	err := serializer.ReadElements(r, &o.TxHash, &o.Index)
	if err != nil {
		return err
	}
	return nil
}

func (o *OutPoint) Hash() (chainhash.Hash, error) {
	buf := bytes.NewBuffer([]byte{})
	err := o.Serialize(buf)
	if err != nil {
		return chainhash.Hash{}, err
	}
	return chainhash.DoubleHashH(buf.Bytes()), nil
}

func NewOutPoint(hash chainhash.Hash, index int64) *OutPoint {
	return &OutPoint{
		TxHash: hash,
		Index:  index,
	}
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

func (m *MsgTx) AddPayload(payload []byte) {
	m.Payload = payload
}
