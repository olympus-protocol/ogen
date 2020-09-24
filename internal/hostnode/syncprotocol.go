package hostnode

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/logger"
	"github.com/olympus-protocol/ogen/pkg/p2p"
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

type peerInfo struct {
	TipBlockSlot uint64
}

var (
	// ErrorBlockAlreadyKnown returns when received a block already known
	ErrorBlockAlreadyKnown = errors.New("block already known")

	// ErrorBlockParentUnknown returns when received a block with an unknown parent
	ErrorBlockParentUnknown = errors.New("unknown block parent")
)

// SyncProtocol is an interface for the syncProtocol
type SyncProtocol interface {
	Notify(notifee SyncNotifee)
	Listen(network.Network, multiaddr.Multiaddr)
	ListenClose(network.Network, multiaddr.Multiaddr)
	Connected(net network.Network, conn network.Conn)
	Disconnected(net network.Network, conn network.Conn)
	OpenedStream(net network.Network, stream network.Stream)
	ClosedStream(network.Network, network.Stream)
	Syncing() bool
}

var _ SyncProtocol = &syncProtocol{}

// SyncProtocol handles syncing for a blockchain.
type syncProtocol struct {
	host   HostNode
	config Config
	ctx    context.Context
	log    logger.Logger

	chain chain.Blockchain

	protocolHandler ProtocolHandler

	notifees     []SyncNotifee
	notifeesLock sync.Mutex

	peersTrack     map[peer.ID]*peerInfo
	peersTrackLock sync.Mutex

	onSync   bool
	withPeer peer.ID
}

func listenToTopic(ctx context.Context, subscription *pubsub.Subscription, handler func(data []byte, id peer.ID)) {
	for {
		msg, err := subscription.Next(ctx)
		if err != nil {
			break
		}

		handler(msg.Data, msg.GetFrom())
	}
}

// NewSyncProtocol constructs a new sync protocol with a given host and chain.
func NewSyncProtocol(ctx context.Context, host HostNode, config Config, chain chain.Blockchain) (SyncProtocol, error) {
	ph := newProtocolHandler(ctx, syncProtocolID, host, config)
	sp := &syncProtocol{
		host:            host,
		config:          config,
		log:             config.Log,
		ctx:             ctx,
		protocolHandler: ph,
		chain:           chain,
		onSync:          true,
		peersTrack:      make(map[peer.ID]*peerInfo),
	}
	if err := ph.RegisterHandler(p2p.MsgVersionCmd, sp.handleVersion); err != nil {
		return nil, err
	}
	if err := ph.RegisterHandler(p2p.MsgGetBlocksCmd, sp.handleGetBlocks); err != nil {
		return nil, err
	}
	if err := ph.RegisterHandler(p2p.MsgBlockCmd, sp.blockHandler); err != nil {
		return nil, err
	}
	if err := ph.RegisterHandler(p2p.MsgSyncEndCmd, sp.syncEndHandler); err != nil {
		return nil, err
	}

	if err := sp.listenForBroadcasts(); err != nil {
		return nil, err
	}

	sp.host.Notify(sp)

	go sp.waitForPeers()

	return sp, nil
}

// waitForPeers will wait for 4 peers connected until start the sync routine.
func (sp *syncProtocol) waitForPeers() {
	for {
		time.Sleep(time.Second * 1)
		if sp.host.PeersConnected() < 2 {
			continue
		}
		break
	}

	go sp.startSync()

	return
}

// startSync will do some contextual checks among peers to evaluate our state and peers state.
func (sp *syncProtocol) startSync() {
	latestSlot := sp.chain.State().Tip().Slot

	peersHigher := make(map[int]peer.ID)
	peersSame := make(map[int]peer.ID)

	index := 0
	for id, p := range sp.peersTrack {
		if p.TipBlockSlot > latestSlot {
			peersHigher[index] = id
		}
		if p.TipBlockSlot == latestSlot {
			peersSame[index] = id

		}
		index++
	}

	if len(peersHigher) > len(peersSame) {

		r := rand.Intn(len(peersHigher))

		peerToSync := peersHigher[r]
		sp.onSync = true
		sp.withPeer = peerToSync

		err := sp.protocolHandler.SendMessage(peerToSync, &p2p.MsgGetBlocks{
			LastBlockHash: sp.chain.State().Chain().Tip().Hash,
		})

		if err != nil {
			sp.log.Error("unable to send block request msg")
		}

	} else {
		sp.onSync = false
		sp.withPeer = ""
		return
	}
}

