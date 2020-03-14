package votes_txpayload

import (
	"bytes"
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/txs/txpayloads"
	"github.com/olympus-protocol/ogen/utils/amount"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
	"io"
)

type PayloadUploadAndUpdate struct {
	WorkerID p2p.OutPoint
	PubKey   [48]byte
	Sig      [96]byte
	GovID    chainhash.Hash
	Approval bool
}

func (p *PayloadUploadAndUpdate) Serialize(w io.Writer) error {
	err := serializer.WriteElements(w, p.WorkerID, p.PubKey, p.Sig, p.GovID, p.Approval)
	if err != nil {
		return err
	}
	return nil
}

func (p *PayloadUploadAndUpdate) Deserialize(r io.Reader) error {
	err := serializer.ReadElements(r, &p.WorkerID, &p.PubKey, &p.Sig, &p.GovID, &p.Approval)
	if err != nil {
		return err
	}
	return nil
}

func (p *PayloadUploadAndUpdate) GetAggPubKey() (*bls.PublicKey, error) {
	return nil, txpayloads.ErrorNoMethodForPayload
}

func (p *PayloadUploadAndUpdate) GetPublicKeys() ([]*bls.PublicKey, error) {
	return nil, txpayloads.ErrorNoMethodForPayload
}

func (p *PayloadUploadAndUpdate) GetPublicKey() (*bls.PublicKey, error) {
	return bls.DeserializePublicKey(p.PubKey)
}

func (p *PayloadUploadAndUpdate) GetSignature() (*bls.Signature, error) {
	return bls.DeserializeSignature(p.Sig)
}

func (p *PayloadUploadAndUpdate) GetMessage() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	err := p.WorkerID.Encode(buf)
	if err != nil {
		return nil, err
	}
	err = serializer.WriteElement(buf, p.Approval)
	if err != nil {
		return nil, err
	}
	return chainhash.DoubleHashB(buf.Bytes()), nil
}

func (p *PayloadUploadAndUpdate) GetHashForDataMatch() (chainhash.Hash, error) {
	return p.WorkerID.Hash()
}

func (p *PayloadUploadAndUpdate) GetHashInvForDataMatch() ([]chainhash.Hash, error) {
	return nil, txpayloads.ErrorNoMethodForPayload
}

func (p *PayloadUploadAndUpdate) GetSpentAmount() (amount.AmountType, error) {
	return 0, txpayloads.ErrorNoMethodForPayload
}

type PayloadRevoke struct {
	WorkerID p2p.OutPoint
	PubKey   [48]byte
	Sig      [96]byte
	GovID    chainhash.Hash
}

func (p *PayloadRevoke) Serialize(w io.Writer) error {
	err := serializer.WriteElements(w, p.WorkerID, p.PubKey, p.Sig, p.GovID)
	if err != nil {
		return err
	}
	return nil
}

func (p *PayloadRevoke) Deserialize(r io.Reader) error {
	err := serializer.ReadElements(r, &p.WorkerID, &p.PubKey, &p.Sig, &p.GovID)
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
	buf := bytes.NewBuffer([]byte{})
	err := p.WorkerID.Encode(buf)
	if err != nil {
		return nil, err
	}
	return chainhash.DoubleHashB(buf.Bytes()), nil
}

func (p *PayloadRevoke) GetHashForDataMatch() (chainhash.Hash, error) {
	return p.WorkerID.Hash()
}

func (p *PayloadRevoke) GetHashInvForDataMatch() ([]chainhash.Hash, error) {
	return nil, txpayloads.ErrorNoMethodForPayload
}

func (p *PayloadRevoke) GetSpentAmount() (amount.AmountType, error) {
	return 0, txpayloads.ErrorNoMethodForPayload
}
