package wallet

import (
	"fmt"

	"github.com/olympus-protocol/ogen/pkg/bech32"
	"github.com/olympus-protocol/ogen/pkg/bls"
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

// StartValidator signs a validator deposit with the current open wallet private key.
func (w *wallet) StartValidator(validatorPrivBytes [32]byte) (*primitives.Deposit, error) {
	if !w.open {
		return nil, errorNotOpen
	}
	priv, err := w.GetSecret()
	if err != nil {
		return nil, err
	}
	pub := priv.PublicKey()

	validatorPriv, err := bls.SecretKeyFromBytes(validatorPrivBytes[:])
	if err != nil {
		return nil, err
	}

	validatorPub := validatorPriv.PublicKey()
	validatorPubBytes := validatorPub.Marshal()
	validatorPubHash := chainhash.HashH(validatorPubBytes[:])

	validatorProofOfPossession := validatorPriv.Sign(validatorPubHash[:])

	addr, err := w.GetAccountRaw()
	if err != nil {
		return nil, err
	}
	var p [48]byte
	var s [96]byte
	copy(p[:], validatorPubBytes)
	copy(s[:], validatorProofOfPossession.Marshal())
	depositData := &primitives.DepositData{
		PublicKey:         p,
		ProofOfPossession: s,
		WithdrawalAddress: addr,
	}

	buf, err := depositData.Marshal()
	if err != nil {
		return nil, err
	}

	depositHash := chainhash.HashH(buf)

	depositSig := priv.Sign(depositHash[:])

	var pubKey [48]byte
	var ds [96]byte
	copy(pubKey[:], pub.Marshal())
	copy(ds[:], depositSig.Marshal())
	deposit := &primitives.Deposit{
		PublicKey: pubKey,
		Signature: ds,
		Data:      depositData,
	}

	currentState := w.chain.State().TipState()

	if err := w.actionMempool.AddDeposit(deposit, currentState); err != nil {
		return nil, err
	}
	w.broadcastDeposit(deposit)
	return deposit, nil
}

// ExitValidator submits an exit transaction for a certain validator with the current wallet private key.
func (w *wallet) ExitValidator(validatorPubKey [48]byte) (*primitives.Exit, error) {
	if !w.open {
		return nil, errorNotOpen
	}
	priv, err := w.GetSecret()
	if err != nil {
		return nil, err
	}

	validatorPub, err := bls.PublicKeyFromBytes(validatorPubKey)
	if err != nil {
		return nil, err
	}

	currentState := w.chain.State().TipState()

	pub := priv.PublicKey()

	msg := fmt.Sprintf("exit %x", validatorPub.Marshal())
	msgHash := chainhash.HashH([]byte(msg))

	sig := priv.Sign(msgHash[:])
	var valp, withp [48]byte
	var s [96]byte
	copy(valp[:], validatorPub.Marshal())
	copy(withp[:], pub.Marshal())
	copy(s[:], sig.Marshal())
	exit := &primitives.Exit{
		ValidatorPubkey: valp,
		WithdrawPubkey:  withp,
		Signature:       s,
	}

	if err := w.actionMempool.AddExit(exit, currentState); err != nil {
		return nil, err
	}

	w.broadcastExit(exit)

	return exit, nil
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
	if err := w.mempool.Add(*tx, &cs); err != nil {
		return nil, err
	}

	w.broadcastTx(tx)

	txHash := tx.Hash()

	return &txHash, nil
}

func (w *wallet) broadcastTx(payload *primitives.Tx) {
	buf, err := payload.Marshal()
	if err != nil {
		w.log.Errorf("error encoding transaction: %s", err)
		return
	}
	if err := w.txTopic.Publish(w.ctx, buf); err != nil {
		w.log.Errorf("error broadcasting transaction: %s", err)
	}
}

func (w *wallet) broadcastDeposit(deposit *primitives.Deposit) {
	buf, err := deposit.Marshal()
	if err != nil {
		w.log.Errorf("error encoding transaction: %s", err)
		return
	}
	if err := w.depositTopic.Publish(w.ctx, buf); err != nil {
		w.log.Errorf("error broadcasting transaction: %s", err)
	}
}

func (w *wallet) broadcastExit(exit *primitives.Exit) {
	buf, err := exit.Marshal()
	if err != nil {
		w.log.Errorf("error encoding transaction: %s", err)
		return
	}
	if err := w.exitTopic.Publish(w.ctx, buf); err != nil {
		w.log.Errorf("error broadcasting transaction: %s", err)
	}
}
