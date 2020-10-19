package indexer

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/mysql"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/olympus-protocol/ogen/pkg/logger"
)

// Database represents an DB connection
type Database struct {
	log  logger.Logger
	db   *sql.DB
}

// NewDBClient creates a db client
func NewDB(dbConnString string, log logger.Logger) *Database {

	db, err := sql.Open("mysql", "root:9563287421@tcp(localhost)/indexer")
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	driver, _ := mysql.WithInstance(db, &mysql.Config{})
	m, _ := migrate.NewWithDatabaseInstance(
		"file://cmd/ogen/indexer/migrations/mysql",
		"mysql",
		driver,
	)

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}

	dbclient := &Database{
		log: log,
		db:  db,
	}

	dbclient.log.Info("Database connection established")

	return dbclient
}
