package chainrpc

import (
	"context"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/olympus-protocol/ogen/api/proto"
	"github.com/olympus-protocol/ogen/internal/hostnode"
)

type networkServer struct {
	hostnode hostnode.HostNode
	proto.UnimplementedNetworkServer
}

func (s *networkServer) GetNetworkInfo(context.Context, *proto.Empty) (*proto.NetworkInfo, error) {
	p := s.hostnode.GetPeerList()
	return &proto.NetworkInfo{Peers: int32(len(p)), ID: s.hostnode.GetHost().ID().String()}, nil
}

func (s *networkServer) GetPeersInfo(context.Context, *proto.Empty) (*proto.Peers, error) {
	peersID := s.hostnode.GetPeerList()
	peersInfo := make([]*proto.Peer, len(peersID))
	for i, p := range peersID {
		addr := s.hostnode.GetPeerDirection(p)
		peersInfo[i] = &proto.Peer{Id: p.Pretty(), Host: &proto.IP{Host: addr.String()}}
	}
	return &proto.Peers{Peers: peersInfo}, nil
}
func (s *networkServer) AddPeer(_ context.Context, peerAddr *proto.IP) (*proto.Success, error) {
	maddr, err := multiaddr.NewMultiaddr(peerAddr.Host)
	if err != nil {
		return nil, err
	}
	pinfo, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		return nil, err
	}
	err = s.hostnode.SavePeer(*pinfo)
	if err != nil {
		return &proto.Success{Success: false, Error: err.Error()}, nil
	}
	return &proto.Success{Success: true}, nil
}
