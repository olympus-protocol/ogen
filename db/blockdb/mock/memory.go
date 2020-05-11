package mock

import (
	"fmt"
	"github.com/olympus-protocol/ogen/db/blockdb"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"sync"
	"time"
)

type MemoryDB struct {
	lock sync.RWMutex

	rawBlocks      map[chainhash.Hash]*primitives.Block
	latestVotes    map[uint32]*primitives.MultiValidatorVote
	tipHash        chainhash.Hash
	finalizedTip   chainhash.Hash
	finalizedState primitives.State
	justifiedTip   chainhash.Hash
	justifiedState primitives.State

	blockRows map[chainhash.Hash]*blockdb.BlockNodeDisk
}

type MemoryDBTransaction struct {
	db *MemoryDB
}

func (m *MemoryDBTransaction) AddRawBlock(block *primitives.Block) error {
	m.db.rawBlocks[block.Hash()] = block
	return nil
}

func (m *MemoryDBTransaction) SetLatestVoteIfNeeded(validators []uint32, vote *primitives.MultiValidatorVote) error {
	for _, v := range validators {
		if oldVote, found := m.db.latestVotes[v]; found {
			if oldVote.Data.Slot >= vote.Data.Slot {
				m.db.latestVotes[v] = vote
			}
		} else {
			m.db.latestVotes[v] = vote
		}
	}

	return nil
}

func (m *MemoryDBTransaction) SetTip(hash chainhash.Hash) error {
	m.db.tipHash = hash
	return nil
}

func (m *MemoryDBTransaction) SetFinalizedState(state *primitives.State) error {
	m.db.finalizedState = *state
	return nil
}

func (m *MemoryDBTransaction) SetJustifiedState(state *primitives.State) error {
	m.db.justifiedState = *state
	return nil
}

func (m *MemoryDBTransaction) SetBlockRow(disk *blockdb.BlockNodeDisk) error {
	m.db.blockRows[disk.Hash] = disk
	return nil
}

func (m *MemoryDBTransaction) SetJustifiedHead(hash chainhash.Hash) error {
	m.db.justifiedTip = hash
	return nil
}

func (m *MemoryDBTransaction) SetFinalizedHead(hash chainhash.Hash) error {
	m.db.finalizedTip = hash
	return nil
}

func (m *MemoryDBTransaction) SetGenesisTime(_ time.Time) error {
	return nil
}

func (m *MemoryDBTransaction) GetRawBlock(hash chainhash.Hash) (*primitives.Block, error) {
	block, ok := m.db.rawBlocks[hash]
	if !ok {
		return nil, fmt.Errorf("could not find block with hash: %s", hash)
	}

	return block, nil
}

func (m *MemoryDBTransaction) GetTip() (chainhash.Hash, error) {
	return m.db.tipHash, nil
}

func (m *MemoryDBTransaction) GetBlockRow(hash chainhash.Hash) (*blockdb.BlockNodeDisk, error) {
	row, ok := m.db.blockRows[hash]
	if !ok {
		return nil, fmt.Errorf("could not find block row with hash %s", hash)
	}

	return row, nil
}

func (m *MemoryDBTransaction) GetJustifiedHead() (chainhash.Hash, error) {
	return m.db.justifiedTip, nil
}

func (m *MemoryDBTransaction) GetFinalizedHead() (chainhash.Hash, error) {
	return m.db.finalizedTip, nil
}

func (m *MemoryDBTransaction) GetLatestVote(validator uint32) (*primitives.MultiValidatorVote, error) {
	vote, ok := m.db.latestVotes[validator]
	if !ok {
		return nil, fmt.Errorf("do not have votes for validator %d", validator)
	}

	return vote, nil
}

func (m *MemoryDBTransaction) GetFinalizedState() (*primitives.State, error) {
	s := m.db.finalizedState.Copy()
	return &s, nil
}

func (m *MemoryDBTransaction) GetJustifiedState() (*primitives.State, error) {
	s := m.db.justifiedState.Copy()
	return &s, nil
}

func (m *MemoryDBTransaction) GetGenesisTime() (time.Time, error) {
	return time.Time{}, fmt.Errorf("genesis time not implemented for memory DB")
}

func (m *MemoryDB) Close() {
	return
}

func (m *MemoryDB) Update(f func(blockdb.DBUpdateTransaction) error) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	return f(&MemoryDBTransaction{db: m})
}

func (m *MemoryDB) View(f func(blockdb.DBViewTransaction) error) error {
	return f(&MemoryDBTransaction{
		db: m,
	})
}

var _ blockdb.DBUpdateTransaction = &MemoryDBTransaction{}
var _ blockdb.DBViewTransaction = &MemoryDBTransaction{}
var _ blockdb.DB = &MemoryDB{}

func NewMemoryDB() *MemoryDB {
	return &MemoryDB{
		rawBlocks:   make(map[chainhash.Hash]*primitives.Block),
		latestVotes: map[uint32]*primitives.MultiValidatorVote{},
		blockRows:   map[chainhash.Hash]*blockdb.BlockNodeDisk{},
	}
}
