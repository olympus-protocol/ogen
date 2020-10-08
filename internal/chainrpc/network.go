package chainrpc

import (
	"context"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/olympus-protocol/ogen/api/proto"
	"github.com/olympus-protocol/ogen/internal/hostnode"
)

type networkServer struct {
	host hostnode.HostNode
	proto.UnimplementedNetworkServer
}

func (s *networkServer) GetNetworkInfo(ctx context.Context, _ *proto.Empty) (*proto.NetworkInfo, error) {
	defer ctx.Done()

	return &proto.NetworkInfo{Peers: int32(len(s.host.GetHost().Network().Peers())), ID: s.host.GetHost().ID().String()}, nil
}

func (s *networkServer) GetPeersInfo(ctx context.Context, _ *proto.Empty) (*proto.Peers, error) {
	defer ctx.Done()

	peersID := s.host.GetHost().Network().Peers()
	peersInfo := make([]*proto.Peer, len(peersID))
	for i, p := range peersID {
		addr := s.host.GetPeerDirection(p)
		peersInfo[i] = &proto.Peer{Id: p.Pretty(), Host: &proto.IP{Host: addr.String()}}
	}
	return &proto.Peers{Peers: peersInfo}, nil
}

func (s *networkServer) AddPeer(ctx context.Context, peerAddr *proto.IP) (*proto.Success, error) {
	maddr, err := multiaddr.NewMultiaddr(peerAddr.Host)
	if err != nil {
		return nil, err
	}
	pinfo, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		return nil, err
	}
	err = s.host.GetHost().Connect(ctx, *pinfo)
	if err != nil {
		return &proto.Success{Success: false, Error: err.Error()}, nil
	}
	return &proto.Success{Success: true}, nil
}
