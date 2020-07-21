package primitives

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/golang/snappy"
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/prysmaticlabs/go-ssz"
)

var (
	// ErrorSingleTransferPayloadSize returned when the single transfer payload size is above MaxSingleTransferPayloadSize
	ErrorSingleTransferPayloadSize = errors.New("single transfer payload too big")
	// ErrorMultiTransferPayloadSize returned when the multi transfer payload size is above MaxMultipleTransferPayloadSize
	ErrorMultiTransferPayloadSize = errors.New("multi transfer payload too big")
	// ErrorTxSingleSize returned when the tx size is above MaxTransactionSingleSize
	ErrorTxSingleSize = errors.New("tx single transfer too big")
	// ErrorTxMultiSize returned when the tx size is above MaxTransactionMultiSize
	ErrorTxMultiSize = errors.New("tx multi transfer too big")
	// ErrorInvalidSignature returned when a tx signature is invalid.
	ErrorInvalidSignature = errors.New("invalid signature")
)

const (
	// MaxSingleTransferPayloadSize is the maximum payload size for a single transfer transaction.
	MaxSingleTransferPayloadSize = 196

	// MaxMultipleTransferPayloadSize is the maximum payload size for a multiple transfer transaction.
	// Multiple Transfers are m-of-n signed transactions. The maximum defined amount of keys available to be used is 7 keys with the signatures.
	MaxMultipleTransferPayloadSize = 1068

	// MaxTransactionSingleSize is the maximum size of the transaction information with a single transfer payload
	MaxTransactionSingleSize = 8 + MaxSingleTransferPayloadSize

	// MaxTransactionMultiSize is the maximum size of the transaction information with a multi transfer payload
	MaxTransactionMultiSize = 8 + MaxMultipleTransferPayloadSize
)

const (
	// TxTransferSingle represents a transaction sending money from a single
	// address to another address.
	TxTransferSingle uint32 = iota + 1

	// TxTransferMulti represents a transaction sending money from a multisig
	// address to another address.
	TxTransferMulti
)

// TransferSinglePayload is a transaction payload for sending money from
// a single address to another address.
type TransferSinglePayload struct {
	To            [20]byte
	FromPublicKey []byte
	Amount        uint64
	Nonce         uint64
	Fee           uint64
	Signature     []byte
}

// Hash calculates the transaction ID of the payload.
func (c TransferSinglePayload) Hash() chainhash.Hash {
	b, _ := c.Marshal()
	return chainhash.HashH(b)
}

// FromPubkeyHash calculates the hash of the from public key.
func (c TransferSinglePayload) FromPubkeyHash() (out [20]byte, err error) {
	pkS := c.FromPublicKey
	h := chainhash.HashH(pkS[:])
	copy(out[:], h[:20])
	return
}

// Marshal encodes the data.
func (c *TransferSinglePayload) Marshal() ([]byte, error) {
	b, err := ssz.Marshal(c)
	if err != nil {
		return nil, err
	}
	if len(b) > MaxSingleTransferPayloadSize {
		return nil, ErrorSingleTransferPayloadSize
	}
	return snappy.Encode(nil, b), nil
}

// Unmarshal decodes the data.
func (c *TransferSinglePayload) Unmarshal(b []byte) error {
	d, err := snappy.Decode(nil, b)
	if err != nil {
		return err
	}
	if len(d) > MaxSingleTransferPayloadSize {
		return ErrorSingleTransferPayloadSize
	}
	return ssz.Unmarshal(d, c)
}

// SignatureMessage gets the message the needs to be signed.
func (c TransferSinglePayload) SignatureMessage() chainhash.Hash {
	cp := c
	cp.Signature = make([]byte, 96)
	b, _ := cp.Marshal()
	return chainhash.HashH(b)
}

// GetSignature returns the bls signature of the transaction.
func (c TransferSinglePayload) GetSignature() (*bls.Signature, error) {
	return bls.SignatureFromBytes(c.Signature)
}

// GetPublic returns the bls public key of the transaction.
func (c TransferSinglePayload) GetPublic() (*bls.PublicKey, error) {
	return bls.PublicKeyFromBytes(c.FromPublicKey)
}

// VerifySig verifies the signatures is valid.
func (c *TransferSinglePayload) VerifySig() error {
	sigMsg := c.SignatureMessage()
	sig, err := c.GetSignature()
	if err != nil {
		return err
	}
	pub, err := c.GetPublic()
	if err != nil {
		return err
	}
	valid := sig.Verify(sigMsg[:], pub)
	if !valid {
		return ErrorInvalidSignature
	}

	return nil
}

// GetNonce gets the transaction nonce.
func (c TransferSinglePayload) GetNonce() uint64 {
	return c.Nonce
}

// GetAmount gets the transaction amount to send.
func (c TransferSinglePayload) GetAmount() uint64 {
	return c.Amount
}

// GetFee gets the transaction fee.
func (c TransferSinglePayload) GetFee() uint64 {
	return c.Fee
}

// GetToAccount gets the receiving acccount.
func (c TransferSinglePayload) GetToAccount() [20]byte {
	return c.To
}

// GetFromAddress gets the from address.
func (c TransferSinglePayload) GetFromAddress() ([20]byte, error) {
	return c.FromPubkeyHash()
}

var _ TxPayload = &TransferSinglePayload{}

// TransferMultiPayload represents a transfer from a multisig to
// another address.
type TransferMultiPayload struct {
	To       [20]byte
	Amount   uint64
	Nonce    uint64
	Fee      uint64
	MultiSig []byte
}

