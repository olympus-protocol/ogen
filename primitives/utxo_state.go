package primitives

import (
	"bytes"
	"io"

	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
)

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

type Utxo struct {
	OutPoint          OutPoint
	PrevInputsPubKeys [][48]byte
	Owner             string
	Amount            int64
}

// Serialize serializes the UtxoRow to a writer.
func (l *Utxo) Serialize(w io.Writer) error {
	err := l.OutPoint.Serialize(w)
	if err != nil {
		return err
	}
	err = serializer.WriteVarString(w, l.Owner)
	if err != nil {
		return err
	}
	err = serializer.WriteVarInt(w, uint64(len(l.PrevInputsPubKeys)))
	if err != nil {
		return err
	}
	for _, pub := range l.PrevInputsPubKeys {
		err = serializer.WriteElements(w, pub)
		if err != nil {
			return err
		}
	}
	err = serializer.WriteVarInt(w, uint64(l.Amount))
	if err != nil {
		return err
	}
	return nil
}

// Deserialize deserializes a UtxoRow from a reader.
func (l *Utxo) Deserialize(r io.Reader) error {
	err := l.OutPoint.Deserialize(r)
	if err != nil {
		return err
	}
	l.Owner, err = serializer.ReadVarString(r)
	if err != nil {
		return err
	}
	count, err := serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	l.PrevInputsPubKeys = make([][48]byte, 0, count)
	for i := uint64(0); i < count; i++ {
		var pubKey [48]byte
		err = serializer.ReadElement(r, &pubKey)
		if err != nil {
			return err
		}
		l.PrevInputsPubKeys = append(l.PrevInputsPubKeys, pubKey)
	}
	amount, err := serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	l.Amount = int64(amount)
	return nil
}

func (l *Utxo) Hash() chainhash.Hash {
	buf := bytes.NewBuffer([]byte{})
	_ = l.OutPoint.Serialize(buf)
	return chainhash.DoubleHashH(buf.Bytes())
}

type UtxoState struct {
	UTXOs map[chainhash.Hash]Utxo
}

// Have checks if a UTXO exists.
func (u *UtxoState) Have(c chainhash.Hash) bool {
	_, found := u.UTXOs[c]
	return found
}

// Get gets the UTXO from state.
func (u *UtxoState) Get(c chainhash.Hash) Utxo {
	return u.UTXOs[c]
}

func (u *UtxoState) Serialize(w io.Writer) error {
	if err := serializer.WriteVarInt(w, uint64(len(u.UTXOs))); err != nil {
		return err
	}

	for h, utxo := range u.UTXOs {
		if _, err := w.Write(h[:]); err != nil {
			return err
		}

		if err := utxo.Serialize(w); err != nil {
			return err
		}
	}

	return nil
}

func (u *UtxoState) Deserialize(r io.Reader) error {
	if u.UTXOs == nil {
		u.UTXOs = make(map[chainhash.Hash]Utxo)
	}

	numUtxos, err := serializer.ReadVarInt(r)

	if err != nil {
		return err
	}

	for i := uint64(0); i < numUtxos; i++ {
		var hash chainhash.Hash
		if _, err := r.Read(hash[:]); err != nil {
			return err
		}

		var utxo Utxo
		if err := utxo.Deserialize(r); err != nil {
			return err
		}

		u.UTXOs[hash] = utxo
	}

	return nil
}
