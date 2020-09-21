package db

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/mattn/go-sqlite3"
	"github.com/olympus-protocol/ogen/pkg/primitives"
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

func (dbc DBClient) insert(tableName string, queryVars []interface{}) error {
	stmt, err := dbc.db.Prepare(dbc.queries["insert_"+tableName])
	if err != nil {
		return err
	}
	_, err = stmt.Exec(queryVars...)
	if err != nil {
		return err
	}
	/*ra, err := res.RowsAffected()
	fmt.Println(ra)*/
	return err

}

func (dbc DBClient) InsertBlock(block primitives.Block, height int) error {
	// insert into blocks table
	var queryVars []interface{}
	bHash := block.Hash().String()
	queryVars = append(queryVars, bHash, hex.EncodeToString(block.Signature[:]),
		hex.EncodeToString(block.RandaoSignature[:]), height)
	err := dbc.insert("blocks", queryVars)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	// blockheaders
	queryVars = nil
	queryVars = append(queryVars, bHash, int(block.Header.Version), int(block.Header.Nonce),
		hex.EncodeToString(block.Header.TxMerkleRoot[:]), hex.EncodeToString(block.Header.TxMultiMerkleRoot[:]),
		hex.EncodeToString(block.Header.VoteMerkleRoot[:]), hex.EncodeToString(block.Header.DepositMerkleRoot[:]),
		hex.EncodeToString(block.Header.ExitMerkleRoot[:]), hex.EncodeToString(block.Header.VoteSlashingMerkleRoot[:]),
		hex.EncodeToString(block.Header.RANDAOSlashingMerkleRoot[:]), hex.EncodeToString(block.Header.ProposerSlashingMerkleRoot[:]),
		hex.EncodeToString(block.Header.GovernanceVotesMerkleRoot[:]), hex.EncodeToString(block.Header.PrevBlockHash[:]),
		int(block.Header.Timestamp), int(block.Header.Slot), hex.EncodeToString(block.Header.StateRoot[:]), hex.EncodeToString(block.Header.FeeAddress[:]))
	err = dbc.insert("block_headers", queryVars)
	if err != nil {
		fmt.Println("blockheaders: " + err.Error())
		return err
	}
	// multivotes
	for _, vote := range block.Votes {
		queryVars = nil
		queryVars = append(queryVars, bHash, hex.EncodeToString(vote.Sig[:]), "bitfield", int(vote.Data.Slot), int(vote.Data.FromEpoch),
			hex.EncodeToString(vote.Data.FromHash[:]), int(vote.Data.ToEpoch), hex.EncodeToString(vote.Data.ToHash[:]), hex.EncodeToString(vote.Data.BeaconBlockHash[:]),
			int(vote.Data.Nonce))
		err = dbc.insert("multi_votes", queryVars)
		if err != nil {
			fmt.Println("multi_votes: " + err.Error())
			return err
		}
	}
	// transactions (single and multi)
	for _, tx := range block.Txs {
		queryVars = nil
		queryVars = append(queryVars, bHash, 0, hex.EncodeToString(tx.To[:]), hex.EncodeToString(tx.FromPublicKey[:]),
			tx.Amount, tx.Nonce, tx.Fee, 0, hex.EncodeToString(tx.Signature[:]))
		err = dbc.insert("transactions_0", queryVars)
		if err != nil {
			fmt.Println("transactions0: " + err.Error())
			return err
		}
	}
	for _, tx := range block.TxsMulti {
		queryVars = nil
		multiSig, err := tx.Signature.MarshalSSZ()
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		queryVars = append(queryVars, bHash, 1, hex.EncodeToString(tx.To[:]),
			tx.Amount, tx.Nonce, tx.Fee, 1, hex.EncodeToString(multiSig))
		err = dbc.insert("transactions_1", queryVars)
		if err != nil {
			fmt.Println("transactions1: " + err.Error())
			return err
		}
	}
	//// validators
	//	queryVars = nil
	//	queryVars = append()
	//	err = dbc.insert("validators", queryVars)
	//	if err != nil {
	//		fmt.Println("validators: " +err.Error())
	//		return err
	//	}
	//}
	// deposits
	for _, depo := range block.Deposits {
		queryVars = nil
		queryVars = append(queryVars, bHash, hex.EncodeToString(depo.PublicKey[:]), hex.EncodeToString(depo.Signature[:]),
			hex.EncodeToString(depo.Data.PublicKey[:]), hex.EncodeToString(depo.Data.ProofOfPossession[:]), hex.EncodeToString(depo.Data.WithdrawalAddress[:]))
		err = dbc.insert("deposits", queryVars)
		if err != nil {
			fmt.Println("deposits: " + err.Error())
			return err
		}
	}
	// exits
	for _, exits := range block.Exits {
		queryVars = nil
		queryVars = append(queryVars, bHash, hex.EncodeToString(exits.ValidatorPubkey[:]), hex.EncodeToString(exits.WithdrawPubkey[:]),
			hex.EncodeToString(exits.Signature[:]))
		err = dbc.insert("exits", queryVars)
		if err != nil {
			fmt.Println("exits: " + err.Error())
			return err
		}
	}
	//// vote_slashings
	//for _, vs := range block.VoteSlashings {
	//	queryVars = nil
	//	queryVars = append(queryVars, bHash, vs.Vote1, vs.Vote2)
	//	err = dbc.insert("vote_slashings", queryVars)
	//	if err != nil {
	//		fmt.Println("vote_slashings: " +err.Error())
	//		return err
	//	}
	//}
	// RANDAO_slashings
	for _, rs := range block.RANDAOSlashings {
		queryVars = nil
		queryVars = append(queryVars, bHash, hex.EncodeToString(rs.RandaoReveal[:]), int(rs.Slot), hex.EncodeToString(rs.ValidatorPubkey[:]))
		err = dbc.insert("RANDAO_slashings", queryVars)
		if err != nil {
			fmt.Println("RANDAO_slashings: " + err.Error())
			return err
		}
	}
	//// proposer_slashings
	//for _, ps := range block.ProposerSlashings {
	//	queryVars = nil
	//	queryVars = append(queryVars, bHash, ps.BlockHeader1, ps.BlockHeader2,hex.EncodeToString(ps.Signature1[:]),
	//		hex.EncodeToString(ps.Signature2[:]))
	//	err = dbc.insert("proposer_slashings", queryVars)
	//	if err != nil {
	//		fmt.Println("proposer_slashings: " +err.Error())
	//		return err
	//	}
	//}
	return nil
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
