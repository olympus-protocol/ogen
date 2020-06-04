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

func (*networkServer) GetNetworkInfo(context.Context, *proto.Empty) (*proto.NetworkInfo, error) {
	return nil, nil
}
func (*networkServer) GetPeersInfo(context.Context, *proto.Empty) (*proto.Peers, error) {
	return nil, nil
}
func (*networkServer) Add(context.Context, *proto.IP) (*proto.Success, error) {
	return nil, nil
}
func (*networkServer) Ban(context.Context, *proto.IP) (*proto.Success, error) {
	return nil, nil
}
