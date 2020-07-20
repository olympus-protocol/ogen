package primitives

import (
	"github.com/golang/snappy"
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/prysmaticlabs/go-ssz"
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

// Marshal encodes the data.
func (d *Deposit) Marshal() ([]byte, error) {
	b, err := ssz.Marshal(d)
	if err != nil {
		return nil, err
	}
	return snappy.Encode(nil, b), nil
}

// Unmarshal decodes the data.
func (d *Deposit) Unmarshal(b []byte) error {
	de, err := snappy.Decode(nil, b)
	if err != nil {
		return err
	}
	return ssz.Unmarshal(de, d)
}

func (d *Deposit) GetPublicKey() (*bls.PublicKey, error) {
	return bls.PublicKeyFromBytes(d.PublicKey)
}

func (d *Deposit) GetSignature() (*bls.Signature, error) {
	return bls.SignatureFromBytes(d.Signature)
}

// Hash calculates the hash of the deposit
func (d *Deposit) Hash() chainhash.Hash {
	hash, _ := ssz.HashTreeRoot(d)
	return chainhash.Hash(hash)
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

// Marshal encodes the data.
func (d *DepositData) Marshal() ([]byte, error) {
	return ssz.Marshal(d)
}

// Unmarshal decodes the data.
func (d *DepositData) Unmarshal(b []byte) error {
	de, err := snappy.Decode(nil, b)
	if err != nil {
		return err
	}
	return ssz.Unmarshal(de, d)
}

func (d *DepositData) GetPublicKey() (*bls.PublicKey, error) {
	return bls.PublicKeyFromBytes(d.PublicKey)
}

func (d *DepositData) GetSignature() (*bls.Signature, error) {
	return bls.SignatureFromBytes(d.ProofOfPossession)
}
