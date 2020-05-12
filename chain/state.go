package chain

import (
	"fmt"
	"sync"

	"github.com/olympus-protocol/ogen/chain/index"
	"github.com/olympus-protocol/ogen/db/blockdb"
	"github.com/olympus-protocol/ogen/logger"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

type stateDerivedFromBlock struct {
	firstSlot      uint64
	firstSlotState *primitives.State

	lastSlot      uint64
	lastSlotState *primitives.State

	lock *sync.Mutex
}

func newStateDerivedFromBlock(stateAfterProcessingBlock *primitives.State) *stateDerivedFromBlock {
	firstSlotState := stateAfterProcessingBlock.Copy()
	return &stateDerivedFromBlock{
		firstSlotState: &firstSlotState,
		firstSlot:      firstSlotState.Slot,
		lastSlotState:  stateAfterProcessingBlock,
		lastSlot:       stateAfterProcessingBlock.Slot,
		lock:           new(sync.Mutex),
	}
}

func (s *stateDerivedFromBlock) deriveState(slot uint64, view primitives.BlockView, p *params.ChainParams, log *logger.Logger) (*primitives.State, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if slot == s.lastSlot {
		return s.lastSlotState, nil
	}

	if slot < s.lastSlot {
		derivedState := s.firstSlotState.Copy()

		err := derivedState.ProcessSlots(slot, view, p, log)
		if err != nil {
			return nil, err
		}
		view.SetTipSlot(slot)

		return &derivedState, nil
	}

	view.SetTipSlot(s.lastSlot)

	err := s.lastSlotState.ProcessSlots(slot, view, p, log)
	if err != nil {
		return nil, err
	}

	s.lastSlot = slot

	return s.lastSlotState, nil
}

type blockNodeAndState struct {
	node  *index.BlockRow
	state primitives.State
}

// StateService keeps track of the blockchain and its state. This is where pruning should eventually be implemented to
// get rid of old states.
type StateService struct {
	log    *logger.Logger
	lock   sync.RWMutex
	params params.ChainParams
	db     blockdb.DB

	blockIndex *index.BlockIndex
	blockChain *Chain
	stateMap   map[chainhash.Hash]*stateDerivedFromBlock

	headLock      sync.Mutex
	finalizedHead blockNodeAndState
	justifiedHead blockNodeAndState
}

func (s *StateService) Chain() *Chain {
	return s.blockChain
}

func (s *StateService) Index() *index.BlockIndex {
	return s.blockIndex
}

func (s *StateService) setFinalizedHead(finalizedHash chainhash.Hash, finalizedState primitives.State) error {
	s.headLock.Lock()
	defer s.headLock.Unlock()

	finalizedNode, found := s.blockIndex.Get(finalizedHash)
	if !found {
		return fmt.Errorf("could not find block with hash %s", finalizedHash)
	}

	s.finalizedHead = blockNodeAndState{finalizedNode, finalizedState}
	return nil
}

// GetFinalizedHead gets the current finalized head.
func (s *StateService) GetFinalizedHead() (*index.BlockRow, primitives.State) {
	s.headLock.Lock()
	defer s.headLock.Unlock()

	return s.finalizedHead.node, s.finalizedHead.state
}

// GetJustifiedHead gets the current justified head.
func (s *StateService) GetJustifiedHead() (*index.BlockRow, primitives.State) {
	s.headLock.Lock()
	defer s.headLock.Unlock()

	return s.justifiedHead.node, s.justifiedHead.state
}

func (s *StateService) setJustifiedHead(justifiedHash chainhash.Hash, justifiedState primitives.State) error {
	s.headLock.Lock()
	defer s.headLock.Unlock()

	justifiedNode, found := s.blockIndex.Get(justifiedHash)
	if !found {
		return fmt.Errorf("could not find block with hash %s", justifiedHash)
	}

	s.justifiedHead = blockNodeAndState{justifiedNode, justifiedState}

	return nil
}

func (s *StateService) initChainState(db blockdb.DB, params params.ChainParams, genesisState primitives.State) error {
	// Get the state snap from db dbindex and deserialize
	s.log.Info("loading chain state...")

	genesisBlock := primitives.GetGenesisBlock(params)
	genesisHash := genesisBlock.Header.Hash()

	// load chain state
	err := s.db.Update(func(txn blockdb.DBUpdateTransaction) error {
		return txn.AddRawBlock(&genesisBlock)
	})
	if err != nil {
		return err
	}

	blockIndex, err := index.InitBlocksIndex(genesisBlock)
	if err != nil {
		return err
	}

	row, _ := blockIndex.Get(genesisHash)

	s.blockIndex = blockIndex
	s.blockChain = NewChain(row)

	return db.Update(func(txn blockdb.DBUpdateTransaction) error {
		if _, err := txn.GetBlockRow(genesisHash); err != nil {
			if err := s.initializeDatabase(txn, row, genesisState); err != nil {
				return err
			}
		} else {
			if err := s.loadBlockchainFromDisk(txn, genesisHash); err != nil {
				return err
			}
		}
		return nil
	})
	return nil
}

// GetStateForHash gets the state for a certain block hash.
func (s *StateService) GetStateForHash(hash chainhash.Hash) (*primitives.State, bool) {
	s.lock.RLock()
	derivedState, found := s.stateMap[hash]
	s.lock.RUnlock()
	if !found {
		return nil, false
	}
	derivedState.lock.Lock()
	defer derivedState.lock.Unlock()
	return derivedState.firstSlotState, true
}

// GetStateForHashAtSlot gets the state for a certain block hash at a certain slot.
func (s *StateService) GetStateForHashAtSlot(hash chainhash.Hash, slot uint64, view primitives.BlockView, p *params.ChainParams) (*primitives.State, error) {
	s.lock.RLock()
	derivedState, found := s.stateMap[hash]
	s.lock.RUnlock()
	if slot > derivedState.lastSlot+1000 {
		return nil, fmt.Errorf("tried to get block too far in future")
	}

	if !found {
		return nil, fmt.Errorf("could not find state for block %s", hash)
	}

	return derivedState.deriveState(slot, view, p, s.log)
}

// Add adds a block to the blockchain.
func (s *StateService) Add(block *primitives.Block) (*primitives.State, error) {
	lastBlockHash := block.Header.PrevBlockHash

	view, err := s.GetSubView(lastBlockHash)
	if err != nil {
		return nil, err
	}

	lastBlockState, err := s.GetStateForHashAtSlot(lastBlockHash, block.Header.Slot, &view, &s.params)
	if err != nil {
		return nil, err
	}

	newState := lastBlockState.Copy()

	err = newState.ProcessBlock(block, &s.params)
	if err != nil {
		return nil, err
	}

	s.setBlockState(block.Hash(), &newState)

	return &newState, nil
}

func (s *StateService) RemoveBeforeSlot(slot uint64) {
	s.lock.Lock()
	defer s.lock.Unlock()
	numRemoved := 0

	for i := range s.stateMap {
		row, found := s.blockIndex.Get(i)
		if !found {
			delete(s.stateMap, i)
			numRemoved++
			continue
		}

		if row.Slot < slot {
			delete(s.stateMap, i)
			numRemoved++
			continue
		}
	}
}

// GetRowByHash gets a specific row by hash.
func (s *StateService) GetRowByHash(h chainhash.Hash) (*index.BlockRow, bool) {
	return s.blockIndex.Get(h)
}

// Height gets the height of the blockchain.
func (s *StateService) Height() uint64 {
	return s.blockChain.Height()
}

// TipState gets the state of the tip of the blockchain.
func (s *StateService) TipState() *primitives.State {
	return s.stateMap[s.blockChain.Tip().Hash].firstSlotState
}

// NewStateService constructs a new state service.
func NewStateService(log *logger.Logger, ip primitives.InitializationParameters, params params.ChainParams, db blockdb.DB) (*StateService, error) {
	genesisBlock := primitives.GetGenesisBlock(params)
	genesisHash := genesisBlock.Hash()

	genesisState, err := primitives.GetGenesisStateWithInitializationParameters(genesisHash, &ip, &params)
	if err != nil {
		return nil, err
	}

	ss := &StateService{
		params: params,
		log:    log,
		stateMap: map[chainhash.Hash]*stateDerivedFromBlock{
			genesisHash: newStateDerivedFromBlock(genesisState),
		},
		db: db,
	}
	err = ss.initChainState(db, params, *genesisState)
	if err != nil {
		return nil, err
	}
	return ss, nil
}
