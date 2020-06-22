package chain

import (
	"fmt"
	"sync"

	"github.com/olympus-protocol/ogen/bdb"
	"github.com/olympus-protocol/ogen/chain/index"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/logger"
)

type stateDerivedFromBlock struct {
	firstSlot      uint64
	firstSlotState *primitives.State

	lastSlot      uint64
	lastSlotState *primitives.State

	totalReceipts []*primitives.EpochReceipt

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

func (s *stateDerivedFromBlock) deriveState(slot uint64, view primitives.BlockView, p *params.ChainParams, log *logger.Logger) (*primitives.State, []*primitives.EpochReceipt, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if slot == s.lastSlot {
		return s.lastSlotState, s.totalReceipts, nil
	}

	if slot < s.lastSlot {
		derivedState := s.firstSlotState.Copy()

		receipts, err := derivedState.ProcessSlots(slot, view, p, log)
		if err != nil {
			return nil, nil, err
		}

		view.SetTipSlot(slot)

		return &derivedState, receipts, nil
	}

	view.SetTipSlot(s.lastSlot)

	receipts, err := s.lastSlotState.ProcessSlots(slot, view, p, log)
	if err != nil {
		return nil, nil, err
	}

	s.totalReceipts = append(s.totalReceipts, receipts...)
	s.lastSlot = slot

	return s.lastSlotState, s.totalReceipts, nil
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
	db     bdb.DB

	blockIndex *index.BlockIndex
	blockChain *Chain
	stateMap   map[chainhash.Hash]*stateDerivedFromBlock

	headLock      sync.Mutex
	finalizedHead blockNodeAndState
	justifiedHead blockNodeAndState

	latestVotes     map[uint32]*primitives.MultiValidatorVote
	latestVotesLock sync.RWMutex
}

// GetLatestVote gets the latest vote for this validator.
func (s *StateService) GetLatestVote(val uint32) (*primitives.MultiValidatorVote, bool) {
	s.latestVotesLock.RLock()
	s.latestVotesLock.RUnlock()

	v, ok := s.latestVotes[val]

	return v, ok
}

// SetLatestVotesIfNeeded sets the latest vote for this validator.
func (s *StateService) SetLatestVotesIfNeeded(vals []uint32, vote *primitives.MultiValidatorVote) {
	s.latestVotesLock.Lock()
	defer s.latestVotesLock.Unlock()
	for _, v := range vals {
		oldVote, ok := s.latestVotes[v]
		if ok && oldVote.Data.Slot >= vote.Data.Slot {
			continue
		}
		s.latestVotes[v] = vote
	}
}

// Chain gets the blockchain.
func (s *StateService) Chain() *Chain {
	return s.blockChain
}

// Index gets the block index.
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

func (s *StateService) initChainState(db bdb.DB, params params.ChainParams, genesisState primitives.State) error {
	// Get the state snap from db dbindex and deserialize
	s.log.Info("loading chain state...")

	genesisBlock := primitives.GetGenesisBlock(params)
	genesisHash, err := genesisBlock.Header.Hash()
	if err != nil {
		return err
	}

	// load chain state
	err = s.db.Update(func(txn bdb.DBUpdateTransaction) error {
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

	return db.Update(func(txn bdb.DBUpdateTransaction) error {
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

var ErrTooFarInFuture = fmt.Errorf("tried to get block too far in future")

// GetStateForHashAtSlot gets the state for a certain block hash at a certain slot.
func (s *StateService) GetStateForHashAtSlot(hash chainhash.Hash, slot uint64, view primitives.BlockView, p *params.ChainParams) (*primitives.State, []*primitives.EpochReceipt, error) {
	s.lock.RLock()
	derivedState, found := s.stateMap[hash]
	s.lock.RUnlock()
	if !found {
		return nil, nil, fmt.Errorf("could not find state for block %s", hash)
	}

	if slot > derivedState.lastSlot+1000 {
		return nil, nil, ErrTooFarInFuture
	}

	return derivedState.deriveState(slot, view, p, s.log)
}

// Add adds a block to the blockchain.
func (s *StateService) Add(block *primitives.Block) (*primitives.State, []*primitives.EpochReceipt, error) {
	lastBlockHash := block.Header.PrevBlockHash

	view, err := s.GetSubView(lastBlockHash)
	if err != nil {
		return nil, nil, err
	}

	lastBlockState, receipts, err := s.GetStateForHashAtSlot(lastBlockHash, block.Header.Slot, &view, &s.params)
	if err != nil {
		return nil, nil, err
	}

	newState := lastBlockState.Copy()

	err = newState.ProcessBlock(block, &s.params)
	if err != nil {
		return nil, nil, err
	}

	hash, err := block.Hash()
	if err != nil {
		return nil, nil, err
	}
	s.setBlockState(hash, &newState)

	return &newState, receipts, nil
}

// RemoveBeforeSlot removes state before a certain slot.
func (s *StateService) RemoveBeforeSlot(slot uint64) {
	s.lock.Lock()
	defer s.lock.Unlock()
	numRemoved := 0

	for i := range s.stateMap {
		row, found := s.blockIndex.Get(i)
		if !found {
			s.log.Debugf("deleting block state for %s", i)
			delete(s.stateMap, i)
			numRemoved++
			continue
		}

		if row.Slot < slot {
			s.log.Debugf("deleting block state for %s", i)
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

// TipStateAtSlot gets the tip state updated to a certain slot.
func (s *StateService) TipStateAtSlot(slot uint64) (*primitives.State, error) {
	tipHash := s.Tip().Hash
	view, err := s.GetSubView(tipHash)
	if err != nil {
		return nil, err
	}
	state, _, err := s.GetStateForHashAtSlot(tipHash, slot, &view, &s.params)
	if err != nil {
		return nil, err
	}

	return state, nil
}

// NewStateService constructs a new state service.
func NewStateService(log *logger.Logger, ip primitives.InitializationParameters, params params.ChainParams, db bdb.DB) (*StateService, error) {
	genesisBlock := primitives.GetGenesisBlock(params)
	genesisHash, err := genesisBlock.Hash()
	if err != nil {
		return nil, err
	}
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
		latestVotes: make(map[uint32]*primitives.MultiValidatorVote),
		db:          db,
	}
	err = ss.initChainState(db, params, *genesisState)
	if err != nil {
		return nil, err
	}
	return ss, nil
}
