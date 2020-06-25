package primitives

import (
	"bytes"
	"fmt"
	"io"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/prysmaticlabs/go-ssz"
)

// TxLocator is a simple struct to find a database referenced to a block without building a full index
type TxLocator struct {
	Hash  chainhash.Hash
	Block chainhash.Hash
	Index uint32
}

// Marshal encodes the data.
func (tl *TxLocator) Marshal() ([]byte, error) {
	return ssz.Marshal(tl)
}

// Unmarshal decodes the data.
func (tl *TxLocator) Unmarshal(b []byte) error {
	return ssz.Unmarshal(b, tl)
}

// TxType represents a type of transaction.
type TxType = int32

const (
	// TxTransferSingle represents a transaction sending money from a single
	// address to another address.
	TxTransferSingle TxType = iota + 1

	// TxTransferMulti represents a transaction sending money from a multisig
	// address to another address.
	TxTransferMulti
)

// TransferSinglePayload is a transaction payload for sending money from
// a single address to another address.
type TransferSinglePayload struct {
	To            [20]byte
	FromPublicKey []byte
	Amount        uint64
	Nonce         uint64
	Fee           uint64
	Signature     []byte
}

// Hash calculates the transaction ID of the payload.
func (c TransferSinglePayload) Hash() chainhash.Hash {
	hash, _ := ssz.HashTreeRoot(c)
	return chainhash.Hash(hash)
}

// FromPubkeyHash calculates the hash of the from public key.
func (c TransferSinglePayload) FromPubkeyHash() (out [20]byte, err error) {
	pkS := c.FromPublicKey
	h := chainhash.HashH(pkS[:])
	copy(out[:], h[:20])
	return
}

// Marshal encodes the data.
func (c *TransferSinglePayload) Marshal() ([]byte, error) {
	return ssz.Marshal(c)
}

// Unmarshal decodes the data.
func (c *TransferSinglePayload) Unmarshal(b []byte) error {
	return ssz.Unmarshal(b, c)
}

// SignatureMessage gets the message the needs to be signed.
func (c TransferSinglePayload) SignatureMessage() chainhash.Hash {
	cp := c
	cp.Signature = make([]byte, 96)
	b, _ := cp.Marshal()
	return chainhash.HashH(b)
}

func (c TransferSinglePayload) GetSignature() (*bls.Signature, error) {
	return bls.SignatureFromBytes(c.Signature)
}

func (c TransferSinglePayload) GetPublic() (*bls.PublicKey, error) {
	return bls.PublicKeyFromBytes(c.FromPublicKey)
}

// VerifySig verifies the signatures is valid.
func (c *TransferSinglePayload) VerifySig() error {
	sigMsg := c.SignatureMessage()
	sig, err := c.GetSignature()
	if err != nil {
		return err
	}
	pub, err := c.GetPublic()
	if err != nil {
		return err
	}
	valid := sig.Verify(sigMsg[:], pub)
	if !valid {
		return fmt.Errorf("signature is not valid")
	}

	return nil
}

// GetNonce gets the transaction nonce.
func (c TransferSinglePayload) GetNonce() uint64 {
	return c.Nonce
}

// GetAmount gets the transaction amount to send.
func (c TransferSinglePayload) GetAmount() uint64 {
	return c.Amount
}

// GetFee gets the transaction fee.
func (c TransferSinglePayload) GetFee() uint64 {
	return c.Fee
}

// GetFromAddress gets the from address.
func (c TransferSinglePayload) GetFromAddress() ([20]byte, error) {
	return c.FromPubkeyHash()
}

var _ TxPayload = &TransferSinglePayload{}

// TransferMultiPayload represents a transfer from a multisig to
// another address.
type TransferMultiPayload struct {
	To       [20]byte
	Amount   uint64
	Nonce    uint64
	Fee      uint64
	MultiSig []byte
}

// Marshal encodes the data.
func (c *TransferMultiPayload) Marshal() ([]byte, error) {
	return ssz.Marshal(c)
}

// Unmarshal decodes the data.
func (c *TransferMultiPayload) Unmarshal(b []byte) error {
	return ssz.Unmarshal(b, c)
}

// Hash calculates the transaction ID of the payload.
func (c TransferMultiPayload) Hash() chainhash.Hash {
	hash, _ := ssz.HashTreeRoot(c)
	return chainhash.Hash(hash)
}

// FromPubkeyHash calculates the hash of the from public key.
func (c TransferMultiPayload) FromPubkeyHash() ([20]byte, error) {
	pub, err := c.GetPublic()
	if err != nil {
		return [20]byte{}, err
	}
	return pub.Hash()
}

// SignatureMessage gets the message the needs to be signed.
func (c TransferMultiPayload) SignatureMessage() chainhash.Hash {
	cp := c
	cp.MultiSig = []byte{}
	b, _ := cp.Marshal()
	return chainhash.HashH(b)
}

func (c TransferMultiPayload) GetPublic() (bls.FunctionalPublicKey, error) {
	sig, err := c.GetSignature()
	if err != nil {
		return nil, err
	}
	return sig.GetPublicKey()
}

func (c TransferMultiPayload) GetSignature() (bls.FunctionalSignature, error) {
	buf := bytes.NewBuffer(c.MultiSig)
	return bls.ReadFunctionalSignature(buf)
}

// VerifySig verifies the signatures is valid.
func (c TransferMultiPayload) VerifySig() error {
	sigMsg := c.SignatureMessage()
	sig, err := c.GetSignature()
	if err != nil {
		return err
	}
	valid := sig.Verify(sigMsg[:])
	if !valid {
		return fmt.Errorf("signature is not valid")
	}

	return nil
}

// GetNonce gets the transaction nonce.
func (c TransferMultiPayload) GetNonce() uint64 {
	return c.Nonce
}

// GetAmount gets the transaction amount to send.
func (c TransferMultiPayload) GetAmount() uint64 {
	return c.Amount
}

// GetFee gets the transaction fee.
func (c TransferMultiPayload) GetFee() uint64 {
	return c.Fee
}

var _ TxPayload = &TransferMultiPayload{}

// GenesisPayload is the payload of the genesis transaction.
type GenesisPayload struct{}

// Encode does nothing for the genesis payload.
func (g *GenesisPayload) Encode(w io.Writer) error { return nil }

// Decode does nothing for the genesis payload.
func (g *GenesisPayload) Decode(r io.Reader) error { return nil }

// TxPayload represents anything that can be used as a payload in a transaction.
type TxPayload interface {
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
	GetNonce() uint64
	GetAmount() uint64
	GetFee() uint64
	FromPubkeyHash() ([20]byte, error)
}

// Tx represents a transaction on the blockchain.
type Tx struct {
	Version int32
	Type    TxType
	Payload []byte
}

func (t *Tx) GetPayload() (TxPayload, error) {
	var pload TxPayload
	err := pload.Unmarshal(t.Payload)
	if err != nil {
		return nil, err
	}
	return pload, nil
}

func (t *Tx) AppendPayload(p TxPayload) error {
	buf, err := p.Marshal()
	if err != nil {
		return err
	}
	t.Payload = buf
	return nil
}

// Marshal encodes the data.
func (t *Tx) Marshal() ([]byte, error) {
	return ssz.Marshal(t)
}

// Unmarshal decodes the data.
func (t *Tx) Unmarshal(b []byte) error {
	return ssz.Unmarshal(b, t)
}

// Hash calculates the transaction hash.
func (t *Tx) Hash() chainhash.Hash {
	b, _ := t.Marshal()
	return chainhash.DoubleHashH(b)
}
