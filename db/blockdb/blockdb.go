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
	log      *logger.Logger
	badgerdb *badger.DB
	params   params.ChainParams
}

func (bdb *BlockDB) Close() {
	_ = bdb.badgerdb.Close()
}

// GetRawBlock gets a block from the database.
func (bdb *BlockDB) GetRawBlock(hash chainhash.Hash) (*primitives.Block, error) {
	blockBytes, err := getKey(bdb.badgerdb, hash[:])
	if err != nil {
		return nil, err
	}

	block := new(primitives.Block)
	err = block.Decode(bytes.NewBuffer(blockBytes))
	return block, err
}

// AddRawBlock adds a raw block to the database.
func (bdb *BlockDB) AddRawBlock(block *primitives.Block) error {
	blockHash := block.Hash()
	blockBytes := bytes.NewBuffer([]byte{})
	err := block.Encode(blockBytes)
	if err != nil {
		return err
	}
	return setKey(bdb.badgerdb, blockHash[:], blockBytes.Bytes())
}

// NewBlockDB returns a database instance with a rawBlockDatabase and BadgerDB to use on the selected path.
func NewBlockDB(path string, params params.ChainParams, log *logger.Logger) (*BlockDB, error) {
	dbOptions := badger.DefaultOptions(path + "/db")
	dbOptions.Logger = nil
	badgerdb, err := badger.Open(dbOptions)
	if err != nil {
		return nil, err
	}
	blockdb := &BlockDB{
		log:      log,
		badgerdb: badgerdb,
		params:   params,
	}
	return blockdb, nil
}
