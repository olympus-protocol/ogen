package indexer

import (
	"context"
	"encoding/hex"
	"github.com/olympus-protocol/ogen/api/proto"
	"github.com/olympus-protocol/ogen/cmd/ogen/initialization"
	"github.com/olympus-protocol/ogen/internal/indexer/db"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/params"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"github.com/olympus-protocol/ogen/pkg/rpcclient"
	"io"
	"os"
	"sync"
	"time"
)

// Indexer is the module that allows operations across multiple services.
type Indexer struct {
	log logger.Logger
	ctx context.Context

	client    *rpcclient.Client
	db        *db.Database
	canClose  *sync.WaitGroup
	netParams *params.ChainParams
	state     state.State
}

func (i *Indexer) ProcessBlock(b *primitives.Block) error {
	err := i.state.ProcessBlock(b)
	if err != nil {
		return err
	}
	return nil
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
	genesis := primitives.GetGenesisBlock()
	genesisHash := genesis.Hash()

	i.log.Infof("Starting initial sync")
initSync:
	time.Sleep(5 * time.Second)
	syncClient, err := i.client.Chain().Sync(context.Background(), &proto.Hash{Hash: genesisHash.String()})
	if err != nil {
		i.log.Warn("Unable to connect to RPC server. Trying again...")
		goto initSync
	}
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
	}
	i.log.Infof("Initial sync finished")
}

func (i *Indexer) subscribeBlocks() {
	subscribe, err := i.client.Chain().SubscribeBlocks(context.Background(), &proto.Empty{})
	if err != nil {
		panic("unable to initialize subscription client")
	}
	for {
		select {
		case <-i.ctx.Done():
			_ = subscribe.CloseSend()
			break
		default:
			res, err := subscribe.Recv()
			if err == io.EOF || err != nil {
				// listener closed restart with sync
				i.initialSync()
				continue
			}
			// To make sure the explorer is always synced, every new block we reinsert the last 5
			blockBytes, err := hex.DecodeString(res.Data)
			if err != nil {
				i.log.Errorf("unable to parse error %s", err.Error())
				continue
			}
			block := new(primitives.Block)
			err = block.Unmarshal(blockBytes)
			if err != nil {
				i.log.Errorf("unable to parse error %s", err.Error())
				continue
			}
			//err = i.db.InsertBlock(block)
			//if err != nil {
			//	if err == db.ErrorPrevBlockHash {
			//		i.log.Error(db.ErrorPrevBlockHash)
			//		i.log.Info("Restarting sync...")
			//		i.initialSync()
			//		continue
			//	}
			//	i.log.Errorf("unable to insert error %s", err.Error())
			//	continue
			//}
			//i.log.Infof("Received new block %s", block.Hash().String())
		}
	}
}

func (i *Indexer) Context() context.Context {
	return i.ctx
}

func (i *Indexer) GetGenesisState() error {
	genesisBlock := primitives.GetGenesisBlock()
	genesisHash := genesisBlock.Hash()

	init, err := initialization.LoadParams(i.netParams.Name)
	if err != nil {
		return err
	}

	initialValidators := make([]initialization.ValidatorInitialization, len(init.Validators))
	for i := range initialValidators {
		v := initialization.ValidatorInitialization{
			PubKey:       init.Validators[i].PublicKey,
			PayeeAddress: init.PremineAddress,
		}
		initialValidators[i] = v
	}

	var genesisTime time.Time
	if init.GenesisTime == 0 {
		genesisTime = time.Now()
	} else {
		genesisTime = time.Unix(init.GenesisTime, 0)
	}

	ip := &initialization.InitializationParameters{
		GenesisTime:       genesisTime,
		InitialValidators: initialValidators,
		PremineAddress:    init.PremineAddress,
	}

	i.state, err = state.GetGenesisStateWithInitializationParameters(genesisHash, ip, i.netParams)
	if err != nil {
		return err
	}

	return nil
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

	err = indexer.GetGenesisState()
	if err != nil {
		return nil, err
	}

	return indexer, nil
}
