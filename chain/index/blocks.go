package index

import (
	"bytes"
	"github.com/grupokindynos/ogen/db/blockdb"
	"github.com/grupokindynos/ogen/p2p"
	"github.com/grupokindynos/ogen/utils/chainhash"
	"github.com/grupokindynos/ogen/utils/serializer"
	"io"
	"sync"
)

type BlockRow struct {
	Header  p2p.BlockHeader
	Locator blockdb.BlockLocation
	Height  int32
}

func (br *BlockRow) Serialize(w io.Writer) error {
	err := br.Locator.Serialize(w)
	if err != nil {
		return err
	}
	err = br.Header.Serialize(w)
	if err != nil {
		return err
	}
	return nil
}

func (br *BlockRow) Deserialize(r io.Reader) error {
	err := br.Locator.Deserialize(r)
	if err != nil {
		return err
	}
	err = br.Header.Deserialize(r)
	if err != nil {
		return err
	}
	return nil
}

func NewBlockRow(locator blockdb.BlockLocation, header p2p.BlockHeader) *BlockRow {
	return &BlockRow{
		Header:  header,
		Locator: locator,
	}
}

type BlockIndex struct {
	lock  sync.Mutex
	Index map[chainhash.Hash]*BlockRow
}

func (i *BlockIndex) Serialize(w io.Writer) error {
	err := serializer.WriteVarInt(w, uint64(len(i.Index)))
	if err != nil {
		return err
	}
	for _, row := range i.Index {
		err = row.Serialize(w)
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *BlockIndex) Deserialize(r io.Reader) error {
	buf, _ := r.(*bytes.Buffer)
	if buf.Len() > 0 {
		count, err := serializer.ReadVarInt(r)
		if err != nil {
			return err
		}
		i.Index = make(map[chainhash.Hash]*BlockRow, count)
		for k := uint64(0); k < count; k++ {
			var row *BlockRow
			err = row.Deserialize(r)
			if err != nil {
				return err
			}
			err = i.Add(row)
			if err != nil {
				return err
			}
		}
		return nil
	}
	return nil
}

func (i *BlockIndex) Have(hash chainhash.Hash) bool {
	i.lock.Lock()
	_, ok := i.Index[hash]
	i.lock.Unlock()
	return ok
}

func (i *BlockIndex) Add(row *BlockRow) error {
	blockHash, err := row.Header.Hash()
	if err != nil {
		return err
	}
	i.lock.Lock()
	i.Index[blockHash] = row
	i.lock.Unlock()
	return nil
}

func InitBlocksIndex() *BlockIndex {
	return &BlockIndex{
		Index: make(map[chainhash.Hash]*BlockRow),
	}
}
