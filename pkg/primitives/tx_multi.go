package primitives

import (
	"encoding/binary"
	"errors"
	"github.com/olympus-protocol/ogen/pkg/bls/multisig"

	"github.com/olympus-protocol/ogen/pkg/chainhash"
)

var (
	// ErrorMultiInvalidSignature returned when a tx signature is invalid.
	ErrorMultiInvalidSignature = errors.New("invalid tx multi signature")
)

const (
	// MaxTransactionSize is the maximum size of the transaction information with a single transfer payload
	MaxTransactionMultiSize = multisig.MaxMultisigSize + 20 + 8 + 8 + 8
)

// TxMulti represents a transaction on the blockchain using a multi signature
type TxMulti struct {
	To        [20]byte
	Amount    uint64
	Nonce     uint64
	Fee       uint64
	Signature *multisig.Multisig
}

// Marshal encodes the data.
func (t *TxMulti) Marshal() ([]byte, error) {
	return t.MarshalSSZ()
}

// Unmarshal decodes the data.
func (t *TxMulti) Unmarshal(b []byte) error {
	return t.UnmarshalSSZ(b)
}

// Hash calculates the transaction hash.
func (t *TxMulti) Hash() chainhash.Hash {
	b, _ := t.Marshal()
	return chainhash.DoubleHashH(b)
}

// FromPubkeyHash calculates the hash of the from public key.
func (t TxMulti) FromPubkeyHash() ([20]byte, error) {
	return t.Signature.PublicKey.Hash()
}

// SignatureMessage gets the message the needs to be signed.
func (t TxMulti) SignatureMessage() chainhash.Hash {
	buf := make([]byte, 44)
	copy(buf[:], t.To[:])
	binary.LittleEndian.PutUint64(buf, t.Nonce)
	binary.LittleEndian.PutUint64(buf, t.Amount)
	binary.LittleEndian.PutUint64(buf, t.Fee)
	return chainhash.HashH(buf)
}

// VerifySig verifies the signatures is valid.
func (t *TxMulti) VerifySig() error {

	sigMsg := t.SignatureMessage()

	valid := t.Signature.Verify(sigMsg[:])

	if !valid {
		return ErrorMultiInvalidSignature
	}
	return nil
}
