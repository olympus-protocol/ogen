package blockdb

import (
	"bytes"

	"github.com/dgraph-io/badger"
	"github.com/olympus-protocol/ogen/logger"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

type BlockDB struct {
	log        *logger.Logger
	rawBlockDb *RawBlockDB
	badgerdb   *badger.DB
	params     params.ChainParams
}

func (bdb *BlockDB) Close() {
	_ = bdb.badgerdb.Close()
}

// GetRawBlock gets a block from the database.
func (bdb *BlockDB) GetRawBlock(locator BlockLocation, hash chainhash.Hash) (*primitives.Block, error) {
	blockBytes, err := bdb.rawBlockDb.read(hash, locator)
	if err != nil {
		return nil, err
	}

	block := new(primitives.Block)
	if err := block.Decode(bytes.NewBuffer(blockBytes)); err != nil {
		return nil, err
	}

	return block, nil
}

func (bdb *BlockDB) AddRawBlock(block *primitives.Block) (*BlockLocation, error) {
	locator, err := bdb.rawBlockDb.AddBlock(block)
	if err != nil {
		return nil, err
	}
	locatorBuf := bytes.NewBuffer([]byte{})
	err = locator.Serialize(locatorBuf)
	if err != nil {
		return nil, err
	}
	err = bdb.addBlock(locatorBuf.Bytes(), block.Header)
	if err != nil {
		return nil, err
	}
	return locator, nil
}

// NewBlockDB returns a database instance with a rawBlockDatabase and BadgerDB to use on the selected path.
func NewBlockDB(path string, params params.ChainParams, log *logger.Logger) (*BlockDB, error) {
	dbOptions := badger.DefaultOptions(path + "/db")
	dbOptions.Logger = nil
	badgerdb, err := badger.Open(dbOptions)
	if err != nil {
		return nil, err
	}
	rawbd, err := NewRawBlockDB(path + "/blocks")
	if err != nil {
		return nil, err
	}
	blockdb := &BlockDB{
		log:        log,
		badgerdb:   badgerdb,
		params:     params,
		rawBlockDb: rawbd,
	}
	return blockdb, nil
}
