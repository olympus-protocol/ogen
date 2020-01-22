package primitives

import (
	"bytes"
	"github.com/grupokindynos/ogen/p2p"
	"github.com/grupokindynos/ogen/utils/chainhash"
)

type Tx struct {
	msgTx        *p2p.MsgTx
	txHash       chainhash.Hash
	txIndex      int64
	serializedTx []byte
}

func (t *Tx) GetVersion() int32 {
	return t.msgTx.TxVersion
}

func (t *Tx) GetType() p2p.TxType {
	return t.msgTx.TxType
}

func (t *Tx) GetAction() p2p.TxAction {
	return t.msgTx.TxAction
}

func (t *Tx) GetTime() int64 {
	return t.msgTx.Time
}

func (t *Tx) GetPayload() []byte {
	return t.msgTx.Payload
}

func (t *Tx) Hash() chainhash.Hash {
	return t.txHash
}

func (t *Tx) Index() int64 {
	return t.txIndex
}

func NewTxFromMsg(msgTx *p2p.MsgTx, txIndex int64) (*Tx, error) {
	serializedTx := bytes.NewBuffer([]byte{})
	err := msgTx.Encode(serializedTx)
	if err != nil {
		return nil, err
	}
	txHash, err := msgTx.TxHash()
	if err != nil {
		return nil, err
	}
	tx := &Tx{
		msgTx:        msgTx,
		txHash:       txHash,
		txIndex:      txIndex,
		serializedTx: serializedTx.Bytes(),
	}
	return tx, nil
}

func NewTxFromBytes(txBytes []byte, txIndex int64) (*Tx, error) {
	var msgTx p2p.MsgTx
	buf := bytes.NewBuffer(txBytes)
	err := msgTx.Decode(buf)
	if err != nil {
		return nil, err
	}
	txHash, err := msgTx.TxHash()
	if err != nil {
		return nil, err
	}
	tx := &Tx{
		msgTx:        &msgTx,
		txHash:       txHash,
		txIndex:      txIndex,
		serializedTx: txBytes,
	}
	return tx, nil
}
