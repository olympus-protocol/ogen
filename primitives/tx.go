package primitives

import (
	"errors"

	"github.com/golang/snappy"
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

var (
	// ErrorTxSize returned when the tx size is above MaxTransactionSingleSize
	ErrorTxSize = errors.New("tx size too big")

	// ErrorInvalidSignature returned when a tx signature is invalid.
	ErrorInvalidSignature = errors.New("invalid signature")
)

const (
	// MaxTransactionSize is the maximum size of the transaction information with a single transfer payload
	MaxTransactionSize = 204
)

// // TransferMultiPayload represents a transfer from a multisig to another address.
// type TransferMultiPayload struct {
// 	To       [20]byte `ssz-max:"20"`
// 	Amount   uint64
// 	Nonce    uint64
// 	Fee      uint64
// 	MultiSig *bls.Multisig
// }

// // Marshal encodes the data.
// func (c *TransferMultiPayload) Marshal() ([]byte, error) {
// 	b, err := c.MarshalSSZ()
// 	if err != nil {
// 		return nil, err
// 	}
// 	if len(b) > MaxMultipleTransferPayloadSize {
// 		return nil, ErrorMultiTransferPayloadSize
// 	}
// 	return snappy.Encode(nil, b), nil
// }

// // Unmarshal decodes the data.
// func (c *TransferMultiPayload) Unmarshal(b []byte) error {
// 	d, err := snappy.Decode(nil, b)
// 	if err != nil {
// 		return err
// 	}
// 	if len(d) > MaxMultipleTransferPayloadSize {
// 		return ErrorMultiTransferPayloadSize
// 	}
// 	return c.UnmarshalSSZ(d)
// }

// // Hash calculates the transaction ID of the payload.
// func (c TransferMultiPayload) Hash() chainhash.Hash {
// 	b, _ := c.Marshal()
// 	return chainhash.HashH(b)
// }

// // FromPubkeyHash calculates the hash of the from public key.
// func (c TransferMultiPayload) FromPubkeyHash() ([20]byte, error) {
// 	pub, err := c.GetPublic()
// 	if err != nil {
// 		return [20]byte{}, err
// 	}
// 	return pub.Hash()
// }

// // SignatureMessage gets the message the needs to be signed.
// func (c TransferMultiPayload) SignatureMessage() chainhash.Hash {
// 	cp := c
// 	cp.MultiSig = nil
// 	b, _ := cp.Marshal()
// 	return chainhash.HashH(b)
// }

// // GetPublic returns the bls public key of the multi signature transaction.
// func (c TransferMultiPayload) GetPublic() (bls.FunctionalPublicKey, error) {
// 	sig, err := c.GetSignature()
// 	if err != nil {
// 		return nil, err
// 	}
// 	return sig.GetPublicKey()
// }

// // GetSignature returns the bls signature of the multi signature transaction.
// func (c TransferMultiPayload) GetSignature() (bls.FunctionalSignature, error) {
// 	return c.MultiSig, nil
// }

// // VerifySig verifies the signatures is valid.
// func (c TransferMultiPayload) VerifySig() error {
// 	sigMsg := c.SignatureMessage()
// 	sig, err := c.GetSignature()
// 	if err != nil {
// 		return err
// 	}
// 	valid := sig.Verify(sigMsg[:])
// 	if !valid {
// 		return ErrorInvalidSignature
// 	}

// 	return nil
// }

// // GetNonce gets the transaction nonce.
// func (c TransferMultiPayload) GetNonce() uint64 {
// 	return c.Nonce
// }

// // GetAmount gets the transaction amount to send.
// func (c TransferMultiPayload) GetAmount() uint64 {
// 	return c.Amount
// }

// // GetFee gets the transaction fee.
// func (c TransferMultiPayload) GetFee() uint64 {
// 	return c.Fee
// }

// // GetToAccount gets the receiving acccount.
// func (c TransferMultiPayload) GetToAccount() [20]byte {
// 	return c.To
// }

// Tx represents a transaction on the blockchain.
type Tx struct {
	Version       uint64
	Type          uint64
	To            [20]byte `ssz-size:"20"`
	FromPublicKey [48]byte `ssz-size:"48"`
	Amount        uint64
	Nonce         uint64
	Fee           uint64
	Signature     [96]byte `ssz-size:"96"`
}

// Marshal encodes the data.
func (t *Tx) Marshal() ([]byte, error) {
	b, err := t.MarshalSSZ()
	if err != nil {
		return nil, err
	}
	if len(b) > MaxTransactionSize {
		return nil, ErrorTxSize
	}
	return snappy.Encode(nil, b), nil
}

// Unmarshal decodes the data.
func (t *Tx) Unmarshal(b []byte) error {
	d, err := snappy.Decode(nil, b)
	if err != nil {
		return err
	}
	if len(d) > MaxTransactionSize {
		return ErrorTxSize
	}
	return t.UnmarshalSSZ(d)
}

// Hash calculates the transaction hash.
func (t *Tx) Hash() chainhash.Hash {
	b, _ := t.Marshal()
	return chainhash.DoubleHashH(b)
}

// FromPubkeyHash calculates the hash of the from public key.
func (t Tx) FromPubkeyHash() ([20]byte, error) {
	pub, err := bls.PublicKeyFromBytes(t.FromPublicKey)
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
func (t Tx) GetSignature() (*bls.Signature, error) {
	return bls.SignatureFromBytes(t.Signature)
}

// GetSignature returns the bls signature of the transaction.
func (t Tx) GetPublic() (*bls.PublicKey, error) {
	return bls.PublicKeyFromBytes(t.FromPublicKey)
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
	valid := sig.Verify(sigMsg[:], pub)
	if !valid {
		return ErrorInvalidSignature
	}

	return nil
}
