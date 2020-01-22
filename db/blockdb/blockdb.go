package blockdb

import (
	"bytes"
	"github.com/dgraph-io/badger"
	"github.com/olympus-protocol/ogen/db/blockdb/dbindex"
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

	blockIndex   *dbindex.Blocks
	govIndex     *dbindex.Gov
	usersIndex   *dbindex.Users
	utxoIndex    *dbindex.Utxos
	votesIndex   *dbindex.Votes
	workersIndex *dbindex.Workers
}

func (bdb *BlockDB) Close() {
	_ = bdb.badgerdb.Close()
}

// RawDataSize returns the size on bytes for the raw block information.
func (bdb *BlockDB) RawDataSize() (int64, error) {
	var size int64
	blockIndexSize, err := bdb.blockIndex.Size()
	if err != nil {
		return 0, err
	}
	size += blockIndexSize
	govIndexSize, err := bdb.govIndex.Size()
	if err != nil {
		return 0, err
	}
	size += govIndexSize
	usersIndexSize, err := bdb.usersIndex.Size()
	if err != nil {
		return 0, err
	}
	size += usersIndexSize
	utxosIndexSize, err := bdb.utxoIndex.Size()
	if err != nil {
		return 0, err
	}
	size += utxosIndexSize
	votesIndexSize, err := bdb.votesIndex.Size()
	if err != nil {
		return 0, err
	}
	size += votesIndexSize
	workersIndexSize, err := bdb.workersIndex.Size()
	if err != nil {
		return 0, err
	}
	size += workersIndexSize
	size += bdb.rawBlockDb.FullDataSize()
	return size, nil
}

func (bdb *BlockDB) GetRawBlock(locator BlockLocation, hash chainhash.Hash) ([]byte, error) {
	return bdb.rawBlockDb.read(hash, locator)
}

func (bdb *BlockDB) AddRawBlock(block *primitives.Block) (BlockLocation, error) {
	locator, err := bdb.rawBlockDb.ConnectBlock(block)
	if err != nil {
		return BlockLocation{}, err
	}
	buf := bytes.NewBuffer([]byte{})
	err = locator.Serialize(buf)
	if err != nil {
		return BlockLocation{}, err
	}
	err = bdb.blockIndex.Add(buf.Bytes(), block.Header())
	if err != nil {
		return BlockLocation{}, err
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

		blockIndex:   &dbindex.Blocks{DB: badgerdb},
		govIndex:     &dbindex.Gov{DB: badgerdb},
		usersIndex:   &dbindex.Users{DB: badgerdb},
		utxoIndex:    &dbindex.Utxos{DB: badgerdb},
		votesIndex:   &dbindex.Votes{DB: badgerdb},
		workersIndex: &dbindex.Workers{DB: badgerdb},
	}
	return blockdb, nil
}
