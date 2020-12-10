package indexer

import (
	"context"
	"encoding/hex"
	"github.com/olympus-protocol/ogen/api/proto"
	"github.com/olympus-protocol/ogen/internal/indexer/db"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/params"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"github.com/olympus-protocol/ogen/pkg/rpcclient"
	"io"
	"os"
	"sync"
)

// Indexer is the module that allows operations across multiple services.
type Indexer struct {
	log logger.Logger
	ctx context.Context

	client    *rpcclient.Client
	db        *db.Database
	canClose  *sync.WaitGroup
	netParams *params.ChainParams
}

func (i *Indexer) Start() {
	i.initialSync()
	i.log.Info("Listening for new blocks")
	go i.subscribeBlocks()
}

func (i *Indexer) Stop() {
	i.db.Close()
}

func (i *Indexer) initialSync() {
	// Get the saved state

	indexState, err := i.db.GetState()
	if err != nil {
		i.log.Fatal(err)
	}

	var latestBHash string
	if indexState.Blocks == 0 && indexState.LastBlockHash == "" {
		// Initialize the database with initialization parameters
		latestBHash, err = i.db.Initialize()
		if err != nil {
			i.log.Fatal("unable to initialize database")
			return
		}
	} else {
		latestBHash = indexState.LastBlockHash
	}

	i.log.Infof("Starting initial sync...")
	syncClient, err := i.client.Chain().Sync(context.Background(), &proto.Hash{Hash: latestBHash})
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
		block := new(primitives.Block)
		err = block.Unmarshal(blockBytes)
		if err != nil {
			i.log.Error("unable to parse block")
			break
		}
		err = i.db.InsertBlock(block)
		if err != nil {
			i.log.Error("unable to insert block")
			break
		} else {
			blockCount++
		}
	}
	i.log.Infof("Initial sync finished, parsed %d blocks", blockCount)
}

func (i *Indexer) subscribeBlocks() {
}

func (i *Indexer) Context() context.Context {
	return i.ctx
}

func NewIndexer(dbConnString, rpcEndpoint string, netParams *params.ChainParams) (*Indexer, error) {
	log := logger.New(os.Stdin)

	rpcClient := rpcclient.NewRPCClient(rpcEndpoint, true)
	var wg sync.WaitGroup
	database := db.NewDB(dbConnString, log, &wg, netParams)

	err := database.Migrate()
	if err != nil {
		return nil, err
	}

	indexer := &Indexer{
		log:       log,
		ctx:       context.Background(),
		client:    rpcClient,
		db:        database,
		canClose:  &wg,
		netParams: netParams,
	}

	return indexer, nil
}
