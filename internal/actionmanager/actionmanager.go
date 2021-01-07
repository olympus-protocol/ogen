package actionmanager

import (
	"context"
	"errors"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/bls/common"
	"github.com/olympus-protocol/ogen/pkg/p2p"
	"math/rand"
	"sync"
	"time"

	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/chainindex"
	"github.com/olympus-protocol/ogen/pkg/params"
	"github.com/olympus-protocol/ogen/pkg/primitives"

	"github.com/olympus-protocol/ogen/internal/hostnode"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/logger"
)

// MaxMessagePropagationTime is the maximum time we're expecting a message to
// take to propagate across the network. We wait double this before allowing a
// validator to start.
const MaxMessagePropagationTime = 60 * time.Second

// LastActionManager is an interface for lastActionManager
type LastActionManager interface {
	NewTip(row *chainindex.BlockRow, block *primitives.Block, state state.State, receipts []*primitives.EpochReceipt)
	StartValidator(valPub [48]byte, sign func(*primitives.ValidatorHelloMessage) common.Signature) bool
	ShouldRun(val [48]byte) bool
	RegisterActionAt(by [48]byte, at time.Time, nonce uint64)
	RegisterAction(by [48]byte, nonce uint64)
	GetNonce() uint64
}

var _ LastActionManager = &lastActionManager{}

// lastActionManager keeps track of the last action recorded by validators.
// This is a very basic protection against slashing. Validators, on startup,
// will broadcast a StartMessage
type lastActionManager struct {
	log logger.Logger

	host hostnode.HostNode
	ctx  context.Context

	nonce uint64

	// lastActions are the last recorded actions by a validator with a certain
	// salted private key hash.
	lastActions     map[[48]byte]time.Time
	lastActionsLock sync.Mutex

	netParams *params.ChainParams
}

func (l *lastActionManager) NewTip(_ *chainindex.BlockRow, block *primitives.Block, state state.State, _ []*primitives.EpochReceipt) {
	slotIndex := (block.Header.Slot + l.netParams.EpochLength - 1) % l.netParams.EpochLength

	proposerIndex := state.GetProposerQueue()[slotIndex]
	proposer := state.GetValidatorRegistry()[proposerIndex]

	l.RegisterActionAt(proposer.PubKey, time.Unix(int64(block.Header.Timestamp), 0), block.Header.Nonce)
}

func (l *lastActionManager) ProposerSlashingConditionViolated(*primitives.ProposerSlashing) {}

// NewLastActionManager creates a new last action manager.
func NewLastActionManager(node hostnode.HostNode, ch chain.Blockchain) (LastActionManager, error) {
	ctx := config.GlobalParams.Context
	log := config.GlobalParams.Logger
	netParams := config.GlobalParams.NetParams

	l := &lastActionManager{
		host:        node,
		ctx:         ctx,
		lastActions: make(map[[48]byte]time.Time),
		log:         log,
		nonce:       rand.Uint64(),
		netParams:   netParams,
	}

	if err := l.host.RegisterTopicHandler(p2p.MsgValidatorStartCmd, l.handleValidatorStart); err != nil {
		return nil, err
	}

	ch.Notify(l)

	return l, nil
}

func (l *lastActionManager) handleValidatorStart(id peer.ID, msg p2p.Message) error {
	if id == l.host.GetHost().ID() {
		return nil
	}

	data, ok := msg.(*p2p.MsgValidatorStart)
	if !ok {
		return errors.New("wrong message on start validator topic")
	}
	sig, err := bls.SignatureFromBytes(data.Data.Signature[:])
	if err != nil {
		return err
	}

	pub, err := bls.PublicKeyFromBytes(data.Data.PublicKey[:])
	if err != nil {
		return err
	}
	if !sig.Verify(pub, data.Data.SignatureMessage()) {
		return err
	}

	return nil
}

// StartValidator requests a validator to be started and returns whether it should be started.
func (l *lastActionManager) StartValidator(valPub [48]byte, sign func(*primitives.ValidatorHelloMessage) common.Signature) bool {
	l.lastActionsLock.Lock()
	defer l.lastActionsLock.Unlock()

	if !l.ShouldRun(valPub) {
		return false
	}

	validatorHello := new(primitives.ValidatorHelloMessage)
	validatorHello.PublicKey = valPub
	validatorHello.Timestamp = uint64(time.Now().Unix())

	signature := sign(validatorHello)
	var sig [96]byte
	copy(sig[:], signature.Marshal())
	validatorHello.Signature = sig

	msg := &p2p.MsgValidatorStart{Data: validatorHello}
	err := l.host.Broadcast(msg)
	if err != nil {
		return false
	}
	return true
}

func (l *lastActionManager) ShouldRun(val [48]byte) bool {
	l.lastActionsLock.Lock()
	defer l.lastActionsLock.Unlock()

	return l.shouldRun(val)
}

func (l *lastActionManager) shouldRun(pubSer [48]byte) bool {
	// no actions observed
	if _, ok := l.lastActions[pubSer]; !ok {
		return true
	}

	lastAction := l.lastActions[pubSer]

	// last action was long enough ago we can start
	if time.Since(lastAction) > MaxMessagePropagationTime*2 {
		return true
	}

	return false
}

// RegisterActionAt registers an action by a validator at a certain time.
func (l *lastActionManager) RegisterActionAt(by [48]byte, at time.Time, nonce uint64) {
	l.lastActionsLock.Lock()
	defer l.lastActionsLock.Unlock()

	if nonce == l.nonce {
		return
	}

	l.lastActions[by] = at
}

// RegisterAction registers an action by a validator.
func (l *lastActionManager) RegisterAction(by [48]byte, nonce uint64) {
	l.RegisterActionAt(by, time.Now(), nonce)
}

func (l *lastActionManager) GetNonce() uint64 {
	return l.nonce
}
