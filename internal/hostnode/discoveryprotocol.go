package hostnode

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p/p2p/discovery"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-kad-dht"
	"github.com/multiformats/go-multiaddr"
	"github.com/olympus-protocol/ogen/internal/logger"
)

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
}

// NewDiscoveryProtocol creates a new discovery service.
func NewDiscoveryProtocol(ctx context.Context, host HostNode, config Config) (DiscoveryProtocol, error) {
	ph := newProtocolHandler(ctx, discoveryProtocolID, host, config)
	d, err := dht.New(ctx, host.GetHost())
	if err != nil {
		return nil, err
	}

	err = d.Bootstrap(ctx)
	if err != nil {
		return nil, err
	}


	dp := &discoveryProtocol{
		host:            host,
		ctx:             ctx,
		config:          config,
		protocolHandler: ph,
		log:             config.Log,
		dht:             d,
	}

	host.Notify(dp)

	go dp.findPeers()
	//go dp.advertise()

	return dp, nil
}

func (cm *discoveryProtocol) handleNewPeer(pi peer.AddrInfo) {
	if pi.ID == cm.host.GetHost().ID() {
		return
	}
	err := cm.Connect(pi)
	if err != nil {
		cm.log.Error("unable to connect to peer %s", pi.ID.String())
	}
}

func (cm *discoveryProtocol) findPeers() {
	for {
		peerChan := initMDNS(cm.ctx, cm.host.GetHost(), "rendevouz")
	peerLoop:
		for {
			select {
			case pi, ok := <-peerChan:
				fmt.Println(pi, ok)
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

//func (cm *discoveryProtocol) advertise() {
//	discovery.Advertise(cm.ctx, cm.discovery, "randezvous")
//}

func (cm *discoveryProtocol) Start() error {
	for _, addr := range cm.config.InitialNodes {
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


type discoveryNotifee struct {
	PeerChan chan peer.AddrInfo
}

//interface to be called when new  peer is found
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	n.PeerChan <- pi
}

//Initialize the MDNS service
func initMDNS(ctx context.Context, peerhost host.Host, rendezvous string) chan peer.AddrInfo {
	// An hour might be a long long period in practical applications. But this is fine for us
	ser, err := discovery.NewMdnsService(ctx, peerhost, time.Hour, rendezvous)
	if err != nil {
		panic(err)
	}

	//register with service so that we get notified about peer discovery
	n := &discoveryNotifee{}
	n.PeerChan = make(chan peer.AddrInfo)

	ser.RegisterNotifee(n)
	return n.PeerChan
}