package users_txpayload

import (
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/txs/txpayloads"
	"github.com/olympus-protocol/ogen/utils/amount"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
	"io"
)

type PayloadUpload struct {
	PubKey [48]byte
	Sig    [96]byte
	Name   string
}

func (p *PayloadUpload) Serialize(w io.Writer) error {
	err := serializer.WriteElements(w, p.PubKey, p.Sig)
	if err != nil {
		return err
	}
	err = serializer.WriteVarString(w, p.Name)
	if err != nil {
		return err
	}
	return nil
}

func (p *PayloadUpload) Deserialize(r io.Reader) error {
	err := serializer.ReadElements(r, &p.PubKey, &p.Sig)
	if err != nil {
		return err
	}
	p.Name, err = serializer.ReadVarString(r)
	if err != nil {
		return err
	}
	return nil
}

func (p *PayloadUpload) GetAggPubKey() (*bls.PublicKey, error) {
	return nil, txpayloads.ErrorNoMethodForPayload
}

func (p *PayloadUpload) GetPublicKeys() ([]*bls.PublicKey, error) {
	return nil, txpayloads.ErrorNoMethodForPayload
}

func (p *PayloadUpload) GetPublicKey() (*bls.PublicKey, error) {
	return bls.DeserializePublicKey(p.PubKey)
}

func (p *PayloadUpload) GetSignature() (*bls.Signature, error) {
	return bls.DeserializeSignature(p.Sig)
}

func (p *PayloadUpload) GetMessage() ([]byte, error) {
	return chainhash.DoubleHashB([]byte(p.Name)), nil
}

func (p *PayloadUpload) GetHashForDataMatch() (chainhash.Hash, error) {
	return chainhash.DoubleHashH([]byte(p.Name)), nil
}

func (p *PayloadUpload) GetHashInvForDataMatch() ([]chainhash.Hash, error) {
	return nil, txpayloads.ErrorNoMethodForPayload
}

func (p *PayloadUpload) GetSpentAmount() (amount.AmountType, error) {
	return 0, txpayloads.ErrorNoMethodForPayload
}

type PayloadUpdate struct {
	NewPubKey [48]byte
	PubKey    [48]byte
	Sig       [96]byte
	Name      string
}

func (p *PayloadUpdate) Serialize(w io.Writer) error {
	err := serializer.WriteElements(w, p.NewPubKey, p.PubKey, p.Sig)
	if err != nil {
		return err
	}
	err = serializer.WriteVarString(w, p.Name)
	if err != nil {
		return err
	}
	return nil
}

func (p *PayloadUpdate) Deserialize(r io.Reader) error {
	err := serializer.ReadElements(r, &p.NewPubKey, &p.PubKey, &p.Sig)
	if err != nil {
		return err
	}
	p.Name, err = serializer.ReadVarString(r)
	if err != nil {
		return err
	}
	return nil
}

func (p *PayloadUpdate) GetAggPubKey() (*bls.PublicKey, error) {
	aggPubKey := bls.NewAggregatePublicKey()
	oldPubKey, err := bls.DeserializePublicKey(p.PubKey)
	if err != nil {
		return nil, err
	}
	newPubKey, err := bls.DeserializePublicKey(p.NewPubKey)
	if err != nil {
		return nil, err
	}
	aggPubKey.AggregatePubKey(oldPubKey)
	aggPubKey.AggregatePubKey(newPubKey)
	return aggPubKey, nil
}

func (p *PayloadUpdate) GetPublicKeys() ([]*bls.PublicKey, error) {
	return nil, txpayloads.ErrorNoMethodForPayload
}

func (p *PayloadUpdate) GetPublicKey() (*bls.PublicKey, error) {
	return bls.DeserializePublicKey(p.PubKey)
}

func (p *PayloadUpdate) GetSignature() (*bls.Signature, error) {
	return bls.DeserializeSignature(p.Sig)
}

func (p *PayloadUpdate) GetMessage() ([]byte, error) {
	return chainhash.DoubleHashB([]byte(p.Name)), nil
}

func (p *PayloadUpdate) GetHashForDataMatch() (chainhash.Hash, error) {
	return chainhash.DoubleHashH([]byte(p.Name)), nil
}

func (p *PayloadUpdate) GetHashInvForDataMatch() ([]chainhash.Hash, error) {
	return nil, txpayloads.ErrorNoMethodForPayload
}

func (p *PayloadUpdate) GetSpentAmount() (amount.AmountType, error) {
	return 0, txpayloads.ErrorNoMethodForPayload
}

type PayloadRevoke struct {
	Sig  [96]byte
	Name string
}

func (p *PayloadRevoke) Serialize(w io.Writer) error {
	err := serializer.WriteElement(w, p.Sig)
	if err != nil {
		return err
	}
	err = serializer.WriteVarString(w, p.Name)
	if err != nil {
		return err
	}
	return nil
}

func (p *PayloadRevoke) Deserialize(r io.Reader) error {
	err := serializer.ReadElement(r, &p.Sig)
	if err != nil {
		return err
	}
	p.Name, err = serializer.ReadVarString(r)
	if err != nil {
		return err
	}
	return nil
}

func (p *PayloadRevoke) GetAggPubKey() (*bls.PublicKey, error) {
	return nil, txpayloads.ErrorNoMethodForPayload
}

func (p *PayloadRevoke) GetPublicKeys() ([]*bls.PublicKey, error) {
	return nil, txpayloads.ErrorNoMethodForPayload
}

func (p *PayloadRevoke) GetPublicKey() (*bls.PublicKey, error) {
	return nil, txpayloads.ErrorNoMethodForPayload
}

func (p *PayloadRevoke) GetSignature() (*bls.Signature, error) {
	return bls.DeserializeSignature(p.Sig)
}

func (p *PayloadRevoke) GetMessage() ([]byte, error) {
	return chainhash.DoubleHashB([]byte(p.Name)), nil
}

func (p *PayloadRevoke) GetHashForDataMatch() (chainhash.Hash, error) {
	return chainhash.DoubleHashH([]byte(p.Name)), nil
}

func (p *PayloadRevoke) GetHashInvForDataMatch() ([]chainhash.Hash, error) {
	return nil, txpayloads.ErrorNoMethodForPayload
}

func (p *PayloadRevoke) GetSpentAmount() (amount.AmountType, error) {
	return 0, txpayloads.ErrorNoMethodForPayload
}
