package indexer

import (
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"
	"github.com/golang-migrate/migrate/database"
	_ "github.com/mattn/go-sqlite3"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/mysql"
	"github.com/golang-migrate/migrate/database/sqlite3"

	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"strconv"
	"sync"
)

// State represents the last block saved
type State struct {
	Blocks        int
	LastBlockHash string
}

// Database represents an DB connection
type Database struct {
	log      logger.Logger
	db       *sql.DB
	canClose *sync.WaitGroup
	driver   string
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

func (d *Database) InsertBlock(block primitives.Block) error {

	nextHeight, prevHash, err := d.getNextHeight()
	if err != nil {
		d.log.Error(err)
		return err
	}

	if nextHeight > 0 && hex.EncodeToString(block.Header.PrevBlockHash[:]) != prevHash {
		d.log.Error(errorPrevBlockHash)
		return errorPrevBlockHash
	}

	// Insert into blocks table
	var queryVars []interface{}
	hash := block.Hash()
	queryVars = append(queryVars, hash.String(), hex.EncodeToString(block.Signature[:]), hex.EncodeToString(block.RandaoSignature[:]), nextHeight)
	err = d.insertRow("blocks", queryVars)
	if err != nil {
		return err
	}

	// Block Headers
	queryVars = nil
	queryVars = append(queryVars, hash.String(), int(block.Header.Version), int(block.Header.Nonce),
		hex.EncodeToString(block.Header.TxMerkleRoot[:]), hex.EncodeToString(block.Header.TxMultiMerkleRoot[:]), hex.EncodeToString(block.Header.VoteMerkleRoot[:]),
		hex.EncodeToString(block.Header.DepositMerkleRoot[:]), hex.EncodeToString(block.Header.ExitMerkleRoot[:]), hex.EncodeToString(block.Header.VoteSlashingMerkleRoot[:]),
		hex.EncodeToString(block.Header.RANDAOSlashingMerkleRoot[:]), hex.EncodeToString(block.Header.ProposerSlashingMerkleRoot[:]),
		hex.EncodeToString(block.Header.GovernanceVotesMerkleRoot[:]), hex.EncodeToString(block.Header.PrevBlockHash[:]),
		int(block.Header.Timestamp), int(block.Header.Slot), hex.EncodeToString(block.Header.StateRoot[:]),
		hex.EncodeToString(block.Header.FeeAddress[:]))
	err = d.insertRow("block_headers", queryVars)
	if err != nil {
		return err
	}

	// Votes
	for _, vote := range block.Votes {
		queryVars = nil
		queryVars = append(queryVars, hash.String(), hex.EncodeToString(vote.Sig[:]), hex.EncodeToString(vote.ParticipationBitfield), int(vote.Data.Slot), int(vote.Data.FromEpoch),
			hex.EncodeToString(vote.Data.FromHash[:]), int(vote.Data.ToEpoch), hex.EncodeToString(vote.Data.ToHash[:]), hex.EncodeToString(vote.Data.BeaconBlockHash[:]),
			int(vote.Data.Nonce), vote.Data.Hash().String())
		err = d.insertRow("votes", queryVars)
		if err != nil {
			continue
		}
	}

	// Transactions (single and multi)
	for _, tx := range block.Txs {
		queryVars = nil
		queryVars = append(queryVars, hash, 0, hex.EncodeToString(tx.To[:]), hex.EncodeToString(tx.FromPublicKey[:]),
			tx.Amount, tx.Nonce, tx.Fee, 0, hex.EncodeToString(tx.Signature[:]))
		err = d.insertRow("transactions_0", queryVars)
		if err != nil {
			continue
		}
	}
	for _, tx := range block.TxsMulti {
		queryVars = nil
		multiSig, err := tx.Signature.MarshalSSZ()
		if err != nil {
			return err
		}
		queryVars = append(queryVars, hash, 1, hex.EncodeToString(tx.To[:]),
			tx.Amount, tx.Nonce, tx.Fee, 1, hex.EncodeToString(multiSig))
		err = d.insertRow("transactions_1", queryVars)
		if err != nil {
			continue
		}
	}

	for _, depo := range block.Deposits {
		queryVars = nil
		queryVars = append(queryVars, hash, hex.EncodeToString(depo.PublicKey[:]), hex.EncodeToString(depo.Signature[:]),
			hex.EncodeToString(depo.Data.PublicKey[:]), hex.EncodeToString(depo.Data.ProofOfPossession[:]), hex.EncodeToString(depo.Data.WithdrawalAddress[:]))
		err = d.insertRow("deposits", queryVars)
		if err != nil {
			continue
		}
		queryVars = nil
		queryVars = append(queryVars, hex.EncodeToString(depo.Data.PublicKey[:]))
		err = d.insertRow("validators", queryVars)
		if err != nil {
			continue
		}
	}

	// Exits
	for _, exits := range block.Exits {
		queryVars = nil
		queryVars = append(queryVars, hash, hex.EncodeToString(exits.ValidatorPubkey[:]), hex.EncodeToString(exits.WithdrawPubkey[:]),
			hex.EncodeToString(exits.Signature[:]))
		err = d.insertRow("exits", queryVars)
		if err != nil {
			continue
		}
	}

	// Vote Slashings
	for _, vs := range block.VoteSlashings {

		// find votes id
		v1, err := d.querySingleRow("select id from multi_votes where vote_hash = " + vs.Vote1.Data.Hash().String())
		if err != nil {
			continue
		}
		v2, err := d.querySingleRow("select id from multi_votes where vote_hash = " + vs.Vote2.Data.Hash().String())
		if err != nil {
			continue
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
		queryVars = append(queryVars, hash, vote1Int, vote2Int)
		err = d.insertRow("vote_slashings", queryVars)
		if err != nil {
			continue
		}
	}

	// RANDAO Slashings
	for _, rs := range block.RANDAOSlashings {
		queryVars = nil
		queryVars = append(queryVars, hash, hex.EncodeToString(rs.RandaoReveal[:]), int(rs.Slot), hex.EncodeToString(rs.ValidatorPubkey[:]))
		err = d.insertRow("RANDAO_slashings", queryVars)
		if err != nil {
			continue
		}
	}

	// Proposer Slashings
	for _, ps := range block.ProposerSlashings {

		// find blockheader id

		bh1, err := d.querySingleRow("select id from block_headers where header_hash = " + ps.BlockHeader1.Hash().String())
		if err != nil {
			continue
		}
		bh1Int, err := strconv.Atoi(bh1)
		bh2, err := d.querySingleRow("select id from block_headers where header_hash = " + ps.BlockHeader2.Hash().String())
		if err != nil {
			continue
		}
		bh2Int, err := strconv.Atoi(bh2)
		queryVars = nil
		queryVars = append(queryVars, hash, bh1Int, bh2Int, hex.EncodeToString(ps.Signature1[:]),
			hex.EncodeToString(ps.Signature2[:]))
		err = d.insertRow("proposer_slashings", queryVars)
		if err != nil {
			continue
		}
	}

	return nil
}

func (d *Database) getNextHeight() (int, string, error) {
	idS, err := d.querySingleRow("select max(height) from blocks;")
	if err != nil {
		return 0, "", nil
	}
	id, err := strconv.Atoi(idS)
	if err != nil {
		return -1, "", err
	}
	lasHash, err := d.querySingleRow("select block_hash from blocks where height = " + idS + ";")
	if err != nil {
		return id + 1, "", err
	}
	return id + 1, lasHash, nil
}

func (d *Database) querySingleRow(query string) (string, error) {
	var res string
	err := d.db.QueryRow(query).Scan(&res)
	if err != nil {
		return "", err
	}
	return res, nil
}

func (d *Database) insertRow(tableName string, queryVars []interface{}) error {

	d.canClose.Add(1)
	defer d.canClose.Done()
	switch tableName {
	case "blocks":
		return d.insertBlockRow(queryVars)
	case "block_headers":
		return d.insertBlockHeadersRow(queryVars)
	case "votes":
		return d.insertVote(queryVars)

	}
	return nil
}

func (d *Database) insertBlockRow(queryVars []interface{}) error {
	dw := goqu.Dialect(d.driver)
	ds := dw.Insert("blocks").Rows(
		goqu.Record{
			"block_hash":             queryVars[0],
			"block_signature":        queryVars[1],
			"block_randao_signature": queryVars[2],
			"height":                 queryVars[3],
		},
	)
	query, _, err := ds.ToSQL()
	if err != nil {
		return err
	}
	_, err = d.db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) insertBlockHeadersRow(queryVars []interface{}) error {
	dw := goqu.Dialect(d.driver)
	ds := dw.Insert("block_headers").Rows(
		goqu.Record{
			"block_hash":                    queryVars[0],
			"version":                       queryVars[1],
			"nonce":                         queryVars[2],
			"tx_mekle_root":                 queryVars[3],
			"tx_multi_merkle_root":          queryVars[4],
			"vote_merkle_root":              queryVars[5],
			"deposit_merkle_root":           queryVars[6],
			"exit_merkle_root":              queryVars[7],
			"vote_slashing_merkle_root":     queryVars[8],
			"randao_slashing_merkle_root":   queryVars[9],
			"proposer_slashing_merkle_root": queryVars[10],
			"governance_votes_merkle_root":  queryVars[11],
			"previous_block_hash":           queryVars[12],
			"timestamp":                     queryVars[13],
			"slot":                          queryVars[14],
			"state_root":                    queryVars[15],
			"fee_address":                   queryVars[16],
		},
	)
	query, _, err := ds.ToSQL()
	if err != nil {
		return err
	}
	_, err = d.db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) insertVote(queryVars []interface{}) error {
	dw := goqu.Dialect(d.driver)
	ds := dw.Insert("votes").Rows(
		goqu.Record{
			"block_hash":             queryVars[0],
			"signature":              queryVars[1],
			"participation_bitfield": queryVars[2],
			"data_slot":              queryVars[3],
			"data_from_epoch":        queryVars[4],
			"data_from_hash":         queryVars[5],
			"data_to_epoch":          queryVars[6],
			"data_to_hash":           queryVars[7],
			"data_beacon_block_hash": queryVars[8],
			"data_nonce":             queryVars[9],
			"vote_hash":              queryVars[10],
		},
	)
	query, _, err := ds.ToSQL()
	if err != nil {
		return err
	}
	_, err = d.db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

// NewDB creates a db client
func NewDB(dbConnString string, log logger.Logger, wg *sync.WaitGroup, driver string) *Database {
	fmt.Println(driver, dbConnString)
	db, err := sql.Open(driver, dbConnString)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	err = runMigrations(driver, db)
	if err != nil {
		log.Fatal(err)
	}

	dbclient := &Database{
		log:      log,
		db:       db,
		canClose: wg,
		driver:   driver,
	}

	dbclient.log.Info("Database connection established")

	return dbclient
}

func runMigrations(driver string, db *sql.DB) error {
	var driverWrapper database.Driver
	var err error
	var migrationsString string
	switch driver {
	case "mysql":
		migrationsString = "file://cmd/ogen/indexer/migrations/mysql"
		driverWrapper, err = mysql.WithInstance(db, &mysql.Config{})
		if err != nil {
			return err
		}
	case "sqlite3":
		migrationsString = "file://cmd/ogen/indexer/migrations/sqlite3"
		driverWrapper, err = sqlite3.WithInstance(db, &sqlite3.Config{})
		if err != nil {
			return err
		}
	default:
		return errors.New("driver not supported")
	}

	m, _ := migrate.NewWithDatabaseInstance(
		migrationsString,
		driver,
		driverWrapper,
	)
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
