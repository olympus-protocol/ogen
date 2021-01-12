package host

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	dsleveldb "github.com/ipfs/go-ds-leveldb"
	libhost "github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/olympus-protocol/ogen/pkg/params"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-peerstore/pstoreds"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/p2p"
	"path"
)

const connectionTimeout = 2000 * time.Millisecond
const connectionWait = 60 * time.Second

// MessageHandler is a handler for a specific message.
type MessageHandler func(id peer.ID, msg p2p.Message) error

type Host interface {
	ID() peer.ID
	Version() *p2p.MsgVersion
	Synced() bool
	ConnectedPeers() int
	GetPeersInfo() []*peerStats
	GetPeerDirection(p peer.ID) network.Direction
	SendMessage(id peer.ID, msg p2p.Message) error

	Notify(n *notify)
	Unnotify(n *notify)

	Disconnect(p peer.ID) error
	Connect(p peer.AddrInfo) error
	HandleConnection(net network.Network, conn network.Conn)

	RegisterTopicHandler(messageName string, handler MessageHandler)
	RegisterHandler(messageName string, handler MessageHandler)

	Broadcast(msg p2p.Message) error

	Stop()

	SetStreamHandler(pid protocol.ID, s network.StreamHandler)

	AddPeerStats(pid peer.ID, msg *p2p.MsgVersion, dir network.Direction)
	IncreasePeerReceivedBytes(p peer.ID, amount uint64)
}

type host struct {
	host     libhost.Host
	ctx      context.Context
	datapath string
	netMagic uint32
	log      logger.Logger
	chain    chain.Blockchain

	lastConnect     map[peer.ID]time.Time
	lastConnectLock sync.Mutex

	topic             *pubsub.Topic
	topicSub          *pubsub.Subscription
	topicHandlersLock sync.Mutex
	topicHandlers     map[string]MessageHandler

	messageHandler      map[string]MessageHandler
	messageHandlersLock sync.Mutex

	outgoingMessages     map[peer.ID]chan p2p.Message
	outgoingMessagesLock sync.Mutex

	stats        *stats
	discovery    *discovery
	synchronizer *synchronizer
}

var _ Host = &host{}

func (h *host) ID() peer.ID {
	return h.host.ID()
}

func (h *host) Version() *p2p.MsgVersion {

	justified, _ := h.chain.State().GetJustifiedHead()
	finalized, _ := h.chain.State().GetFinalizedHead()

	tip := h.chain.State().Chain().Tip()

	buf := make([]byte, 8)
	_, _ = rand.Read(buf)

	msg := &p2p.MsgVersion{
		Tip:             tip.Height,
		TipHash:         tip.Hash,
		Nonce:           binary.LittleEndian.Uint64(buf),
		Timestamp:       uint64(time.Now().Unix()),
		JustifiedSlot:   justified.Slot,
		JustifiedHeight: justified.Height,
		JustifiedHash:   justified.Hash,
		FinalizedSlot:   finalized.Slot,
		FinalizedHeight: finalized.Height,
		FinalizedHash:   finalized.Hash,
	}
	return msg
}

func (h *host) Synced() bool {
	return h.synchronizer.sync
}

func (h *host) ConnectedPeers() int {
	return len(h.host.Network().Peers())
}

func (h *host) GetPeersInfo() []*peerStats {
	peers := h.host.Network().Peers()
	var s []*peerStats
	for _, p := range peers {
		if stat, ok := h.stats.GetPeerStats(p); ok {
			s = append(s, stat)
		}
	}
	return s
}

func (h *host) GetPeerDirection(p peer.ID) network.Direction {
	conns := h.host.Network().ConnsToPeer(p)

	if len(conns) != 1 {
		return network.DirUnknown
	}
	return conns[0].Stat().Direction
}

func (h *host) Notify(n *notify) {
	h.host.Network().Notify(n)
}

func (h *host) Unnotify(n *notify) {
	h.host.Network().StopNotify(n)
}

// Disconnect disconnects to a peer
func (h *host) Disconnect(p peer.ID) error {
	err := h.host.Network().ClosePeer(p)
	if err != nil {
		return err
	}
	return nil
}

// Connect connects to a peer.
func (h *host) Connect(pi peer.AddrInfo) error {
	h.lastConnectLock.Lock()
	defer h.lastConnectLock.Unlock()
	lastConnect, found := h.lastConnect[pi.ID]
	if !found || time.Since(lastConnect) > connectionWait {
		h.lastConnect[pi.ID] = time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
		defer cancel()
		return h.host.Connect(ctx, pi)
	}
	return nil
}

func (h *host) HandleConnection(_ network.Network, conn network.Conn) {
	if conn.Stat().Direction != network.DirOutbound {
		return
	}

	s, err := h.host.NewStream(h.ctx, conn.RemotePeer(), params.ProtocolID(config.GlobalParams.NetParams.Name))
	if err != nil {
		h.log.Errorf("could not open stream for connection: %s", err)
	}

	h.handleStream(s)

	err = h.SendMessage(s.Conn().RemotePeer(), h.Version())
	if err != nil {
		h.log.Error(err)
	}
}

