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

// Client represents an RPC connection to a server.
type Client struct {
	address string
	conn    *grpc.ClientConn

	chain      proto.ChainClient
	validators proto.ValidatorsClient
	utils      proto.UtilsClient
	network    proto.NetworkClient
	wallet     proto.WalletClient
}

func (c *Client) Chain() proto.ChainClient {
	return c.chain
}

func (c *Client) Validators() proto.ValidatorsClient {
	return c.validators
}

func (c *Client) Utils() proto.UtilsClient {
	return c.utils
}

func (c *Client) Network() proto.NetworkClient {
	return c.network
}

func (c *Client) Wallet() proto.WalletClient {
	return c.wallet
}

// NewRPCClient creates a new RPC client.
func NewRPCClient(addr string, datadir string, insecure bool) *Client {
	var creds credentials.TransportCredentials
	if insecure {
		creds = credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: true,
		})
	} else {
		certPool, err := chainrpc.LoadCerts(datadir)
		if err != nil {
			return nil
		}
		creds = credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: false,
			RootCAs:            certPool,
		})
	}

	if addr == "" {
		fmt.Println("Missing address")
		os.Exit(1)
	}
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		panic("unable to connect to rpc server")
	}
	client := &Client{
		address:    addr,
		chain:      proto.NewChainClient(conn),
		validators: proto.NewValidatorsClient(conn),
		utils:      proto.NewUtilsClient(conn),
		network:    proto.NewNetworkClient(conn),
		wallet:     proto.NewWalletClient(conn),
	}
	return client
}
