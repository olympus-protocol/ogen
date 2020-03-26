package gov_txpayload

import (
	"bytes"
	"io"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/txs/txpayloads"
	"github.com/olympus-protocol/ogen/utils/amount"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
)

type PayloadUpload struct {
	BurnedUtxo    primitives.OutPoint
	PubKey        [48]byte
	Sig           [96]byte
	Name          string
	URL           string
	PayoutAddress string
	Amount        int64
	Cycles        int32
}

func (p *PayloadUpload) Serialize(w io.Writer) error {
	err := serializer.WriteElements(w, p.BurnedUtxo, p.PubKey, p.Sig)
	if err != nil {
		return err
	}
	err = serializer.WriteVarString(w, p.Name)
	if err != nil {
		return err
	}
	err = serializer.WriteVarString(w, p.URL)
	if err != nil {
		return err
	}
	err = serializer.WriteVarString(w, p.PayoutAddress)
	if err != nil {
		return err
	}
	err = serializer.WriteElements(w, p.Amount, p.Cycles)
	if err != nil {
		return err
	}
	return nil
}

func (p *PayloadUpload) Deserialize(r io.Reader) error {
	err := serializer.ReadElements(r, &p.BurnedUtxo, &p.PubKey, &p.Sig)
	if err != nil {
		return err
	}
	p.Name, err = serializer.ReadVarString(r)
	if err != nil {
		return err
	}
	p.URL, err = serializer.ReadVarString(r)
	if err != nil {
		return err
	}
	p.PayoutAddress, err = serializer.ReadVarString(r)
	if err != nil {
		return err
	}
	err = serializer.ReadElements(r, &p.Amount, &p.Cycles)
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
	var payloadCopy PayloadUpload
	payloadCopy = *p
	buf := bytes.NewBuffer([]byte{})
	payloadCopy.Sig = [96]byte{}
	err := payloadCopy.Serialize(buf)
	if err != nil {
		return nil, err
	}
	return chainhash.DoubleHashB(buf.Bytes()), nil
}

func (p *PayloadUpload) GetHashForDataMatch() (chainhash.Hash, error) {
	return p.BurnedUtxo.Hash()
}

func (p *PayloadUpload) GetHashInvForDataMatch() ([]chainhash.Hash, error) {
	return nil, txpayloads.ErrorNoMethodForPayload
}

func (p *PayloadUpload) GetSpentAmount() (amount.AmountType, error) {
	return 0, txpayloads.ErrorNoMethodForPayload
}

type PayloadRevoke struct {
	GovID  chainhash.Hash
	PubKey [48]byte
	Sig    [96]byte
}

func (p *PayloadRevoke) Serialize(w io.Writer) error {
	err := serializer.WriteElements(w, p.GovID, p.PubKey, p.Sig)
	if err != nil {
		return err
	}
	return nil
}

func (p *PayloadRevoke) Deserialize(r io.Reader) error {
	err := serializer.ReadElements(r, &p.GovID, &p.PubKey, &p.Sig)
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
	return bls.DeserializePublicKey(p.PubKey)
}

func (p *PayloadRevoke) GetSignature() (*bls.Signature, error) {
	return bls.DeserializeSignature(p.Sig)
}

func (p *PayloadRevoke) GetMessage() ([]byte, error) {
	return p.GovID.CloneBytes(), nil
}

func (p *PayloadRevoke) GetHashForDataMatch() (chainhash.Hash, error) {
	return p.GovID, nil
}

func (p *PayloadRevoke) GetHashInvForDataMatch() ([]chainhash.Hash, error) {
	return nil, txpayloads.ErrorNoMethodForPayload
}

func (p *PayloadRevoke) GetSpentAmount() (amount.AmountType, error) {
	return 0, txpayloads.ErrorNoMethodForPayload
}
