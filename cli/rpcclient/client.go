package rpcclient

import (
	"github.com/olympus-protocol/ogen/chainrpc/proto"
	"google.golang.org/grpc"
)

// RPCClient represents an RPC connection to a server.
type RPCClient struct {
	address string
	conn    *grpc.ClientConn

	chain      proto.ChainClient
	validators proto.ValidatorsClient
}

// NewRPCClient creates a new RPC client.
func NewRPCClient(addr string) *RPCClient {
	conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		panic("unable to connect to rpc server")
	}
	client := &RPCClient{
		address:    addr,
		chain:      proto.NewChainClient(conn),
		validators: proto.NewValidatorsClient(conn),
	}
	return client
}
