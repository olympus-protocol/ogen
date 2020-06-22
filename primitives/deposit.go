package primitives

import (
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

// Deposit is a deposit a user can submit to queue as a validator.
type Deposit struct {
	// PublicKey is the public key of the address that is depositing.
	PublicKey []byte

	// Signature is the signature signing the deposit data.
	Signature []byte

	// Data is the data that describes the new validator.
	Data DepositData
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
	PublicKey []byte

	// ProofOfPossession is the public key signed by the private key to prove that you
	// own the address and prevent rogue public-key attacks.
	ProofOfPossession []byte

	// WithdrawalAddress is the address to withdraw to.
	WithdrawalAddress [20]byte
}
