package primitives

import (
	"bytes"
	"fmt"
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

// UtxoState is the state that we
type UtxoState struct {
	Balances map[[20]byte]uint64
	Nonces   map[[20]byte]uint64
}

// ApplyTransaction applies a transaction to the coin state.
func (u *UtxoState) ApplyTransaction(tx *CoinPayload, blockWithdrawalAddress [20]byte) error {
	pkh := tx.FromPubkeyHash()
	if u.Balances[pkh] < tx.Amount+tx.Fee {
		return fmt.Errorf("insufficient balance of %d for %d transaction", u.Balances[pkh], tx.Amount)
	}

	if u.Nonces[pkh] >= tx.Nonce {
		return fmt.Errorf("nonce is too small (already processed: %d, trying: %d)", u.Nonces[pkh], tx.Nonce)
	}

	if err := tx.VerifySig(); err != nil {
		return err
	}

	u.Balances[pkh] -= tx.Amount + tx.Fee
	u.Balances[tx.To] += tx.Amount
	u.Balances[blockWithdrawalAddress] += tx.Fee
	u.Nonces[pkh] = tx.Nonce
	return nil
}

// Copy copies UtxoState and returns a new one.
func (u *UtxoState) Copy() UtxoState {
	u2 := *u
	u2.Balances = make(map[[20]byte]uint64)
	u2.Nonces = make(map[[20]byte]uint64)
	for i, c := range u.Balances {
		u2.Balances[i] = c
	}
	for i, c := range u.Nonces {
		u2.Nonces[i] = c
	}
	return u2
}

func (u *UtxoState) Serialize(w io.Writer) error {
	if err := serializer.WriteVarInt(w, uint64(len(u.Balances))); err != nil {
		return err
	}

	for h, b := range u.Balances {
		if _, err := w.Write(h[:]); err != nil {
			return err
		}

		if err := serializer.WriteElement(w, b); err != nil {
			return err
		}
	}

	if err := serializer.WriteVarInt(w, uint64(len(u.Nonces))); err != nil {
		return err
	}

	for h, b := range u.Nonces {
		if _, err := w.Write(h[:]); err != nil {
			return err
		}

		if err := serializer.WriteElement(w, b); err != nil {
			return err
		}
	}

	return nil
}

func (u *UtxoState) Deserialize(r io.Reader) error {
	if u.Balances == nil {
		u.Balances = make(map[[20]byte]uint64)
	}
	if u.Nonces == nil {
		u.Nonces = make(map[[20]byte]uint64)
	}

	numBalances, err := serializer.ReadVarInt(r)
	if err != nil {
		return err
	}

	for i := uint64(0); i < numBalances; i++ {
		var hash [20]byte
		if _, err := r.Read(hash[:]); err != nil {
			return err
		}

		var balance uint64
		if err := serializer.ReadElement(r, &balance); err != nil {
			return err
		}

		u.Balances[hash] = balance
	}

	numNonces, err := serializer.ReadVarInt(r)
	if err != nil {
		return err
	}

	for i := uint64(0); i < numNonces; i++ {
		var hash [20]byte
		if _, err := r.Read(hash[:]); err != nil {
			return err
		}

		var nonce uint64
		if err := serializer.ReadElement(r, &nonce); err != nil {
			return err
		}

		u.Nonces[hash] = nonce
	}

	return nil
}
