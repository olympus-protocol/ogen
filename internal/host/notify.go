package host

import (
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/multiformats/go-multiaddr"
)

type notify struct {
	h Host
	s *stats
}

func (n notify) Listen(network network.Network, multiaddr multiaddr.Multiaddr) {
	panic("implement me")
}

func (n notify) ListenClose(network network.Network, multiaddr multiaddr.Multiaddr) {
	panic("implement me")
}

func (n notify) Connected(network network.Network, conn network.Conn) {
	panic("implement me")
}

func (n notify) Disconnected(network network.Network, conn network.Conn) {
	panic("implement me")
}

func (n notify) OpenedStream(network network.Network, stream network.Stream) {
	panic("implement me")
}

func (n notify) ClosedStream(network network.Network, stream network.Stream) {
	panic("implement me")
}

func NewNotify(h Host, s *stats) *notify {
	return &notify{
		h: h,
		s: s,
	}
}
