package host

import (
	"context"
	libhost "github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	discover "github.com/libp2p/go-libp2p-discovery"
	"github.com/multiformats/go-multiaddr"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/params"
	"time"

	"github.com/libp2p/go-libp2p-kad-dht"
)

type discovery struct {
	ctx       context.Context
	netParams *params.ChainParams
	log       logger.Logger

	h         Host
	dht       *dht.IpfsDHT
	discovery *discover.RoutingDiscovery
}

func (d *discovery) findPeers() {
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

func (d *discovery) advertise() {
	discover.Advertise(d.ctx, d.discovery, d.netParams.GetRendevouzString())
}

func (d *discovery) handleNewPeer(pi peer.AddrInfo) {
	if pi.ID == d.h.ID() {
		return
	}

	err := d.h.Connect(pi)
	if err != nil {
		d.log.Infof("unable to connect to peer %s error: %s", pi.ID.String(), err.Error())
	}
}

func (d *discovery) getRelayers() []peer.AddrInfo {
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

func NewDiscovery(ctx context.Context, h Host, lh libhost.Host) (*discovery, error) {
	netParams := config.GlobalParams.NetParams
	log := config.GlobalParams.Logger

	d := &discovery{
		ctx:       ctx,
		netParams: netParams,
		log:       log,
		h:         h,
	}

	dhts := d.getRelayers()

	dh, err := dht.New(ctx, lh, dht.Mode(dht.ModeAutoServer), dht.BootstrapPeers(dhts...))
	if err != nil {
		return nil, err
	}

	err = dh.Bootstrap(ctx)
	if err != nil {
		return nil, err
	}

	r := discover.NewRoutingDiscovery(dh)

	d.dht = dh
	d.discovery = r

	go d.advertise()
	go d.findPeers()

	return d, nil
}
