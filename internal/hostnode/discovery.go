package hostnode

import (
	"context"
	"github.com/dgraph-io/ristretto"
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

var maxPeers = int64(100000)

var BadPeersCache, _ = ristretto.NewCache(&ristretto.Config{
	NumCounters: maxPeers,
	MaxCost:     1 << 22, // ~4mb is cache max size
	BufferItems: 64,
})

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

	peersIDs := dp.host.GetHost().Peerstore().Peers()
	peerstorePeers := make([]peer.AddrInfo, len(peersIDs))

	for i := range peersIDs {
		peerstorePeers[i] = dp.host.GetHost().Peerstore().PeerInfo(peersIDs[i])
	}

	var initialNodes []peer.AddrInfo
	initialNodes = append(initialNodes, dp.getRelayers()...)

	if len(peerstorePeers) < 8 {
		initialNodes = append(initialNodes, peerstorePeers...)
	} else {
		initialNodes = append(initialNodes, peerstorePeers[0:7]...)
	}

	for _, addr := range initialNodes {
		if err := dp.host.GetHost().Connect(dp.ctx, addr); err != nil {
			m, err := addr.MarshalJSON()
			if err != nil {
				continue
			}
			BadPeersCache.Set(addr.ID.String(), m, int64(len(m)))
			dp.host.GetHost().Peerstore().ClearAddrs(addr.ID)
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
	if _, ok := BadPeersCache.Get(pi.ID.String()); ok {
		return
	}
	err := d.Connect(pi)
	if err != nil {
		m, err := pi.MarshalJSON()
		if err != nil {
			return
		}
		BadPeersCache.Set(pi.ID.String(), m, int64(len(m)))
		d.host.GetHost().Peerstore().ClearAddrs(pi.ID)
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

const connectionTimeout = 1500 * time.Millisecond
const connectionWait = 10 * time.Second

// Connect connects to a peer.
func (d *discover) Connect(pi peer.AddrInfo) error {
	d.lastConnectLock.Lock()
	defer d.lastConnectLock.Unlock()
	lastConnect, found := d.lastConnect[pi.ID]
	if !found || time.Since(lastConnect) > connectionWait {
		d.lastConnect[pi.ID] = time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
		defer cancel()
		return d.host.GetHost().Connect(ctx, pi)
	}
	return nil
}
