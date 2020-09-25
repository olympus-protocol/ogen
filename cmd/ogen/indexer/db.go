package indexer

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/mattn/go-sqlite3"
	"github.com/olympus-protocol/ogen/internal/logger"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Database represents an DB connection
type Database struct {
	log     logger.Logger
	name    string
	db      *sql.DB
	params  *Config
	queries map[string]string
}

// State represents the last block saved
type State struct {
	Blocks        int
	LastBlockHash string
}

func (d *Database) Ping() error {
	return d.db.Ping()
}

func (d *Database) InitializeTables() error {
	// extract queries
	path, err := os.Getwd()
	if err != nil {
		return err
	}
	jsonFile, err := os.Open(filepath.Join(path, "cmd/ogen-d/db/queries.json"))
	if err != nil {
		return err
	}

	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	err = json.Unmarshal(byteValue, &d.queries)
	if err != nil {
		return err
	}

	err = d.CreateTable("blocks")
	if err != nil {
		return err
	}
	err = d.CreateTable("block_headers")
	if err != nil {
		return err
	}
	err = d.CreateTable("multi_votes")
	if err != nil {
		return err
	}
	err = d.CreateTable("transactions")
	if err != nil {
		return err
	}
	err = d.CreateTable("validators")
	if err != nil {
		return err
	}
	err = d.CreateTable("deposits")
	if err != nil {
		return err
	}
	err = d.CreateTable("exits")
	if err != nil {
		return err
	}
	err = d.CreateTable("vote_slashings")
	if err != nil {
		return err
	}
	err = d.CreateTable("RANDAO_slashings")
	if err != nil {
		return err
	}
	err = d.CreateTable("proposer_slashings")
	if err != nil {
		return err
	}

	return nil
}

func (d *Database) CreateTable(tableName string) error {
	rows, exists := d.db.Query("select * from " + tableName + ";")
	if exists == nil {
		defer rows.Close()
		return nil
	} else {
		stmt, err := d.db.Prepare(d.queries["create_"+tableName])
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

func (d *Database) insert(tableName string, queryVars []interface{}) error {
	stmt, err := d.db.Prepare(d.queries["insert_"+tableName])
	if err != nil {
		return err
	}
	_, err = stmt.Exec(queryVars...)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return errors.New("skip block")
		}
		return err
	}
	return err

}

// funcion to run a query that is expected to return one row and returns the first column
func (d *Database) querySingleRow(query string) (string, error) {
	var res string
	err := d.db.QueryRow(query).Scan(&res)
	if err != nil {
		return "", err
	}
	return res, nil
}

