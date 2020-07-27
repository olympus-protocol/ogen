package conflict

import (
	"context"
	"encoding/binary"
	"github.com/golang/snappy"
	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/chain/index"
	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/prysmaticlabs/go-ssz"
	"math/rand"
	"sync"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/peers"
	"github.com/olympus-protocol/ogen/utils/logger"
)

// ValidatorHelloMessage is a message sent by validators to indicate that they
// are coming online.
type ValidatorHelloMessage struct {
	PublicKey []byte
	Timestamp uint64
	Nonce     uint64
	Signature []byte
}

func (v *ValidatorHelloMessage) MaxPayloadLength() uint32 {
	return 160 // 48 + 8 + 8 + 96 =
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

// Serialize serializes the hello message to the given writer.
func (v *ValidatorHelloMessage) Serialize() ([]byte, error) {
	b, err := ssz.Marshal(v)
	if err != nil {
		return nil, err
	}
	if uint32(len(b)) > v.MaxPayloadLength() {
		return nil, p2p.ErrorSizeExceed
	}
	return snappy.Encode(nil, b), nil
}

// Deserialize deserializes the validator hello message from the reader.
func (v *ValidatorHelloMessage) Unmarshal(b []byte) error {
	d, err := snappy.Decode(nil, b)
	if err != nil {
		return err
	}
	if uint32(len(d)) > v.MaxPayloadLength() {
		return p2p.ErrorSizeExceed
	}
	return ssz.Unmarshal(d, v)
}

// MaxMessagePropagationTime is the maximum time we're expecting a message to
// take to propogate across the network. We wait double this before allowing a
// validator to start.
const MaxMessagePropagationTime = 60 * time.Second

type lastPing struct {
	nonce uint64
	time  uint64
}

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
	lastActions     map[[48]byte]lastPing
	lastActionsLock sync.RWMutex

	startTopic *pubsub.Topic

	params *params.ChainParams
}

func (l *LastActionManager) NewTip(row *index.BlockRow, block *primitives.Block, state *primitives.State, receipts []*primitives.EpochReceipt) {
	slotIndex := (block.Header.Slot + l.params.EpochLength - 1) % l.params.EpochLength

	proposerIndex := state.ProposerQueue[slotIndex]
	proposer := state.ValidatorRegistry[proposerIndex]

	var pubkey [48]byte
	copy(pubkey[:], proposer.PubKey)

	l.RegisterActionAt(pubkey, block.Header.Timestamp, block.Header.Nonce)
}

func (l *LastActionManager) ProposerSlashingConditionViolated(slashing primitives.ProposerSlashing) {}

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
		lastActions: make(map[[48]byte]lastPing),
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

		pub, err := bls.PublicKeyFromBytes(validatorHello.Signature)
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
func (l *LastActionManager) StartValidator(valPub []byte, sign func(*ValidatorHelloMessage) *bls.Signature) bool {
	l.lastActionsLock.RLock()
	defer l.lastActionsLock.RUnlock()

	pubSer := [48]byte{}
	copy(pubSer[:], valPub)

	if !l.ShouldRun(pubSer) {
		return false
	}

	validatorHello := new(ValidatorHelloMessage)
	validatorHello.PublicKey = valPub
	validatorHello.Timestamp = uint64(time.Now().Unix())

	signature := sign(validatorHello)
	validatorHello.Signature = signature.Marshal()

	msgBytes, _ := validatorHello.Serialize()

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
	lastActionTime := time.Unix(int64(lastAction.time), 0)

	// last action was long enough ago we can start
	if lastAction.nonce != l.nonce && time.Since(lastActionTime) > MaxMessagePropagationTime*2 {
		return true
	}

	return false
}

// RegisterActionAt registers an action by a validator at a certain time.
func (l *LastActionManager) RegisterActionAt(by [48]byte, at uint64, nonce uint64) {
	l.lastActionsLock.Lock()
	defer l.lastActionsLock.Unlock()

	l.lastActions[by] = lastPing{
		time:  at,
		nonce: nonce,
	}
}

// RegisterAction registers an action by a validator.
func (l *LastActionManager) RegisterAction(by [48]byte, nonce uint64) {
	l.RegisterActionAt(by, uint64(time.Now().Unix()), nonce)
}

func (l *LastActionManager) GetNonce() uint64 {
	return l.nonce
}
