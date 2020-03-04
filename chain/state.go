package chain

import (
	"bytes"
	"fmt"
	"github.com/olympus-protocol/ogen/state"
	"io"
	"sync"

	"github.com/dgraph-io/badger"
	"github.com/olympus-protocol/ogen/chain/index"
	"github.com/olympus-protocol/ogen/db/blockdb"
	"github.com/olympus-protocol/ogen/logger"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
)

type StateSnap struct {
	Hash          chainhash.Hash // 32 bytes
	Height        int32          // 4 bytes
	Txs           int64          // 8 bytes
	Workers       int64          // 8 bytes
	Users         int64          // 8 bytes
	GovObjects    int64          // 8 bytes
	LastBlockTime int64          // 8 bytes
}

func (snap *StateSnap) Serialize(w io.Writer) error {
	err := serializer.WriteElements(w, snap.Hash, snap.Height, snap.Txs, snap.Workers, snap.Users, snap.GovObjects, snap.LastBlockTime)
	if err != nil {
		return err
	}
	return nil
}

func (snap *StateSnap) Deserialize(r io.Reader) error {
	err := serializer.ReadElements(r, &snap.Hash, &snap.Height, &snap.Txs, &snap.Workers, &snap.Users, &snap.GovObjects, &snap.LastBlockTime)
	if err != nil {
		return err
	}
	return nil
}

func (snap *StateSnap) String() string {
	return fmt.Sprintf("StateService Snapshot: Hash: %s, Height: %v, Txs: %v, Workers: %v, Users: %v, GovObjects: %v, LastBlockTime: %v",
		snap.Hash.String(), snap.Height, snap.Txs, snap.Workers, snap.Users, snap.GovObjects, snap.LastBlockTime)
}

// StateService keeps track of the blockchain and its state. This is where pruning should eventually be implemented to
// get rid of old states.
type StateService struct {
	log      *logger.Logger
	snapshot StateSnap
	lock     sync.RWMutex
	params   params.ChainParams

	View     *ChainView
	stateMap map[chainhash.Hash]state.State

	sync bool
}

func (s *StateService) IsSync() bool {
	return s.sync
}

func (s *StateService) SetSyncStatus(sync bool) {
	s.sync = sync
	return
}

func (s *StateService) Snapshot() *StateSnap {
	s.lock.Lock()
	snap := s.snapshot
	s.lock.Unlock()
	return &snap
}

func (s *StateService) updateStateSnap(block *primitives.Block, workers int64, users int64, govObjects int64) error {
	s.lock.Lock()
	s.snapshot = StateSnap{
		Hash:          block.Hash,
		Height:        int32(block.Height),
		Txs:           s.snapshot.Txs + int64(len(block.Txs)),
		Workers:       workers,
		Users:         users,
		GovObjects:    govObjects,
		LastBlockTime: block.GetTime().Unix(),
	}
	s.lock.Unlock()
	return nil
}

func (s *StateService) initChainState(db *blockdb.BlockDB, params params.ChainParams) error {
start:
	// Get the state snap from db dbindex and deserialize
	s.log.Info("loading chain state...")
	rawState, err := db.GetStateSnap()
	if err != nil {
		if err == badger.ErrKeyNotFound {
			newStateSnap := StateSnap{
				Hash:          params.GenesisHash,
				Height:        0,
				Txs:           0,
				Workers:       0,
				Users:         0,
				GovObjects:    0,
				LastBlockTime: params.GenesisBlock.Header.Timestamp.Unix(),
			}
			buf := bytes.NewBuffer([]byte{})
			err := newStateSnap.Serialize(buf)
			if err != nil {
				return err
			}
			err = db.SetStateSnap(buf.Bytes())
			if err != nil {
				return err
			}
			genBlock, err := primitives.NewBlockFromMsg(&params.GenesisBlock, 0)
			if err != nil {
				return err
			}
			loc, err := db.AddRawBlock(genBlock)
			view, err := NewChainView(params.GenesisBlock.Header, loc)
			if err != nil {
				return err
			}
			s.View = view
			goto start
		}
		return err
	}
	bufState := bytes.NewBuffer(rawState)
	err = s.snapshot.Deserialize(bufState)
	if err != nil {
		return err
	}
	// Get block dbindex raw data and deserialize
	s.log.Info("loading block index...")
	searchHash := s.params.GenesisHash
	lastBlockHeight := 0
	for {
		blockRow := index.BlockRow{
			Height: int32(lastBlockHeight),
		}
		rawBlockRow, err := db.GetBlockIndex(searchHash)
		if err != nil {
			if err == badger.ErrKeyNotFound {
				break
			}
			return err
		}
		buf := bytes.NewBuffer(rawBlockRow)
		err = blockRow.Deserialize(buf)
		if err != nil {
			return err
		}
		_, err = s.View.Add(blockRow.Header, blockRow.Locator)
		if err != nil {
			return err
		}
		searchHash = blockRow.Header.Hash()
		lastBlockHeight = lastBlockHeight + 1
	}
	return nil
}

func (s *StateService) TipState() state.State {
	tip := s.View.Tip()
	return s.stateMap[tip.Hash]
}

func NewStateService(log *logger.Logger, params params.ChainParams, db *blockdb.BlockDB) (*StateService, error) {
	ss := &StateService{
		params:   params,
		log:      log,
		snapshot: StateSnap{},
		sync:     false,
	}
	err := ss.initChainState(db, params)
	if err != nil {
		return nil, err
	}
	return ss, nil
}