//Inserts a block into the db.
func (d *Database) InsertBlock(block primitives.Block) error {
	nextHeight, prevHash, err := d.getNextHeight()
	if err != nil {
		fmt.Println("height: " + err.Error())
		return err
	}
	fmt.Println(nextHeight)
	if nextHeight > 0 && hex.EncodeToString(block.Header.PrevBlockHash[:]) != prevHash {
		return handleError("blocks", errors.New("skip block"))
	}
	// insert into blocks table
	var queryVars []interface{}
	bHash := block.Hash().String()
	queryVars = append(queryVars, bHash, hex.EncodeToString(block.Signature[:]),
		hex.EncodeToString(block.RandaoSignature[:]), nextHeight)
	err = d.insert("blocks", queryVars)
	if err != nil {
		return handleError("blocks", err)
	}
	// blockheaders
	queryVars = nil
	queryVars = append(queryVars, bHash, int(block.Header.Version), int(block.Header.Nonce),
		hex.EncodeToString(block.Header.TxMerkleRoot[:]), hex.EncodeToString(block.Header.TxMultiMerkleRoot[:]),
		hex.EncodeToString(block.Header.VoteMerkleRoot[:]), hex.EncodeToString(block.Header.DepositMerkleRoot[:]),
		hex.EncodeToString(block.Header.ExitMerkleRoot[:]), hex.EncodeToString(block.Header.VoteSlashingMerkleRoot[:]),
		hex.EncodeToString(block.Header.RANDAOSlashingMerkleRoot[:]), hex.EncodeToString(block.Header.ProposerSlashingMerkleRoot[:]),
		hex.EncodeToString(block.Header.GovernanceVotesMerkleRoot[:]), hex.EncodeToString(block.Header.PrevBlockHash[:]),
		int(block.Header.Timestamp), int(block.Header.Slot), hex.EncodeToString(block.Header.StateRoot[:]),
		hex.EncodeToString(block.Header.FeeAddress[:]), block.Header.Hash().String())
	err = d.insert("block_headers", queryVars)
	if err != nil {
		return handleError("block_headers", err)
	}
	// multivotes
	for _, vote := range block.Votes {
		queryVars = nil
		queryVars = append(queryVars, bHash, hex.EncodeToString(vote.Sig[:]), hex.EncodeToString(vote.ParticipationBitfield), int(vote.Data.Slot), int(vote.Data.FromEpoch),
			hex.EncodeToString(vote.Data.FromHash[:]), int(vote.Data.ToEpoch), hex.EncodeToString(vote.Data.ToHash[:]), hex.EncodeToString(vote.Data.BeaconBlockHash[:]),
			int(vote.Data.Nonce), vote.Hash().String())
		err = d.insert("multi_votes", queryVars)
		if err != nil {
			return handleError("votes", err)
		}
	}
	// transactions (single and multi)
	for _, tx := range block.Txs {
		queryVars = nil
		queryVars = append(queryVars, bHash, 0, hex.EncodeToString(tx.To[:]), hex.EncodeToString(tx.FromPublicKey[:]),
			tx.Amount, tx.Nonce, tx.Fee, 0, hex.EncodeToString(tx.Signature[:]))
		err = d.insert("transactions_0", queryVars)
		if err != nil {
			return handleError("single_tx", err)
		}
	}
	for _, tx := range block.TxsMulti {
		queryVars = nil
		multiSig, err := tx.Signature.MarshalSSZ()
		if err != nil {
			return err
		}
		queryVars = append(queryVars, bHash, 1, hex.EncodeToString(tx.To[:]),
			tx.Amount, tx.Nonce, tx.Fee, 1, hex.EncodeToString(multiSig))
		err = d.insert("transactions_1", queryVars)
		if err != nil {
			return handleError("multi_tx", err)
		}
	}

	for _, depo := range block.Deposits {
		queryVars = nil
		queryVars = append(queryVars, bHash, hex.EncodeToString(depo.PublicKey[:]), hex.EncodeToString(depo.Signature[:]),
			hex.EncodeToString(depo.Data.PublicKey[:]), hex.EncodeToString(depo.Data.ProofOfPossession[:]), hex.EncodeToString(depo.Data.WithdrawalAddress[:]))
		err = d.insert("deposits", queryVars)
		if err != nil {
			return handleError("deposits", err)
		}
		queryVars = nil
		queryVars = append(queryVars, hex.EncodeToString(depo.Data.PublicKey[:]))
		err = d.insert("validators", queryVars)
		if err != nil {
			return handleError("validators", err)
		}
	}
	// exits
	for _, exits := range block.Exits {
		queryVars = nil
		queryVars = append(queryVars, bHash, hex.EncodeToString(exits.ValidatorPubkey[:]), hex.EncodeToString(exits.WithdrawPubkey[:]),
			hex.EncodeToString(exits.Signature[:]))
		err = d.insert("exits", queryVars)
		if err != nil {
			return handleError("exits", err)
		}
	}
	// vote_slashings
	for _, vs := range block.VoteSlashings {
		// find votes id
		v1, err := d.querySingleRow("select id from multi_votes where vote_hash = " + vs.Vote1.Hash().String())
		if err != nil {
			return handleError("vote_slashings", err)
		}
		v2, err := d.querySingleRow("select id from multi_votes where vote_hash = " + vs.Vote2.Hash().String())
		if err != nil {
			return handleError("vote_slashings", err)
		}
		vote1Int, err := strconv.Atoi(v1)
		if err != nil {
			return err
		}
		vote2Int, err := strconv.Atoi(v2)
		if err != nil {
			return err
		}
		queryVars = nil
		queryVars = append(queryVars, bHash, vote1Int, vote2Int)
		err = d.insert("vote_slashings", queryVars)
		if err != nil {
			return handleError("vote_slashings", err)
		}
	}
	// RANDAO_slashings
	for _, rs := range block.RANDAOSlashings {
		queryVars = nil
		queryVars = append(queryVars, bHash, hex.EncodeToString(rs.RandaoReveal[:]), int(rs.Slot), hex.EncodeToString(rs.ValidatorPubkey[:]))
		err = d.insert("RANDAO_slashings", queryVars)
		if err != nil {
			return handleError("RANDAO_slashings", err)
		}
	}
	// proposer_slashings
	for _, ps := range block.ProposerSlashings {
		// find blockheader id
		bh1, err := d.querySingleRow("select id from block_headers where header_hash = " + ps.BlockHeader1.Hash().String())
		if err != nil {
			return handleError("proposer_slashings", err)
		}
		bh1Int, err := strconv.Atoi(bh1)
		bh2, err := d.querySingleRow("select id from block_headers where header_hash = " + ps.BlockHeader2.Hash().String())
		if err != nil {
			return handleError("proposer_slashings", err)
		}
		bh2Int, err := strconv.Atoi(bh2)
		queryVars = nil
		queryVars = append(queryVars, bHash, bh1Int, bh2Int, hex.EncodeToString(ps.Signature1[:]),
			hex.EncodeToString(ps.Signature2[:]))
		err = d.insert("proposer_slashings", queryVars)
		if err != nil {
			return handleError("proposer_slashings", err)
		}
	}
	return nil
}

