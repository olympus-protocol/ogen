package rpc

import (
	"context"
	"fmt"
	"github.com/olympus-protocol/ogen/api/proto"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"github.com/spf13/viper"
	"io"
	"os"
	"path"
)

// Empty is the empty request.
type Empty struct{}

// CLI is the module that allows operations across multiple services.
type CLI struct {
	rpcClient *RPCClient
}

// Run runs the CLI.
func (c *CLI) Run(optArgs []string) {

	fmt.Println(c.rpcClient.address)

	//Here Runs the RPC
	//check db for tip?

	genesis := primitives.GetGenesisBlock()
	genesisHash := genesis.Hash()

	syncClient, err := c.rpcClient.sync(genesisHash.String())
	if err != nil {
		fmt.Println("unable to initialize sync client")
		os.Exit(0)
	}

	blockCount := 0
	for {
		res, err := syncClient.Recv()
		if err == io.EOF || err != nil {
			fmt.Println(err)
			_ = syncClient.CloseSend()
			break
		}
		fmt.Println(res.Data)
		blockCount++
	}

	fmt.Printf("updated %v blocks", blockCount)
}

func (c *RPCClient) sync(hash string) (proto.Chain_SyncClient, error) {

	syncClient, err := c.chain.Sync(context.Background(), &proto.Hash{Hash: hash})
	if err != nil {
		return nil, err
	}
	return syncClient, err
}

func newCli(rpcClient *RPCClient) *CLI {
	return &CLI{
		rpcClient: rpcClient,
	}
}

func Run(host string, args []string) {
	DataFolder := viper.GetString("datadir")
	if DataFolder != "" {
		// Use config file from the flag.
		viper.AddConfigPath(DataFolder)
		viper.SetConfigName("config")
	} else {
		configDir, err := os.UserConfigDir()
		if err != nil {
			panic(err)
		}

		ogenDir := path.Join(configDir, "ogen")

		if _, err := os.Stat(ogenDir); os.IsNotExist(err) {
			err = os.Mkdir(ogenDir, 0744)
			if err != nil {
				panic(err)
			}
		}

		DataFolder = ogenDir

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(ogenDir)
		viper.SetConfigName("config")
	}
	rpcClient := NewRPCClient(host, DataFolder)
	cli := newCli(rpcClient)
	cli.Run(args)
}
