package actionmanager

import (
	"context"
	"errors"
	"github.com/VictoriaMetrics/fastcache"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/chainindex"
	"github.com/olympus-protocol/ogen/internal/host"
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
	ShouldRun(val [48]byte) (bool, error)
	RegisterActionAt(by [48]byte, at time.Time, nonce uint64) error
	RegisterAction(by [48]byte, nonce uint64) error
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

	err := l.RegisterActionAt(proposer.PubKey, time.Unix(int64(block.Header.Timestamp), 0), block.Header.Nonce)
	if err != nil {
		l.log.Error(err)
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

	l.host.RegisterTopicHandler(p2p.MsgValidatorStartCmd, l.handleValidatorStart)

	ch.Notify(l)

	return l, nil
}

func (l *lastActionManager) handleValidatorStart(id peer.ID, msg p2p.Message) error {

	if id == l.host.ID() {
		return nil
	}

	l.host.IncreasePeerReceivedBytes(id, msg.PayloadLength())

	data, ok := msg.(*p2p.MsgValidatorStart)
	if !ok {
		return errors.New("wrong message on start validator topic")
	}

	sig, err := bls.SignatureFromBytes(data.Data.Signature[:])
	if err != nil {
		return err
	}

	validators := data.Data.Validators.BitIndices()

	lastIdx := validators[len(validators)-1]

	if len(validators) > lastIdx {
		var pubs []common.PublicKey

		for _, validatorIdx := range validators {
			pub, err := bls.PublicKeyFromBytes(l.ch.State().TipState().GetValidatorRegistry()[validatorIdx].PubKey[:])
			if err != nil {
				return err
			}
			pubs = append(pubs, pub)
		}

		if !sig.FastAggregateVerify(pubs, data.Data.SignatureMessage()) {
			return err
		}

		for _, valPub := range pubs {
			var pub [48]byte
			copy(pub[:], valPub.Marshal())
			ok, err := l.ShouldRun(pub)
			if err != nil {
				return err
			}
			if ok {
				err := l.RegisterActionAt(pub, time.Unix(int64(data.Data.Timestamp), 0), data.Data.Nonce)
				if err != nil {
					return err
				}
			}
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
		ok, err := l.ShouldRun(pub)
		if err != nil {
			l.log.Error(err)
			continue
		}
		if ok {
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
func (l *lastActionManager) RegisterActionAt(by [48]byte, at time.Time, nonce uint64) error {
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

// RegisterAction registers an action by a validator.
func (l *lastActionManager) RegisterAction(by [48]byte, nonce uint64) error {
	err := l.RegisterActionAt(by, time.Now(), nonce)
	if err != nil {
		return err
	}
	return nil
}

func (l *lastActionManager) GetNonce() uint64 {
	return l.nonce
}
