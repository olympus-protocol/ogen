package wallet

import (
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/p2p"
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

func (w *wallet) StartValidatorBulk(valSecKeys []*bls.SecretKey) (bool, error) {
	if !w.open {
		return false, errorNotOpen
	}
	priv, err := w.GetSecret()
	if err != nil {
		return false, err
	}

	addr, err := w.GetAccountRaw()
	if err != nil {
		return false, err
	}

	deposits := make([]*primitives.Deposit, len(valSecKeys))
	currentState := w.chain.State().TipState()

	for i := range deposits {
		deposit, err := w.createDeposit(priv, addr, valSecKeys[i])
		if err != nil {
			return false, err
		}
		deposits[i] = deposit

		if err := w.actionsmempool.AddDeposit(deposit, currentState); err != nil {
			return false, err
		}
	}

	msg := &p2p.MsgDeposits{Data: deposits}

	err = w.host.Broadcast(msg)
	if err != nil {
		return false, nil
	}

	return true, nil
}

// StartValidator signs a validator deposit with the current open wallet private key.
func (w *wallet) StartValidator(valPrivBytes *bls.SecretKey) (bool, error) {

	if !w.open {
		return false, errorNotOpen
	}
	priv, err := w.GetSecret()
	if err != nil {
		return false, err
	}

	addr, err := w.GetAccountRaw()
	if err != nil {
		return false, err
	}

	deposit, err := w.createDeposit(priv, addr, valPrivBytes)

	currentState := w.chain.State().TipState()

	if err := w.actionsmempool.AddDeposit(deposit, currentState); err != nil {
		return false, err
	}

	msg := &p2p.MsgDeposit{Data: deposit}

	err = w.host.Broadcast(msg)
	if err != nil {
		return false, nil
	}

	return true, nil
}

func (w *wallet) createDeposit(priv *bls.SecretKey, addr [20]byte, validatorPriv *bls.SecretKey) (*primitives.Deposit, error) {
	pub := priv.PublicKey()

	validatorPub := validatorPriv.PublicKey()
	validatorPubBytes := validatorPub.Marshal()
	validatorPubHash := chainhash.HashH(validatorPubBytes[:])
	validatorProofOfPossession := validatorPriv.Sign(validatorPubHash[:])

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

	return &primitives.Deposit{
		PublicKey: pubKey,
		Signature: ds,
		Data:      depositData,
	}, nil
}

// ExitValidatorBulk submits an exit transaction for a certain validator with the current wallet private key.
func (w *wallet) ExitValidatorBulk(valPubKeys []*bls.PublicKey) (bool, error) {

	if !w.open {
		return false, errorNotOpen
	}

	priv, err := w.GetSecret()
	if err != nil {
		return false, err
	}

	exits := make([]*primitives.Exit, len(valPubKeys))

	currentState := w.chain.State().TipState()

	for i := range exits {
		exit, err := w.createExit(priv, valPubKeys[i])
		if err != nil {
			return false, err
		}
		exits[i] = exit

		if err := w.actionsmempool.AddExit(exit, currentState); err != nil {
			return false, err
		}
	}

	msg := &p2p.MsgExits{Data: exits}
	err = w.host.Broadcast(msg)
	if err != nil {
		return false, nil
	}

	return true, nil
}

// ExitValidator submits an exit transaction for a certain validator with the current wallet private key.
func (w *wallet) ExitValidator(valPubKey *bls.PublicKey) (bool, error) {

	if !w.open {
		return false, errorNotOpen
	}

	priv, err := w.GetSecret()
	if err != nil {
		return false, err
	}

	exit, err := w.createExit(priv, valPubKey)
	if err != nil {
		return false, err
	}

	currentState := w.chain.State().TipState()

	if err := w.actionsmempool.AddExit(exit, currentState); err != nil {
		return false, err
	}

	msg := &p2p.MsgExit{Data: exit}
	err = w.host.Broadcast(msg)
	if err != nil {
		return false, nil
	}

	return true, nil
}

func (w *wallet) createExit(priv *bls.SecretKey, valPubKey *bls.PublicKey) (*primitives.Exit, error) {
	pub := priv.PublicKey()

	msgHash := chainhash.HashH(valPubKey.Marshal())

	sig := priv.Sign(msgHash[:])
	var valp, withp [48]byte
	var s [96]byte
	copy(valp[:], valPubKey.Marshal())
	copy(withp[:], pub.Marshal())
	copy(s[:], sig.Marshal())
	return &primitives.Exit{
		ValidatorPubkey: valp,
		WithdrawPubkey:  withp,
		Signature:       s,
	}, nil
}
