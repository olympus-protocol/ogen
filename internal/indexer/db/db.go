package db

import (
	"errors"
	"github.com/olympus-protocol/ogen/internal/chainindex"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/params"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"sync"
)

var stateKey = "state"

// Database represents an DB connection
type Database struct {
	log logger.Logger
	DB  *gorm.DB

	canClose  *sync.WaitGroup
	netParams *params.ChainParams
}

func (d *Database) SetFinalized(e uint64) error {
	res := d.DB.Model(&Epoch{}).Where("epoch = ?", e).Update("finalized", true)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (d *Database) SetJustified(e uint64) error {
	res := d.DB.Model(&Epoch{}).Where("epoch = ?", e).Update("justified", true)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (d *Database) AddEpoch(e *Epoch) error {
	res := d.DB.Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(e)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (d *Database) AddSlot(s *Slot) error {
	res := d.DB.Create(s)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (d *Database) MarkSlotProposed(s *Slot) error {
	res := d.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "slot"}},
		DoUpdates: []clause.Assignment{{
			Column: clause.Column{Name: "block_hash"},
			Value:  s.BlockHash,
		}, {
			Column: clause.Column{Name: "proposed"},
			Value:  true,
		}},
		UpdateAll: false,
	}).Create(s)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (d *Database) AddAccounts(a *[]Account) error {

	res := d.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "account"}},
		UpdateAll: true,
	}).Create(a)

	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (d *Database) AddValidators(v *[]Validator) error {

	res := d.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "pub_key"}},
		UpdateAll: true,
	}).Create(v)
	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (d *Database) StoreState(s state.State, lastBlock *chainindex.BlockRow) error {
	ser := s.ToSerializable()
	buf, err := ser.Marshal()
	if err != nil {
		return err
	}

	dbState := &State{
		Key:             stateKey,
		Raw:             buf,
		LastBlock:       lastBlock.Hash[:],
		LastBlockHeight: lastBlock.Height,
	}

	res := d.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		UpdateAll: true,
	}).Create(dbState)

	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (d *Database) GetState() (state.State, chainhash.Hash, uint64, error) {
	var s State
	res := d.DB.Find(&State{}, &State{
		Key: stateKey,
	}).Scan(&s)
	if res.Error != nil || res.RowsAffected == -1 {
		return nil, [32]byte{}, 0, errors.New("no state found")
	}
	storedState := state.NewEmptyState()
	err := storedState.Unmarshal(s.Raw)
	if err != nil {
		return nil, [32]byte{}, 0, res.Error
	}
	var lastBlockHash [32]byte
	copy(lastBlockHash[:], s.LastBlock)

	return storedState, lastBlockHash, s.LastBlockHeight, nil
}

func (d *Database) AddBlock(b *Block) error {
	res := d.DB.Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(b)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (d *Database) GetRawBlock(hash chainhash.Hash) (*primitives.Block, uint64, error) {
	var block Block
	res := d.DB.Find(&Block{}, &Block{Hash: hash[:]}).Scan(&block)
	if res.Error != nil {
		return nil, 0, res.Error
	}
	b := new(primitives.Block)
	err := b.Unmarshal(block.RawBlock)
	if err != nil {
		return nil, 0, err
	}
	return b, block.Height, nil
}

func (d *Database) Close() {
	d.canClose.Wait()
	return
}

func (d *Database) Migrate() error {

	err := d.DB.AutoMigrate(&Block{})
	if err != nil {
		return err
	}

	err = d.DB.AutoMigrate(&Deposit{})
	if err != nil {
		return err
	}

	err = d.DB.AutoMigrate(&Tx{})
	if err != nil {
		return err
	}

	err = d.DB.AutoMigrate(&Vote{})
	if err != nil {
		return err
	}

	err = d.DB.AutoMigrate(&Epoch{})
	if err != nil {
		return err
	}

	err = d.DB.AutoMigrate(&Exit{})
	if err != nil {
		return err
	}

	err = d.DB.AutoMigrate(&Account{})
	if err != nil {
		return err
	}

	err = d.DB.AutoMigrate(&Validator{})
	if err != nil {
		return err
	}

	err = d.DB.AutoMigrate(&State{})
	if err != nil {
		return err
	}

	err = d.DB.AutoMigrate(&Slot{})
	if err != nil {
		return err
	}

	return nil
}

// NewDB creates a db client
func NewDB(dbConnString string, log logger.Logger, wg *sync.WaitGroup, netParams *params.ChainParams) *Database {

	gdb, err := gorm.Open(postgres.Open(dbConnString), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database")
	}

	dbclient := &Database{
		log:       log,
		DB:        gdb,
		canClose:  wg,
		netParams: netParams,
	}

	dbclient.log.Info("Database connection established")

	return dbclient
}