package primitives

import (
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
)

// MaxDepositSize is the maximum amount of bytes a deposit can contain.
const MaxDepositSize = MaxDepositDataSize + 48 + 96

// Deposit is a deposit a user can submit to queue as a validator.
type Deposit struct {
	// PublicKey is the public key of the address that is depositing.
	PublicKey [48]byte

	// Signature is the signature signing the deposit data.
	Signature [96]byte

	// Data is the data that describes the new validator.
	Data *DepositData
}

// Marshal encodes the data.
func (d *Deposit) Marshal() ([]byte, error) {
	return d.MarshalSSZ()
}

// Unmarshal decodes the data.
func (d *Deposit) Unmarshal(b []byte) error {
	return d.UnmarshalSSZ(b)
}

// GetPublicKey returns the bls public key of the deposit.
func (d *Deposit) GetPublicKey() (*bls.PublicKey, error) {
	return bls.PublicKeyFromBytes(d.PublicKey[:])
}

// GetSignature returns the bls signature of the deposit.
func (d *Deposit) GetSignature() (*bls.Signature, error) {
	return bls.SignatureFromBytes(d.Signature[:])
}

// Hash calculates the hash of the deposit
func (d *Deposit) Hash() chainhash.Hash {
	b, _ := d.Marshal()
	return chainhash.HashH(b)
}

// MaxDepositDataSize is the maximum amount of bytes the deposit data can contain.
const MaxDepositDataSize = 164

// DepositData is the part of the deposit that is signed
type DepositData struct {
	// PublicKey is the key used for the validator.
	PublicKey [48]byte

	// ProofOfPossession is the public key signed by the private key to prove that you own the address and prevent rogue public-key attacks.
	ProofOfPossession [96]byte

	// WithdrawalAddress is the address to withdraw to.
	WithdrawalAddress [20]byte
}

// Marshal encodes the data.
func (d *DepositData) Marshal() ([]byte, error) {
	return d.MarshalSSZ()
}

// Unmarshal decodes the data.
func (d *DepositData) Unmarshal(b []byte) error {
	return d.UnmarshalSSZ(b)
}

// GetPublicKey returns the bls public key of the deposit data.
func (d *DepositData) GetPublicKey() (*bls.PublicKey, error) {
	return bls.PublicKeyFromBytes(d.PublicKey[:])
}

// GetSignature returns the bls signature of the deposit data.
func (d *DepositData) GetSignature() (*bls.Signature, error) {
	return bls.SignatureFromBytes(d.ProofOfPossession[:])
}
