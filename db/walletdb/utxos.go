package walletdb

import (
	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/utils/serializer"
	"io"
)

var walletUtxosBucketKey = []byte("wallet-utxos")

type WalletUtxo struct {
	OutPoint p2p.OutPoint
	Path     string
	Owner    string
	Value    int64
}

func (utxo *WalletUtxo) Serialize(w io.Writer) error {
	err := utxo.OutPoint.Encode(w)
	if err != nil {
		return err
	}
	err = serializer.WriteVarString(w, utxo.Path)
	if err != nil {
		return err
	}
	err = serializer.WriteVarString(w, utxo.Owner)
	if err != nil {
		return err
	}
	err = serializer.WriteElement(w, utxo.Value)
	if err != nil {
		return err
	}
	return nil
}

func (utxo *WalletUtxo) Deserialize(r io.Reader) error {
	err := utxo.OutPoint.Decode(r)
	if err != nil {
		return err
	}
	utxo.Path, err = serializer.ReadVarString(r)
	if err != nil {
		return err
	}
	utxo.Owner, err = serializer.ReadVarString(r)
	if err != nil {
		return err
	}
	err = serializer.ReadElement(r, &utxo.Value)
	if err != nil {
		return err
	}
	return nil
}
