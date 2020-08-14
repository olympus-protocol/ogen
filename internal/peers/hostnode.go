package peers

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/logger"

	"github.com/libp2p/go-libp2p-peerstore/pstoremem"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
	mnet "github.com/multiformats/go-multiaddr-net"
)

type Config struct {
	Log          logger.Logger
	Port         string
	InitialNodes []peer.AddrInfo
	Path         string
	PrivateKey   crypto.PrivKey
}

const (
	OgenVersion       = "0.0.1"
	timeoutInterval   = 60 * time.Second
	heartbeatInterval = 20 * time.Second
)

//HostNode is an interface for hostNode
type HostNode interface {
	SyncProtocol() SyncProtocol
	Topic(topic string) (*pubsub.Topic, error)
	Syncing() bool
	GetContext() context.Context
	GetHost() host.Host
	GetNetMagic() uint32
	removePeer(p peer.ID)
	DisconnectPeer(p peer.ID) error
	IsConnected() bool
	PeersConnected() int
	GetPeerList() []peer.ID
	GetPeerInfos() []peer.AddrInfo
	ConnectedToPeer(id peer.ID) bool
	Notify(notifee network.Notifiee)
	setStreamHandler(id protocol.ID, handleStream func(s network.Stream))
	CountPeers(id protocol.ID) int
	GetPeerDirection(id peer.ID) network.Direction
	Start() error
	SavePeer(pma multiaddr.Multiaddr) error
	BanScorePeer(id peer.ID, weight int) error
	IsPeerBanned(id peer.ID) (bool, error)
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

	timeoutInterval   time.Duration
	heartbeatInterval time.Duration

	netMagic uint32

	log logger.Logger

	// discoveryProtocol handles peer discovery (mDNS, DHT, etc)
	discoveryProtocol DiscoveryProtocol

	// syncProtocol handles peer syncing
	syncProtocol SyncProtocol
	db           Database
}

// NewHostNode creates a host node
func NewHostNode(ctx context.Context, config Config, blockchain chain.Blockchain) (HostNode, error) {
	ps := pstoremem.NewPeerstore()
	db, err := NewDatabase(config.Path)
	if err != nil {
		return nil, err
	}

	err = db.Initialize()
	if err != nil {
		return nil, err
	}

	priv, err := db.GetPrivKey()
	if err != nil {
		return nil, err
	}
	// get saved peers
	savedAddresses, err := db.GetSavedPeers()
	if err != nil {
		config.Log.Errorf("error retrieving saved peers: %s", err)
	}

	netAddr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:"+config.Port)
	if err != nil {
		return nil, err
	}

	listen, err := mnet.FromNetAddr(netAddr)
	if err != nil {
		return nil, err
	}
	listenAddress := []multiaddr.Multiaddr{listen}

	//append saved addresses
	listenAddress = append(listenAddress, savedAddresses...)

	h, err := libp2p.New(
		ctx,
		libp2p.ListenAddrs(listenAddress...),
		libp2p.Identity(priv),
		libp2p.EnableRelay(),
		libp2p.Peerstore(ps),
	)

	if err != nil {
		return nil, err
	}

	addrs, err := peer.AddrInfoToP2pAddrs(&peer.AddrInfo{
		ID:    h.ID(),
		Addrs: listenAddress,
	})
	if err != nil {
		return nil, err
	}

	for _, a := range addrs {
		config.Log.Infof("binding to address: %s", a)
	}

	// setup gossip sub protocol
	g, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		return nil, err
	}

	node := &hostNode{
		privateKey:        config.PrivateKey,
		host:              h,
		gossipSub:         g,
		ctx:               ctx,
		timeoutInterval:   timeoutInterval,
		heartbeatInterval: heartbeatInterval,
		log:               config.Log,
		topics:            map[string]*pubsub.Topic{},
		db:                db,
	}

	discovery, err := NewDiscoveryProtocol(ctx, node, config)
	if err != nil {
		return nil, err
	}
	node.discoveryProtocol = discovery

	syncProtocol, err := NewSyncProtocol(ctx, node, config, blockchain)
	if err != nil {
		return nil, err
	}
	node.syncProtocol = syncProtocol

	return node, nil
}

func (node *hostNode) SyncProtocol() SyncProtocol {
	return node.syncProtocol
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

	node.topics[topic] = t
	return t, nil
}

// Syncing returns a boolean if the chain is on sync mode
func (node *hostNode) Syncing() bool {
	return node.syncProtocol.syncing()
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

func (node *hostNode) removePeer(p peer.ID) {
	node.host.Peerstore().ClearAddrs(p)
}

// DisconnectPeer disconnects a peer
func (node *hostNode) DisconnectPeer(p peer.ID) error {
	return node.host.Network().ClosePeer(p)
}

// IsConnected checks if the host node is connected.
func (node *hostNode) IsConnected() bool {
	return node.PeersConnected() > 0
}

// PeersConnected checks how many peers are connected.
func (node *hostNode) PeersConnected() int {
	return len(node.host.Network().Peers())
}

// GetPeerList returns a list of all peers.
func (node *hostNode) GetPeerList() []peer.ID {
	return node.host.Network().Peers()
}

// GetPeerInfos gets peer infos of connected peers.
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

// setStreamHandler sets a stream handler for the host node.
func (node *hostNode) setStreamHandler(id protocol.ID, handleStream func(s network.Stream)) {
	node.host.SetStreamHandler(id, handleStream)
}

// CountPeers counts the number of peers that support the protocol.
func (node *hostNode) CountPeers(id protocol.ID) int {
	count := 0
	for _, n := range node.host.Peerstore().Peers() {
		if sup, err := node.host.Peerstore().SupportsProtocols(n, string(id)); err != nil && len(sup) != 0 {
			count++
		}
	}
	return count
}

// GetPeerDirection gets the direction of the peer.
func (node *hostNode) GetPeerDirection(id peer.ID) network.Direction {
	conns := node.host.Network().ConnsToPeer(id)

	if len(conns) != 1 {
		return network.DirUnknown
	}
	return conns[0].Stat().Direction
}

// Start the host node and start discovering peers.
func (node *hostNode) Start() error {
	if err := node.discoveryProtocol.Start(); err != nil {
		return err
	}

	return nil
}

// Database <-> hostNode Functions

func (node *hostNode) SavePeer(pma multiaddr.Multiaddr) error {
	if node.db == nil {
		return errors.New("no initialized db in node")
	}
	return node.db.SavePeer(pma)
}

func (node *hostNode) BanScorePeer(id peer.ID, weight int) error {
	if node.db == nil {
		return errors.New("no initialized db in node")
	}
	if node.host.ID() == id {
		return errors.New("trying to ban itself")
	}
	banned, err := node.db.BanscorePeer(id, weight)
	if err == nil {
		if banned {
			// disconnect
			_ = node.DisconnectPeer(id)
		}
	}
	return err
}

func (node *hostNode) IsPeerBanned(id peer.ID) (bool, error) {
	if node.db == nil {
		return false, errors.New("no initialized db in node")
	}
	return node.db.IsPeerBanned(id)
}
