package db

import (
	"github.com/olympus-protocol/ogen/internal/indexer/models"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/params"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"sync"
)

// Database represents an DB connection
type Database struct {
	log logger.Logger
	db  *gorm.DB

	canClose  *sync.WaitGroup
	netParams *params.ChainParams

	state state.State
}

func (d *Database) Close() {
	d.canClose.Wait()
	return
}

func (d *Database) Migrate() error {
	err := d.db.AutoMigrate(&models.Block{})
	if err != nil {
		return err
	}
	err = d.db.AutoMigrate(&models.BlockHeader{})
	if err != nil {
		return err
	}
	return nil
}

// NewDB creates a db client
func NewDB(dbConnString string, log logger.Logger, wg *sync.WaitGroup, netParams *params.ChainParams) *Database {

	gdb, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
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
