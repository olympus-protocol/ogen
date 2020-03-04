package primitives

import (
	"bytes"
	"fmt"
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
	"io"
	"time"
)

const (
	maxBlockSize = 1024 * 512 // 512 kilobytes
)

type Block struct {
	Header    BlockHeader
	Txs       []Tx
	PubKey    [48]byte
	Signature [96]byte
}

func (b *Block) MinerPubKey() (*bls.PublicKey, error) {
	return bls.DeserializePublicKey(b.PubKey)
}

func (b *Block) MinerSig() (*bls.Signature, error) {
	return bls.DeserializeSignature(b.Signature)
}

func (b *Block) GetTime() time.Time {
	return b.Header.Timestamp
}

func (b *Block) GetTx(index int32) *Tx {
	return &b.Txs[index]
}

func (b *Block) Hash() chainhash.Hash {
	return b.Header.Hash()
}

func (m *Block) Encode(w io.Writer) error {
	err := m.Header.Serialize(w)
	if err != nil {
		return err
	}
	err = serializer.WriteVarInt(w, uint64(len(m.Txs)))
	if err != nil {
		return err
	}
	for _, tx := range m.Txs {
		err := tx.Encode(w)
		if err != nil {
			return err
		}
	}
	err = serializer.WriteElements(w, m.PubKey, m.Signature)
	if err != nil {
		return err
	}
	return nil
}



func (m *Block) Decode(r io.Reader) error {
	buf, ok := r.(*bytes.Buffer)
	if !ok {
		return fmt.Errorf("MsgBlock.Decode reader is not a " +
			"*bytes.Buffer")
	}
	err := m.Header.Deserialize(r)
	if err != nil {
		return err
	}
	txCount, err := serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	m.Txs = make([]Tx, txCount)
	for i := uint64(0); i < txCount; i++ {
		var tx Tx
		err := tx.Decode(r)
		if err != nil {
			return err
		}
		m.Txs[i] = tx
	}
	err = serializer.ReadElements(buf, &m.PubKey, &m.Signature)
	if err != nil {
		return err
	}
	return nil
}
