package db

import (
	"database/sql"
	"encoding/hex"
	"errors"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"strconv"
	"sync"
)

var ErrorPrevBlockHash = errors.New("block previous hash doesn't match")

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
		d.log.Error(ErrorPrevBlockHash)
		return ErrorPrevBlockHash
	}

	// Insert into blocks table
	var queryVars []interface{}
	hash := block.Hash()
	queryVars = append(queryVars, hash.String(), hex.EncodeToString(block.Signature[:]), hex.EncodeToString(block.RandaoSignature[:]), nextHeight)
	err = d.insertRow("blocks", queryVars)
	if err != nil {
		d.log.Error(err)
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
		d.log.Error(err)
	}

	// Votes
	for _, vote := range block.Votes {
		queryVars = nil
		queryVars = append(queryVars, hash.String(), hex.EncodeToString(vote.Sig[:]), hex.EncodeToString(vote.ParticipationBitfield), int(vote.Data.Slot), int(vote.Data.FromEpoch),
			hex.EncodeToString(vote.Data.FromHash[:]), int(vote.Data.ToEpoch), hex.EncodeToString(vote.Data.ToHash[:]), hex.EncodeToString(vote.Data.BeaconBlockHash[:]),
			int(vote.Data.Nonce), vote.Data.Hash().String())
		err = d.insertRow("votes", queryVars)
		if err != nil {
			d.log.Error(err)
			continue
		}
	}

	// Transactions Single
	for _, tx := range block.Txs {
		queryVars = nil
		queryVars = append(queryVars, tx.Hash().String(), hash.String(), 0, hex.EncodeToString(tx.To[:]), hex.EncodeToString(tx.FromPublicKey[:]),
			int(tx.Amount), int(tx.Nonce), int(tx.Fee), hex.EncodeToString(tx.Signature[:]))
		err = d.insertRow("tx_single", queryVars)
		if err != nil {
			d.log.Error(err)
			continue
		}

		// update the receiver account
		queryVars = nil
		queryVars = append(queryVars, hex.EncodeToString(tx.To[:]), int(tx.Amount), 0, int(tx.Amount), int(tx.Amount), 0, int(tx.Amount))
		err = d.insertRow("accounts", queryVars)
		if err != nil {
			continue
		}

		// update the sender account
		queryVars = nil
		fromAccount, err := tx.FromPubkeyHash()
		queryVars = append(queryVars, hex.EncodeToString(fromAccount[:]), (-1)*int(tx.Amount), int(tx.Amount), 0, (-1)*int(tx.Amount), int(tx.Amount), 0)
		err = d.insertRow("accounts", queryVars)
		if err != nil {
			continue
		}
	}

	for _, deposit := range block.Deposits {
		queryVars = nil
		queryVars = append(queryVars, hash.String(), hex.EncodeToString(deposit.PublicKey[:]), hex.EncodeToString(deposit.Signature[:]),
			hex.EncodeToString(deposit.Data.PublicKey[:]), hex.EncodeToString(deposit.Data.ProofOfPossession[:]), hex.EncodeToString(deposit.Data.WithdrawalAddress[:]))
		err = d.insertRow("deposits", queryVars)
		if err != nil {
			d.log.Error(err)
			continue
		}
	}

	// Exits
	for _, exits := range block.Exits {
		queryVars = nil
		queryVars = append(queryVars, hash.String(), hex.EncodeToString(exits.ValidatorPubkey[:]), hex.EncodeToString(exits.WithdrawPubkey[:]),
			hex.EncodeToString(exits.Signature[:]))
		err = d.insertRow("exits", queryVars)
		if err != nil {
			d.log.Error(err)
			continue
		}
	}

	// Vote Slashings
	for _, vs := range block.VoteSlashings {
		queryVars = nil
		queryVars = append(queryVars, hash.String(), vs.Vote1.Data.Hash().String(), vs.Vote2.Data.Hash().String())
		err = d.insertRow("vote_slashings", queryVars)
		if err != nil {
			d.log.Error(err)
			continue
		}
	}

	// RANDAO Slashings
	for _, rs := range block.RANDAOSlashings {
		queryVars = nil
		queryVars = append(queryVars, hash.String(), hex.EncodeToString(rs.RandaoReveal[:]), int(rs.Slot), hex.EncodeToString(rs.ValidatorPubkey[:]))
		err = d.insertRow("randao_slashings", queryVars)
		if err != nil {
			d.log.Error(err)
			continue
		}
	}

	// Proposer Slashings
	for _, ps := range block.ProposerSlashings {
		queryVars = nil
		queryVars = append(queryVars, hash.String(), ps.BlockHeader1.Hash().String(), ps.BlockHeader2.Hash().String(), hex.EncodeToString(ps.Signature1[:]),
			hex.EncodeToString(ps.Signature2[:]), hex.EncodeToString(ps.ValidatorPublicKey[:]))
		err = d.insertRow("proposer_slashings", queryVars)
		if err != nil {
			d.log.Error(err)
			continue
		}
	}

	return nil
}

