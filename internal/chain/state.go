package chain

import (
	"errors"
	"fmt"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/internal/state"
	"sync"

	"github.com/olympus-protocol/ogen/internal/blockdb"
	"github.com/olympus-protocol/ogen/internal/chainindex"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/params"
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

type stateDerivedFromBlock struct {
	firstSlot      uint64
	firstSlotState state.State

	lastSlot      uint64
	lastSlotState state.State

	totalReceipts []*primitives.EpochReceipt

	lock *sync.Mutex
}

func newStateDerivedFromBlock(stateAfterProcessingBlock state.State) *stateDerivedFromBlock {
	firstSlotState := stateAfterProcessingBlock.Copy()
	return &stateDerivedFromBlock{
		firstSlotState: firstSlotState,
		firstSlot:      firstSlotState.GetSlot(),
		lastSlotState:  stateAfterProcessingBlock,
		lastSlot:       stateAfterProcessingBlock.GetSlot(),
		lock:           new(sync.Mutex),
	}
}

func (s *stateDerivedFromBlock) deriveState(slot uint64, view state.BlockView) (state.State, []*primitives.EpochReceipt, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if slot == s.lastSlot {
		return s.lastSlotState, s.totalReceipts, nil
	}

	if slot < s.lastSlot {
		derivedState := s.firstSlotState.Copy()

		receipts, err := derivedState.ProcessSlots(slot, view)
		if err != nil {
			return nil, nil, err
		}

		view.SetTipSlot(slot)

		return derivedState, receipts, nil
	}

	view.SetTipSlot(s.lastSlot)

	receipts, err := s.lastSlotState.ProcessSlots(slot, view)
	if err != nil {
		return nil, nil, err
	}

	s.totalReceipts = append(s.totalReceipts, receipts...)
	s.lastSlot = slot

	return s.lastSlotState, s.totalReceipts, nil
}

type blockNodeAndState struct {
	node  *chainindex.BlockRow
	state state.State
}

type StateService interface {
	GetLatestVote(val uint64) (*primitives.MultiValidatorVote, bool)
	SetLatestVotesIfNeeded(vals []uint64, vote *primitives.MultiValidatorVote)
	Chain() *Chain
	Index() *chainindex.BlockIndex
	SetFinalizedHead(finalizedHash chainhash.Hash, finalizedState state.State) error
	GetFinalizedHead() (*chainindex.BlockRow, state.State)
	SetJustifiedHead(justifiedHash chainhash.Hash, justifiedState state.State) error
	GetJustifiedHead() (*chainindex.BlockRow, state.State)
	GetStateForHash(hash chainhash.Hash) (state.State, bool)
	GetStateForHashAtSlot(hash chainhash.Hash, slot uint64, view state.BlockView) (state.State, []*primitives.EpochReceipt, error)
	Add(block *primitives.Block) (state.State, []*primitives.EpochReceipt, error)
	RemoveBeforeSlot(slot uint64)
	GetRowByHash(h chainhash.Hash) (*chainindex.BlockRow, bool)
	Height() uint64
	TipState() state.State
	TipStateAtSlot(slot uint64) (state.State, error)
	GetSubView(tip chainhash.Hash) (View, error)
	Tip() *chainindex.BlockRow
}

// stateService keeps track of the blockchain and its state. This is where pruning should eventually be implemented to
// get rid of old states.
type stateService struct {
	log       logger.Logger
	netParams *params.ChainParams
	db        blockdb.Database

	index *chainindex.BlockIndex
	chain *Chain

	stateMap     map[chainhash.Hash]*stateDerivedFromBlock
	stateMapLock sync.Mutex

	headLock      sync.Mutex
	finalizedHead blockNodeAndState
	justifiedHead blockNodeAndState

	latestVotes     map[uint64]*primitives.MultiValidatorVote
	latestVotesLock sync.Mutex
}

var _ StateService = &stateService{}

// GetLatestVote gets the latest vote for this validator.
func (s *stateService) GetLatestVote(val uint64) (*primitives.MultiValidatorVote, bool) {
	s.latestVotesLock.Lock()
	defer s.latestVotesLock.Unlock()
	v, ok := s.latestVotes[val]
	return v, ok
}

// SetLatestVotesIfNeeded sets the latest vote for this validator.
func (s *stateService) SetLatestVotesIfNeeded(vals []uint64, vote *primitives.MultiValidatorVote) {
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
func (s *stateService) Chain() *Chain {
	return s.chain
}

// Index gets the block chainindex.
func (s *stateService) Index() *chainindex.BlockIndex {
	return s.index
}

func (s *stateService) SetFinalizedHead(finalizedHash chainhash.Hash, finalizedState state.State) error {
	s.headLock.Lock()
	defer s.headLock.Unlock()

	finalizedNode, found := s.index.Get(finalizedHash)
	if !found {
		return fmt.Errorf("could not find block with hash %s", finalizedHash)
	}

	s.finalizedHead = blockNodeAndState{finalizedNode, finalizedState}
	return nil
}

// GetFinalizedHead gets the current finalized head.
func (s *stateService) GetFinalizedHead() (*chainindex.BlockRow, state.State) {
	s.headLock.Lock()
	defer s.headLock.Unlock()

	return s.finalizedHead.node, s.finalizedHead.state
}

// GetJustifiedHead gets the current justified head.
func (s *stateService) GetJustifiedHead() (*chainindex.BlockRow, state.State) {
	s.headLock.Lock()
	defer s.headLock.Unlock()

	return s.justifiedHead.node, s.justifiedHead.state
}

func (s *stateService) SetJustifiedHead(justifiedHash chainhash.Hash, justifiedState state.State) error {
	s.headLock.Lock()
	defer s.headLock.Unlock()

	justifiedNode, found := s.index.Get(justifiedHash)
	if !found {
		return fmt.Errorf("could not find block with hash %s", justifiedHash)
	}

	s.justifiedHead = blockNodeAndState{justifiedNode, justifiedState}

	return nil
}

func (s *stateService) initChainState(db blockdb.Database, genesisState state.State) error {
	// Get the state snap from db dbindex and deserialize
	s.log.Info("Loading chain state...")

	genesisBlock := primitives.GetGenesisBlock()
	genesisHash := genesisBlock.Header.Hash()

	// load chain state
	err := db.AddRawBlock(&genesisBlock)
	if err != nil {
		return err
	}

	blockIndex, err := chainindex.InitBlocksIndex(genesisBlock)
	if err != nil {
		return err
	}

	row, _ := blockIndex.Get(genesisHash)

	s.index = blockIndex
	s.chain = NewChain(row)

	if _, err := db.GetBlockRow(genesisHash); err != nil {
		if err := s.initializeDatabase(db, row, genesisState); err != nil {
			return err
		}
	} else {
		if err := s.loadBlockchainFromDisk(db, genesisHash); err != nil {
			return err
		}
	}
	return nil
}

// GetStateForHash gets the state for a certain block hash.
func (s *stateService) GetStateForHash(hash chainhash.Hash) (state.State, bool) {
	s.stateMapLock.Lock()
	derivedState, found := s.stateMap[hash]
	s.stateMapLock.Unlock()
	if !found {
		return nil, false
	}
	derivedState.lock.Lock()
	defer derivedState.lock.Unlock()
	return derivedState.firstSlotState, true
}

var ErrTooFarInFuture = fmt.Errorf("tried to get block too far in future")

// GetStateForHashAtSlot gets the state for a certain block hash at a certain slot.
func (s *stateService) GetStateForHashAtSlot(hash chainhash.Hash, slot uint64, view state.BlockView) (state.State, []*primitives.EpochReceipt, error) {
	s.stateMapLock.Lock()
	derivedState, found := s.stateMap[hash]
	s.stateMapLock.Unlock()
	if !found {
		return nil, nil, fmt.Errorf("could not find state for block %s", hash)
	}

	if slot > derivedState.lastSlot+1000 {
		return nil, nil, ErrTooFarInFuture
	}

	return derivedState.deriveState(slot, view)
}

// Add adds a block to the blockchain.
func (s *stateService) Add(block *primitives.Block) (state.State, []*primitives.EpochReceipt, error) {
	lastBlockHash := block.Header.PrevBlockHash

	view, err := s.GetSubView(lastBlockHash)
	if err != nil {
		return nil, nil, err
	}

	lastBlockState, receipts, err := s.GetStateForHashAtSlot(lastBlockHash, block.Header.Slot, &view)
	if err != nil {
		return nil, nil, err
	}

	newState := lastBlockState.Copy()

	err = newState.ProcessBlock(block)
	if err != nil {
		return nil, nil, err
	}

	s.setBlockState(block.Hash(), newState)

	return newState, receipts, nil
}

// RemoveBeforeSlot removes state before a certain slot.
func (s *stateService) RemoveBeforeSlot(slot uint64) {
	s.stateMapLock.Lock()
	defer s.stateMapLock.Unlock()
	if slot < 20 {
		return
	}
	slot -= 20
	statesCleaned := 0
	for i := range s.stateMap {
		row, found := s.index.Get(i)
		if !found {
			s.log.Debugf("deleting block state for %s", i)
			//delete(s.stateMap, i)
			statesCleaned++
			continue
		}

		if row.Slot < slot {
			s.log.Debugf("deleting block state for %s", i)
			delete(s.stateMap, i)
			statesCleaned++
			continue
		}
	}
	s.log.Debugf("removed %d states on map observing %d states", statesCleaned, len(s.stateMap))
}

// GetRowByHash gets a specific row by hash.
func (s *stateService) GetRowByHash(h chainhash.Hash) (*chainindex.BlockRow, bool) {
	return s.index.Get(h)
}

// Height gets the height of the blockchain.
func (s *stateService) Height() uint64 {
	return s.chain.Height()
}

// TipState gets the state of the tip of the blockchain.
func (s *stateService) TipState() state.State {
	return s.stateMap[s.chain.Tip().Hash].firstSlotState
}

// TipStateAtSlot gets the tip state updated to a certain slot.
func (s *stateService) TipStateAtSlot(slot uint64) (state.State, error) {
	tipHash := s.Tip().Hash
	view, err := s.GetSubView(tipHash)
	if err != nil {
		return nil, err
	}
	st, _, err := s.GetStateForHashAtSlot(tipHash, slot, &view)
	if err != nil {
		return nil, err
	}

	return st, nil
}

// NewStateService constructs a new state service.
func NewStateService(db blockdb.Database) (StateService, error) {
	ip := config.GlobalParams.InitParams
	netParams := config.GlobalParams.NetParams
	log := config.GlobalParams.Logger

	genesisBlock := primitives.GetGenesisBlock()
	genesisHash := genesisBlock.Hash()

	genesisState, err := state.GetGenesisStateWithInitializationParameters(genesisHash, ip, netParams)
	if err != nil {
		return nil, err
	}

	ss := &stateService{
		netParams: netParams,
		log:       log,
		stateMap: map[chainhash.Hash]*stateDerivedFromBlock{
			genesisHash: newStateDerivedFromBlock(genesisState),
		},
		latestVotes: make(map[uint64]*primitives.MultiValidatorVote),
		db:          db,
	}

	err = ss.initChainState(db, genesisState)
	if err != nil {
		return nil, err
	}
	return ss, nil
}

// GetSubView gets a view of the blockchain at a certain tip.
func (s *stateService) GetSubView(tip chainhash.Hash) (View, error) {
	tipNode, found := s.index.Get(tip)
	if !found {
		return View{}, errors.New("could not find tip node")
	}
	return NewChainView(tipNode), nil
}

// Tip gets the tip of the blockchain.
func (s *stateService) Tip() *chainindex.BlockRow {
	return s.chain.Tip()
}
