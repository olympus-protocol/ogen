package host

import (
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/multiformats/go-multiaddr"
)

type notify struct {
	h Host
	s *stats
}

func (n notify) Listen(_ network.Network, _ multiaddr.Multiaddr) {}

func (n notify) ListenClose(_ network.Network, _ multiaddr.Multiaddr) {}

func (n notify) Connected(net network.Network, conn network.Conn) {
	n.h.HandleConnection(net, conn)
}

func (n notify) Disconnected(network network.Network, conn network.Conn) {
	n.h.RemovePeerStats(conn.RemotePeer())
}

func (n notify) OpenedStream(network network.Network, stream network.Stream) {}

func (n notify) ClosedStream(network network.Network, stream network.Stream) {}

func NewNotify(h Host, s *stats) *notify {
	return &notify{
		h: h,
		s: s,
	}
}
