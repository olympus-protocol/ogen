package actionmanager

import (
	"context"
	"github.com/VictoriaMetrics/fastcache"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/chainindex"
	"github.com/olympus-protocol/ogen/internal/host"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/params"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"math/rand"
	"time"
)

// MaxMessagePropagationTime is the maximum time we're expecting a message to
// take to propagate across the network. We wait double this before allowing a
// validator to start.
const MaxMessagePropagationTime = 60 * time.Second

// LastActionManager is an interface for lastActionManager
type LastActionManager interface {
	NewTip(row *chainindex.BlockRow, block *primitives.Block, state state.State, receipts []*primitives.EpochReceipt)
	ShouldRun(val [48]byte) (bool, error)
	GetNonce() uint64
}

var _ LastActionManager = &lastActionManager{}

// lastActionManager keeps track of the last action recorded by validators.
// This is a very basic protection against slashing. Validators, on startup,
// will broadcast a StartMessage
type lastActionManager struct {
	log logger.Logger

	host host.Host
	ch   chain.Blockchain
	ctx  context.Context

	nonce uint64

	// lastActions are the last recorded actions by a validator with a certain
	// salted private key hash.
	lastActions *fastcache.Cache

	netParams *params.ChainParams
}

func (l *lastActionManager) NewTip(_ *chainindex.BlockRow, block *primitives.Block, state state.State, _ []*primitives.EpochReceipt) {
	slotIndex := (block.Header.Slot + l.netParams.EpochLength - 1) % l.netParams.EpochLength

	proposerIndex := state.GetProposerQueue()[slotIndex]
	proposer := state.GetValidatorRegistry()[proposerIndex]

	err := l.registerAction(proposer.PubKey, time.Unix(int64(block.Header.Timestamp), 0), block.Header.Nonce)
	if err != nil {
		l.log.Error(err)
	}

	validators := state.GetValidatorRegistry()

	for _, v := range block.Votes {
		idx := v.ParticipationBitfield.BitIndices()
		for _, i := range idx {
			val := validators[i]
			err := l.registerAction(val.PubKey, time.Unix(int64(block.Header.Timestamp), 0), v.Data.Nonce)
			if err != nil {
				l.log.Error(err)
			}
		}
	}
}

func (l *lastActionManager) ProposerSlashingConditionViolated(*primitives.ProposerSlashing) {}

// NewLastActionManager creates a new last action manager.
func NewLastActionManager(h host.Host, ch chain.Blockchain) (LastActionManager, error) {
	ctx := config.GlobalParams.Context
	log := config.GlobalParams.Logger
	netParams := config.GlobalParams.NetParams

	l := &lastActionManager{
		host:        h,
		ch:          ch,
		ctx:         ctx,
		lastActions: fastcache.New(128 * 1024 * 1024),
		log:         log,
		nonce:       rand.Uint64(),
		netParams:   netParams,
	}

	ch.Notify(l)

	return l, nil
}

func (l *lastActionManager) ShouldRun(val [48]byte) (bool, error) {
	var lastActionBytes []byte
	lastActionBytes, ok := l.lastActions.HasGet(lastActionBytes, val[:])
	if !ok {
		return true, nil
	}

	var lastAction time.Time
	err := lastAction.UnmarshalBinary(lastActionBytes)
	if err != nil {
		return false, err
	}

	// last action was long enough ago we can start
	if time.Since(lastAction) > MaxMessagePropagationTime*2 {
		return true, nil
	}

	return false, nil
}

// RegisterActionAt registers an action by a validator at a certain time.
func (l *lastActionManager) registerAction(by [48]byte, at time.Time, nonce uint64) error {
	if nonce == l.nonce {
		return nil
	}
	timeBytes, err := at.MarshalBinary()
	if err != nil {
		return err
	}

	l.lastActions.Set(by[:], timeBytes)

	return nil
}

func (l *lastActionManager) GetNonce() uint64 {
	return l.nonce
}
