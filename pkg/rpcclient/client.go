package rpcclient

import (
	"crypto/tls"
	"fmt"
	"os"

	"github.com/olympus-protocol/ogen/api/proto"
	"github.com/olympus-protocol/ogen/internal/chainrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// RPCClient represents an RPC connection to a server.
type RPCClient struct {
	address string
	conn    *grpc.ClientConn

	chain      proto.ChainClient
	validators proto.ValidatorsClient
	utils      proto.UtilsClient
	network    proto.NetworkClient
	wallet     proto.WalletClient
}

// NewRPCClient creates a new RPC client.
func NewRPCClient(addr string, datadir string) *RPCClient {
	certPool, err := chainrpc.LoadCerts(datadir)
	if err != nil {
		return nil
	}
	creds := credentials.NewTLS(&tls.Config{
		InsecureSkipVerify: false,
		RootCAs:            certPool,
	})
	if addr == "" {
		fmt.Println("Missing address")
		os.Exit(1)
	}
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		panic("unable to connect to rpc server")
	}
	client := &RPCClient{
		address:    addr,
		chain:      proto.NewChainClient(conn),
		validators: proto.NewValidatorsClient(conn),
		utils:      proto.NewUtilsClient(conn),
		network:    proto.NewNetworkClient(conn),
		wallet:     proto.NewWalletClient(conn),
	}
	return client
}
