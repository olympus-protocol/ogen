package primitives

import (
	"bytes"
	"io"

	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
)

type TxType = int32

const (
	TxCoins TxType = iota + 1
	TxWorker
	TxGovernance
	TxVotes
	TxUsers
)

type TxAction = int32

const (
	Transfer TxAction = iota + 1
	Revoke
	Update
	Upload
	Generate
	Pay
)

type Tx struct {
	Time      int64
	TxVersion int32
	TxType    TxType
	TxAction  TxAction
	Payload   []byte
}

func (t *Tx) Encode(w io.Writer) error {
	err := serializer.WriteElements(w, t.TxVersion, t.TxType, t.TxAction, t.Time)
	if err != nil {
		return err
	}
	err = serializer.WriteVarBytes(w, t.Payload)
	if err != nil {
		return err
	}
	return nil
}

func (t *Tx) Decode(r io.Reader) error {
	err := serializer.ReadElements(r, &t.TxVersion, &t.TxType, &t.TxAction, &t.Time)
	if err != nil {
		return err
	}
	t.Payload, err = serializer.ReadVarBytes(r)
	if err != nil {
		return err
	}
	return nil
}

func (t *Tx) Hash() chainhash.Hash {
	buf := bytes.NewBuffer([]byte{})
	_ = t.Encode(buf)
	return chainhash.DoubleHashH(buf.Bytes())
}
