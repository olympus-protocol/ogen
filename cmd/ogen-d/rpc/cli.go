package rpc

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/olympus-protocol/ogen/api/proto"
	"github.com/olympus-protocol/ogen/cmd/ogen-d/db"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"github.com/spf13/viper"
	"io"
	"os"
	"path"
	"sync"
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

	var wg sync.WaitGroup
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		c.blockSync(wg)
	}(&wg)
	wg.Wait()

}

func (c *CLI) initialSync() {
	// ensure the tables are created for the db
	err := c.dbClient.InitializeTables()
	if err != nil {
		panic(err)
	}

	//get the saved state
	indexState, err := c.dbClient.GetCurrentState()
	if err != nil {
		panic(err)
	}

	var latestBHash string
	if indexState.Blocks == 0 && indexState.LastBlockHash == "" {
		genesis := primitives.GetGenesisBlock()
		genesisHash := genesis.Hash()
		err = c.dbClient.InsertBlock(genesis)
		if err != nil {
			fmt.Println("unable to register genesis block")
			return
		}
		latestBHash = genesisHash.String()
	} else {
		latestBHash = indexState.LastBlockHash
	}
	syncClient, err := c.rpcClient.chain.Sync(context.Background(), &proto.Hash{Hash: latestBHash})
	if err != nil {
		panic("unable to initialize sync client")
	}

	blockCount := 0
	for {
		res, err := syncClient.Recv()
		if err == io.EOF || err != nil {
			fmt.Println(err)
			_ = syncClient.CloseSend()
			break
		}
		blockBytes, err := hex.DecodeString(res.Data)
		if err != nil {
			fmt.Println("unable to parse block")
			break
		}
		var blockOgen primitives.Block
		err = blockOgen.Unmarshal(blockBytes)
		if err != nil {
			fmt.Println("unable to parse block")
			break
		}
		err = c.dbClient.InsertBlock(blockOgen)
		if err != nil {
			fmt.Println("unable to insert")
			break
		} else {
			blockCount++
		}
	}
	fmt.Printf("registered %v blocks", blockCount)
}

func (c *CLI) blockSync(wg *sync.WaitGroup) {
sync:
	// check the connection to the db
	err := c.dbClient.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("You are Successfully connected!")
	c.initialSync()
	subscribe, err := c.rpcClient.chain.SubscribeBlocks(context.Background(), &proto.Empty{})
	if err != nil {
		panic("unable to initialize subscription client")
	}
	wg.Done()
	for {
		res, err := subscribe.Recv()
		if err == io.EOF || err != nil {
			// listener closed restart with sync
			goto sync
		}
		// To make sure the explorer is always synced, every new block we reinsert the last 5
		blockBytes, err := hex.DecodeString(res.Data)
		if err != nil {
			fmt.Println("unable to parse block")
		}
		var blockOgen primitives.Block
		err = blockOgen.Unmarshal(blockBytes)
		if err != nil {
			fmt.Println("unable to unmarshal block")
		}
		err = c.dbClient.InsertBlock(blockOgen)
		if err != nil {
			fmt.Println("unable to insert")
			break
		}
		fmt.Println("received and parsed new block")
	}
}

func (c *CLI) customSync(blockGap int) {

	//get the saved state
	customState, err := c.dbClient.GetSpecificState(blockGap)
	if err != nil {
		panic(err)
	}

	syncClient, err := c.rpcClient.chain.Sync(context.Background(), &proto.Hash{Hash: customState.LastBlockHash})
	if err != nil {
		panic("unable to initialize sync client")
	}

	blockCount := 0
	for {
		res, err := syncClient.Recv()
		if err == io.EOF || err != nil {
			fmt.Println(err)
			_ = syncClient.CloseSend()
			break
		}
		blockBytes, err := hex.DecodeString(res.Data)
		if err != nil {
			fmt.Println("unable to parse block")
			break
		}
		var blockOgen primitives.Block
		err = blockOgen.Unmarshal(blockBytes)
		if err != nil {
			fmt.Println("unable to parse block")
			break
		}
		err = c.dbClient.InsertBlock(blockOgen)
		if err != nil {
			fmt.Println("unable to insert")
			break
		} else {
			blockCount++
		}
	}
	fmt.Printf("registered %v blocks", blockCount)
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
