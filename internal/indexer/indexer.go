package indexer

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"github.com/olympus-protocol/ogen/api/proto"
	"github.com/olympus-protocol/ogen/cmd/ogen/initialization"
	"github.com/olympus-protocol/ogen/internal/chain"
	"github.com/olympus-protocol/ogen/internal/chainindex"
	"github.com/olympus-protocol/ogen/internal/indexer/db"
	"github.com/olympus-protocol/ogen/internal/indexer/models"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
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
	index     *chainindex.BlockIndex
}

func (i *Indexer) ProcessBlock(b *primitives.Block) error {
	i.log.Infof("Processing block at slot %d", b.Header.Slot)

	tip, _ := i.index.Get(b.Header.PrevBlockHash)
	v := chain.NewChainView(tip)

	_, err := i.state.ProcessSlots(b.Header.Slot, &v)
	if err != nil {
		return err
	}

	if b.Header.Slot/i.netParams.EpochLength > i.state.GetEpochIndex() {
		
	}

	err = i.state.ProcessBlock(b)
	if err != nil {
		return err
	}

	row, err := i.index.Add(b)
	if err != nil {
		return err
	}

	nonce := make([]byte, 8)
	binary.LittleEndian.PutUint64(nonce, b.Header.Nonce)

	dbBlock := &models.Block{
		Hash:   hex.EncodeToString(row.Hash[:]),
		Height: row.Height,
		Slot:   row.Slot,
		Header: models.BlockHeader{
			Hash:                       hex.EncodeToString(row.Hash[:]),
			Version:                    b.Header.Version,
			Nonce:                      hex.EncodeToString(nonce),
			TxMerkleRoot:               hex.EncodeToString(b.Header.TxMerkleRoot[:]),
			TxMultiMerkleRoot:          hex.EncodeToString(b.Header.TxMultiMerkleRoot[:]),
			VoteMerkleRoot:             hex.EncodeToString(b.Header.VoteMerkleRoot[:]),
			DepositMerkleRoot:          hex.EncodeToString(b.Header.DepositMerkleRoot[:]),
			ExitMerkleRoot:             hex.EncodeToString(b.Header.ExitMerkleRoot[:]),
			VoteSlashingMerkleRoot:     hex.EncodeToString(b.Header.VoteSlashingMerkleRoot[:]),
			RandaoSlashingMerkleRoot:   hex.EncodeToString(b.Header.RANDAOSlashingMerkleRoot[:]),
			ProposerSlashingMerkleRoot: hex.EncodeToString(b.Header.ProposerSlashingMerkleRoot[:]),
			GovernanceVotesMerkleRoot:  hex.EncodeToString(b.Header.GovernanceVotesMerkleRoot[:]),
			PreviousBlockHash:          hex.EncodeToString(b.Header.PrevBlockHash[:]),
			Timestamp:                  time.Unix(int64(b.Header.Timestamp), 0),
			Slot:                       b.Header.Slot,
			StateRoot:                  hex.EncodeToString(b.Header.StateRoot[:]),
			FeeAddress:                 hex.EncodeToString(b.Header.FeeAddress[:]),
		},
	}

	if len(b.Txs) > 0 {
		dbTxs := make([]models.Tx, len(b.Txs))
		for i := range b.Txs {
			txHash := b.Txs[i].Hash()
			fpkh, err := b.Txs[i].FromPubkeyHash()
			if err != nil {
				return err
			}
			dbTxs[i] = models.Tx{
				BlockHash:         hex.EncodeToString(row.Hash[:]),
				Hash:              hex.EncodeToString(txHash[:]),
				ToAddress:         hex.EncodeToString(b.Txs[i].To[:]),
				FromPublicKey:     hex.EncodeToString(b.Txs[i].FromPublicKey[:]),
				FromPublicKeyHash: hex.EncodeToString(fpkh[:]),
				Amount:            b.Txs[i].Amount,
				Nonce:             b.Txs[i].Nonce,
				Fee:               b.Txs[i].Fee,
			}
		}
		dbBlock.Txs = dbTxs
	}

	if len(b.Deposits) > 0 {
		dbDeposits := make([]models.Deposit, len(b.Deposits))
		for i := range b.Deposits {
			hash := chainhash.HashB(b.Deposits[i].Data.PublicKey[:])
			dbDeposits[i] = models.Deposit{
				Hash:      hex.EncodeToString(hash[:]),
				BlockHash: hex.EncodeToString(row.Hash[:]),
				PublicKey: hex.EncodeToString(b.Deposits[i].PublicKey[:]),
				Data: models.DepositData{
					Hash:              hex.EncodeToString(hash[:]),
					PublicKey:         hex.EncodeToString(b.Deposits[i].Data.PublicKey[:]),
					ProofOfPossession: hex.EncodeToString(b.Deposits[i].Data.ProofOfPossession[:]),
					WithdrawalAddress: hex.EncodeToString(b.Deposits[i].Data.WithdrawalAddress[:]),
				},
			}
		}
		dbBlock.Deposits = dbDeposits
	}

	if len(b.Votes) > 0 {
		dbVotes := make([]models.Vote, len(b.Votes))
		for i := range b.Votes {
			nonce := make([]byte, 8)
			binary.LittleEndian.PutUint64(nonce, b.Votes[i].Data.Nonce)
			hash := b.Votes[i].Data.Hash()
			dbVotes[i] = models.Vote{
				Hash:                  hex.EncodeToString(hash[:]),
				BlockHash:             hex.EncodeToString(row.Hash[:]),
				ParticipationBitfield: hex.EncodeToString(b.Votes[i].ParticipationBitfield[:]),
				Data: models.VoteData{
					Hash:            hex.EncodeToString(hash[:]),
					Slot:            b.Votes[i].Data.Slot,
					FromEpoch:       b.Votes[i].Data.FromEpoch,
					FromHash:        hex.EncodeToString(b.Votes[i].Data.FromHash[:]),
					ToEpoch:         b.Votes[i].Data.ToEpoch,
					ToHash:          hex.EncodeToString(b.Votes[i].Data.ToHash[:]),
					BeaconBlockHash: hex.EncodeToString(b.Votes[i].Data.BeaconBlockHash[:]),
					Nonce:           hex.EncodeToString(nonce),
				},
			}
		}
		dbBlock.Votes = dbVotes
	}

	err = i.db.AddBlock(dbBlock)
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
		err = i.ProcessBlock(block)
		if err != nil {
			i.log.Error("unable to process block")
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

			err = i.ProcessBlock(block)
			if err != nil {
				i.log.Error("unable to process block")
				break
			}
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

	i.state.ProcessSlot(genesisHash)

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

	genesisBlock := primitives.GetGenesisBlock()

	idx, err := chainindex.InitBlocksIndex(genesisBlock)
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
		index:     idx,
	}

	err = indexer.GetGenesisState()
	if err != nil {
		return nil, err
	}

	return indexer, nil
}
