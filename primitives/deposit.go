package primitives

import "github.com/olympus-protocol/ogen/utils/chainhash"

// DepositData is the part of the deposit that is signed
type DepositData struct {
	// PublicKey is the key used for the validator.
	PublicKey []byte `ssz-size:"48"`

	// ProofOfPossession is the public key signed by the private key to prove that you
	// own the address and prevent rogue public-key attacks.
	ProofOfPossession []byte `ssz-size:"96"`

	// WithdrawalAddress is the address to withdraw to.
	WithdrawalAddress []byte `ssz-size:"20"`
}

// Deposit is a deposit a user can submit to queue as a validator.
type Deposit struct {
	// PublicKey is the public key of the address that is depositing.
	PublicKey []byte `ssz-size:"48"`

	// Signature is the signature signing the deposit data.
	Signature []byte `ssz-size:"96"`

	// Data is the data that describes the new validator.
	Data *DepositData
}

// Hash calculates the hash of the deposit
func (d *Deposit) Hash() chainhash.Hash {
	b, _ := d.MarshalSSZ()
	return chainhash.HashH(b)
}
