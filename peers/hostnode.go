package peers

import (
	"context"
	"crypto/rand"
	"net"
	"path"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/utils/logger"
	"go.etcd.io/bbolt"

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
	Log          *logger.Logger
	Port         string
	AddNodes     []peer.AddrInfo
	MaxPeers     int32
	Path         string
	PrivateKey   crypto.PrivKey
	MaximumPeers int
}

const timeoutInterval = 60 * time.Second
const heartbeatInterval = 20 * time.Second

var configBucketKey = []byte("config")
var privKeyDbKey = []byte("privkey")

// HostNode is the node for p2p host
// It's the low level P2P communication layer, the App class handles high level protocols
// The RPC communication is hanlded by App, not HostNode
type HostNode struct {
	privateKey crypto.PrivKey

	host      host.Host
	gossipSub *pubsub.PubSub
	ctx       context.Context

	topics     map[string]*pubsub.Topic
	topicsLock sync.RWMutex

	timeoutInterval   time.Duration
	heartbeatInterval time.Duration

	netMagic uint32

	log *logger.Logger

	// discoveryProtocol handles peer discovery (mDNS, DHT, etc)
	discoveryProtocol *DiscoveryProtocol

	// syncProtocol handles peer syncing
	syncProtocol *SyncProtocol

	// database buckets
	configBucket *bbolt.Bucket
}

// NewHostNode creates a host node
func NewHostNode(ctx context.Context, config Config, blockchain *chain.Blockchain) (node *HostNode, err error) {
	ps := pstoremem.NewPeerstore()
	netDB, err := bbolt.Open(path.Join(config.Path, "net.db"), 0600, nil)
	if err != nil {
		return nil, err
	}
	var priv crypto.PrivKey
	var configBucket *bbolt.Bucket
	err = netDB.Update(func(tx *bbolt.Tx) error {
		configBucket = tx.Bucket(configBucketKey)
		// If the bucket doesn't exist, initialize the database
		if configBucket == nil {
			configBucket, err = tx.CreateBucketIfNotExists(configBucketKey)
			if err != nil {
				return err
			}

		}
		var keyBytes []byte
		keyBytes = configBucket.Get(privKeyDbKey)
		if keyBytes == nil {
			priv, _, err = crypto.GenerateEd25519Key(rand.Reader)
			if err != nil {
				return err
			}
			privBytes, err := crypto.MarshalPrivateKey(priv)
			if err != nil {
				return err
			}
			err = configBucket.Put(privKeyDbKey, privBytes)
			if err != nil {
				return err
			}
			keyBytes = privBytes
		}
		key, err := crypto.UnmarshalPrivateKey(keyBytes)
		if err != nil {
			return err
		}

		priv = key
		return nil
	})
	if err != nil {
		return nil, err
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

	hostNode := &HostNode{
		privateKey:        config.PrivateKey,
		host:              h,
		gossipSub:         g,
		ctx:               ctx,
		timeoutInterval:   timeoutInterval,
		heartbeatInterval: heartbeatInterval,
		log:               config.Log,
		topics:            map[string]*pubsub.Topic{},
		configBucket:      configBucket,
	}

	discovery, err := NewDiscoveryProtocol(ctx, hostNode, config)
	if err != nil {
		return nil, err
	}
	hostNode.discoveryProtocol = discovery

	syncProtocol, err := NewSyncProtocol(ctx, hostNode, config, blockchain)
	if err != nil {
		return nil, err
	}
	hostNode.syncProtocol = syncProtocol

	return hostNode, nil
}

func (node *HostNode) SyncProtocol() *SyncProtocol {
	return node.syncProtocol
}

func (node *HostNode) Topic(topic string) (*pubsub.Topic, error) {
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
func (node *HostNode) Syncing() bool {
	return node.syncProtocol.syncInfo.syncing
}

// GetContext returns the context
func (node *HostNode) GetContext() context.Context {
	return node.ctx
}

// GetHost returns the host
func (node *HostNode) GetHost() host.Host {
	return node.host
}

// SubscribeMessage registers a handler for a network topic.
func (node *HostNode) SubscribeMessage(topic string, handler func([]byte, peer.ID)) (*pubsub.Subscription, error) {
	subscription, err := node.gossipSub.Subscribe(topic)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			msg, err := subscription.Next(node.ctx)
			node.log.Debugf("received broadcast of size %d on topic %s", len(msg.Data), topic)
			if err != nil {
				node.log.Errorf("error when receiving message: %s", err)
				continue
			}

			handler(msg.Data, msg.GetFrom())
		}
	}()

	return subscription, nil
}

// UnsubscribeMessage cancels a subscription to a topic.
func (node *HostNode) UnsubscribeMessage(subscription *pubsub.Subscription) {
	subscription.Cancel()
}

func (node *HostNode) removePeer(p peer.ID) {
	node.host.Peerstore().ClearAddrs(p)
}

// DisconnectPeer disconnects a peer
func (node *HostNode) DisconnectPeer(p peer.ID) error {
	return node.host.Network().ClosePeer(p)
}

// IsConnected checks if the host node is connected.
func (node *HostNode) IsConnected() bool {
	return node.PeersConnected() > 0
}

// PeersConnected checks how many peers are connected.
func (node *HostNode) PeersConnected() int {
	return len(node.host.Network().Peers())
}

// GetPeerList returns a list of all peers.
func (node *HostNode) GetPeerList() []peer.ID {
	return node.host.Network().Peers()
}

// GetPeerInfos gets peer infos of connected peers.
func (node *HostNode) GetPeerInfos() []peer.AddrInfo {
	peers := node.host.Network().Peers()
	infos := make([]peer.AddrInfo, 0, len(peers))
	for _, p := range peers {
		addrInfo := node.host.Peerstore().PeerInfo(p)
		infos = append(infos, addrInfo)
	}

	return infos
}

// ConnectedToPeer returns true if we're connected to the peer.
func (node *HostNode) ConnectedToPeer(id peer.ID) bool {
	connectedness := node.host.Network().Connectedness(id)
	return connectedness == network.Connected
}

// Notify notifies a notifee for network events.
func (node *HostNode) Notify(notifee network.Notifiee) {
	node.host.Network().Notify(notifee)
}

// setStreamHandler sets a stream handler for the host node.
func (node *HostNode) setStreamHandler(id protocol.ID, handleStream func(s network.Stream)) {
	node.host.SetStreamHandler(id, handleStream)
}

// CountPeers counts the number of peers that support the protocol.
func (node *HostNode) CountPeers(id protocol.ID) int {
	count := 0
	for _, n := range node.host.Peerstore().Peers() {
		if sup, err := node.host.Peerstore().SupportsProtocols(n, string(id)); err != nil && len(sup) != 0 {
			count++
		}
	}
	return count
}

// GetPeerDirection gets the direction of the peer.
func (node *HostNode) GetPeerDirection(id peer.ID) network.Direction {
	conns := node.host.Network().ConnsToPeer(id)

	if len(conns) != 1 {
		return network.DirUnknown
	}
	return conns[0].Stat().Direction
}

// Start the host node and start discovering peers.
func (node *HostNode) Start() error {
	if err := node.discoveryProtocol.Start(); err != nil {
		return err
	}

	return nil
}
