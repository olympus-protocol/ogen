package indexer

import (
	"context"
	"github.com/olympus-protocol/ogen/internal/indexer/db"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/params"
	"github.com/olympus-protocol/ogen/pkg/rpcclient"
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

}

func (i *Indexer) Stop() {
	i.db.Close()
}

func (i *Indexer) initialSync() {}

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
