package commands

import (
	"context"
	"encoding/hex"
	"github.com/olympus-protocol/ogen/api/proto"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/cmd/ogen/indexer"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"github.com/olympus-protocol/ogen/pkg/rpcclient"
	"github.com/spf13/cobra"
	"io"
	"os"
)

// TODO provide a dynanmic way for the user to load the info

const (
	hostname = "localhost"
	hostport = 5432
	username = "postgres"
	password = "testpass"
	dbname   = "olympus_data"
	driver   = "sqlite3"
)

func init() {
	indexerCmd.Flags().StringVar(&rpcHost, "rpc_host", "127.0.0.1:24127", "IP and port of the RPC Server to connect")

	rootCmd.AddCommand(indexerCmd)
}

// Indexer is the module that allows operations across multiple services.
type Indexer struct {
	log       logger.Logger
	ctx       context.Context
	rpcClient *rpcclient.Client
	dbClient  *indexer.Database
	path      string
}

func (i *Indexer) Run() {
	i.blockSync()
}

func (i *Indexer) blockSync() {
sync:
	i.initialSync()
	subscribe, err := i.rpcClient.Chain().SubscribeBlocks(context.Background(), &proto.Empty{})
	if err != nil {
		panic("unable to initialize subscription client")
	}
	for {
		res, err := subscribe.Recv()
		if err == io.EOF || err != nil {
			// listener closed restart with sync
			goto sync
		}
		// To make sure the explorer is always synced, every new block we reinsert the last 5
		blockBytes, err := hex.DecodeString(res.Data)
		if err != nil {
			i.log.Error("unable to parse block")
			continue
		}
		var blockOgen primitives.Block
		err = blockOgen.Unmarshal(blockBytes)
		if err != nil {
			i.log.Error("unable to parse block")
			continue
		}
		err = i.dbClient.InsertBlock(blockOgen)
		if err != nil {
			i.log.Error("unable to insert new block")
			continue
		}
		i.log.Info("received and parsed new block")
	}
}

func (i *Indexer) initialSync() {

	// ensure the tables are created for the db
	err := i.dbClient.InitializeTables()
	if err != nil {
		panic(err)
	}

	// get the saved state
	indexState, err := i.dbClient.GetCurrentState()
	if err != nil {
		panic(err)
	}

	var latestBHash string
	if indexState.Blocks == 0 && indexState.LastBlockHash == "" {
		genesis := primitives.GetGenesisBlock()
		genesisHash := genesis.Hash()
		err = i.dbClient.InsertBlock(genesis)
		if err != nil {
			i.log.Error("unable to register genesis block")
			return
		}
		latestBHash = genesisHash.String()
	} else {
		latestBHash = indexState.LastBlockHash
	}

	syncClient, err := i.rpcClient.Chain().Sync(context.Background(), &proto.Hash{Hash: latestBHash})
	if err != nil {
		i.log.Fatal("unable to initialize initial sync")
		return
	}

	blockCount := 0
	for {
		res, err := syncClient.Recv()
		if err != nil {
			if err == io.EOF {
				_ = syncClient.CloseSend()
				break
			}
			i.log.Error(err)
			break
		}
		blockBytes, err := hex.DecodeString(res.Data)
		if err != nil {
			i.log.Error("unable to parse block")
			break
		}
		var blockOgen primitives.Block
		err = blockOgen.Unmarshal(blockBytes)
		if err != nil {
			i.log.Error("unable to parse block")
			break
		}
		err = i.dbClient.InsertBlock(blockOgen)
		if err != nil {
			i.log.Error("unable to insert block")
			break
		} else {
			blockCount++
		}
	}
	i.log.Infof("registered %v blocks", blockCount)
}

func (i *Indexer) customSync(blockGap int) {

	//get the saved state
	customState, err := i.dbClient.GetSpecificState(blockGap)
	if err != nil {
		panic(err)
	}

	syncClient, err := i.rpcClient.Chain().Sync(context.Background(), &proto.Hash{Hash: customState.LastBlockHash})
	if err != nil {
		panic("unable to initialize sync client")
	}

	blockCount := 0
	for {
		res, err := syncClient.Recv()
		if err != nil {
			if err == io.EOF {
				_ = syncClient.CloseSend()
				break
			}
			i.log.Error(err)
			break
		}
		blockBytes, err := hex.DecodeString(res.Data)
		if err != nil {
			i.log.Error("unable to parse block")
			break
		}
		var blockOgen primitives.Block
		err = blockOgen.Unmarshal(blockBytes)
		if err != nil {
			i.log.Error("unable to parse block")
			break
		}
		err = i.dbClient.InsertBlock(blockOgen)
		if err != nil {
			i.log.Error("unable to insert")
			break
		} else {
			blockCount++
		}
	}
	i.log.Infof("registered %v blocks", blockCount)
}

var indexerCmd = &cobra.Command{
	Use:   "indexer",
	Short: "Execute the and indexer to organize the blockchain information through RPC",
	Long:  `Execute the and indexer to organize the blockchain information through RPC`,
	Run: func(cmd *cobra.Command, args []string) {
		log := logger.New(os.Stdin)

		rpcClient := rpcclient.NewRPCClient(rpcHost, config.DataPath, true)

		dbp := &indexer.Config{
			Hostname:     hostname,
			HostPort:     hostport,
			Username:     username,
			Password:     password,
			DatabaseName: dbname,
			DriverName:   driver,
		}
		dbClient := indexer.NewDBClient(dbp, config.DataPath, log)

		indexer := Indexer{
			log:       log,
			ctx:       context.Background(),
			rpcClient: rpcClient,
			dbClient:  dbClient,
		}

		go indexer.Run()
		<-indexer.ctx.Done()
	},
}
