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
	TxHash chainhash.Hash
	Block  chainhash.Hash
	Index  uint32
}

// Marshal serializes the struct to bytes
func (tl *TxLocator) Marshal() ([]byte, error) {
	return ssz.Marshal(tl)
}

// Unmarshal deserializes the struct from bytes
func (tl *TxLocator) Unmarshal(b []byte) error {
	return ssz.Unmarshal(b, tl)
}

const (
	// TxTransferSingle represents a transaction sending money from a single
	// address to another address.
	TxTransferSingle uint32 = iota + 1

	// TxTransferMulti represents a transaction sending money from a multisig
	// address to another address.
	TxTransferMulti
)

// TransferSinglePayload is a transaction payload for sending money from
// a single address to another address.
type TransferSinglePayload struct {
	To            [20]byte `ssz:"size=20"`
	FromPublicKey []byte   `ssz:"size=48"`
	Amount        uint64
	Nonce         uint64
	Fee           uint64
	Signature     []byte `ssz:"size=96"`
}

// Marshal serializes the struct to bytes
func (c *TransferSinglePayload) Marshal() ([]byte, error) {
	return ssz.Marshal(c)
}

// Unmarshal deserializes the struct from bytes
func (c *TransferSinglePayload) Unmarshal(b []byte) error {
	return ssz.Unmarshal(b, c)
}

// Hash calculates the transaction ID of the payload.
func (c *TransferSinglePayload) Hash() (chainhash.Hash, error) {
	ser, err := c.Marshal()
	if err != nil {
		return chainhash.Hash{}, err
	}
	return chainhash.HashH(ser), nil
}

// FromPubkeyHash calculates the hash of the from public key.
func (c *TransferSinglePayload) FromPubkeyHash() (out [20]byte, err error) {
	h := chainhash.HashH(c.FromPublicKey[:])
	copy(out[:], h[:20])
	return
}

// SignatureMessage gets the message the needs to be signed.
func (c *TransferSinglePayload) SignatureMessage() chainhash.Hash {
	buf := bytes.NewBuffer([]byte{})
	// TODO
	//_ = serializer.WriteElements(buf, c.To, c.Nonce, c.FromPublicKey, c.Amount, c.Fee)
	return chainhash.HashH(buf.Bytes())
}

// VerifySig verifies the signatures is valid.
func (c *TransferSinglePayload) VerifySig() error {
	sigMsg := c.SignatureMessage()
	pubkey, err := bls.PublicKeyFromBytes(c.FromPublicKey)
	if err != nil {
		return err
	}
	sign, err := bls.SignatureFromBytes(c.Signature)
	if err != nil {
		return err
	}
	valid := sign.Verify(sigMsg[:], pubkey)
	if !valid {
		return fmt.Errorf("signature is not valid")
	}
	return nil
}

// GetNonce gets the transaction nonce.
func (c *TransferSinglePayload) GetNonce() uint64 {
	return c.Nonce
}

// GetAmount gets the transaction amount to send.
func (c *TransferSinglePayload) GetAmount() uint64 {
	return c.Amount
}

// GetFee gets the transaction fee.
func (c *TransferSinglePayload) GetFee() uint64 {
	return c.Fee
}

// GetFromAddress gets the from address.
func (c *TransferSinglePayload) GetFromAddress() ([20]byte, error) {
	return c.FromPubkeyHash()
}

var _ TxPayload = &TransferSinglePayload{}

// TransferMultiPayload represents a transfer from a multisig to
// another address.
type TransferMultiPayload struct {
	To        [20]byte `ssz:"size=20"`
	Amount    uint64
	Nonce     uint64
	Fee       uint64
	Signature bls.Multisig
}

// Marshal serializes the struct to bytes
func (c *TransferMultiPayload) Marshal() ([]byte, error) {
	return ssz.Marshal(c)
}

// Unmarshal deserializes the struct from bytes
func (c *TransferMultiPayload) Unmarshal(b []byte) error {
	return ssz.Unmarshal(b, c)
}

// Hash calculates the transaction ID of the payload.
func (c *TransferMultiPayload) Hash() (chainhash.Hash, error) {
	ser, err := c.Marshal()
	if err != nil {
		return chainhash.Hash{}, err
	}
	return chainhash.HashH(ser), nil
}

// FromPubkeyHash calculates the hash of the from public key.
func (c *TransferMultiPayload) FromPubkeyHash() ([20]byte, error) {
	hash := c.Signature.PublicKey.Hash()
	return hash, nil
}

// SignatureMessage gets the message the needs to be signed.
func (c *TransferMultiPayload) SignatureMessage() chainhash.Hash {
	buf := bytes.NewBuffer([]byte{})
	//_ = serializer.WriteElements(buf, c.To, c.Nonce, c.FromPubkeyHash(), c.Amount, c.Fee)
	return chainhash.HashH(buf.Bytes())
}

// VerifySig verifies the signatures is valid.
func (c *TransferMultiPayload) VerifySig() error {
	sigMsg := c.SignatureMessage()

	valid := c.Signature.Verify(sigMsg[:])
	if !valid {
		return fmt.Errorf("signature is not valid")
	}

	return nil
}

// GetNonce gets the transaction nonce.
func (c *TransferMultiPayload) GetNonce() uint64 {
	return c.Nonce
}

// GetAmount gets the transaction amount to send.
func (c *TransferMultiPayload) GetAmount() uint64 {
	return c.Amount
}

// GetFee gets the transaction fee.
func (c *TransferMultiPayload) GetFee() uint64 {
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
	GetNonce() uint64
	GetAmount() uint64
	GetFee() uint64
	FromPubkeyHash() ([20]byte, error)
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}

// Tx represents a transaction on the blockchain.
type Tx struct {
	Version uint32
	Type    uint32
	Payload TxPayload
}

// Marshal serializes the struct to bytes
func (t *Tx) Marshal() ([]byte, error) {
	return ssz.Marshal(t)
}

// Unmarshal deserializes the struct from bytes
func (t *Tx) Unmarshal(b []byte) error {
	return ssz.Unmarshal(b, t)
}

// Hash calculates the transaction hash.
func (t *Tx) Hash() (chainhash.Hash, error) {
	ser, err := t.Marshal()
	if err != nil {
		return chainhash.Hash{}, err
	}
	return chainhash.DoubleHashH(ser), nil
}
