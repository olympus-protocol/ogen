package db

import (
	"database/sql"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/params"

	"github.com/olympus-protocol/ogen/pkg/logger"
	"sync"
)

// Database represents an DB connection
type Database struct {
	log logger.Logger
	db  *sql.DB

	canClose  *sync.WaitGroup
	netParams *params.ChainParams

	state state.State
}

func (d *Database) Close() {
	d.canClose.Wait()
	_ = d.db.Close()
	return
}

func (d *Database) Migrate() error {
	migrationsString := "file://cmd/ogen/indexer/db/migrations"
	dbdriver, err := postgres.WithInstance(d.db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrationsString,
		"postgres",
		dbdriver,
	)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

// NewDB creates a db client
func NewDB(dbConnString string, log logger.Logger, wg *sync.WaitGroup, netParams *params.ChainParams) *Database {
	db, err := sql.Open("postgres", dbConnString)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	dbclient := &Database{
		log:       log,
		db:        db,
		canClose:  wg,
		netParams: netParams,
	}

	dbclient.log.Info("Database connection established")

	return dbclient
}
