package hostnode

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	dsleveldb "github.com/ipfs/go-ds-leveldb"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-peerstore/pstoreds"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/p2p"
	"github.com/olympus-protocol/ogen/pkg/params"
	"io/ioutil"
	"os"
	"path"
	"time"
)

// HostNode is an interface for hostNode
type HostNode interface {
	Syncing() bool
	Stop()
	GetHost() host.Host
	GetNetMagic() uint32
	DisconnectPeer(p peer.ID) error
	GetPeerInfos() []peer.AddrInfo
	GetPeerDirection(id peer.ID) network.Direction
	GetPeerInfo(id peer.ID) *peer.AddrInfo
	RegisterHandler(message string, handler MessageHandler) error
	RegisterTopicHandler(message string, handler MessageHandler) error
	HandleStream(s network.Stream)
	SendMessage(id peer.ID, msg p2p.Message) error
	Broadcast(msg p2p.Message) error
	StatsService() *statsService
	VersionMsg() *p2p.MsgVersion
}

var _ HostNode = &hostNode{}

// HostNode is the node for p2p host
// It's the low level P2P communication layer, the App class handles high level protocols
// The RPC communication is hanlded by App, not HostNode
type hostNode struct {
	host     host.Host
	ctx      context.Context
	datapath string
	netMagic uint32
	log      logger.Logger
	chain    chain.Blockchain

	discover      *discover
	synchronizer  *synchronizer
	handler       *handler
	statesSerivce *statsService

	topic     *pubsub.Topic
	topicSub  *pubsub.Subscription
	listening bool
}

func (node *hostNode) Syncing() bool {
	return node.synchronizer.sync
}

// GetHost returns the host
func (node *hostNode) GetHost() host.Host {
	return node.host
}

func (node *hostNode) GetNetMagic() uint32 {
	return node.netMagic
}

// DisconnectPeer disconnects a peer
func (node *hostNode) DisconnectPeer(p peer.ID) error {
	return node.host.Network().ClosePeer(p)
}

// GetPeerInfos gets peer infos of connected hostnode.
func (node *hostNode) GetPeerInfos() []peer.AddrInfo {
	peers := node.host.Network().Peers()
	infos := make([]peer.AddrInfo, 0, len(peers))
	for _, p := range peers {
		addrInfo := node.host.Peerstore().PeerInfo(p)
		infos = append(infos, addrInfo)
	}

	return infos
}

// GetPeerDirection gets the direction of the peer.
func (node *hostNode) GetPeerDirection(id peer.ID) network.Direction {
	conns := node.host.Network().ConnsToPeer(id)

	if len(conns) != 1 {
		return network.DirUnknown
	}
	return conns[0].Stat().Direction
}

func (node *hostNode) GetPeerInfo(id peer.ID) *peer.AddrInfo {
	pinfo := node.host.Peerstore().PeerInfo(id)
	return &pinfo
}

func (node *hostNode) RegisterHandler(message string, handler MessageHandler) error {
	return node.handler.RegisterHandler(message, handler)
}

func (node *hostNode) RegisterTopicHandler(message string, handler MessageHandler) error {
	return node.handler.RegisterTopicHandler(message, handler)
}

func (node *hostNode) HandleStream(s network.Stream) {
	node.handler.handleStream(s)
}

func (node *hostNode) SendMessage(id peer.ID, msg p2p.Message) error {
	return node.handler.SendMessage(id, msg)
}

func (node *hostNode) Broadcast(msg p2p.Message) error {
	buf := bytes.NewBuffer([]byte{})
	err := p2p.WriteMessage(buf, msg, node.netMagic)
	if err != nil {
		return err
	}
	return node.topic.Publish(node.ctx, buf.Bytes())
}

func (node *hostNode) loadPrivateKey() (crypto.PrivKey, error) {
	keyBytes, err := ioutil.ReadFile(path.Join(node.datapath, "node_key.dat"))
	if err != nil {
		return node.createPrivateKey()
	}

	key, err := crypto.UnmarshalPrivateKey(keyBytes)
	if err != nil {
		return node.createPrivateKey()
	}
	return key, nil
}

