package actionmanager

import (
	"context"
	"errors"
	"github.com/VictoriaMetrics/fastcache"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/chainindex"
	"github.com/olympus-protocol/ogen/internal/hostnode"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/bitfield"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/bls/common"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/p2p"
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
	StartValidators(validators map[common.PublicKey]common.SecretKey) error
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
		ch:          ch,
		ctx:         ctx,
		lastActions: fastcache.New(128 * 1024 * 1024),
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

	if l.host.Syncing() {
		return nil
	}

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

	validators := data.Data.Validators.BitIndices()
	pubs := make([]common.PublicKey, len(validators))

	lastIdx := validators[len(validators)-1]

	if len(validators) > lastIdx {
		for i, validatorIdx := range validators {
			pub, err := bls.PublicKeyFromBytes(l.ch.State().TipState().GetValidatorRegistry()[validatorIdx].PubKey[:])
			if err != nil {
				return err
			}
			pubs[i] = pub
		}
	}
	
	if !sig.FastAggregateVerify(pubs, data.Data.SignatureMessage()) {
		return err
	}

	for _, valPub := range pubs {
		var pub [48]byte
		copy(pub[:], valPub.Marshal())
		if ok := l.ShouldRun(pub); ok {
			l.RegisterActionAt(pub, time.Unix(int64(data.Data.Timestamp), 0), data.Data.Nonce)
		}
	}

	return nil
}

// StartValidators requests a validator to be started and returns whether it should be started.
func (l *lastActionManager) StartValidators(validators map[common.PublicKey]common.SecretKey) error {

	safeRunValidators := make(map[[48]byte]common.SecretKey)
	for public, secret := range validators {
		var pub [48]byte
		copy(pub[:], public.Marshal())
		if l.ShouldRun(pub) {
			safeRunValidators[pub] = secret
		}
	}

	registry := l.ch.State().TipState().GetValidatorRegistry()

	bitlist := bitfield.NewBitlist(uint64(len(registry)))

	for i, val := range registry {
		_, ok := safeRunValidators[val.PubKey]
		if ok {
			bitlist.Set(uint(i))
		}
	}

	validatorHello := new(primitives.ValidatorHelloMessage)
	validatorHello.Nonce = l.nonce
	validatorHello.Timestamp = uint64(time.Now().Unix())
	validatorHello.Validators = bitlist

	var sigs []common.Signature
	msg := validatorHello.SignatureMessage()
	for _, k := range safeRunValidators {
		sigs = append(sigs, k.Sign(msg[:]))
	}

	signature := bls.AggregateSignatures(sigs)
	var sig [96]byte
	copy(sig[:], signature.Marshal())
	validatorHello.Signature = sig

	helloMsg := &p2p.MsgValidatorStart{Data: validatorHello}
	err := l.host.Broadcast(helloMsg)
	if err != nil {
		return err
	}
	return nil
}

func (l *lastActionManager) ShouldRun(val [48]byte) bool {
	var lastActionBytes []byte
	lastActionBytes, ok := l.lastActions.HasGet(lastActionBytes, val[:])
	if !ok {
		return true
	}

	var lastAction time.Time
	err := lastAction.UnmarshalBinary(lastActionBytes)
	if err != nil {
		return true
	}

	// last action was long enough ago we can start
	if time.Since(lastAction) > MaxMessagePropagationTime*2 {
		return true
	}

	return false
}

// RegisterActionAt registers an action by a validator at a certain time.
func (l *lastActionManager) RegisterActionAt(by [48]byte, at time.Time, nonce uint64) {
	if nonce == l.nonce {
		return
	}
	timeBytes, _ := at.MarshalBinary()
	l.lastActions.Set(by[:], timeBytes)

	return
}

// RegisterAction registers an action by a validator.
func (l *lastActionManager) RegisterAction(by [48]byte, nonce uint64) {
	l.RegisterActionAt(by, time.Now(), nonce)
}

func (l *lastActionManager) GetNonce() uint64 {
	return l.nonce
}
