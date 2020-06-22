package primitives

import (
	ssz "github.com/ferranbt/fastssz"
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

// Deposit is a deposit a user can submit to queue as a validator.
type Deposit struct {
	// PublicKey is the public key of the address that is depositing.
	PublicKey bls.PublicKey

	// Signature is the signature signing the deposit data.
	Signature bls.Signature

	// Data is the data that describes the new validator.
	Data DepositData
	ssz.Marshaler
	ssz.Unmarshaler
}

// Marshal serializes the struct to bytes
func (d *Deposit) Marshal() ([]byte, error) {
	return d.MarshalSSZ()
}

// Unmarshal deserializes the struct from bytes
func (d *Deposit) Unmarshal(b []byte) error {
	return d.UnmarshalSSZ(b)
}

// Hash calculates the hash of the deposit
func (d *Deposit) Hash() (chainhash.Hash, error) {
	b, err := d.Marshal()
	if err != nil {
		return chainhash.Hash{}, err
	}
	return chainhash.HashH(b), nil
}

// DepositData is the part of the deposit that is signed
type DepositData struct {
	// PublicKey is the key used for the validator.
	PublicKey bls.PublicKey

	// ProofOfPossession is the public key signed by the private key to prove that you
	// own the address and prevent rogue public-key attacks.
	ProofOfPossession bls.Signature

	// WithdrawalAddress is the address to withdraw to.
	WithdrawalAddress [20]byte

	ssz.Marshaler
	ssz.Unmarshaler
}

// Marshal serializes the struct to bytes
func (d *DepositData) Marshal() ([]byte, error) {
	return d.Marshal()
}

// Unmarshal deserializes the struct from bytes
func (d *DepositData) Unmarshal(b []byte) error {
	return d.Unmarshal(b)
}
