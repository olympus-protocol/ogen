package coins_txpayload

import (
	"bytes"
	"github.com/grupokindynos/ogen/bls"
	"github.com/grupokindynos/ogen/p2p"
	"github.com/grupokindynos/ogen/txs/txpayloads"
	"github.com/grupokindynos/ogen/utils/amount"
	"github.com/grupokindynos/ogen/utils/chainhash"
	"github.com/grupokindynos/ogen/utils/serializer"
	"io"
)

type Input struct {
	PrevOutpoint p2p.OutPoint
	Sig          [96]byte
	PubKey       [48]byte
}

func (in *Input) Serialize(w io.Writer) error {
	err := in.PrevOutpoint.Serialize(w)
	if err != nil {
		return err
	}
	err = serializer.WriteElements(w, in.Sig, in.PubKey)
	if err != nil {
		return err
	}
	return nil
}

func (in *Input) Deserialize(r io.Reader) error {
	err := in.PrevOutpoint.Deserialize(r)
	if err != nil {
		return err
	}
	err = serializer.ReadElements(r, &in.Sig, &in.PubKey)
	if err != nil {
		return err
	}
	return nil
}

func NewInput(prevOut p2p.OutPoint, sig [96]byte, pubKey [48]byte) Input {
	return Input{
		PrevOutpoint: prevOut,
		Sig:          sig,
		PubKey:       pubKey,
	}
}

type Output struct {
	Value   int64
	Address string
}

func (out *Output) Serialize(w io.Writer) error {
	err := serializer.WriteElement(w, out.Value)
	if err != nil {
		return err
	}
	err = serializer.WriteVarString(w, out.Address)
	if err != nil {
		return err
	}
	return nil
}

func (out *Output) Deserialize(r io.Reader) error {
	err := serializer.ReadElements(r, &out.Value)
	if err != nil {
		return err
	}
	out.Address, err = serializer.ReadVarString(r)
	if err != nil {
		return err
	}
	return nil
}

func NewOutput(value int64, addr string) Output {
	return Output{
		Value:   value,
		Address: addr,
	}
}

type PayloadTransfer struct {
	AggSig [96]byte
	TxIn   []Input
	TxOut  []Output
}

func (p *PayloadTransfer) Serialize(w io.Writer) (err error) {
	err = serializer.WriteElements(w, p.AggSig)
	if err != nil {
		return err
	}
	err = serializer.WriteVarInt(w, uint64(len(p.TxIn)))
	if err != nil {
		return err
	}
	for _, in := range p.TxIn {
		err = in.Serialize(w)
		if err != nil {
			return err
		}
	}
	err = serializer.WriteVarInt(w, uint64(len(p.TxOut)))
	if err != nil {
		return err
	}
	for _, out := range p.TxOut {
		err = out.Serialize(w)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *PayloadTransfer) Deserialize(r io.Reader) (err error) {
	err = serializer.ReadElements(r, &p.AggSig)
	if err != nil {
		return err
	}
	txInCount, err := serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	for i := uint64(0); i < txInCount; i++ {
		var in Input
		err := in.Deserialize(r)
		if err != nil {
			return err
		}
		p.TxIn = append(p.TxIn, in)
	}
	txOutCount, err := serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	for i := uint64(0); i < txOutCount; i++ {
		var out Output
		err := out.Deserialize(r)
		if err != nil {
			return err
		}
		p.TxOut = append(p.TxOut, out)
	}
	return nil
}

func (p *PayloadTransfer) GetAggPubKey() (*bls.PublicKey, error) {
	aggPubKey := bls.NewAggregatePublicKey()
	for _, input := range p.TxIn {
		pubKey, err := bls.DeserializePublicKey(input.PubKey)
		if err != nil {
			return nil, err
		}
		aggPubKey.AggregatePubKey(pubKey)
	}
	return aggPubKey, nil
}

func (p *PayloadTransfer) GetPublicKeys() ([]*bls.PublicKey, error) {
	var pubKeys []*bls.PublicKey
	for _, input := range p.TxIn {
		pubKey, err := bls.DeserializePublicKey(input.PubKey)
		if err != nil {
			return nil, err
		}
		pubKeys = append(pubKeys, pubKey)
	}
	return pubKeys, nil
}

func (p *PayloadTransfer) GetPublicKey() (*bls.PublicKey, error) {
	return nil, txpayloads.ErrorNoMethodForPayload
}

func (p *PayloadTransfer) GetSignature() (*bls.Signature, error) {
	return bls.DeserializeSignature(p.AggSig)
}

func (p *PayloadTransfer) GetMessage() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	var payloadCopy PayloadTransfer
	for _, input := range p.TxIn {
		input.Sig = [96]byte{}
		payloadCopy.TxIn = append(payloadCopy.TxIn, input)
	}
	for _, output := range p.TxOut {
		payloadCopy.TxOut = append(payloadCopy.TxOut, output)
	}
	payloadCopy.AggSig = [96]byte{}
	err := payloadCopy.Serialize(buf)
	if err != nil {
		return nil, err
	}
	return chainhash.DoubleHashB(buf.Bytes()), nil
}

func (p *PayloadTransfer) GetHashForDataMatch() (chainhash.Hash, error) {
	return chainhash.Hash{}, txpayloads.ErrorNoMethodForPayload
}

