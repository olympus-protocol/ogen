package wallet

import (
	"bytes"
	"fmt"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/bech32"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

func (w *Wallet) GetBalance() (uint64, error) {
	if !w.open {
		return 0, errorNotOpen
	}
	out, ok := w.chain.State().TipState().CoinsState.Balances[w.info.account]
	if !ok {
		return 0, nil
	}
	return out, nil
}

// StartValidator signs a validator deposit.
func (w *Wallet) StartValidator(validatorPrivBytes [32]byte) (*primitives.Deposit, error) {
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

	depositData := &primitives.DepositData{
		PublicKey:         *validatorPub,
		ProofOfPossession: *validatorProofOfPossession,
		WithdrawalAddress: addr,
	}

	buf := bytes.NewBuffer([]byte{})

	if err := depositData.Encode(buf); err != nil {
		return nil, err
	}

	depositHash := chainhash.HashH(buf.Bytes())

	depositSig := priv.Sign(depositHash[:])

	deposit := &primitives.Deposit{
		PublicKey: *pub,
		Signature: *depositSig,
		Data:      *depositData,
	}

	currentState := w.chain.State().TipState()

	if err := w.actionMempool.AddDeposit(deposit, currentState); err != nil {
		return nil, err
	}
	w.broadcastDeposit(deposit)
	return deposit, nil
}

// ExitValidator submits an exit transaction for a certain validator.
func (w *Wallet) ExitValidator(validatorPubKey [48]byte) (*primitives.Exit, error) {
	if !w.open {
		return nil, errorNotOpen
	}
	priv, err := w.GetSecret()
	if err != nil {
		return nil, err
	}

	validatorPub, err := bls.PublicKeyFromBytes(validatorPubKey[:])
	if err != nil {
		return nil, err
	}

	currentState := w.chain.State().TipState()

	pub := priv.PublicKey()

	msg := fmt.Sprintf("exit %x", validatorPub.Marshal())
	msgHash := chainhash.HashH([]byte(msg))

	sig := priv.Sign(msgHash[:])

	exit := &primitives.Exit{
		ValidatorPubkey: *validatorPub,
		WithdrawPubkey:  *pub,
		Signature:       *sig,
	}

	if err := w.actionMempool.AddExit(exit, currentState); err != nil {
		return nil, err
	}

	w.broadcastExit(exit)

	return exit, nil
}

// SendToAddress sends an amount to an address with the given password and parameters.
func (w *Wallet) SendToAddress(to string, amount uint64) (*chainhash.Hash, error) {
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

	w.lastNonceLock.Lock()
	w.info.lastNonce++
	nonce := w.info.lastNonce
	w.lastNonceLock.Unlock()

	payload := &primitives.TransferSinglePayload{
		To:            toPkh,
		FromPublicKey: *pub,
		Amount:        amount,
		Nonce:         nonce,
		Fee:           1,
	}

	sigMsg := payload.SignatureMessage()
	sig := priv.Sign(sigMsg[:])

	payload.Signature = *sig

	tx := &primitives.Tx{
		Type:    primitives.TxTransferSingle,
		Version: 0,
		Payload: payload,
	}

	currentState := w.chain.State().TipState()

	if err := w.mempool.Add(*tx, &currentState.CoinsState); err != nil {
		return nil, err
	}

	w.broadcastTx(tx)

	txHash := tx.Hash()

	return &txHash, nil
}

func (w *Wallet) broadcastTx(payload *primitives.Tx) {
	buf, err := payload.Marshal()
	if err != nil {
		w.log.Errorf("error encoding transaction: %s", err)
		return
	}
	if err := w.txTopic.Publish(w.ctx, buf); err != nil {
		w.log.Errorf("error broadcasting transaction: %s", err)
	}
}

func (w *Wallet) broadcastDeposit(deposit *primitives.Deposit) {
	buf := bytes.NewBuffer([]byte{})
	err := deposit.Encode(buf)
	if err != nil {
		w.log.Errorf("error encoding transaction: %s", err)
		return
	}
	if err := w.depositTopic.Publish(w.ctx, buf.Bytes()); err != nil {
		w.log.Errorf("error broadcasting transaction: %s", err)
	}
}

func (w *Wallet) broadcastExit(exit *primitives.Exit) {
	buf, err := exit.Marshal()
	if err != nil {
		w.log.Errorf("error encoding transaction: %s", err)
		return
	}
	if err := w.exitTopic.Publish(w.ctx, buf); err != nil {
		w.log.Errorf("error broadcasting transaction: %s", err)
	}
}
