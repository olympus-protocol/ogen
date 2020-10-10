package blockdb

import (
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"time"
)

var (
	tipKey      = []byte("tip")
	finHeadKey  = []byte("finalized_head")
	jusHeadKey  = []byte("justified_head")
	finStateKey = []byte("finalized_state")
	jusStateKey = []byte("justified_state")
	genTimeKey  = []byte("genesis_key")

	blockRowPrefix = []byte("block-row-")
)

type Database interface {
	Close()
	GetBlock(hash chainhash.Hash) (*primitives.Block, error)
	GetRawBlock(hash chainhash.Hash) ([]byte, error)
	AddRawBlock(block *primitives.Block) error
	SetTip(c chainhash.Hash) error
	GetTip() (chainhash.Hash, error)
	SetFinalizedState(s state.State) error
	GetFinalizedState() (state.State, error)
	SetJustifiedState(s state.State) error
	GetJustifiedState() (state.State, error)
	SetBlockRow(disk *primitives.BlockNodeDisk) error
	GetBlockRow(c chainhash.Hash) (*primitives.BlockNodeDisk, error)
	SetJustifiedHead(c chainhash.Hash) error
	GetJustifiedHead() (chainhash.Hash, error)
	SetFinalizedHead(c chainhash.Hash) error
	GetFinalizedHead() (chainhash.Hash, error)
	SetGenesisTime(t time.Time) error
	GetGenesisTime() (time.Time, error)
}

var _ Database = &levelDB{}
