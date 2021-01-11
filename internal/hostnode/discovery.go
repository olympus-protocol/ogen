package hostnode

import (
	"context"
	"github.com/libp2p/go-libp2p-core/network"
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

	d, err := dht.New(ctx, host.GetHost(), dht.Mode(dht.ModeServer))
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

	go dp.initialConnect()
	go dp.advertise()
	go dp.findPeers()

	return dp, nil
}

func (d *discover) initialConnect() {
	peersIDs := d.host.GetHost().Peerstore().Peers()
	peerstorePeers := make([]peer.AddrInfo, len(peersIDs))

	for i := range peersIDs {
		peerstorePeers[i] = d.host.GetHost().Peerstore().PeerInfo(peersIDs[i])
	}

	dhts := dht.GetDefaultBootstrapPeerAddrInfos()
	for _, bp := range dhts {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond * 1000)
		err := d.host.GetHost().Connect(ctx, bp)
		if err != nil {
			d.log.Errorf("unable to connect to relayer %s", bp.ID.String())
		}
		cancel()
	}

	var initialNodes []peer.AddrInfo

	if len(peerstorePeers) < 8 {
		initialNodes = append(initialNodes, peerstorePeers...)
	} else {
		initialNodes = append(initialNodes, peerstorePeers[0:7]...)
	}

	for _, addr := range initialNodes {
		if addr.ID == d.ID {
			continue
		}
		if err := d.Connect(addr); err != nil {
			d.host.GetHost().Peerstore().ClearAddrs(addr.ID)
			d.log.Infof("unable to connect to peer %s error: %s", addr.ID, err.Error())
		}
	}
}

func (d *discover) handleNewPeer(pi peer.AddrInfo) {
	if pi.ID == d.ID {
		return
	}

	ok, err := d.host.StatsService().IsBanned(pi.ID)
	if ok {
		return
	}
	if err != nil {
		d.log.Error(err)
		return
	}

	err = d.Connect(pi)
	if err != nil {
		d.host.GetHost().Peerstore().ClearAddrs(pi.ID)
		d.log.Infof("unable to connect to peer %s error: %s", pi.ID.String(), err.Error())
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

const connectionTimeout = 2000 * time.Millisecond
const connectionWait = 60 * time.Second

func (d *discover) handleStream(s network.Stream) {
	d.log.Infof("handling messages from relayer %s for protocol %s", s.Conn().RemotePeer(), s.Protocol())
}

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
