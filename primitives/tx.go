package primitives

import (
	"bytes"
	"fmt"
	"io"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
)

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
	FromPublicKey bls.PublicKey
	Amount        uint64
	Nonce         uint64
	Fee           uint64
	Signature     bls.Signature
}

// Hash calculates the transaction ID of the payload.
func (c *TransferSinglePayload) Hash() chainhash.Hash {
	buf := bytes.NewBuffer([]byte{})
	_ = c.Encode(buf)
	return chainhash.HashH(buf.Bytes())
}

// FromPubkeyHash calculates the hash of the from public key.
func (c *TransferSinglePayload) FromPubkeyHash() (out [20]byte) {
	pkS := c.FromPublicKey.Marshal()
	h := chainhash.HashH(pkS[:])
	copy(out[:], h[:20])
	return
}

// Encode enccodes the transaction to the writer.
func (c *TransferSinglePayload) Encode(w io.Writer) error {
	if err := serializer.WriteElements(w, c.To); err != nil {
		return err
	}
	pubBytes := c.FromPublicKey.Marshal()
	sigBytes := c.Signature.Marshal()
	if _, err := w.Write(pubBytes[:]); err != nil {
		return err
	}
	if _, err := w.Write(sigBytes[:]); err != nil {
		return err
	}
	if err := serializer.WriteVarInt(w, c.Amount); err != nil {
		return err
	}
	if err := serializer.WriteVarInt(w, c.Nonce); err != nil {
		return err
	}
	if err := serializer.WriteVarInt(w, c.Fee); err != nil {
		return err
	}
	return nil
}

// SignatureMessage gets the message the needs to be signed.
func (c *TransferSinglePayload) SignatureMessage() chainhash.Hash {
	buf := bytes.NewBuffer([]byte{})
	_ = serializer.WriteElements(buf, c.To, c.Nonce, c.FromPublicKey, c.Amount, c.Fee)
	return chainhash.HashH(buf.Bytes())
}

// VerifySig verifies the signatures is valid.
func (c *TransferSinglePayload) VerifySig() error {
	sigMsg := c.SignatureMessage()

	valid := c.Signature.Verify(sigMsg[:], &c.FromPublicKey)
	if !valid {
		return fmt.Errorf("signature is not valid")
	}

	return nil
}

// Decode decodes the transaction payload from the given reader.
func (c *TransferSinglePayload) Decode(r io.Reader) error {
	if err := serializer.ReadElements(r, &c.To); err != nil {
		return err
	}
	sigBytes := make([]byte, 96)
	pubBytes := make([]byte, 48)
	_, err := r.Read(pubBytes[:])
	if err != nil {
		return err
	}
	pub, err := bls.PublicKeyFromBytes(pubBytes)
	if err != nil {
		return err
	}
	c.FromPublicKey = *pub
	_, err = r.Read(sigBytes[:])
	if err != nil {
		return err
	}
	sig, err := bls.SignatureFromBytes(sigBytes)
	if err != nil {
		return err
	}
	c.Signature = *sig
	c.Amount, err = serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	c.Nonce, err = serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	c.Fee, err = serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	return nil
}

var _ TxPayload = &TransferSinglePayload{}

// TransferMultiPayload represents a transfer from a multisig to
// another address.
type TransferMultiPayload struct {
	To        [20]byte
	Amount    uint64
	Nonce     uint64
	Fee       uint64
	Signature bls.Multisig
}

// Hash calculates the transaction ID of the payload.
func (c *TransferMultiPayload) Hash() chainhash.Hash {
	buf := bytes.NewBuffer([]byte{})
	_ = c.Encode(buf)
	return chainhash.HashH(buf.Bytes())
}

// FromPubkeyHash calculates the hash of the from public key.
func (c *TransferMultiPayload) FromPubkeyHash() []byte {
	return c.Signature.PublicKey.ToHash()
}

// Encode enccodes the transaction to the writer.
func (c *TransferMultiPayload) Encode(w io.Writer) error {
	if err := serializer.WriteElements(w, c.To); err != nil {
		return err
	}
	if err := c.Signature.Encode(w); err != nil {
		return err
	}
	if err := serializer.WriteVarInt(w, c.Amount); err != nil {
		return err
	}
	if err := serializer.WriteVarInt(w, c.Nonce); err != nil {
		return err
	}
	if err := serializer.WriteVarInt(w, c.Fee); err != nil {
		return err
	}
	return nil
}

// SignatureMessage gets the message the needs to be signed.
func (c *TransferMultiPayload) SignatureMessage() chainhash.Hash {
	buf := bytes.NewBuffer([]byte{})
	_ = serializer.WriteElements(buf, c.To, c.Nonce, c.FromPubkeyHash(), c.Amount, c.Fee)
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

// Decode decodes the transaction payload from the given reader.
func (c *TransferMultiPayload) Decode(r io.Reader) error {
	if err := serializer.ReadElements(r, &c.To); err != nil {
		return err
	}
	if err := c.Signature.Decode(r); err != nil {
		return err
	}
	var err error
	c.Amount, err = serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	c.Nonce, err = serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	c.Fee, err = serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	return nil
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
	Encode(w io.Writer) error
	Decode(r io.Reader) error
}

// Tx represents a transaction on the blockchain.
type Tx struct {
	TxVersion int32
	TxType    TxType
	Payload   TxPayload
}

// Encode encodes the transaction to the given writer.
func (t *Tx) Encode(w io.Writer) error {
	err := serializer.WriteElements(w, t.TxVersion, t.TxType)
	if err != nil {
		return err
	}
	if t.Payload == nil {
		return fmt.Errorf("transaction missing payload")
	}
	return t.Payload.Encode(w)
}

// Decode decodes a transaction from the given reader.
func (t *Tx) Decode(r io.Reader) error {
	err := serializer.ReadElements(r, &t.TxVersion, &t.TxType)
	if err != nil {
		return err
	}
	switch t.TxType {
	case TxTransferSingle:
		t.Payload = &TransferSinglePayload{}
		return t.Payload.Decode(r)
	case TxTransferMulti:
		t.Payload = &TransferMultiPayload{}
		return t.Payload.Decode(r)
	default:
		return fmt.Errorf("could not decode transaction with type: %d", t.TxType)
	}
}

// Hash calculates the transaction hash.
func (t *Tx) Hash() chainhash.Hash {
	buf := bytes.NewBuffer([]byte{})
	_ = t.Encode(buf)
	return chainhash.DoubleHashH(buf.Bytes())
}
