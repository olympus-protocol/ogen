package primitives

import (
	"fmt"
	"io"

	"github.com/olympus-protocol/ogen/utils/serializer"
)

// CoinsState is the state that we
type CoinsState struct {
	Balances map[[20]byte]uint64
	Nonces   map[[20]byte]uint64
}

// ApplyTransactionSingle applies a transaction to the coin state.
func (u *CoinsState) ApplyTransactionSingle(tx *TransferSinglePayload, blockWithdrawalAddress [20]byte) error {
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

// ApplyTransaction applies a transaction to the coin state.
func (u *CoinsState) ApplyTransactionMulti(tx *TransferMultiPayload, blockWithdrawalAddress [20]byte) error {
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

// Copy copies CoinsState and returns a new one.
func (u *CoinsState) Copy() CoinsState {
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

func (u *CoinsState) Serialize(w io.Writer) error {
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

func (u *CoinsState) Deserialize(r io.Reader) error {
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
