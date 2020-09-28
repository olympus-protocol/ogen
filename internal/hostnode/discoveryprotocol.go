package hostnode

import (
	"context"
	discovery "github.com/libp2p/go-libp2p-discovery"
	"github.com/olympus-protocol/ogen/pkg/params"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-kad-dht"
	"github.com/multiformats/go-multiaddr"
	"github.com/olympus-protocol/ogen/pkg/logger"
)

func (d *discoveryProtocol) getRelayers() []peer.AddrInfo {
	var r []peer.AddrInfo
	for _, node := range d.params.Relayers {
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

	ID              peer.ID
	protocolHandler ProtocolHandler
	dht             *dht.IpfsDHT
	discovery       *discovery.RoutingDiscovery
}

// NewDiscoveryProtocol creates a new discovery service.
func NewDiscoveryProtocol(ctx context.Context, host HostNode, config Config, p *params.ChainParams) (DiscoveryProtocol, error) {
	d, err := dht.New(ctx, host.GetHost(), dht.Mode(dht.ModeAutoServer))
	if err != nil {
		return nil, err
	}

	err = d.Bootstrap(ctx)
	if err != nil {
		return nil, err
	}

	r := discovery.NewRoutingDiscovery(d)

	dp := &discoveryProtocol{
		host:        host,
		ctx:         ctx,
		config:      config,
		log:         config.Log,
		dht:         d,
		discovery:   r,
		params:      p,
		ID:          host.GetHost().ID(),
		lastConnect: make(map[peer.ID]time.Time),
	}

	go dp.findPeers()
	go dp.advertise()

	return dp, nil
}

func (d *discoveryProtocol) handleNewPeer(pi peer.AddrInfo) {
	if pi.ID == d.ID {
		return
	}
	if d.host.ConnectedToPeer(pi.ID) {
		return
	}
	err := d.Connect(pi)
	if err != nil {
		d.log.Errorf("unable to connect to peer %s: %s", pi.ID.String(), err.Error())
	}
}

func (d *discoveryProtocol) findPeers() {
	for {
		peers, err := d.discovery.FindPeers(d.ctx, d.params.GetRendevouzString())
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

				d.handleNewPeer(pi)
			case <-d.ctx.Done():
				return
			}
		}
	}
}

func (d *discoveryProtocol) advertise() {
	discovery.Advertise(d.ctx, d.discovery, d.params.GetRendevouzString())
}

func (d *discoveryProtocol) Start() error {
	peersIDs := d.host.GetHost().Peerstore().Peers()
	var peerstorePeers []peer.AddrInfo
	for _, id := range peersIDs {
		peerstorePeers = append(peerstorePeers, d.host.GetHost().Peerstore().PeerInfo(id))
	}
	var initialNodes []peer.AddrInfo
	initialNodes = append(initialNodes, d.getRelayers()...)
	initialNodes = append(initialNodes, peerstorePeers...)
	for _, addr := range initialNodes {
		if err := d.host.GetHost().Connect(d.ctx, addr); err != nil {
			d.log.Error(err)
		}
	}
	return nil
}

const connectionTimeout = 10 * time.Second
const connectionCooldown = 60 * time.Second

// Connect connects to a peer.
func (d *discoveryProtocol) Connect(pi peer.AddrInfo) error {
	d.lastConnectLock.Lock()
	defer d.lastConnectLock.Unlock()
	lastConnect, found := d.lastConnect[pi.ID]
	if !found || time.Since(lastConnect) > connectionCooldown {
		d.lastConnect[pi.ID] = time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
		defer cancel()
		return d.host.GetHost().Connect(ctx, pi)
	}
	return nil
}
