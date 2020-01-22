package chain

import (
	"bytes"
	"fmt"
	"github.com/dgraph-io/badger"
	"github.com/olympus-protocol/ogen/chain/index"
	"github.com/olympus-protocol/ogen/db/blockdb"
	"github.com/olympus-protocol/ogen/logger"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
	"io"
	"sync"
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
	return fmt.Sprintf("State Snapshot: Hash: %s, Height: %v, Txs: %v, Workers: %v, Users: %v, GovObjects: %v, LastBlockTime: %v",
		snap.Hash.String(), snap.Height, snap.Txs, snap.Workers, snap.Users, snap.GovObjects, snap.LastBlockTime)
}

type State struct {
	log        *logger.Logger
	snapshot   StateSnap
	lock       sync.RWMutex
	params     params.ChainParams
	blockIndex *index.BlockIndex

	utxosIndex  *index.UtxosIndex
	govIndex    *index.GovIndex
	usersIndex  *index.UserIndex
	workerIndex *index.WorkerIndex

	sync bool
}

func (s *State) IsSync() bool {
	return s.sync
}

func (s *State) SetSyncStatus(sync bool) {
	s.sync = sync
	return
}

func (s *State) Snapshot() *StateSnap {
	s.lock.Lock()
	snap := s.snapshot
	s.lock.Unlock()
	return &snap
}

func (s *State) updateStateSnap(block *primitives.Block, workers int64, users int64, govObjects int64) error {
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

func (s *State) InitChainState(db *blockdb.BlockDB, params params.ChainParams) error {
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
			newRow := index.NewBlockRow(loc, params.GenesisBlock.Header)
			err = s.blockIndex.Add(newRow)
			if err != nil {
				return err
			}
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
		err = s.blockIndex.Add(&blockRow)
		if err != nil {
			return err
		}
		searchHash, err = blockRow.Header.Hash()
		if err != nil {
			return err
		}
		lastBlockHeight = lastBlockHeight + 1
	}
	// Get utxo dbindex raw data and deserialize
	s.log.Info("loading utxo index...")
	rawUtxoIndex, err := db.GetUtxoIndex()
	if err != nil {
		return err
	}
	bufUtxos := bytes.NewBuffer(rawUtxoIndex)
	err = s.utxosIndex.Deserialize(bufUtxos)
	if err != nil {
		return err
	}
	// Get gov dbindex raw data and deserialize
	s.log.Info("loading gov index...")
	rawGovIndex, err := db.GetGovIndex()
	if err != nil {
		return err
	}
	bufGov := bytes.NewBuffer(rawGovIndex)
	err = s.govIndex.Deserialize(bufGov)
	if err != nil {
		return err
	}
	// Get users dbindex raw data and deserialize
	s.log.Info("loading users index...")
	rawUserIndex, err := db.GetUserIndex()
	bufUsers := bytes.NewBuffer(rawUserIndex)
	if err != nil {
		return err
	}
	err = s.usersIndex.Deserialize(bufUsers)
	if err != nil {
		return err
	}
	// Get workers dbindex raw data and deserialize
	s.log.Info("loading workers index...")
	rawWorkersIndex, err := db.GetWorkersIndex()
	if err != nil {
		return err
	}
	bufWorkers := bytes.NewBuffer(rawWorkersIndex)
	err = s.workerIndex.Deserialize(bufWorkers)
	if err != nil {
		return err
	}
	return nil
}

func NewChainState(indexers *index.Indexers, log *logger.Logger, params params.ChainParams) *State {
	state := &State{
		params:      params,
		log:         log,
		snapshot:    StateSnap{},
		blockIndex:  indexers.BlockIndex,
		utxosIndex:  indexers.UtxoIndex,
		govIndex:    indexers.GovIndex,
		usersIndex:  indexers.UserIndex,
		workerIndex: indexers.WorkerIndex,
		sync:        false,
	}
	return state
}