func (node *hostNode) createPrivateKey() (crypto.PrivKey, error) {
	_ = os.RemoveAll(path.Join(node.datapath, "node_key.dat"))

	priv, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return nil, err
	}

	keyBytes, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile(path.Join(node.datapath, "node_key.dat"), keyBytes, 0700)
	if err != nil {
		return nil, err
	}

	return priv, nil
}

func (node *hostNode) listenerWatcher() {
	for {
		time.Sleep(time.Second * 5)
		if node.listening {
			continue
		} else {
			go node.listenTopics()
		}
	}
}

func (node *hostNode) listenTopics() {
	node.listening = true
	defer func() {
		node.listening = false
	}()

	for {
		msg, err := node.topicSub.Next(node.ctx)
		if err != nil {
			if err != node.ctx.Err() {
				node.log.Warnf("error getting next message in votes topic: %s", err)
				continue
			}
			continue
		}

		if msg.GetFrom() == node.host.ID() {
			continue
		}

		buf := bytes.NewBuffer(msg.Data)

		msgData, err := p2p.ReadMessage(buf, node.netMagic)
		if err != nil {
			node.log.Warnf("unable to decode message: %s", err)
			continue
		}

		cmd := msgData.Command()
		node.handler.topicHandlersLock.Lock()
		handler, found := node.handler.topicHandlers[cmd]
		if !found {
			continue
		}
		err = handler(msg.GetFrom(), msgData)
		if err != nil {
			node.log.Error(err)
		}
		node.handler.topicHandlersLock.Unlock()
	}
}
func (node *hostNode) Stop() {
	node.statesSerivce.Close()
}

func (node *hostNode) StatsService() *statsService {
	return node.statesSerivce
}

func (node *hostNode) VersionMsg() *p2p.MsgVersion {

	justified, _ := node.chain.State().GetJustifiedHead()
	finalized, _ := node.chain.State().GetFinalizedHead()

	tip := node.chain.State().Chain().Tip()

	buf := make([]byte, 8)
	rand.Read(buf)
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

// NewHostNode creates a host node
func NewHostNode(blockchain chain.Blockchain) (HostNode, error) {
	ctx := config.GlobalParams.Context
	log := config.GlobalParams.Logger
	netParams := config.GlobalParams.NetParams
	datapath := config.GlobalFlags.DataPath

	node := &hostNode{
		ctx:      ctx,
		log:      log,
		netMagic: netParams.NetMagic,
		datapath: datapath,
		chain:    blockchain,
	}

	pstats, err := NewPeersStatsService(node)
	if err != nil {
		return nil, err
	}
	node.statesSerivce = pstats

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

	listenAddress, err := ma.NewMultiaddr("/ip4/0.0.0.0/tcp/" + netParams.DefaultP2PPort)
	if err != nil {
		return nil, err
	}

	ip := ipAddr()
	opts := buildOptions(ip, priv, ps)
	h, err := libp2p.New(
		ctx,
		opts...,
	)
	if err != nil {
		return nil, err
	}

	node.host = h

	addrs, err := peer.AddrInfoToP2pAddrs(&peer.AddrInfo{
		ID:    h.ID(),
		Addrs: []ma.Multiaddr{listenAddress},
	})
	if err != nil {
		return nil, err
	}

	for _, a := range addrs {
		log.Infof("binding to address: %s", a)
	}

	g, err := pubsub.NewGossipSub(node.ctx, node.host)
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

	handler, err := newHandler(params.ProtocolID(config.GlobalParams.NetParams.Name), node)
	if err != nil {
		return nil, err
	}
	node.handler = handler

	go node.listenTopics()
	go node.listenerWatcher()

	synchronizer, err := NewSyncronizer(node, blockchain)
	if err != nil {
		return nil, err
	}
	node.synchronizer = synchronizer

	discovery, err := NewDiscover(node)
	if err != nil {
		return nil, err
	}
	node.discover = discovery

	return node, nil
}
