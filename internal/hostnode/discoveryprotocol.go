package hostnode

import (
	"context"
	discovery "github.com/libp2p/go-libp2p-discovery"
	"github.com/olympus-protocol/ogen/pkg/params"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-kad-dht"
	"github.com/multiformats/go-multiaddr"
	"github.com/olympus-protocol/ogen/pkg/logger"
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
	params *params.ChainParams

	lastConnect     map[peer.ID]time.Time
	lastConnectLock sync.RWMutex

	protocolHandler ProtocolHandler
	dht             *dht.IpfsDHT
	discovery       *discovery.RoutingDiscovery
}

// NewDiscoveryProtocol creates a new discovery service.
func NewDiscoveryProtocol(ctx context.Context, host HostNode, config Config, p *params.ChainParams) (DiscoveryProtocol, error) {
	ph := newProtocolHandler(ctx, params.DiscoveryProtocolID, host, config)
	d, err := dht.New(ctx, host.GetHost(), dht.Mode(dht.ModeAuto))
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
		params:          p,
		lastConnect:     make(map[peer.ID]time.Time),
	}

	host.Notify(dp)

	go dp.findPeers()
	go dp.advertise()

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
		peers, err := cm.discovery.FindPeers(cm.ctx, cm.params.GetRendevouzString())
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

func (cm *discoveryProtocol) advertise() {
	discovery.Advertise(cm.ctx, cm.discovery, cm.params.GetRendevouzString())
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
	s, err := cm.host.GetHost().NewStream(cm.ctx, conn.RemotePeer(), params.DiscoveryProtocolID)
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
