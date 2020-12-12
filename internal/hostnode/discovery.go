package hostnode

import (
	"context"
	discovery "github.com/libp2p/go-libp2p-discovery"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/pkg/params"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-kad-dht"
	"github.com/multiformats/go-multiaddr"
	"github.com/olympus-protocol/ogen/pkg/logger"
)

func (d *discover) getRelayers() []peer.AddrInfo {
	var r []peer.AddrInfo
	for _, node := range d.netParams.Relayers {
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

// discover is the routine to announce and discover new peers for the hostnode.
type discover struct {
	host      HostNode
	ctx       context.Context
	log       logger.Logger
	netParams *params.ChainParams

	lastConnect     map[peer.ID]time.Time
	lastConnectLock sync.Mutex

	ID        peer.ID
	dht       *dht.IpfsDHT
	discovery *discovery.RoutingDiscovery
}

// NewDiscover creates a new discovery service.
func NewDiscover(host HostNode) (*discover, error) {
	ctx := config.GlobalParams.Context
	log := config.GlobalParams.Logger
	netParams := config.GlobalParams.NetParams

	d, err := dht.New(ctx, host.GetHost(), dht.Mode(dht.ModeAutoServer))
	if err != nil {
		return nil, err
	}

	err = d.Bootstrap(ctx)
	if err != nil {
		return nil, err
	}

	r := discovery.NewRoutingDiscovery(d)

	dp := &discover{
		host:        host,
		ctx:         ctx,
		log:         log,
		dht:         d,
		discovery:   r,
		netParams:   netParams,
		ID:          host.GetHost().ID(),
		lastConnect: make(map[peer.ID]time.Time),
	}

	var initialNodes []peer.AddrInfo
	peersIDs := dp.host.GetHost().Peerstore().Peers()
	var peerstorePeers []peer.AddrInfo
	for _, id := range peersIDs {
		peerstorePeers = append(peerstorePeers, dp.host.GetHost().Peerstore().PeerInfo(id))
	}
	initialNodes = append(initialNodes, dp.getRelayers()...)
	initialNodes = append(initialNodes, peerstorePeers...)
	for _, addr := range initialNodes {
		if err := dp.host.GetHost().Connect(dp.ctx, addr); err != nil {
			dp.log.Infof("unable to connect to peer %s", addr.ID)
		}
	}

	go dp.advertise()
	go dp.findPeers()

	return dp, nil
}

func (d *discover) handleNewPeer(pi peer.AddrInfo) {
	if pi.ID == d.ID {
		return
	}
	err := d.Connect(pi)
	if err != nil {
		d.log.Infof("unable to connect to peer %s", pi.ID.String())
	}
}

func (d *discover) findPeers() {
	for {
		peers, err := d.discovery.FindPeers(d.ctx, d.netParams.GetRendevouzString())
		if err != nil {
			break
		}
	peerLoop:
		for {
			select {
			case pi, ok := <-peers:
				if !ok {
					time.Sleep(time.Second * 5)
					break peerLoop
				}

				d.handleNewPeer(pi)
			case <-d.ctx.Done():
				return
			}
		}
	}
}

func (d *discover) advertise() {
	discovery.Advertise(d.ctx, d.discovery, d.netParams.GetRendevouzString())
}

const connectionTimeout = 3 * time.Second
const connectionCooldown = 30 * time.Second

// Connect connects to a peer.
func (d *discover) Connect(pi peer.AddrInfo) error {
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
