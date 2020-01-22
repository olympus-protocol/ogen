package primitives

import (
	"bytes"
	"github.com/grupokindynos/ogen/bls"
	"github.com/grupokindynos/ogen/p2p"
	"github.com/grupokindynos/ogen/utils/chainhash"
	"time"
)

const (
	maxBlockSize = 1024 * 512 // 512 kilobytes
)

type Block struct {
	MsgBlock *p2p.MsgBlock
	Height   uint32
	Bytes    []byte
	Hash     chainhash.Hash
	Txs      []*Tx
}

func (b *Block) SetHeight(height uint32) {
	b.Height = height
}

func (b *Block) Header() *p2p.BlockHeader {
	return &b.MsgBlock.Header
}

func (b *Block) MinerPubKey() (*bls.PublicKey, error) {
	return bls.DeserializePublicKey(b.MsgBlock.PubKey)
}

func (b *Block) MinerSig() (*bls.Signature, error) {
	return bls.DeserializeSignature(b.MsgBlock.Signature)
}

func (b *Block) GetTime() time.Time {
	return b.MsgBlock.Header.Timestamp
}

func (b *Block) GetTx(index int32) *Tx {
	return b.Txs[index]
}

func NewBlockFromMsg(blockMsg *p2p.MsgBlock, blockHeight uint32) (*Block, error) {
	serializedBlock := bytes.NewBuffer([]byte{})
	err := blockMsg.Encode(serializedBlock)
	if err != nil {
		return nil, err
	}
	blockHash, err := blockMsg.Header.Hash()
	if err != nil {
		return nil, err
	}
	var txs []*Tx
	for i, txMsg := range blockMsg.Txs {
		tx, err := NewTxFromMsg(txMsg, int64(i))
		if err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	block := &Block{
		MsgBlock: blockMsg,
		Height:   blockHeight,
		Bytes:    serializedBlock.Bytes(),
		Hash:     blockHash,
		Txs:      txs,
	}
	return block, nil
}

func NewBlockFromBytes(blockBytes []byte, blockHeight uint32) (*Block, error) {
	buf := bytes.NewBuffer(blockBytes)
	var blockMsg p2p.MsgBlock
	err := blockMsg.Decode(buf)
	if err != nil {
		return nil, err
	}
	blockHash, err := blockMsg.Header.Hash()
	if err != nil {
		return nil, err
	}
	var txs []*Tx
	for i, txMsg := range blockMsg.Txs {
		tx, err := NewTxFromMsg(txMsg, int64(i))
		if err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	block := &Block{
		MsgBlock: &blockMsg,
		Height:   blockHeight,
		Bytes:    blockBytes,
		Hash:     blockHash,
		Txs:      txs,
	}
	return block, nil
}
