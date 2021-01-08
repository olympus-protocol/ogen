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

// MultiSignatureTx represents a transaction on the blockchain using a multi signature
type MultiSignatureTx struct {
	To        [20]byte
	Amount    uint64
	Nonce     uint64
	Fee       uint64
	Signature *multisig.Multisig
}

// Marshal encodes the data.
func (m *MultiSignatureTx) Marshal() ([]byte, error) {
	return m.MarshalSSZ()
}

// Unmarshal decodes the data.
func (m *MultiSignatureTx) Unmarshal(b []byte) error {
	return m.UnmarshalSSZ(b)
}

// Hash calculates the transaction hash.
func (m *MultiSignatureTx) Hash() chainhash.Hash {
	b, _ := m.Marshal()
	return chainhash.DoubleHashH(b)
}

// FromPubkeyHash calculates the hash of the from public key.
func (m MultiSignatureTx) FromPubkeyHash() ([20]byte, error) {
	return m.Signature.PublicKey.Hash()
}

// SignatureMessage gets the message the needs to be signed.
func (m MultiSignatureTx) SignatureMessage() chainhash.Hash {
	buf := make([]byte, 44)
	copy(buf[:], m.To[:])
	binary.LittleEndian.PutUint64(buf, m.Nonce)
	binary.LittleEndian.PutUint64(buf, m.Amount)
	binary.LittleEndian.PutUint64(buf, m.Fee)
	return chainhash.HashH(buf)
}

// VerifySig verifies the signatures is valid.
func (m *MultiSignatureTx) VerifySig() error {

	sigMsg := m.SignatureMessage()

	valid := m.Signature.Verify(sigMsg[:])

	if !valid {
		return ErrorMultiInvalidSignature
	}
	return nil
}
