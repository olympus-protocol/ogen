package primitives

import (
	"bytes"
	"fmt"
	"io"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
)

type TxType = int32

const (
	TxCoins TxType = iota + 1
	TxWorker
	TxGovernance
	TxVotes
	TxUsers
)

type CoinPayload struct {
	To            [20]byte
	FromPublicKey bls.PublicKey
	Amount        uint64
	Nonce         uint64
	Fee           uint64
	Signature     bls.Signature
}

func (c *CoinPayload) Hash() chainhash.Hash {
	buf := bytes.NewBuffer([]byte{})
	_ = c.Encode(buf)
	return chainhash.HashH(buf.Bytes())
}

func (c *CoinPayload) FromPubkeyHash() (out [20]byte) {
	pkS := c.FromPublicKey.Serialize()
	h := chainhash.HashH(pkS[:])
	copy(out[:], h[:20])
	return
}

func (c *CoinPayload) Encode(w io.Writer) error {
	if err := serializer.WriteElements(w, c.To); err != nil {
		return err
	}
	pubBytes := c.FromPublicKey.Serialize()
	sigBytes := c.Signature.Serialize()
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

func (c *CoinPayload) SignatureMessage() chainhash.Hash {
	buf := bytes.NewBuffer([]byte{})
	_ = serializer.WriteElements(buf, c.To, c.Nonce, c.FromPublicKey, c.Amount, c.Fee)
	return chainhash.HashH(buf.Bytes())
}

func (c *CoinPayload) VerifySig() error {
	sigMsg := c.SignatureMessage()

	valid, err := bls.VerifySig(&c.FromPublicKey, sigMsg[:], &c.Signature)
	if err != nil {
		return err
	}
	if !valid {
		return fmt.Errorf("signature is not valid")
	}

	return nil
}

func (c *CoinPayload) Decode(r io.Reader) error {
	if err := serializer.ReadElements(r, &c.To); err != nil {
		return err
	}
	var sigBytes [96]byte
	var pubBytes [48]byte
	_, err := r.Read(pubBytes[:])
	if err != nil {
		return err
	}
	pub, err := bls.DeserializePublicKey(pubBytes)
	if err != nil {
		return err
	}
	c.FromPublicKey = *pub
	_, err = r.Read(sigBytes[:])
	if err != nil {
		return err
	}
	sig, err := bls.DeserializeSignature(sigBytes)
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

var _ TxPayload = &CoinPayload{}

type GenesisPayload struct{}

func (g *GenesisPayload) Encode(w io.Writer) error { return nil }
func (g *GenesisPayload) Decode(r io.Reader) error { return nil }

type TxPayload interface {
	Encode(w io.Writer) error
	Decode(r io.Reader) error
}

type Tx struct {
	TxVersion int32
	TxType    TxType
	Payload   TxPayload
}

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

func (t *Tx) Decode(r io.Reader) error {
	err := serializer.ReadElements(r, &t.TxVersion, &t.TxType)
	if err != nil {
		return err
	}
	switch t.TxType {
	case TxCoins:
		t.Payload = &CoinPayload{}
		return t.Payload.Decode(r)
	default:
		return fmt.Errorf("could not decode transaction with type: %d", t.TxType)
	}
}

func (t *Tx) Hash() chainhash.Hash {
	buf := bytes.NewBuffer([]byte{})
	_ = t.Encode(buf)
	return chainhash.DoubleHashH(buf.Bytes())
}