type SyncNotifee interface{}

func (sp *syncProtocol) Notify(notifee SyncNotifee) {
	sp.notifeesLock.Lock()
	defer sp.notifeesLock.Unlock()
	sp.notifees = append(sp.notifees, notifee)
}

func (sp *syncProtocol) listenForBroadcasts() error {
	blockTopic, err := sp.host.Topic(p2p.MsgBlockCmd)
	if err != nil {
		return err
	}

	blockSub, err := blockTopic.Subscribe()
	if err != nil {
		return err
	}

	go listenToTopic(sp.ctx, blockSub, func(data []byte, id peer.ID) {
		if id == sp.host.GetHost().ID() {
			return
		}

		buf := bytes.NewBuffer(data)

		msg, err := p2p.ReadMessage(buf, sp.host.GetNetMagic())

		if err != nil {
			sp.log.Errorf("error decoding msg from peer %s: %s", id, err)
			return
		}

		block, ok := msg.(*p2p.MsgBlock)
		if !ok {
			sp.log.Errorf("wrong message type on block subscription from peer %s: %s", id, err)
			return
		}

		if err := sp.handleBlock(id, block.Data); err != nil {
			sp.log.Errorf("error handling incoming block from peer: %s", err)
		}
	})

	return nil
}

func (sp *syncProtocol) blockHandler(id peer.ID, msg p2p.Message) error {
	bmsg, ok := msg.(*p2p.MsgBlock)
	if !ok {
		return errors.New("non block msg")
	}
	return sp.handleBlock(id, bmsg.Data)
}

func (sp *syncProtocol) syncEndHandler(id peer.ID, msg p2p.Message) error {
	_, ok := msg.(*p2p.MsgSyncEnd)
	if !ok {
		return errors.New("non syncend msg")
	}
	sp.log.Info("syncing finished")
	if !sp.onSync {
		return nil
	}
	if sp.withPeer == id {
		sp.onSync = false
		sp.withPeer = ""
	}
	return nil
}

func (sp *syncProtocol) handleBlock(id peer.ID, block *primitives.Block) error {
	sp.peersTrackLock.Lock()
	_, ok := sp.peersTrack[id]
	if ok {
		sp.peersTrack[id].TipBlockSlot = block.Header.Slot
	}
	sp.peersTrackLock.Unlock()
	if sp.onSync && sp.withPeer != id {
		return nil
	}
	err := sp.processBlock(block)
	if err != nil {
		if err == ErrorBlockAlreadyKnown {
			sp.log.Error(err)
			return nil
		}
		if err == ErrorBlockParentUnknown {
			if !sp.onSync {
				sp.log.Error(err)
				sp.log.Info("restarting sync process")
				go sp.startSync()
				return nil
			}
			return nil
		}
	}
	return nil
}

func (sp *syncProtocol) processBlock(block *primitives.Block) error {

	// Check if we already have this block
	if sp.chain.State().Index().Have(block.Hash()) {
		return ErrorBlockAlreadyKnown
	}

	// Check if the parent block is known.
	if !sp.chain.State().Index().Have(block.Header.PrevBlockHash) {
		return ErrorBlockParentUnknown
	}

	// Process block
	sp.log.Debugf("processing block %s", block.Hash())
	if err := sp.chain.ProcessBlock(block); err != nil {
		return err
	}

	return nil
}

