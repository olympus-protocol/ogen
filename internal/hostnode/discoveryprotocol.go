package hostnode

import (
	"context"
	discovery "github.com/libp2p/go-libp2p-discovery"
	"math"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-kad-dht"
	"github.com/multiformats/go-multiaddr"
	"github.com/olympus-protocol/ogen/internal/logger"
)

var relayerNodes = map[string]string{
	"cronos-1": "/ip4/206.189.231.51/tcp/25000/p2p/12D3KooWJiD1mSdJTYxoTwRrmG2D2zPnzpHe6vpS5T5FcX3J7HCM",
	"cronos-2": "/ip4/104.248.120.150/tcp/25000/p2p/12D3KooWCu1XLbzDN6TASFpvo4QMtHmLPG652VwEq11bfWGj8Tag",
}

func getRelayers() []peer.AddrInfo {
	var r []peer.AddrInfo
	for _, node := range relayerNodes {
		ma, err := multiaddr.NewMultiaddr(node)
		if err != nil {
			continue
		}
		addr, err := peer.AddrInfoFromP2pAddr(ma)
		if err != nil {
			continue
		}
		r = append(r, *addr)
	}
	return r
}

var rendevouzString = map[int]string{
	0: "do_not_go_gentle_into_that_good_night",
}

// GetRendevouzString is a function to return a rendevouz string for a certain version range
// to make sure peers find each other depending on their version.
func GetRendevouzString() string {
	ver := VersionNumber
	var selectedIndex int
	var diffSelected int
	for n := range rendevouzString {
		diff := int(math.Abs(float64(ver - n)))
		if diff < diffSelected {
			selectedIndex = n
			diffSelected = diff
		}
	}
	return rendevouzString[selectedIndex]
}

// DiscoveryProtocol is an interface for discoveryProtocol
type DiscoveryProtocol interface {
	Start() error
	Listen(network.Network, multiaddr.Multiaddr)
	ListenClose(network.Network, multiaddr.Multiaddr)
	Connected(net network.Network, conn network.Conn)
	Disconnected(net network.Network, conn network.Conn)
	OpenedStream(network.Network, network.Stream)
	ClosedStream(network.Network, network.Stream)
}

var _ DiscoveryProtocol = &discoveryProtocol{}

// discoveryProtocol is the service to discover other hostnode.
type discoveryProtocol struct {
	host   HostNode
	config Config
	ctx    context.Context
	log    logger.Logger

	lastConnect     map[peer.ID]time.Time
	lastConnectLock sync.RWMutex

	protocolHandler ProtocolHandler
	dht             *dht.IpfsDHT
	discovery       *discovery.RoutingDiscovery
}

// NewDiscoveryProtocol creates a new discovery service.
func NewDiscoveryProtocol(ctx context.Context, host HostNode, config Config, relayer bool) (DiscoveryProtocol, error) {
	ph := newProtocolHandler(ctx, discoveryProtocolID, host, config)
	d, err := dht.New(ctx, host.GetHost(), dht.Mode(dht.ModeServer))
	if err != nil {
		return nil, err
	}

	err = d.Bootstrap(ctx)
	if err != nil {
		return nil, err
	}

	r := discovery.NewRoutingDiscovery(d)

	dp := &discoveryProtocol{
		host:            host,
		ctx:             ctx,
		config:          config,
		protocolHandler: ph,
		log:             config.Log,
		dht:             d,
		discovery:       r,
		lastConnect:     make(map[peer.ID]time.Time),
	}

	host.Notify(dp)

	if relayer {
		go dp.findPeers()
	} else {
		go dp.findPeersAsRelayer()
	}

	go dp.advertise(relayer)

	return dp, nil
}

func (cm *discoveryProtocol) handleNewPeer(pi peer.AddrInfo) {
	if pi.ID == cm.host.GetHost().ID() {
		return
	}
	if cm.host.ConnectedToPeer(pi.ID) {
		return
	}
	err := cm.Connect(pi)
	if err != nil {
		cm.log.Errorf("unable to connect to peer %s: %s", pi.ID.String(), err.Error())
	}
}

