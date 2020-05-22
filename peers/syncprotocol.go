package peers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/logger"
	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
)

const syncProtocolID = protocol.ID("/ogen/sync/1.0.0")

// SyncProtocol handles syncing for a blockchain.
type SyncProtocol struct {
	host   *HostNode
	config Config
	ctx    context.Context
	log    *logger.Logger

	chain *chain.Blockchain

	protocolHandler *ProtocolHandler

	notifees     []SyncNotifee
	notifeesLock sync.Mutex
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
func NewSyncProtocol(ctx context.Context, host *HostNode, config Config, chain *chain.Blockchain) (*SyncProtocol, error) {
	ph := newProtocolHandler(ctx, syncProtocolID, host, config)
	sp := &SyncProtocol{
		host:            host,
		config:          config,
		log:             config.Log,
		ctx:             ctx,
		protocolHandler: ph,
		chain:           chain,
	}

	if err := ph.RegisterHandler(p2p.MsgVersionCmd, sp.handleVersion); err != nil {
		return nil, err
	}
	if err := ph.RegisterHandler(p2p.MsgGetBlocksCmd, sp.handleGetBlocks); err != nil {
		return nil, err
	}
	if err := ph.RegisterHandler(p2p.MsgBlocksCmd, sp.handleBlocks); err != nil {
		return nil, err
	}

	if err := sp.listenForBroadcasts(); err != nil {
		return nil, err
	}

	host.Notify(sp)

	return sp, nil
}

type SyncNotifee interface {
}

func (sp *SyncProtocol) Notify(notifee SyncNotifee) {
	sp.notifeesLock.Lock()
	defer sp.notifeesLock.Unlock()
	sp.notifees = append(sp.notifees, notifee)
}

func (sp *SyncProtocol) listenForBroadcasts() error {
	blockTopic, err := sp.host.Topic("blocks")
	if err != nil {
		return err
	}

	blockSub, err := blockTopic.Subscribe()
	if err != nil {
		return err
	}

	go listenToTopic(sp.ctx, blockSub, func(data []byte, id peer.ID) {
		buf := bytes.NewReader(data)
		var block primitives.Block

		if err := block.Decode(buf); err != nil {
			sp.log.Errorf("error decoding block from peer %s: %s", id, err)
			return
		}

		if err := sp.handleBlock(id, &block); err != nil {
			sp.log.Errorf("error handling incoming block from peer: %s", err)
		}
	})

	return nil
}

func (sp *SyncProtocol) handleBlock(id peer.ID, block *primitives.Block) error {
	bh := block.Hash()
	if !sp.chain.State().Index().Have(block.Header.PrevBlockHash) {
		sp.log.Debugf("don't have block %s, requesting", block.Header.PrevBlockHash)
		// TODO: better peer selection for syncing
		peers := sp.host.GetPeerList()
		if len(peers) > 0 {
			err := sp.protocolHandler.SendMessage(peers[0], &p2p.MsgGetBlocks{
				LocatorHashes: sp.chain.GetLocatorHashes(),
				HashStop:      bh,
			})
			return err
		} else {
			return nil
		}
	}

	if sp.chain.State().Index().Have(bh) {
		return nil
	}

	sp.log.Debugf("processing block %s", bh)
	if err := sp.chain.ProcessBlock(block); err != nil {
		return err
	}

	return nil
}

func (sp *SyncProtocol) handleBlocks(id peer.ID, rawMsg p2p.Message) error {
	msg, ok := rawMsg.(*p2p.MsgBlocks)
	if !ok {
		return errors.New("did not receive blocks message")
	}

	sp.log.Tracef("received blocks msg from peer %v", id)
	for _, b := range msg.Blocks {
		if err := sp.handleBlock(id, &b); err != nil {
			return err
		}
	}
	return nil
}

func (sp *SyncProtocol) handleGetBlocks(id peer.ID, rawMsg p2p.Message) error {
	msg, ok := rawMsg.(*p2p.MsgGetBlocks)
	if !ok {
		return errors.New("did not receive get blocks message")
	}

	sp.log.Debug("received getblocks")

	// first block is tip, so we check each block in order and check if the block matches
	firstCommon := sp.chain.State().Chain().Genesis()
	locatorHashesGenesis := &msg.LocatorHashes[len(msg.LocatorHashes)-1]

	if !firstCommon.Hash.IsEqual(locatorHashesGenesis) {
		return fmt.Errorf("incorrect genesis block (got: %s, expected: %s)", locatorHashesGenesis, firstCommon.Hash)
	}

	for _, b := range msg.LocatorHashes {
		if b, found := sp.chain.State().Index().Get(b); found {
			firstCommon = b
			break
		}
	}

	sp.log.Debugf("found first common block %s", firstCommon.Hash)

	blocksToSend := make([]primitives.Block, 0, 500)

	if firstCommon.Hash.IsEqual(locatorHashesGenesis) {
		fc, ok := sp.chain.State().Chain().Next(firstCommon)
		if !ok {
			return nil
		}
		firstCommon = fc
	}

	for firstCommon != nil && len(blocksToSend) < p2p.MaxBlocksPerMsg {
		block, err := sp.chain.GetBlock(firstCommon.Hash)
		if err != nil {
			return err
		}

		blocksToSend = append(blocksToSend, *block)

		if firstCommon.Hash.IsEqual(&msg.HashStop) {
			break
		}
		var ok bool
		firstCommon, ok = sp.chain.State().Chain().Next(firstCommon)
		if !ok {
			break
		}
	}

	sp.log.Debugf("sending %d blocks", len(blocksToSend))

	return sp.protocolHandler.SendMessage(id, &p2p.MsgBlocks{
		Blocks: blocksToSend,
	})
}

func (sp *SyncProtocol) handleVersion(id peer.ID, msg p2p.Message) error {
	theirVersion, ok := msg.(*p2p.MsgVersion)
	if !ok {
		return fmt.Errorf("did not receive version message")
	}

	sp.log.Infof("received version message from %s", id)

	// validate version message here
	ourVersion := sp.versionMsg()
	direction := sp.host.GetPeerDirection(id)

	if direction == network.DirInbound {
		if err := sp.protocolHandler.SendMessage(id, ourVersion); err != nil {
			return err
		}
	}

	if theirVersion.LastBlock > ourVersion.LastBlock {
		err := sp.protocolHandler.SendMessage(id, &p2p.MsgGetBlocks{
			HashStop:      chainhash.Hash{},
			LocatorHashes: sp.chain.GetLocatorHashes(),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (sp *SyncProtocol) versionMsg() *p2p.MsgVersion {
	lastBlockHeight := sp.chain.State().Tip().Height
	nonce, _ := serializer.RandomUint64()
	msg := p2p.NewMsgVersion(nonce, lastBlockHeight)
	return msg
}

func (sp *SyncProtocol) sendVersion(id peer.ID) {
	if err := sp.protocolHandler.SendMessage(id, sp.versionMsg()); err != nil {
		sp.log.Errorf("error sending version message: %s", err)
		sp.host.DisconnectPeer(id)
	}
}

// Listen is called when we start listening on a multipraddr.
func (sp *SyncProtocol) Listen(network.Network, multiaddr.Multiaddr) {}

// ListenClose is called when we stop listening on a multiaddr.
func (sp *SyncProtocol) ListenClose(network.Network, multiaddr.Multiaddr) {}

// Connected is called when we connect to a peer.
func (sp *SyncProtocol) Connected(net network.Network, conn network.Conn) {
	if conn.Stat().Direction != network.DirOutbound {
		return
	}

	// open a stream for the discovery protocol:
	s, err := sp.host.host.NewStream(sp.ctx, conn.RemotePeer(), syncProtocolID)
	if err != nil {
		sp.log.Errorf("could not open stream for connection: %s", err)
	}

	sp.protocolHandler.handleStream(s)

	sp.sendVersion(conn.RemotePeer())
}

// Disconnected is called when we disconnect from a peer.
func (sp *SyncProtocol) Disconnected(net network.Network, conn network.Conn) {}

// OpenedStream is called when we open a stream.
func (sp *SyncProtocol) OpenedStream(net network.Network, stream network.Stream) {
	// start the sync process now
}

// ClosedStream is called when we close a stream.
func (sp *SyncProtocol) ClosedStream(network.Network, network.Stream) {}
