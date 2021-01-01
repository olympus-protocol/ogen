package wallet

import (
	"fmt"
	"github.com/olympus-protocol/ogen/internal/mempool"
	"github.com/olympus-protocol/ogen/pkg/bech32"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/p2p"
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

type Balance struct {
	Confirmed uint64
	Pending   uint64
}

// GetBalance returns the balance of the current open wallet.
func (w *wallet) GetBalance() (*Balance, error) {
	if !w.open {
		return nil, errorNotOpen
	}
	acc, err := w.GetAccountRaw()
	if err != nil {
		return nil, err
	}

	confirmed := w.chain.State().TipState().GetCoinsState().Balances[acc]

	return &Balance{Confirmed: confirmed}, nil
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

	var latestNonce uint64
	latestNonce, err = w.coinsmempool.GetMempoolNonce(acc)
	if err != nil {
		if err == mempool.ErrorAccountNotOnMempool {
			latestNonce = w.chain.State().TipState().GetCoinsState().Nonces[acc]
		} else {
			return nil, err
		}
	}

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
