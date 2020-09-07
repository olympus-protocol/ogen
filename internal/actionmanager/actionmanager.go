package actionmanager

import (
	"bytes"
	"context"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/p2p"
	"math/rand"
	"sync"
	"time"

	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/chainindex"
	"github.com/olympus-protocol/ogen/pkg/params"
	"github.com/olympus-protocol/ogen/pkg/primitives"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/olympus-protocol/ogen/internal/hostnode"
	"github.com/olympus-protocol/ogen/internal/logger"
	"github.com/olympus-protocol/ogen/pkg/bls"
)

// MaxMessagePropagationTime is the maximum time we're expecting a message to
// take to propagate across the network. We wait double this before allowing a
// validator to start.
const MaxMessagePropagationTime = 60 * time.Second

// LastActionManager is an interface for lastActionManager
type LastActionManager interface {
	NewTip(row *chainindex.BlockRow, block *primitives.Block, state state.State, receipts []*primitives.EpochReceipt)
	StartValidator(valPub [48]byte, sign func(*primitives.ValidatorHelloMessage) *bls.Signature) bool
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

	hostNode hostnode.HostNode
	ctx      context.Context

	nonce uint64

	// lastActions are the last recorded actions by a validator with a certain
	// salted private key hash.
	lastActions     map[[48]byte]time.Time
	lastActionsLock sync.RWMutex

	startTopic *pubsub.Topic

	params *params.ChainParams
}

func (l *lastActionManager) NewTip(_ *chainindex.BlockRow, block *primitives.Block, state state.State, receipts []*primitives.EpochReceipt) {
	slotIndex := (block.Header.Slot + l.params.EpochLength - 1) % l.params.EpochLength

	proposerIndex := state.GetProposerQueue()[slotIndex]
	proposer := state.GetValidatorRegistry()[proposerIndex]

	l.RegisterActionAt(proposer.PubKey, time.Unix(int64(block.Header.Timestamp), 0), block.Header.Nonce)
}

func (l *lastActionManager) ProposerSlashingConditionViolated(*primitives.ProposerSlashing) {}

// NewLastActionManager creates a new last action manager.
func NewLastActionManager(ctx context.Context, node hostnode.HostNode, log logger.Logger, ch chain.Blockchain, params *params.ChainParams) (LastActionManager, error) {
	topic, err := node.Topic(p2p.MsgValidatorStartCmd)
	if err != nil {
		return nil, err
	}

	topicSub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}

	l := &lastActionManager{
		hostNode:    node,
		ctx:         ctx,
		lastActions: make(map[[48]byte]time.Time),
		log:         log,
		startTopic:  topic,
		nonce:       rand.Uint64(),
		params:      params,
	}

	ch.Notify(l)

	go l.handleStartTopic(topicSub)

	return l, nil
}

func (l *lastActionManager) handleStartTopic(topic *pubsub.Subscription) {
	for {
		msg, err := topic.Next(l.ctx)
		if err != nil {
			l.log.Warnf("error getting next message in start validator topic: %s", err)
			return
		}

		buf := bytes.NewBuffer(msg.Data)

		p2pMsg, err := p2p.ReadMessage(buf, l.hostNode.GetNetMagic())

		if err != nil {
			return
		}

		validatorHello, ok := p2pMsg.(*p2p.MsgValidatorStart)
		if !ok {
			return
		}

		sig, err := bls.SignatureFromBytes(validatorHello.Data.Signature[:])
		if err != nil {
			l.log.Warnf("invalid signature: %s", err)
		}

		pub, err := bls.PublicKeyFromBytes(validatorHello.Data.PublicKey[:])
		if err != nil {
			l.log.Warnf("invalid pubkey: %s", err)
		}

		if !sig.Verify(pub, validatorHello.Data.SignatureMessage()) {
			l.log.Warnf("validator hello signature did not verify")
			return
		}
	}
}

// StartValidator requests a validator to be started and returns whether it should be started.
func (l *lastActionManager) StartValidator(valPub [48]byte, sign func(*primitives.ValidatorHelloMessage) *bls.Signature) bool {
	l.lastActionsLock.RLock()
	defer l.lastActionsLock.RUnlock()

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
	buf := bytes.NewBuffer([]byte{})
	err := p2p.WriteMessage(buf, msg, l.hostNode.GetNetMagic())
	if err != nil {
		return false
	}

	err = l.startTopic.Publish(l.ctx, buf.Bytes())

	if err != nil {
		return false
	}

	return true
}

func (l *lastActionManager) ShouldRun(val [48]byte) bool {
	l.lastActionsLock.RLock()
	defer l.lastActionsLock.RUnlock()

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
