package wallet

import (
	"fmt"
	"github.com/olympus-protocol/ogen/pkg/bech32"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/p2p"
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

// GetBalance returns the balance of the current open wallet.
func (w *wallet) GetBalance() (uint64, uint64, error) {
	if !w.open {
		return 0, 0, errorNotOpen
	}
	acc, err := w.GetAccountRaw()
	if err != nil {
		return 0, 0, err
	}

	confirmed := w.chain.State().TipState().GetCoinsState().Balances[acc]

	return confirmed, 0, nil
}

// SendToAddress sends an amount to an account using the current open wallet private key.
func (w *wallet) SendToAddress(to string, amount uint64) (*chainhash.Hash, error) {
	if !w.open {
		return nil, errorNotOpen
	}
	priv, err := w.GetSecret()
	if err != nil {
		return nil, err
	}
	_, data, err := bech32.Decode(to)
	if err != nil {
		return nil, err
	}

	if len(data) != 20 {
		return nil, fmt.Errorf("invalid address")
	}

	var toPkh [20]byte

	copy(toPkh[:], data)

	pub := priv.PublicKey()

	acc, err := w.GetAccountRaw()
	if err != nil {
		return nil, err
	}

	latestNonce := w.chain.State().TipState().GetCoinsState().Nonces[acc]

	var p [48]byte
	copy(p[:], pub.Marshal())

	tx := &primitives.Tx{
		To:            toPkh,
		FromPublicKey: p,
		Amount:        amount,
		Nonce:         latestNonce + 1,
		Fee:           5000,
	}

	sigMsg := tx.SignatureMessage()
	sig := priv.Sign(sigMsg[:])
	var s [96]byte
	copy(s[:], sig.Marshal())
	tx.Signature = s

	if err := w.coinsmempool.Add(tx); err != nil {
		return nil, err
	}

	msg := &p2p.MsgTx{Data: tx}

	err = w.host.Broadcast(msg)
	if err != nil {
		return nil, nil
	}

	txHash := tx.Hash()

	return &txHash, nil
}
