package wallet

import (
	"bytes"
	"fmt"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/bech32"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

// GetBalance gets the balance of an address or of the wallet address.
func (b *Wallet) GetBalance(addr string) (uint64, error) {
	if addr == "" {
		addr = b.info.address
	}
	_, pkh, err := bech32.Decode(addr)
	if err != nil {
		return 0, err
	}
	if len(pkh) != 20 {
		return 0, fmt.Errorf("expecting address to be 20 bytes, but got %d bytes", len(pkh))
	}
	var pkhBytes [20]byte
	copy(pkhBytes[:], pkh)
	out, ok := b.chain.State().TipState().UtxoState.Balances[pkhBytes]
	if !ok {
		return 0, nil
	}
	return out, nil
}

// GetAddressRaw returns the pubkey hash of the wallet.
func (b *Wallet) GetAddressRaw() ([20]byte, error) {
	_, pkh, err := bech32.Decode(b.info.address)
	if err != nil {
		return [20]byte{}, err
	}
	if len(pkh) != 20 {
		return [20]byte{}, fmt.Errorf("expecting address to be 20 bytes, but got %d bytes", len(pkh))
	}
	var pkhBytes [20]byte
	copy(pkhBytes[:], pkh)
	return pkhBytes, nil
}

func (b *Wallet) broadcastTx(payload *primitives.CoinPayload) {
	buf := bytes.NewBuffer([]byte{})
	err := payload.Encode(buf)
	if err != nil {
		b.log.Errorf("error encoding transaction: %s", err)
		return
	}
	if err := b.txTopic.Publish(b.ctx, buf.Bytes()); err != nil {
		b.log.Errorf("error broadcasting transaction: %s", err)
	}
}

// SendToAddress sends an amount to an address with the given password and parameters.
func (b *Wallet) SendToAddress(authentication []byte, to string, amount uint64) (*chainhash.Hash, error) {
	priv, err := b.unlockIfNeeded(authentication)
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

	pub := priv.DerivePublicKey()

	b.lastNonceLock.Lock()
	b.info.lastNonce++
	nonce := b.info.lastNonce
	b.lastNonceLock.Unlock()

	payload := &primitives.CoinPayload{
		To:            toPkh,
		FromPublicKey: *pub,
		Amount:        amount,
		Nonce:         nonce,
		Fee:           1,
	}

	sigMsg := payload.SignatureMessage()
	sig, err := bls.Sign(priv, sigMsg[:])
	if err != nil {
		return nil, err
	}

	payload.Signature = *sig

	tx := &primitives.Tx{
		TxType:    primitives.TxCoins,
		TxVersion: 0,
		Payload:   payload,
	}

	currentState := b.chain.State().TipState()

	if err := b.mempool.Add(*payload, &currentState.UtxoState); err != nil {
		return nil, err
	}

	b.broadcastTx(payload)

	txHash := tx.Hash()

	return &txHash, nil
}