func handleError(s string, err error) error {
	if err.Error() == "skip block" {
		fmt.Println("skip block")
		return nil
	}
	fmt.Println(s + ": " + err.Error())
	return err
}

func (d *Database) CloseDB() {
	defer d.db.Close()
	return
}

func (d *Database) OpenDB(parameters *Config) error {
	connString, err := getConnString(parameters)
	if err != nil {
		panic(err)
	}
	db, err := sql.Open(parameters.DriverName, connString)
	d.db = db
	return err
}

func (d *Database) GetCurrentState() (State, error) {
	nextH, prevH, err := d.getNextHeight()
	if err != nil {
		return State{}, err
	}
	if nextH == 0 {
		return State{
			Blocks:        0,
			LastBlockHash: "",
		}, nil
	} else {
		return State{Blocks: nextH - 1, LastBlockHash: prevH}, nil
	}
}

// returns state from height - gap
func (d *Database) GetSpecificState(gap int) (State, error) {
	nextH, _, err := d.getNextHeight()
	if err != nil {
		return State{}, err
	}
	var height int
	var hash string
	if nextH-1 <= gap {
		hash, err = d.getHeight(0)
		if err != nil {
			return State{}, err
		}
		height = 0
	} else {
		hash, err = d.getHeight(nextH - 1 - gap)
		if err != nil {
			return State{}, err
		}
		height = nextH - 1 - gap
	}
	return State{Blocks: height, LastBlockHash: hash}, nil
}

func (d *Database) getNextHeight() (int, string, error) {
	idS, err := d.querySingleRow("select max(rowid) from blocks;")
	if err != nil {
		if err.Error() == "sql: Scan error on column index 0, name \"max(rowid)\": converting NULL to string is unsupported" {
			return 0, "", nil
		}
		return -1, "", err
	}
	id, err := strconv.Atoi(idS)
	if err != nil {
		return -1, "", err
	}
	lasHash, err := d.querySingleRow("select block_hash from blocks where rowid = " + idS + ";")
	if err != nil {
		return id + 1, "", err
	}
	return id + 1, lasHash, nil
}

// getheight returns the hash at the specified height
func (d *Database) getHeight(i int) (string, error) {
	idS := strconv.Itoa(i)
	hash, err := d.querySingleRow("select block_hash from blocks where rowid = " + idS + ";")
	if err != nil {
		return "", err
	}
	return hash, nil
}

func getConnString(params *Config) (string, error) {
	var connString string
	switch params.DriverName {
	case "pgx":
		connString = fmt.Sprintf("port=%d host=%s user=%s "+
			"password=%s dbname=%s sslmode=disable",
			params.HostPort, params.Hostname, params.Username, params.Password, params.DatabaseName)
	case "sqlite3":
		path, err := os.Getwd()
		if err != nil {
			return "", err
		}
		connString = filepath.Join(path, "cmd/ogen-d/db") + "/" + params.DatabaseName + ".db?_foreign_keys=on"
	}
	if connString == "" {
		return "", errors.New("dbms not specified")
	}
	return connString, nil

}

type Config struct {
	Hostname     string
	HostPort     int
	Username     string
	Password     string
	DatabaseName string
	DriverName   string
}

// NewDBClient creates a db client
func NewDBClient(parameters *Config) *Database {
	connString, err := getConnString(parameters)
	if err != nil {
		panic(err)
	}
	fmt.Println(connString)
	db, err := sql.Open(parameters.DriverName, connString)
	if err != nil {
		panic(err)
	}
	// check the connection to the db
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("You are Successfully connected!")
	client := &Database{
		name:    parameters.DatabaseName,
		db:      db,
		queries: map[string]string{},
	}
	return client
}
