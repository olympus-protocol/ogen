package chainrpc

import (
	"context"

	"github.com/olympus-protocol/ogen/peers"
	"github.com/olympus-protocol/ogen/proto"
)

type networkServer struct {
	host *peers.HostNode
	proto.UnimplementedNetworkServer
}

func (s *networkServer) GetNetworkInfo(context.Context, *proto.Empty) (*proto.NetworkInfo, error) {
	return nil, nil
}
func (s *networkServer) GetPeersInfo(context.Context, *proto.Empty) (*proto.Peers, error) {
	return nil, nil
}
func (s *networkServer) AddPeer(context.Context, *proto.IP) (*proto.Success, error) {
	return nil, nil
}
func (s *networkServer) BanPeer(context.Context, *proto.IP) (*proto.Success, error) {
	return nil, nil
}