// Marshal encodes the data.
func (c *TransferMultiPayload) Marshal() ([]byte, error) {
	b, err := ssz.Marshal(c)
	if err != nil {
		return nil, err
	}
	if len(b) > MaxMultipleTransferPayloadSize {
		return nil, ErrorMultiTransferPayloadSize
	}
	return snappy.Encode(nil, b), nil
}

// Unmarshal decodes the data.
func (c *TransferMultiPayload) Unmarshal(b []byte) error {
	d, err := snappy.Decode(nil, b)
	if err != nil {
		return err
	}
	if len(d) > MaxMultipleTransferPayloadSize {
		return ErrorMultiTransferPayloadSize
	}
	return ssz.Unmarshal(d, c)
}

// Hash calculates the transaction ID of the payload.
func (c TransferMultiPayload) Hash() chainhash.Hash {
	b, _ := c.Marshal()
	return chainhash.HashH(b)
}

// FromPubkeyHash calculates the hash of the from public key.
func (c TransferMultiPayload) FromPubkeyHash() ([20]byte, error) {
	pub, err := c.GetPublic()
	if err != nil {
		return [20]byte{}, err
	}
	return pub.Hash()
}

// SignatureMessage gets the message the needs to be signed.
func (c TransferMultiPayload) SignatureMessage() chainhash.Hash {
	cp := c
	cp.MultiSig = []byte{}
	b, _ := cp.Marshal()
	return chainhash.HashH(b)
}

// GetPublic returns the bls public key of the multi signature transaction.
func (c TransferMultiPayload) GetPublic() (bls.FunctionalPublicKey, error) {
	sig, err := c.GetSignature()
	if err != nil {
		return nil, err
	}
	return sig.GetPublicKey()
}

// GetSignature returns the bls signature of the multi signature transaction.
func (c TransferMultiPayload) GetSignature() (bls.FunctionalSignature, error) {
	buf := bytes.NewBuffer(c.MultiSig)
	return bls.ReadFunctionalSignature(buf)
}

// VerifySig verifies the signatures is valid.
func (c TransferMultiPayload) VerifySig() error {
	sigMsg := c.SignatureMessage()
	sig, err := c.GetSignature()
	if err != nil {
		return err
	}
	valid := sig.Verify(sigMsg[:])
	if !valid {
		return ErrorInvalidSignature
	}

	return nil
}

// GetNonce gets the transaction nonce.
func (c TransferMultiPayload) GetNonce() uint64 {
	return c.Nonce
}

// GetAmount gets the transaction amount to send.
func (c TransferMultiPayload) GetAmount() uint64 {
	return c.Amount
}

// GetFee gets the transaction fee.
func (c TransferMultiPayload) GetFee() uint64 {
	return c.Fee
}

// GetToAccount gets the receiving acccount.
func (c TransferMultiPayload) GetToAccount() [20]byte {
	return c.To
}

var _ TxPayload = &TransferMultiPayload{}

// GenesisPayload is the payload of the genesis transaction.
type GenesisPayload struct{}

// TxPayload represents anything that can be used as a payload in a transaction.
type TxPayload interface {
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
	GetNonce() uint64
	GetAmount() uint64
	GetFee() uint64
	GetToAccount() [20]byte
	FromPubkeyHash() ([20]byte, error)
}

// Tx represents a transaction on the blockchain.
type Tx struct {
	Version uint32
	Type    uint32
	Payload []byte
}

// GetPayload returns the payload of the transaction.
func (t *Tx) GetPayload() (TxPayload, error) {
	switch t.Type {
	case TxTransferMulti:
		payload := new(TransferMultiPayload)
		payload.Unmarshal(t.Payload)
		return payload, nil
	case TxTransferSingle:
		payload := new(TransferSinglePayload)
		payload.Unmarshal(t.Payload)
		return payload, nil
	default:
		return nil, errors.New("unknown transaction type")
	}
}

// AppendPayload adds a specific payload to the transaction.
func (t *Tx) AppendPayload(p TxPayload) error {
	buf, err := p.Marshal()
	if err != nil {
		return err
	}
	t.Payload = buf
	return nil
}

// Marshal encodes the data.
func (t *Tx) Marshal() ([]byte, error) {
	b, err := ssz.Marshal(t)
	if err != nil {
		return nil, err
	}
	switch t.Type {
	case TxTransferSingle:
		if len(b) > MaxTransactionSingleSize {
			return nil, ErrorTxSingleSize
		}
	case TxTransferMulti:
		if len(b) > MaxTransactionMultiSize {
			return nil, ErrorTxMultiSize
		}
	}
	return snappy.Encode(nil, b), nil
}

// Unmarshal decodes the data.
func (t *Tx) Unmarshal(b []byte) error {
	d, err := snappy.Decode(nil, b)
	if err != nil {
		return err
	}
	txType := binary.LittleEndian.Uint32(d[8:])
	switch txType {
	case TxTransferSingle:
		if len(b) > MaxTransactionSingleSize {
			return ErrorTxSingleSize
		}
	case TxTransferMulti:
		if len(b) > MaxTransactionMultiSize {
			return ErrorTxMultiSize
		}
	}
	return ssz.Unmarshal(d, t)
}

// Hash calculates the transaction hash.
func (t *Tx) Hash() chainhash.Hash {
	b, _ := t.Marshal()
	return chainhash.DoubleHashH(b)
}
