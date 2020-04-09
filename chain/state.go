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

func (s *stateDerivedFromBlock) deriveState(slot uint64, view primitives.BlockView, p *params.ChainParams) (*primitives.State, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if slot == s.lastSlot {
		return s.lastSlotState, nil
	}

	if slot < s.lastSlot {
		derivedState := s.firstSlotState.Copy()

		err := derivedState.ProcessSlots(slot, view, p)
		if err != nil {
			return nil, err
		}
		view.SetTipSlot(slot)

		return &derivedState, nil
	}

	view.SetTipSlot(s.lastSlot)

	err := s.lastSlotState.ProcessSlots(slot, view, p)
	if err != nil {
		return nil, err
	}

	s.lastSlot = slot

	return s.lastSlotState, nil
}

// StateService keeps track of the blockchain and its state. This is where pruning should eventually be implemented to
// get rid of old states.
type StateService struct {
	log    *logger.Logger
	lock   sync.RWMutex
	params params.ChainParams

	blockIndex *index.BlockIndex
	blockChain *Chain
	stateMap   map[chainhash.Hash]*stateDerivedFromBlock
}

func (s *StateService) initChainState(db blockdb.DB, params params.ChainParams) error {
	// Get the state snap from db dbindex and deserialize
	s.log.Info("loading chain state...")

	genesisBlock := primitives.GetGenesisBlock(params)

	// load chain state
	loc, err := db.AddRawBlock(&genesisBlock)
	if err != nil {
		return err
	}

	blockIndex, err := index.InitBlocksIndex(genesisBlock.Header, *loc)
	if err != nil {
		return err
	}

	genesisHash := genesisBlock.Header.Hash()
	row, _ := blockIndex.Get(genesisHash)

	s.blockIndex = blockIndex
	s.blockChain = NewChain(row)

	// TODO: load block index
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
	if !found {
		return nil, fmt.Errorf("could not find state for block %s", hash)
	}

	return derivedState.deriveState(slot, view, p)
}

// Add adds a block to the blockchain.
func (s *StateService) Add(block *primitives.Block, newTip bool) (*primitives.State, error) {
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

	row, err := s.blockIndex.Add(block.Header)
	if err != nil {
		return nil, err
	}
	rowHash := row.Header.Hash()
	s.lock.Lock()
	s.stateMap[rowHash] = newStateDerivedFromBlock(&newState)
	s.lock.Unlock()

	if newTip {
		s.blockChain.SetTip(row)
	}
	return &newState, nil
}

// GetRowByHash gets a specific row by hash.
func (s *StateService) GetRowByHash(h chainhash.Hash) (*index.BlockRow, bool) {
	return s.blockIndex.Get(h)
}

func (s *StateService) Height() int32 {
	return s.blockChain.Height()
}

func (s *StateService) TipState() *primitives.State {
	return s.stateMap[s.blockChain.Tip().Hash].firstSlotState
}

// NewStateService constructs a new state service.
func NewStateService(log *logger.Logger, ip primitives.InitializationParameters, params params.ChainParams, db blockdb.DB) (*StateService, error) {
	genesisBlock := primitives.GetGenesisBlock(params)
	genesisHash := genesisBlock.Hash()

	genesisState := primitives.GetGenesisStateWithInitializationParameters(genesisHash, &ip, &params)

	ss := &StateService{
		params: params,
		log:    log,
		stateMap: map[chainhash.Hash]*stateDerivedFromBlock{
			genesisHash: newStateDerivedFromBlock(genesisState),
		},
	}
	err := ss.initChainState(db, params)
	if err != nil {
		return nil, err
	}
	return ss, nil
}
