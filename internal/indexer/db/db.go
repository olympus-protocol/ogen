package db

import (
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
	db  *gorm.DB

	canClose  *sync.WaitGroup
	netParams *params.ChainParams
}

func (d *Database) AddSlot(b *Slot) error {
	res := d.db.Create(b)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (d *Database) AddAccounts(a *[]Account) error {
	res := d.db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&a)

	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (d *Database) StoreState(s state.State) error {
	ser := s.ToSerializable()
	buf, err := ser.Marshal()
	if err != nil {
		return err
	}

	dbState := State{
		Key: "state",
		Raw: buf,
	}

	res := d.db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&dbState)

	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (d *Database) GetState() (state.State, error) {
	var s State
	res := d.db.Order("raw").Find(&State{}).Scan(&s)
	if res.Error != nil {
		return nil, res.Error
	}
	storedState := state.NewEmptyState()
	err := storedState.Unmarshal(s.Raw)
	if err != nil {
		return nil, err
	}
	return storedState, nil
}

func (d *Database) AddValidators(v *[]Validator) error {
	res := d.db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&v)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (d *Database) AddEpoch(b *Epoch) error {
	res := d.db.Create(b)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (d *Database) AddBlock(b *Block) error {
	res := d.db.Create(b)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (d *Database) GetRawBlock(hash chainhash.Hash) (*primitives.Block, error) {
	var block Block
	res := d.db.First(&Block{}, hash).Scan(&block)
	if res.Error != nil {
		return nil, res.Error
	}
	b := new(primitives.Block)
	err := b.Unmarshal(block.RawBlock)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (d *Database) Close() {
	d.canClose.Wait()
	return
}

func (d *Database) Migrate() error {

	err := d.db.AutoMigrate(&Block{})
	if err != nil {
		return err
	}

	err = d.db.AutoMigrate(&Deposit{})
	if err != nil {
		return err
	}

	err = d.db.AutoMigrate(&Tx{})
	if err != nil {
		return err
	}

	err = d.db.AutoMigrate(&Vote{})
	if err != nil {
		return err
	}

	err = d.db.AutoMigrate(&Epoch{})
	if err != nil {
		return err
	}

	err = d.db.AutoMigrate(&Exit{})
	if err != nil {
		return err
	}

	err = d.db.AutoMigrate(&Account{})
	if err != nil {
		return err
	}

	err = d.db.AutoMigrate(&Validator{})
	if err != nil {
		return err
	}

	err = d.db.AutoMigrate(&State{})
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
		db:        gdb,
		canClose:  wg,
		netParams: netParams,
	}

	dbclient.log.Info("Database connection established")

	return dbclient
}
