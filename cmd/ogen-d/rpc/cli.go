package rpc

import (
	"context"
	"fmt"
	"github.com/olympus-protocol/ogen/api/proto"
	"github.com/olympus-protocol/ogen/cmd/ogen-d/db"
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
	dbClient  *db.DBClient
}

// Run runs the CLI.
func (c *CLI) Run(optArgs []string) {

	fmt.Println(c.rpcClient.address)

	err := c.dbClient.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("You are Successfully connected!")

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

func newCli(rpcClient *RPCClient, dbClient *db.DBClient) *CLI {
	return &CLI{
		rpcClient: rpcClient,
		dbClient:  dbClient,
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
	dbp := db.DbParameters{
		Hostname:     hostname,
		HostPort:     host_port,
		Username:     username,
		Password:     password,
		DatabaseName: database_name,
		DriverName:   driver_name,
	}
	dbClient := db.NewDBClient(dbp)
	cli := newCli(rpcClient, dbClient)
	cli.Run(args)
}

const (
	hostname      = "localhost"
	host_port     = 5432
	username      = "postgres"
	password      = "testpass"
	database_name = "chaindb"
	driver_name   = "sqlite3"
)