func (d *Database) getNextHeight() (int, string, error) {
	dw := goqu.Dialect(d.driver)
	ds := dw.From("blocks").Select(goqu.MAX("height"))
	query, _, err := ds.ToSQL()
	if err != nil {
		return -1, "", err
	}

	// This will fail when the db is empty return 0 to load genesis
	var height string
	err = d.db.QueryRow(query).Scan(&height)
	if err != nil {
		return 0, "", nil
	}
	heightNum, err := strconv.Atoi(height)
	if err != nil {
		return -1, "", err
	}

	dw = goqu.Dialect(d.driver)
	ds = dw.From("blocks").Select("block_hash").Where(goqu.C("height").Eq(height))
	query, _, err = ds.ToSQL()
	if err != nil {
		return heightNum + 1, "", err
	}
	var blockhash string
	err = d.db.QueryRow(query).Scan(&blockhash)
	if err != nil {
		return heightNum + 1, "", err
	}

	return heightNum + 1, blockhash, nil
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
	case "tx_single":
		return d.insertTxSingle(queryVars)
	case "deposits":
		return d.insertDeposit(queryVars)
	case "exits":
		return d.insertExit(queryVars)
	case "vote_slashings":
		return d.insertVoteSlashing(queryVars)
	case "randao_slashings":
		return d.insertRandaoSlashing(queryVars)
	case "proposer_slashings":
		return d.insertProposerSlashing(queryVars)
	case "accounts":
		return d.modifyAccountRow(queryVars)
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
			"tx_merkle_root":                queryVars[3],
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

func (d *Database) insertTxSingle(queryVars []interface{}) error {
	dw := goqu.Dialect(d.driver)
	ds := dw.Insert("tx_single").Rows(
		goqu.Record{
			"hash":            queryVars[0],
			"block_hash":      queryVars[1],
			"tx_type":         queryVars[2],
			"to_addr":         queryVars[3],
			"from_public_key": queryVars[4],
			"amount":          queryVars[5],
			"nonce":           queryVars[6],
			"fee":             queryVars[7],
			"signature":       queryVars[8],
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

func (d *Database) insertDeposit(queryVars []interface{}) error {
	dw := goqu.Dialect(d.driver)
	ds := dw.Insert("deposits").Rows(
		goqu.Record{
			"block_hash":               queryVars[0],
			"public_key":               queryVars[1],
			"signature":                queryVars[2],
			"data_public_key":          queryVars[3],
			"data_proof_of_possession": queryVars[4],
			"data_withdrawal_address":  queryVars[5],
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

	return d.addValidator(queryVars[3])
}

func (d *Database) insertExit(queryVars []interface{}) error {
	dw := goqu.Dialect(d.driver)
	ds := dw.Insert("exits").Rows(
		goqu.Record{
			"block_hash":            queryVars[0],
			"validator_public_key":  queryVars[1],
			"withdrawal_public_key": queryVars[2],
			"signature":             queryVars[3],
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
	return d.exitValidator(queryVars[1])
}

func (d *Database) insertVoteSlashing(queryVars []interface{}) error {
	dw := goqu.Dialect(d.driver)
	ds := dw.Insert("vote_slashings").Rows(
		goqu.Record{
			"block_hash": queryVars[0],
			"vote_1":     queryVars[1],
			"vote_2":     queryVars[2],
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

func (d *Database) insertRandaoSlashing(queryVars []interface{}) error {
	dw := goqu.Dialect(d.driver)
	ds := dw.Insert("randao_slashing").Rows(
		goqu.Record{
			"block_hash":           queryVars[0],
			"randao_reveal":        queryVars[1],
			"slot":                 queryVars[2],
			"validator_public_key": queryVars[3],
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
	return d.exitPenalizeValidator(queryVars[3])
}

func (d *Database) insertProposerSlashing(queryVars []interface{}) error {
	dw := goqu.Dialect(d.driver)
	ds := dw.Insert("proposer_slashing").Rows(
		goqu.Record{
			"block_hash":           queryVars[0],
			"blockheader_1":        queryVars[1],
			"blockheader_2":        queryVars[2],
			"signature_1":          queryVars[3],
			"signature_2":          queryVars[4],
			"validator_public_key": queryVars[5],
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
	return d.exitPenalizeValidator(queryVars[5])
}

func (d *Database) addValidator(valPubKey interface{}) error {
	dw := goqu.Dialect(d.driver)
	ds := dw.Insert("validators").Rows(
		goqu.Record{
			"public_key": valPubKey,
			"exit":       false,
			"penalized":  false,
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

func (d *Database) exitValidator(valPubKey interface{}) error {
	dw := goqu.Dialect(d.driver)
	ds := dw.Update("validators").Set(
		goqu.Record{
			"exit": true,
		}).Where(
		goqu.Ex{
			"public_key": valPubKey,
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

func (d *Database) exitPenalizeValidator(valPubKey interface{}) error {
	dw := goqu.Dialect(d.driver)
	ds := dw.Update("validators").Set(
		goqu.Record{
			"exit":      true,
			"penalized": true,
		}).Where(
		goqu.Ex{
			"public_key": valPubKey,
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

func (d *Database) modifyAccountRow(queryVars []interface{}) error {
	// TODO
	// Modify query to match:
	// "insert_accounts": "insert into accounts(addr, balance, total_sent, total_received) values(?,?,?,?) on conflict(addr) do update set balance=balance+?, total_sent=total_sent+?,total_received=total_received+?;"

	//dw := goqu.Dialect(d.driver)
	//ds := dw.Insert("accounts").Rows(
	//	goqu.Record{
	//		"block_hash":           queryVars[0],
	//		"blockheader_1":        queryVars[1],
	//		"blockheader_2":        queryVars[2],
	//		"signature_1":          queryVars[3],
	//		"signature_2":          queryVars[4],
	//		"validator_public_key": queryVars[5],
	//	})
	//
	//query, _, err := ds.ToSQL()
	//if err != nil {
	//	return err
	//}
	//_, err = d.db.Exec(query)
	//if err != nil {
	//	return err
	//}
	return nil
}

func (d *Database) Close() {
	d.canClose.Wait()
	_ = d.db.Close()
	return
}

func (d *Database) Migrate() error {
	var dbdriver database.Driver
	var err error
	var migrationsString string
	switch d.driver {
	case "sqlite3":
		migrationsString = "file://cmd/ogen/indexer/db/migrations/sqlite3"
		dbdriver, err = sqlite3.WithInstance(d.db, &sqlite3.Config{})
		if err != nil {
			return err
		}
	case "postgres":
		migrationsString = "file://cmd/ogen/indexer/db/migrations/postgres"
		dbdriver, err = postgres.WithInstance(d.db, &postgres.Config{})
		if err != nil {
			return err
		}
	case "mysql":
		migrationsString = "file://cmd/ogen/indexer/db/migrations/mysql"
		dbdriver, err = mysql.WithInstance(d.db, &mysql.Config{})
		if err != nil {
			return err
		}
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrationsString,
		d.driver,
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
func NewDB(dbConnString string, log logger.Logger, wg *sync.WaitGroup, driver string) *Database {
	db, err := sql.Open(driver, dbConnString)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
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