func (cm *discoveryProtocol) findPeers() {
	for {
		peers, err := cm.discovery.FindPeers(cm.ctx, GetRendevouzString())
		if err != nil {
			break
		}
	peerLoop:
		for {
			select {
			case pi, ok := <-peers:
				if !ok {
					time.Sleep(time.Second * 10)
					break peerLoop
				}
				cm.handleNewPeer(pi)
			case <-cm.ctx.Done():
				return
			}
		}
	}
}

func (cm *discoveryProtocol) findPeersAsRelayer() {
	for _, s := range rendevouzString {
		go func(s string) {
			for {
				peers, err := cm.discovery.FindPeers(cm.ctx, s)
				if err != nil {
					break
				}
			peerLoop:
				for {
					select {
					case pi, ok := <-peers:
						if !ok {
							time.Sleep(time.Second * 10)
							break peerLoop
						}
						cm.handleNewPeer(pi)
					case <-cm.ctx.Done():
						return
					}
				}
			}
		}(s)
	}
}

func (cm *discoveryProtocol) advertise(relayer bool) {
	// When relayer mode is active we advertise all versions
	if relayer {

	} else {

	}
	discovery.Advertise(cm.ctx, cm.discovery, GetRendevouzString())
}

func (cm *discoveryProtocol) Start() error {
	peersIDs := cm.host.GetHost().Peerstore().Peers()
	var peerstorePeers []peer.AddrInfo
	for _, id := range peersIDs {
		peerstorePeers = append(peerstorePeers, cm.host.GetHost().Peerstore().PeerInfo(id))
	}
	var initialNodes []peer.AddrInfo
	initialNodes = append(initialNodes, getRelayers()...)
	initialNodes = append(initialNodes, peerstorePeers...)
	for _, addr := range initialNodes {
		if err := cm.host.GetHost().Connect(cm.ctx, addr); err != nil {
			cm.log.Error(err)
		}
	}
	return nil
}

const connectionTimeout = 10 * time.Second
const connectionCooldown = 60 * time.Second

// Connect connects to a peer.
func (cm *discoveryProtocol) Connect(pi peer.AddrInfo) error {
	cm.lastConnectLock.Lock()
	defer cm.lastConnectLock.Unlock()
	lastConnect, found := cm.lastConnect[pi.ID]
	if !found || time.Since(lastConnect) > connectionCooldown {
		cm.lastConnect[pi.ID] = time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
		defer cancel()
		return cm.host.GetHost().Connect(ctx, pi)
	}
	return nil
}

// Listen is called when we start listening on a multiaddr.
func (cm *discoveryProtocol) Listen(network.Network, multiaddr.Multiaddr) {}

// ListenClose is called when we stop listening on a multiaddr.
func (cm *discoveryProtocol) ListenClose(network.Network, multiaddr.Multiaddr) {}

// Connected is called when we connect to a peer.
func (cm *discoveryProtocol) Connected(_ network.Network, conn network.Conn) {
	if conn.Stat().Direction != network.DirOutbound {
		return
	}

	// open a stream for the discovery protocol:
	s, err := cm.host.GetHost().NewStream(cm.ctx, conn.RemotePeer(), discoveryProtocolID)
	if err != nil {
		cm.log.Errorf("could not open stream for connection: %s", err)
	}

	cm.protocolHandler.HandleStream(s)
}

// Disconnected is called when we disconnect from a peer.
func (cm *discoveryProtocol) Disconnected(network.Network, network.Conn) {}

// OpenedStream is called when we open a stream.
func (cm *discoveryProtocol) OpenedStream(network.Network, network.Stream) {}

// ClosedStream is called when we close a stream.
func (cm *discoveryProtocol) ClosedStream(network.Network, network.Stream) {}
