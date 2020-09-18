package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"os"
)

// DBClient represents an DB connectio
type DBClient struct {
	Name    string
	db      *sql.DB
	params  DbParameters
	queries map[string]string
}

func (dbc DBClient) Ping() error {
	return dbc.db.Ping()
}

func (dbc DBClient) InitializeTables() error {
	// extract queries
	jsonFile, err := os.Open("./db/queries.json")
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	err = json.Unmarshal(byteValue, &dbc.queries)
	if err != nil {
		return err
	}

	err = dbc.CreateTable("blocks")
	if err != nil {
		return err
	}
	err = dbc.CreateTable("block_headers")
	if err != nil {
		return err
	}
	err = dbc.CreateTable("multi_votes")
	if err != nil {
		return err
	}
	err = dbc.CreateTable("transactions")
	if err != nil {
		return err
	}
	err = dbc.CreateTable("validators")
	if err != nil {
		return err
	}
	err = dbc.CreateTable("deposits")
	if err != nil {
		return err
	}
	err = dbc.CreateTable("exits")
	if err != nil {
		return err
	}
	err = dbc.CreateTable("vote_slashings")
	if err != nil {
		return err
	}
	err = dbc.CreateTable("RANDAO_slashings")
	if err != nil {
		return err
	}
	err = dbc.CreateTable("proposer_slashings")
	if err != nil {
		return err
	}

	return nil
}

func (dbc DBClient) CreateTable(tableName string) error {
	rows, exists := dbc.db.Query("select * from " + tableName + ";")
	if exists == nil {
		defer rows.Close()
		return nil
	} else {
		stmt, err := dbc.db.Prepare(dbc.queries["create_"+tableName])
		if err != nil {
			return err
		}
		_, err = stmt.Exec()
		if err != nil {
			return err
		}
		return err
	}
}

func (dbc DBClient) Insert(tableName string, queryVars []interface{}) error {
	stmt, err := dbc.db.Prepare(dbc.queries["insert_"+tableName])
	if err != nil {
		return err
	}
	res, err := stmt.Exec(queryVars...)
	if err != nil {
		return err
	}
	ra, err := res.RowsAffected()
	fmt.Println(ra)
	return err

}

func (dbc DBClient) CloseDB() {
	defer dbc.db.Close()
	return
}

func (dbc DBClient) OpenDB(parameters DbParameters) error {
	connString, err := getConnString(parameters)
	if err != nil {
		panic(err)
	}
	db, err := sql.Open(parameters.DriverName, connString)
	dbc.db = db
	return err
}

func getConnString(params DbParameters) (string, error) {
	var connString string
	switch params.DriverName {
	case "pgx":
		connString = fmt.Sprintf("port=%d host=%s user=%s "+
			"password=%s dbname=%s sslmode=disable",
			params.HostPort, params.Hostname, params.Username, params.Password, params.DatabaseName)
	case "sqlite3":
		connString = "./db/" + params.DatabaseName + ".db?_foreign_keys=on"
	}
	if connString == "" {
		return "", errors.New("dbms not specified")
	}
	return connString, nil

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
	connString, err := getConnString(parameters)
	if err != nil {
		panic(err)
	}

	db, err := sql.Open(parameters.DriverName, connString)
	if err != nil {
		panic(err)
	}

	client := &DBClient{
		Name:    parameters.DatabaseName,
		db:      db,
		queries: map[string]string{},
	}
	return client
}
