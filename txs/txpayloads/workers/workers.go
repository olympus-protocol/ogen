package workers_txpayload

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
	Utxo   p2p.OutPoint
	PubKey [48]byte
	Sig    [96]byte
	IP     string
}

func (p *PayloadUploadAndUpdate) Serialize(w io.Writer) error {
	err := serializer.WriteElements(w, p.Utxo, p.PubKey, p.Sig)
	if err != nil {
		return err
	}
	err = serializer.WriteVarString(w, p.IP)
	if err != nil {
		return err
	}
	return nil
}

func (p *PayloadUploadAndUpdate) Deserialize(r io.Reader) error {
	err := serializer.ReadElements(r, &p.Utxo, &p.PubKey, &p.Sig)
	if err != nil {
		return err
	}
	p.IP, err = serializer.ReadVarString(r)
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
	err := p.Utxo.Serialize(buf)
	if err != nil {
		return nil, err
	}
	return chainhash.DoubleHashB(buf.Bytes()), nil
}

func (p *PayloadUploadAndUpdate) GetHashForDataMatch() (chainhash.Hash, error) {
	return p.Utxo.Hash()
}

func (p *PayloadUploadAndUpdate) GetHashInvForDataMatch() ([]chainhash.Hash, error) {
	return nil, txpayloads.ErrorNoMethodForPayload
}

func (p *PayloadUploadAndUpdate) GetSpentAmount() (amount.AmountType, error) {
	return 0, txpayloads.ErrorNoMethodForPayload
}

type PayloadRevoke struct {
	PubKey   [48]byte
	Sig      [96]byte
	WorkerID chainhash.Hash
}

func (p *PayloadRevoke) Serialize(w io.Writer) error {
	err := serializer.WriteElements(w, p.PubKey, p.Sig, p.WorkerID)
	if err != nil {
		return err
	}
	return nil
}

func (p *PayloadRevoke) Deserialize(r io.Reader) error {
	err := serializer.ReadElements(r, &p.PubKey, &p.Sig, &p.WorkerID)
	if err != nil {
		return err
	}
	return nil
}

func (p *PayloadRevoke) GetPublicKeys() ([]*bls.PublicKey, error) {
	return nil, txpayloads.ErrorNoMethodForPayload
}

func (p *PayloadRevoke) GetAggPubKey() (*bls.PublicKey, error) {
	return nil, txpayloads.ErrorNoMethodForPayload
}

func (p *PayloadRevoke) GetPublicKey() (*bls.PublicKey, error) {
	return bls.DeserializePublicKey(p.PubKey)
}

func (p *PayloadRevoke) GetSignature() (*bls.Signature, error) {
	return bls.DeserializeSignature(p.Sig)
}

func (p *PayloadRevoke) GetMessage() ([]byte, error) {
	return p.WorkerID.CloneBytes(), nil
}

func (p *PayloadRevoke) GetHashForDataMatch() (chainhash.Hash, error) {
	return p.WorkerID, nil
}

func (p *PayloadRevoke) GetHashInvForDataMatch() ([]chainhash.Hash, error) {
	return nil, txpayloads.ErrorNoMethodForPayload
}

func (p *PayloadRevoke) GetSpentAmount() (amount.AmountType, error) {
	return 0, txpayloads.ErrorNoMethodForPayload
}
