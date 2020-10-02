package hostnode

import (
	"context"
	"crypto/rand"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"time"

	dsbadger "github.com/ipfs/go-ds-badger"
	"github.com/libp2p/go-libp2p"
	circuit "github.com/libp2p/go-libp2p-circuit"
	"github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/libp2p/go-libp2p-peerstore/pstoreds"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/p2p"
)

// HostNode is an interface for hostNode
type HostNode interface {
	Topic(topic string) (*pubsub.Topic, error)
	Syncing() bool
	GetContext() context.Context
	GetHost() host.Host
	GetNetMagic() uint32
	DisconnectPeer(p peer.ID) error
	GetPeerList() []peer.ID
	GetPeerInfos() []peer.AddrInfo
	ConnectedToPeer(id peer.ID) bool
	Notify(notifee network.Notifiee)
	GetPeerDirection(id peer.ID) network.Direction
	Stop()
	SetStreamHandler(id protocol.ID, handleStream func(s network.Stream))
	GetPeerInfo(id peer.ID) *peer.AddrInfo
}

var _ HostNode = &hostNode{}

// HostNode is the node for p2p host
// It's the low level P2P communication layer, the App class handles high level protocols
// The RPC communication is hanlded by App, not HostNode
type hostNode struct {
	privateKey crypto.PrivKey

	host      host.Host
	gossipSub *pubsub.PubSub
	ctx       context.Context

	topics     map[string]*pubsub.Topic
	topicsLock sync.RWMutex

	netMagic uint32

	log      logger.Logger
	datapath string

	// discoveryProtocol handles peer discovery (mDNS, DHT, etc)
	discoveryProtocol *discoveryProtocol

	// syncProtocol handles peer syncing
	syncProtocol *syncProtocol
}

// NewHostNode creates a host node
func NewHostNode(blockchain chain.Blockchain) (HostNode, error) {
	ctx := config.GlobalParams.Context
	log := config.GlobalParams.Logger
	netParams := config.GlobalParams.NetParams

	node := &hostNode{
		ctx:      ctx,
		log:      log,
		topics:   map[string]*pubsub.Topic{},
		netMagic: netParams.NetMagic,
		datapath: config.GlobalFlags.DataPath,
	}

	ds, err := dsbadger.NewDatastore(path.Join(node.datapath, "peerstore"), nil)
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
	node.privateKey = priv

	listenAddress, err := ma.NewMultiaddr("/ip4/0.0.0.0/tcp/" + netParams.DefaultP2PPort)
	if err != nil {
		return nil, err
	}

	connman := connmgr.NewConnManager(2, 64, time.Second*60)
	h, err := libp2p.New(
		ctx,
		libp2p.ListenAddrs([]ma.Multiaddr{listenAddress}...),
		libp2p.Identity(priv),
		libp2p.EnableRelay(circuit.OptActive, circuit.OptHop),
		libp2p.Peerstore(ps),
		libp2p.ConnectionManager(connman),
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

	g, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		return nil, err
	}
	node.gossipSub = g

	syncProtocol, err := NewSyncProtocol(node, blockchain)
	if err != nil {
		return nil, err
	}
	node.syncProtocol = syncProtocol

	discovery, err := NewDiscoveryProtocol(node)
	if err != nil {
		return nil, err
	}
	node.discoveryProtocol = discovery

	return node, nil
}

func (node *hostNode) Topic(topic string) (*pubsub.Topic, error) {
	node.topicsLock.Lock()
	defer node.topicsLock.Unlock()

	if t, ok := node.topics[topic]; ok {
		return t, nil
	}

	t, err := node.gossipSub.Join(topic)
	if err != nil {
		return nil, err
	}

	node.relay(topic, t)

	node.topics[topic] = t
	return t, nil
}

func (node *hostNode) relay(topic string, pub *pubsub.Topic) {
	switch topic {
	case p2p.MsgBlockCmd:
		_, err := pub.Relay()
		if err != nil {
			node.log.Error(err)
		}
	case p2p.MsgTxCmd:
		_, err := pub.Relay()
		if err != nil {
			node.log.Error(err)
		}
	case p2p.MsgTxMultiCmd:
		_, err := pub.Relay()
		if err != nil {
			node.log.Error(err)
		}
	case p2p.MsgDepositCmd:
		_, err := pub.Relay()
		if err != nil {
			node.log.Error(err)
		}
	case p2p.MsgDepositsCmd:
		_, err := pub.Relay()
		if err != nil {
			node.log.Error(err)
		}
	case p2p.MsgExitCmd:
		_, err := pub.Relay()
		if err != nil {
			node.log.Error(err)
		}
	case p2p.MsgExitsCmd:
		_, err := pub.Relay()
		if err != nil {
			node.log.Error(err)
		}
	case p2p.MsgGovernanceCmd:
		_, err := pub.Relay()
		if err != nil {
			node.log.Error(err)
		}
	case p2p.MsgVoteCmd:
		_, err := pub.Relay()
		if err != nil {
			node.log.Error(err)
		}
	case p2p.MsgValidatorStartCmd:
		_, err := pub.Relay()
		if err != nil {
			node.log.Error(err)
		}
	}
}

// Syncing returns a boolean if the chain is on sync mode
func (node *hostNode) Syncing() bool {
	return node.syncProtocol.Syncing()
}

// GetContext returns the context
func (node *hostNode) GetContext() context.Context {
	return node.ctx
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

// PeersConnected checks how many hostnode are connected.
func (node *hostNode) PeersConnected() int {
	return len(node.host.Network().Peers())
}

// GetPeerList returns a list of all hostnode.
func (node *hostNode) GetPeerList() []peer.ID {
	return node.host.Network().Peers()
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

// ConnectedToPeer returns true if we're connected to the peer.
func (node *hostNode) ConnectedToPeer(id peer.ID) bool {
	connectedness := node.host.Network().Connectedness(id)
	return connectedness == network.Connected
}

// Notify notifies a notifee for network events.
func (node *hostNode) Notify(notifee network.Notifiee) {
	node.host.Network().Notify(notifee)
}

// SetStreamHandler sets a stream handler for the host node.
func (node *hostNode) SetStreamHandler(id protocol.ID, handleStream func(s network.Stream)) {
	node.host.SetStreamHandler(id, handleStream)
}

// GetPeerDirection gets the direction of the peer.
func (node *hostNode) GetPeerDirection(id peer.ID) network.Direction {
	conns := node.host.Network().ConnsToPeer(id)

	if len(conns) != 1 {
		return network.DirUnknown
	}
	return conns[0].Stat().Direction
}

// Stop closes all topics before closing the server.
func (node *hostNode) Stop() {
	for _, topic := range node.topics {
		_ = topic.Close()
	}
}

func (node *hostNode) GetPeerInfo(id peer.ID) *peer.AddrInfo {
	pinfo := node.host.Peerstore().PeerInfo(id)
	return &pinfo
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