func (p *PayloadTransfer) GetHashInvForDataMatch() ([]chainhash.Hash, error) {
	var utxosHashes []chainhash.Hash
	for _, input := range p.TxIn {
		buf := bytes.NewBuffer([]byte{})
		err := input.PrevOutpoint.Serialize(buf)
		if err != nil {
			return nil, err
		}
		utxosHashes = append(utxosHashes, chainhash.DoubleHashH(buf.Bytes()))
	}
	return utxosHashes, nil
}

func (p *PayloadTransfer) GetSpentAmount() (amount.AmountType, error) {
	var spent amount.AmountType
	for _, out := range p.TxOut {
		spent += amount.AmountType(out.Value)
	}
	return spent, nil
}

type PayloadGenerate struct {
	TxOut []Output
}

func (p *PayloadGenerate) Serialize(w io.Writer) (err error) {
	err = serializer.WriteVarInt(w, uint64(len(p.TxOut)))
	if err != nil {
		return err
	}
	for _, out := range p.TxOut {
		err = out.Serialize(w)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *PayloadGenerate) Deserialize(r io.Reader) (err error) {
	txOutCount, err := serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	for i := uint64(0); i < txOutCount; i++ {
		var out Output
		err := out.Deserialize(r)
		if err != nil {
			return err
		}
		p.TxOut = append(p.TxOut, out)
	}
	return nil
}

func (p *PayloadGenerate) GetAggPubKey() (*bls.PublicKey, error) {
	return nil, txpayloads.ErrorNoMethodForPayload
}

func (p *PayloadGenerate) GetPublicKeys() ([]*bls.PublicKey, error) {
	return nil, txpayloads.ErrorNoMethodForPayload
}

func (p *PayloadGenerate) GetPublicKey() (*bls.PublicKey, error) {
	return nil, txpayloads.ErrorNoMethodForPayload
}

func (p *PayloadGenerate) GetSignature() (*bls.Signature, error) {
	return nil, txpayloads.ErrorNoMethodForPayload
}

func (p *PayloadGenerate) GetMessage() ([]byte, error) {
	return nil, txpayloads.ErrorNoMethodForPayload
}

func (p *PayloadGenerate) GetHashForDataMatch() (chainhash.Hash, error) {
	return chainhash.Hash{}, txpayloads.ErrorNoMethodForPayload
}

func (p *PayloadGenerate) GetHashInvForDataMatch() ([]chainhash.Hash, error) {
	return nil, txpayloads.ErrorNoMethodForPayload
}

func (p *PayloadGenerate) GetSpentAmount() (amount.AmountType, error) {
	return 0, txpayloads.ErrorNoMethodForPayload
}

type PaymentType int32

const (
	PayToNetwork PaymentType = iota
	PayToProposals
	PayProfits
)

type PayloadPay struct {
	PaymentType
	TxIn  []Input
	TxOut []Output
}

func (p *PayloadPay) Serialize(w io.Writer) (err error) {
	err = serializer.WriteElements(w, p.PaymentType)
	if err != nil {
		return err
	}
	err = serializer.WriteVarInt(w, uint64(len(p.TxIn)))
	if err != nil {
		return err
	}
	for _, in := range p.TxIn {
		err = in.Serialize(w)
		if err != nil {
			return err
		}
	}
	err = serializer.WriteVarInt(w, uint64(len(p.TxOut)))
	if err != nil {
		return err
	}
	for _, out := range p.TxOut {
		err = out.Serialize(w)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *PayloadPay) Deserialize(r io.Reader) (err error) {
	err = serializer.ReadElements(r, &p.PaymentType)
	if err != nil {
		return err
	}
	txInCount, err := serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	for i := uint64(0); i < txInCount; i++ {
		var in Input
		err := in.Deserialize(r)
		if err != nil {
			return err
		}
		p.TxIn = append(p.TxIn, in)
	}
	txOutCount, err := serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	for i := uint64(0); i < txOutCount; i++ {
		var out Output
		err := out.Deserialize(r)
		if err != nil {
			return err
		}
		p.TxOut = append(p.TxOut, out)
	}
	return nil
}

func (p *PayloadPay) GetAggPubKey() (*bls.PublicKey, error) {
	return nil, txpayloads.ErrorNoMethodForPayload
}

func (p *PayloadPay) GetPublicKeys() ([]*bls.PublicKey, error) {
	return nil, txpayloads.ErrorNoMethodForPayload
}

func (p *PayloadPay) GetPublicKey() (*bls.PublicKey, error) {
	return nil, txpayloads.ErrorNoMethodForPayload
}

func (p *PayloadPay) GetSignature() (*bls.Signature, error) {
	return nil, txpayloads.ErrorNoMethodForPayload
}

func (p *PayloadPay) GetMessage() ([]byte, error) {
	return nil, txpayloads.ErrorNoMethodForPayload
}

func (p *PayloadPay) GetHashForDataMatch() (chainhash.Hash, error) {
	return chainhash.Hash{}, txpayloads.ErrorNoMethodForPayload
}

func (p *PayloadPay) GetHashInvForDataMatch() ([]chainhash.Hash, error) {
	return nil, txpayloads.ErrorNoMethodForPayload
}

func (p *PayloadPay) GetSpentAmount() (amount.AmountType, error) {
	return 0, txpayloads.ErrorNoMethodForPayload
}
