package primitives

import (
	"errors"
	bls_interface "github.com/olympus-protocol/ogen/pkg/bls/interface"

	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
)

var (
	// ErrorTxSize returned when the tx size is above MaxTransactionSingleSize
	ErrorTxSize = errors.New("tx size too big")

	// ErrorInvalidSignature returned when a tx signature is invalid.
	ErrorInvalidSignature = errors.New("invalid tx signature")
)

const (
	// MaxTransactionSize is the maximum size of the transaction information with a single transfer payload
	MaxTransactionSize = 188
)

// Tx represents a transaction on the blockchain.
type Tx struct {
	To            [20]byte
	FromPublicKey [48]byte
	Amount        uint64
	Nonce         uint64
	Fee           uint64
	Signature     [96]byte
}

// Marshal encodes the data.
func (t *Tx) Marshal() ([]byte, error) {
	return t.MarshalSSZ()
}

// Unmarshal decodes the data.
func (t *Tx) Unmarshal(b []byte) error {
	return t.UnmarshalSSZ(b)
}

// Hash calculates the transaction hash.
func (t *Tx) Hash() chainhash.Hash {
	b, _ := t.Marshal()
	return chainhash.DoubleHashH(b)
}

// FromPubkeyHash calculates the hash of the from public key.
func (t Tx) FromPubkeyHash() ([20]byte, error) {
	pub, err := bls.CurrImplementation.PublicKeyFromBytes(t.FromPublicKey[:])
	if err != nil {
		return [20]byte{}, nil
	}
	return pub.Hash()
}

// SignatureMessage gets the message the needs to be signed.
func (t Tx) SignatureMessage() chainhash.Hash {
	cp := t
	cp.Signature = [96]byte{}
	b, _ := cp.Marshal()
	return chainhash.HashH(b)
}

// GetSignature returns the bls signature of the transaction.
func (t Tx) GetSignature() (bls_interface.Signature, error) {
	return bls.CurrImplementation.SignatureFromBytes(t.Signature[:])
}

// GetPublic returns the bls public key of the transaction.
func (t Tx) GetPublic() (bls_interface.PublicKey, error) {
	return bls.CurrImplementation.PublicKeyFromBytes(t.FromPublicKey[:])
}

// VerifySig verifies the signatures is valid.
func (t *Tx) VerifySig() error {

	sigMsg := t.SignatureMessage()

	sig, err := t.GetSignature()

	if err != nil {
		return err
	}

	pub, err := t.GetPublic()

	if err != nil {
		return err
	}

	valid := sig.Verify(pub, sigMsg[:])

	if !valid {
		return ErrorInvalidSignature
	}
	return nil
}