func (sp *syncProtocol) handleGetBlocks(id peer.ID, rawMsg p2p.Message) error {
	msg, ok := rawMsg.(*p2p.MsgGetBlocks)
	if !ok {
		return errors.New("did not receive get blocks message")
	}

	sp.log.Debug("received getblocks")

	// Get the announced last block to make sure we have a common point
	firstCommon, ok := sp.chain.State().Index().Get(msg.LastBlockHash)
	if !ok {
		sp.log.Errorf("unable to find common point for peer %s", id)
		return nil
	}

	for {
		var ok bool
		firstCommon, ok = sp.chain.State().Chain().Next(firstCommon)
		if !ok {
			break
		}

		block, err := sp.chain.GetBlock(firstCommon.Hash)
		if err != nil {
			return err
		}

		err = sp.protocolHandler.SendMessage(id, &p2p.MsgBlock{
			Data: block,
		})

		if err != nil {
			return err
		}

	}
	err := sp.protocolHandler.SendMessage(id, &p2p.MsgSyncEnd{})
	if err != nil {
		return err
	}
	return nil
}

func (sp *syncProtocol) handleVersion(id peer.ID, msg p2p.Message) error {
	theirVersion, ok := msg.(*p2p.MsgVersion)
	if !ok {
		return fmt.Errorf("did not receive version message")
	}

	sp.log.Infof("received version message from %s", id)

	// Send our version message if required
	ourVersion := sp.versionMsg()
	direction := sp.host.GetPeerDirection(id)

	if direction == network.DirInbound {
		if err := sp.protocolHandler.SendMessage(id, ourVersion); err != nil {
			return err
		}
	}

	sp.peersTrackLock.Lock()
	sp.peersTrack[id] = &peerInfo{
		TipBlockSlot: theirVersion.TipSlot,
	}
	sp.peersTrackLock.Unlock()

	return nil
}

func (sp *syncProtocol) versionMsg() *p2p.MsgVersion {
	lastSlot := sp.chain.State().Tip().Slot
	tipState := sp.chain.State().TipState()
	buf := make([]byte, 8)
	rand.Read(buf)
	msg := &p2p.MsgVersion{
		Nonce:              binary.LittleEndian.Uint64(buf),
		TipSlot:            lastSlot,
		Timestamp:          uint64(time.Now().Unix()),
		LastJustifiedHash:  tipState.GetJustifiedEpochHash(),
		LastJustifiedEpoch: tipState.GetJustifiedEpoch(),
	}
	return msg
}

func (sp *syncProtocol) sendVersion(id peer.ID) {
	if err := sp.protocolHandler.SendMessage(id, sp.versionMsg()); err != nil {
		sp.log.Errorf("error sending version message: %s", err)
		sp.host.DisconnectPeer(id)
	}
}

// Listen is called when we start listening on a multipraddr.
func (sp *syncProtocol) Listen(network.Network, multiaddr.Multiaddr) {}

// ListenClose is called when we stop listening on a multiaddr.
func (sp *syncProtocol) ListenClose(network.Network, multiaddr.Multiaddr) {}

// Connected is called when we connect to a peer.
func (sp *syncProtocol) Connected(net network.Network, conn network.Conn) {
	if conn.Stat().Direction != network.DirOutbound {
		return
	}

	// open a stream for the discovery protocol:
	s, err := sp.host.GetHost().NewStream(sp.ctx, conn.RemotePeer(), syncProtocolID)
	if err != nil {
		sp.log.Errorf("could not open stream for connection: %s", err)
	}

	sp.protocolHandler.HandleStream(s)

	sp.sendVersion(conn.RemotePeer())
}

// Disconnected is called when we disconnect from a peer.
func (sp *syncProtocol) Disconnected(net network.Network, conn network.Conn) {}

// OpenedStream is called when we open a stream.
func (sp *syncProtocol) OpenedStream(net network.Network, stream network.Stream) {}

// ClosedStream is called when we close a stream.
func (sp *syncProtocol) ClosedStream(network.Network, network.Stream) {}

func (sp *syncProtocol) Syncing() bool {
	return sp.onSync
}
