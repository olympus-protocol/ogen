package conflict

import (
	"context"
	"encoding/binary"
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/golang/snappy"
	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/chain/index"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/primitives"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/peers"
	"github.com/olympus-protocol/ogen/utils/logger"
)

var (
	// ErrorValidatorHelloMessageSize returns when serialized MaxValidatorHelloMessageSize size exceed MaxValidatorHelloMessageSize.
	ErrorValidatorHelloMessageSize = errors.New("validator hello message too big")
)

const (
	// MaxValidatorHelloMessageSize is the maximum amount of bytes a CombinedSignature can contain.
	MaxValidatorHelloMessageSize = 160
)

// ValidatorHelloMessage is a message sent by validators to indicate that they are coming online.
type ValidatorHelloMessage struct {
	PublicKey [48]byte
	Timestamp uint64
	Nonce     uint64
	Signature [96]byte
}

// SignatureMessage gets the signed portion of the message.
func (v *ValidatorHelloMessage) SignatureMessage() []byte {
	timeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timeBytes, v.Timestamp)

	nonceBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBytes, v.Nonce)

	msg := []byte{}
	msg = append(msg, v.PublicKey[:]...)
	msg = append(msg, timeBytes...)
	msg = append(msg, nonceBytes...)

	return msg
}

// Marshal serializes the hello message to the given writer.
func (v *ValidatorHelloMessage) Marshal() ([]byte, error) {
	b, err := v.MarshalSSZ()
	if err != nil {
		return nil, err
	}
	if len(b) > MaxValidatorHelloMessageSize {
		return nil, ErrorValidatorHelloMessageSize
	}
	return snappy.Encode(nil, b), nil
}

// Unmarshal deserializes the validator hello message from the reader.
func (v *ValidatorHelloMessage) Unmarshal(b []byte) error {
	d, err := snappy.Decode(nil, b)
	if err != nil {
		return err
	}
	if len(d) > MaxValidatorHelloMessageSize {
		return ErrorValidatorHelloMessageSize
	}
	return v.UnmarshalSSZ(d)
}

// MaxMessagePropagationTime is the maximum time we're expecting a message to
// take to propagate across the network. We wait double this before allowing a
// validator to start.
const MaxMessagePropagationTime = 60 * time.Second

// LastActionManager keeps track of the last action recorded by validators.
// This is a very basic protection against slashing. Validators, on startup,
// will broadcast a StartMessage
type LastActionManager struct {
	log *logger.Logger

	hostNode *peers.HostNode
	ctx      context.Context

	nonce uint64

	// lastActions are the last recorded actions by a validator with a certain
	// salted private key hash.
	lastActions     map[[48]byte]time.Time
	lastActionsLock sync.RWMutex

	startTopic *pubsub.Topic

	params *params.ChainParams
}

func (l *LastActionManager) NewTip(row *index.BlockRow, block *primitives.Block, state *primitives.State, receipts []*primitives.EpochReceipt) {
	slotIndex := (block.Header.Slot + l.params.EpochLength - 1) % l.params.EpochLength

	proposerIndex := state.ProposerQueue[slotIndex]
	proposer := state.ValidatorRegistry[proposerIndex]

	l.RegisterActionAt(proposer.PubKey, time.Unix(int64(block.Header.Timestamp), 0), block.Header.Nonce)
}

func (l *LastActionManager) ProposerSlashingConditionViolated(slashing *primitives.ProposerSlashing) {
}

const validatorStartTopic = "validatorStart"

// NewLastActionManager creates a new last action manager.
func NewLastActionManager(ctx context.Context, node *peers.HostNode, log *logger.Logger, ch *chain.Blockchain, params *params.ChainParams) (*LastActionManager, error) {
	topic, err := node.Topic(validatorStartTopic)
	if err != nil {
		return nil, err
	}

	topicSub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}

	l := &LastActionManager{
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

func (l *LastActionManager) handleStartTopic(topic *pubsub.Subscription) {
	for {
		msg, err := topic.Next(l.ctx)
		if err != nil {
			l.log.Warnf("error getting next message in start validator topic: %s", err)
			return
		}

		validatorHello := new(ValidatorHelloMessage)

		if err := validatorHello.Unmarshal(msg.Data); err != nil {
			l.log.Warnf("invalid validator hello: %s", err)
			return
		}

		sig, err := bls.SignatureFromBytes(validatorHello.Signature)
		if err != nil {
			l.log.Warnf("invalid signature: %s", err)
		}

		pub, err := bls.PublicKeyFromBytes(validatorHello.PublicKey)
		if err != nil {
			l.log.Warnf("invalid pubkey: %s", err)
		}

		if !sig.Verify(validatorHello.SignatureMessage(), pub) {
			l.log.Warnf("validator hello signature did not verify")
			return
		}
	}
}

// StartValidator requests a validator to be started and returns whether it should be started.
func (l *LastActionManager) StartValidator(valPub [48]byte, sign func(*ValidatorHelloMessage) *bls.Signature) bool {
	l.lastActionsLock.RLock()
	defer l.lastActionsLock.RUnlock()

	if !l.ShouldRun(valPub) {
		return false
	}

	validatorHello := new(ValidatorHelloMessage)
	validatorHello.PublicKey = valPub
	validatorHello.Timestamp = uint64(time.Now().Unix())

	signature := sign(validatorHello)
	var sig [96]byte
	copy(sig[:], signature.Marshal())
	validatorHello.Signature = sig

	msgBytes, _ := validatorHello.Marshal()

	l.startTopic.Publish(l.ctx, msgBytes)

	return true
}

func (l *LastActionManager) ShouldRun(val [48]byte) bool {
	l.lastActionsLock.RLock()
	defer l.lastActionsLock.RUnlock()

	return l.shouldRun(val)
}

func (l *LastActionManager) shouldRun(pubSer [48]byte) bool {
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
func (l *LastActionManager) RegisterActionAt(by [48]byte, at time.Time, nonce uint64) {
	l.lastActionsLock.Lock()
	defer l.lastActionsLock.Unlock()

	if nonce == l.nonce {
		return
	}

	l.lastActions[by] = at
}

// RegisterAction registers an action by a validator.
func (l *LastActionManager) RegisterAction(by [48]byte, nonce uint64) {
	l.RegisterActionAt(by, time.Now(), nonce)
}

func (l *LastActionManager) GetNonce() uint64 {
	return l.nonce
}
