package actionmanager

import (
	"context"
	"encoding/binary"
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

type timeWithNonce struct {
	Time  time.Time
	Nonce uint64
}

func (t *timeWithNonce) Marshal() []byte {
	u := t.Time.Unix()
	buf := make([]byte, 16)
	binary.LittleEndian.PutUint64(buf[:8], uint64(u))
	binary.LittleEndian.PutUint64(buf[8:], t.Nonce)
	return buf
}

func (t *timeWithNonce) Unmarshal(b []byte) {
	u := binary.LittleEndian.Uint64(b[:8])
	t.Time = time.Unix(int64(u), 0)
	t.Nonce = binary.LittleEndian.Uint64(b[8:])
	return
}

// MaxMessagePropagationTime is the maximum time we're expecting a message to
// take to propagate across the network. We wait double this before allowing a
// validator to start.
const MaxMessagePropagationTime = 60 * time.Second

// LastActionManager is an interface for lastActionManager
type LastActionManager interface {
	NewTip(row *chainindex.BlockRow, block *primitives.Block, state state.State, receipts []*primitives.EpochReceipt)
	ShouldRun(val [48]byte) bool
	GetNonce() uint64
	RegisterAction(b [48]byte, at time.Time, nonce uint64)
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

	l.RegisterAction(proposer.PubKey, time.Unix(int64(block.Header.Timestamp), 0), block.Header.Nonce)

	validators := state.GetValidatorRegistry()
	committee, err := state.GetVoteCommittee(block.Header.Slot)
	if err != nil {
		l.log.Error(err)
	} else {
		for _, v := range block.Votes {
			for idx, valIdx := range committee {
				if v.ParticipationBitfield.Get(uint(idx)) {
					val := validators[valIdx]
					l.RegisterAction(val.PubKey, time.Unix(int64(block.Header.Timestamp), 0), v.Data.Nonce)
				}
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

func (l *lastActionManager) ShouldRun(val [48]byte) bool {
	lastActionBytes, ok := l.lastActions.HasGet(nil, val[:])
	if !ok {
		return true
	}

	d := new(timeWithNonce)
	d.Unmarshal(lastActionBytes)

	if d.Nonce == l.nonce {
		return true
	}

	// last action was long enough ago we can start
	if time.Since(d.Time) > MaxMessagePropagationTime*2 {
		return true
	}

	return false
}

// RegisterAction registers an action by a validator at a certain time.
func (l *lastActionManager) RegisterAction(b [48]byte, at time.Time, nonce uint64) {
	if !l.ShouldRun(b) {
		return
	}

	d := &timeWithNonce{Time: at, Nonce: nonce}

	l.lastActions.Set(b[:], d.Marshal())
}

func (l *lastActionManager) GetNonce() uint64 {
	return l.nonce
}