// RegisterTopicHandler registers a handler for a msg type on the pubsub channel.
func (h *host) RegisterTopicHandler(messageName string, handler MessageHandler) {
	h.topicHandlersLock.Lock()
	defer h.topicHandlersLock.Unlock()
	_, found := h.topicHandlers[messageName]
	if !found {
		h.topicHandlers[messageName] = handler
	}
	return
}

// RegisterHandler registers a handler for a msg type on the conn channel.
func (h *host) RegisterHandler(messageName string, handler MessageHandler) {
	h.messageHandlersLock.Lock()
	defer h.messageHandlersLock.Unlock()
	_, found := h.messageHandler[messageName]
	if !found {
		h.messageHandler[messageName] = handler
	}
	return
}

func (h *host) Broadcast(msg p2p.Message) error {
	buf := bytes.NewBuffer([]byte{})
	err := p2p.WriteMessage(buf, msg, h.netMagic)
	if err != nil {
		return err
	}
	return h.topic.Publish(h.ctx, buf.Bytes())
}

func (h *host) Stop() {
	h.stats.Close()
}

func (h *host) SetStreamHandler(pid protocol.ID, s network.StreamHandler) {
	h.host.SetStreamHandler(pid, s)
}

func (h *host) listenTopics() {
	for {
		msg, err := h.topicSub.Next(h.ctx)
		if err != nil {
			if err != h.ctx.Err() {
				h.log.Warnf("error getting next message: %s", err)
				continue
			}
			continue
		}

		if msg.GetFrom() == h.host.ID() {
			continue
		}

		buf := bytes.NewBuffer(msg.Data)

		msgData, err := p2p.ReadMessage(buf, h.netMagic)
		if err != nil {
			h.log.Warnf("unable to decode message: %s", err)
			continue
		}

		cmd := msgData.Command()
		h.topicHandlersLock.Lock()
		handler, found := h.topicHandlers[cmd]
		if !found {
			continue
		}
		h.topicHandlersLock.Unlock()
		err = handler(msg.GetFrom(), msgData)
		if err != nil {
			h.log.Error(err)
		}
	}
}

func (h *host) AddPeerStats(p peer.ID, ver *p2p.MsgVersion, dir network.Direction) {
	h.stats.Add(p, ver, dir)
}
func (h *host) IncreasePeerReceivedBytes(p peer.ID, amount uint64) {
	h.stats.IncreasePeerReceivedBytes(p, amount)
}

func NewHostNode(ch chain.Blockchain) (Host, error) {
	ctx := config.GlobalParams.Context
	log := config.GlobalParams.Logger
	netParams := config.GlobalParams.NetParams
	datapath := config.GlobalFlags.DataPath

	node := &host{
		ctx:              ctx,
		log:              log,
		netMagic:         netParams.NetMagic,
		datapath:         datapath,
		chain:            ch,
		topicHandlers:    make(map[string]MessageHandler),
		messageHandler:   make(map[string]MessageHandler),
		lastConnect:      make(map[peer.ID]time.Time),
		outgoingMessages: make(map[peer.ID]chan p2p.Message),
	}

	ds, err := dsleveldb.NewDatastore(path.Join(node.datapath, "peerstore"), nil)
	if err != nil {
		return nil, err
	}

	ps, err := pstoreds.NewPeerstore(node.ctx, ds, pstoreds.DefaultOpts())
	if err != nil {
		return nil, err
	}

	priv, err := node.loadPrivateKey()
	if err != nil {
		return nil, err
	}

	opts := buildOptions(priv, ps)
	h, err := libp2p.New(
		ctx,
		opts...,
	)
	if err != nil {
		return nil, err
	}

	node.host = h

	for _, a := range h.Addrs() {
		log.Infof("binding to address: %s", a)
	}

	g, err := pubsub.NewGossipSub(node.ctx, h)
	if err != nil {
		return nil, err
	}

	node.topic, err = g.Join("pub_channel")
	if err != nil {
		return nil, err
	}

	_, err = node.topic.Relay()
	if err != nil {
		return nil, err
	}

	node.topicSub, err = node.topic.Subscribe()
	if err != nil {
		return nil, err
	}

	go node.listenTopics()

	d, err := NewDiscovery(node.ctx, node, node.host)
	if err != nil {
		return nil, err
	}
	node.discovery = d

	s, err := NewStatsService(node)
	if err != nil {
		return nil, err
	}
	node.stats = s

	sy, err := NewSynchronizer(node, ch)
	if err != nil {
		return nil, err
	}
	node.synchronizer = sy

	n := NewNotify(node, s)

	node.Notify(n)

	node.SetStreamHandler(params.ProtocolID(netParams.Name), node.handleStream)

	return node, nil
}
