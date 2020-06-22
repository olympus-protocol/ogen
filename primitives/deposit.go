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

// Hash calculates the hash of the deposit
func (d *Deposit) Hash() (chainhash.Hash, error) {
	b, err := d.MarshalSSZ()
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
