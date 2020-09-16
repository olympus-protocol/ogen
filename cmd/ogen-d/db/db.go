package db

import (
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/mattn/go-sqlite3"
)

// DBClient represents an DB connectio
type DBClient struct {
	Name   string
	db     *sql.DB
	params DbParameters
}

func (dbc DBClient) Ping() error {
	return dbc.db.Ping()
}

func (dbc DBClient) CloseDB() {
	defer dbc.db.Close()
	return
}

func (dbc DBClient) OpenDB(parameters DbParameters) error {
	/*connectionString := fmt.Sprintf("port=%d host=%s user=%s "+
	"password=%s dbname=%s sslmode=disable",
	parameters.HostPort, parameters.Hostname, parameters.Username, parameters.Password, parameters.DatabaseName)*/
	db, err := sql.Open("sqlite3", "./foo.db")
	dbc.db = db
	return err
}

type DbParameters struct {
	Hostname     string
	HostPort     int
	Username     string
	Password     string
	DatabaseName string
	DriverName   string
}

// NewDBClient creates a db client
func NewDBClient(parameters DbParameters) *DBClient {
	var connString string
	switch parameters.DriverName {
	case "pgx":
		connString = fmt.Sprintf("port=%d host=%s user=%s "+
			"password=%s dbname=%s sslmode=disable",
			parameters.HostPort, parameters.Hostname, parameters.Username, parameters.Password, parameters.DatabaseName)
	case "sqlite3":
		connString = "./" + parameters.DatabaseName + ".db"
	}

	fmt.Println(connString)
	db, err := sql.Open(parameters.DriverName, connString)
	if err != nil {
		panic(err)
	}

	client := &DBClient{
		Name: parameters.DatabaseName,
		db:   db,
	}
	return client
}
