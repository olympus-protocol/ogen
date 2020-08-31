package wallet

import (
	"bytes"
	"fmt"
	"github.com/olympus-protocol/ogen/pkg/p2p"

	"github.com/olympus-protocol/ogen/pkg/bech32"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

// GetBalance returns the balance of the current open wallet.
func (w *wallet) GetBalance() (uint64, error) {
	if !w.open {
		return 0, errorNotOpen
	}
	acc, err := w.GetAccountRaw()
	if err != nil {
		return 0, err
	}
	out := w.chain.State().TipState().GetCoinsState().Balances[acc]
	return out, nil
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

	nonce := w.chain.State().TipState().GetCoinsState().Nonces[acc] + 1
	var p [48]byte
	copy(p[:], pub.Marshal())

	tx := &primitives.Tx{
		To:            toPkh,
		FromPublicKey: p,
		Amount:        amount,
		Nonce:         nonce,
		Fee:           5000,
	}

	sigMsg := tx.SignatureMessage()
	sig := priv.Sign(sigMsg[:])
	var s [96]byte
	copy(s[:], sig.Marshal())
	tx.Signature = s

	currentState := w.chain.State().TipState()
	cs := currentState.GetCoinsState()
	if err := w.mempool.Add(tx, &cs); err != nil {
		return nil, err
	}

	w.broadcastTx(tx)

	txHash := tx.Hash()

	return &txHash, nil
}

func (w *wallet) broadcastTx(payload *primitives.Tx) {
	msg := &p2p.MsgTx{Data: payload}
	buf := bytes.NewBuffer([]byte{})
	err := p2p.WriteMessage(buf, msg, w.hostnode.GetNetMagic())
	if err != nil {
		w.log.Errorf("error encoding tx: %s", err)
		return
	}
	if err := w.txTopic.Publish(w.ctx, buf.Bytes()); err != nil {
		w.log.Errorf("error broadcasting transaction: %s", err)
	}
}
