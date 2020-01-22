package blockdb

import (
	"github.com/grupokindynos/ogen/utils/chainhash"
)

func (bdb *BlockDB) GetGovIndex() ([]byte, error) {
	return bdb.govIndex.GetAll()
}

func (bdb *BlockDB) GetUtxoIndex() ([]byte, error) {
	return bdb.utxoIndex.GetAll()
}

func (bdb *BlockDB) GetUserIndex() ([]byte, error) {
	return bdb.usersIndex.GetAll()
}

func (bdb *BlockDB) GetVotesIndex() ([]byte, error) {
	return bdb.votesIndex.GetAll()
}

func (bdb *BlockDB) GetWorkersIndex() ([]byte, error) {
	return bdb.workersIndex.GetAll()
}

func (bdb *BlockDB) GetBlockIndex(blockHash chainhash.Hash) ([]byte, error) {
	return bdb.blockIndex.Get(blockHash)
}
