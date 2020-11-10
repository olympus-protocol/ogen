package db

import (
	"database/sql"
	"encoding/hex"
	"errors"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/olympus-protocol/ogen/api/proto"
	"github.com/olympus-protocol/ogen/cmd/ogen/initialization"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/bech32"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/params"

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
	log       logger.Logger
	db        *sql.DB
	canClose  *sync.WaitGroup
	driver    string
	netParams *params.ChainParams
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

func (d *Database) InsertBlock(block *primitives.Block) error {
	// TODO fix, initialize epoch/slot tables for non produced block tables.

	blockSlot := int64(block.Header.Slot) - 1

	currentEpoch := blockSlot / int64(d.netParams.EpochLength)

	// Check if the remainder of the blockSlot - 1 and the EpochLength is 0, if it is, then we are on a epoch transition
	if blockSlot > 0 && blockSlot%int64(d.netParams.EpochLength) == 0 {
		epoch := blockSlot / int64(d.netParams.EpochLength)
		err := d.initializeEpoch(epoch + 1)
		if err != nil {
			return err
		}
	}

	// Check if this block finishes an epoch and creates a new one

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

	fee := 0
	for _, tx := range block.Txs {
		fee += int(tx.Fee)
	}
	var feeReceiver = &AccountInfo{
		Account:       hex.EncodeToString(block.Header.FeeAddress[:]),
		Confirmed:     fee,
		TotalReceived: fee,
	}

	err = d.modifyAccountRow(feeReceiver)
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
		pkh, err := tx.FromPubkeyHash()
		if err != nil {
			d.log.Error(err)
			continue
		}

		var receiverAccInfo = &AccountInfo{
			Account:       hex.EncodeToString(tx.To[:]),
			Confirmed:     int(tx.Amount),
			TotalReceived: int(tx.Amount),
		}

		var senderAccInfo = &AccountInfo{
			Account:   hex.EncodeToString(pkh[:]),
			Confirmed: -1 * int(tx.Amount+tx.Fee),
			TotalSent: int(tx.Amount + tx.Fee),
		}

		err = d.modifyAccountRow(receiverAccInfo)
		if err != nil {
			d.log.Error(err)
			continue
		}

		err = d.modifyAccountRow(senderAccInfo)
		if err != nil {
			d.log.Error(err)
			continue
		}

		queryVars = append(queryVars, tx.Hash().String(), hash.String(), 0, hex.EncodeToString(tx.To[:]), hex.EncodeToString(tx.FromPublicKey[:]), hex.EncodeToString(pkh[:]),
			int(tx.Amount), int(tx.Nonce), int(tx.Fee), hex.EncodeToString(tx.Signature[:]))
		err = d.insertRow("tx_single", queryVars)
		if err != nil {
			d.log.Error(err)
			continue
		}

	}

	for _, deposit := range block.Deposits {

		var lockedAccountInfo = &AccountInfo{
			Account:   hex.EncodeToString(deposit.Data.WithdrawalAddress[:]),
			Confirmed: -1 * int(100*1e8),
			Locked:    1 * int(100*1e8),
		}

		err = d.modifyAccountRow(lockedAccountInfo)
		if err != nil {
			d.log.Error(err)
			continue
		}

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

		balance, err := d.GetValidatorBalance(hex.EncodeToString(exits.ValidatorPubkey[:]))
		if err != nil {
			d.log.Error(err)
			continue
		}
		var pkh [20]byte
		pkhHash := chainhash.HashB(exits.WithdrawPubkey[:])
		copy(pkh[:], pkhHash[:])

		var unlockedAccountInfo = &AccountInfo{
			Account:   hex.EncodeToString(pkh[:]),
			Confirmed: balance,
			Locked:    -1 * balance,
		}

		err = d.modifyAccountRow(unlockedAccountInfo)
		if err != nil {
			d.log.Error(err)
			continue
		}

		queryVars = nil
		queryVars = append(queryVars, hash.String(), hex.EncodeToString(exits.ValidatorPubkey[:]), hex.EncodeToString(exits.WithdrawPubkey[:]),
			hex.EncodeToString(exits.Signature[:]), currentEpoch)
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

	err = d.ProcessSlot(block)
	if err != nil {
		return err
	}

	err = d.ProcessEpoch(block, currentEpoch)
	if err != nil {
		return err
	}

	return nil
}

func (d *Database) initializeEpoch(epoch int64) error {

	lastSlotFromEpoch := int(epoch)*int(d.netParams.EpochLength) + 1

	slots := make([]int, 5)

	slots[4] = lastSlotFromEpoch

	for i := 0; i < int(d.netParams.EpochLength)-1; i++ {
		slots[i] = lastSlotFromEpoch - (int(d.netParams.EpochLength) - (i + 1))
	}

	//proposers, err := d.getEpochProposers(epoch)
	//if err != nil {
	//	return err
	//}

	for _, slot := range slots {
		d.log.Infof("initializing slot %d", slot)

		dw := goqu.Dialect(d.driver)
		ds := dw.Insert("slots").Rows(
			goqu.Record{
				"slot":           slot,
				"block_hash":     "",
				"proposer_index": -1,
				"proposed":       false,
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

	}

	hash := chainhash.Hash{}

	d.log.Infof("initializing epoch %d", epoch)
	dw := goqu.Dialect(d.driver)
	ds := dw.Insert("epochs").Rows(
		goqu.Record{
			"epoch":                    int(epoch),
			"slot_1":                   slots[0],
			"slot_2":                   slots[1],
			"slot_3":                   slots[2],
			"slot_4":                   slots[3],
			"slot_5":                   slots[4],
			"participation_percentage": 0,
			"finalized":                false,
			"justified":                false,
			"randao":                   hash.String(),
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

func (d *Database) getEpochProposers(epoch int64) ([]uint64, error) {
	validators, err := d.GetActiveValidatorIndices()
	if err != nil {
		return nil, err
	}

	randao, err := d.getEpochRandao(epoch - 1)
	if err != nil {
		return nil, err
	}
	return state.DetermineNextProposers(randao, validators), nil
}

func (d *Database) getEpochRandao(epoch int64) (chainhash.Hash, error) {
	if epoch == 0 {
		return chainhash.Hash{}, nil
	}
	dw := goqu.Dialect(d.driver)
	ds := dw.From("epochs").Select("randao").Where(goqu.Ex{
		"epoch": int(epoch),
	})
	query, _, err := ds.ToSQL()
	if err != nil {
		return chainhash.Hash{}, err
	}
	var hashB []byte
	err = d.db.QueryRow(query).Scan(&hashB)
	if err != nil {
		return chainhash.Hash{}, err
	}
	var hash [32]byte
	copy(hash[:], hashB)
	return hash, nil
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
			"hash":                 queryVars[0],
			"block_hash":           queryVars[1],
			"tx_type":              queryVars[2],
			"to_addr":              queryVars[3],
			"from_public_key":      queryVars[4],
			"from_public_key_hash": queryVars[5],
			"amount":               queryVars[6],
			"nonce":                queryVars[7],
			"fee":                  queryVars[8],
			"signature":            queryVars[9],
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

	return d.addValidator(queryVars[3], queryVars[5], nil)
}

func (d *Database) insertMockDeposit(genesisHash string, valPubKey string, preminepkh string) error {
	emptySig := [96]byte{}
	emptyPub := [48]byte{}

	dw := goqu.Dialect(d.driver)
	ds := dw.Insert("deposits").Rows(
		goqu.Record{
			"block_hash":               genesisHash,
			"public_key":               hex.EncodeToString(emptyPub[:]),
			"signature":                hex.EncodeToString(emptySig[:]),
			"data_public_key":          valPubKey,
			"data_proof_of_possession": hex.EncodeToString(emptySig[:]),
			"data_withdrawal_address":  preminepkh,
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
	return d.exitValidator(queryVars[1], queryVars[4])
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

func (d *Database) addValidator(valPubKey interface{}, payee interface{}, status interface{}) error {
	var addStatus uint64
	if status == nil {
		addStatus = primitives.StatusStarting
	} else {
		statusUint, ok := status.(uint64)
		if !ok {
			return errors.New("wrong status interface parse")
		}
		addStatus = statusUint
	}

	dw := goqu.Dialect(d.driver)
	ds := dw.Insert("validators").Rows(
		goqu.Record{
			"public_key":         valPubKey,
			"status":             int(addStatus),
			"balance":            100 * 1e8,
			"exit":               false,
			"penalized":          false,
			"payee_address":      payee,
			"first_active_epoch": 0,
			"last_active_epoch":  0,
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

func (d *Database) exitValidator(valPubKey interface{}, epoch interface{}) error {
	dw := goqu.Dialect(d.driver)
	ds := dw.Update("validators").Set(
		goqu.Record{
			"exit":              true,
			"last_active_epoch": epoch,
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

func (d *Database) modifyAccountRowForMempoolTxs(accInfo *AccountInfo) error {
	dw := goqu.Dialect(d.driver)

	ds := dw.From("accounts").Select("*").Where(goqu.Ex{
		"account": accInfo.Account,
	})

	query, _, err := ds.ToSQL()
	if err != nil {
		return err
	}

	row := d.db.QueryRow(query)

	var accountResult AccountInfo

	err = row.Scan(&accountResult.Account, &accountResult.Confirmed, &accountResult.Unconfirmed, &accountResult.Locked, &accountResult.TotalSent, &accountResult.TotalReceived)
	if err != nil {
		if err == sql.ErrNoRows {
			ds := dw.Insert("accounts").Rows(
				goqu.Record{
					"account":        accInfo.Account,
					"confirmed":      0,
					"unconfirmed":    accInfo.Unconfirmed,
					"locked":         0,
					"total_sent":     0,
					"total_received": 0,
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
		return err
	}

	newAccountData := &AccountInfo{
		Account:     accInfo.Account,
		Unconfirmed: accountResult.Unconfirmed + accInfo.Unconfirmed,
	}

	nds := dw.Update("accounts").Set(
		goqu.Record{
			"unconfirmed": newAccountData.Unconfirmed,
		}).Where(
		goqu.Ex{
			"account": accountResult.Account,
		},
	)

	nquery, _, err := nds.ToSQL()
	if err != nil {
		return err
	}

	_, err = d.db.Exec(nquery)
	if err != nil {
		return err
	}

	return nil
}

func (d *Database) modifyAccountRow(accInfo *AccountInfo) error {

	dw := goqu.Dialect(d.driver)

	ds := dw.From("accounts").Select("*").Where(goqu.Ex{
		"account": accInfo.Account,
	})

	query, _, err := ds.ToSQL()
	if err != nil {
		return err
	}

	row := d.db.QueryRow(query)

	var accountResult AccountInfo

	err = row.Scan(&accountResult.Account, &accountResult.Confirmed, &accountResult.Unconfirmed, &accountResult.Locked, &accountResult.TotalSent, &accountResult.TotalReceived)
	if err != nil {
		if err == sql.ErrNoRows {
			ds := dw.Insert("accounts").Rows(
				goqu.Record{
					"account":        accInfo.Account,
					"confirmed":      accInfo.Confirmed,
					"total_sent":     accInfo.TotalSent,
					"total_received": accInfo.TotalReceived,
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
		return err
	}

	newAccountData := &AccountInfo{
		Account:       accInfo.Account,
		Confirmed:     accountResult.Confirmed + accInfo.Confirmed,
		Locked:        accountResult.Locked + accInfo.Locked,
		TotalSent:     accountResult.TotalSent + accInfo.TotalSent,
		TotalReceived: accountResult.TotalReceived + accInfo.TotalReceived,
	}

	if accountResult.Unconfirmed != 0 {
		newAccountData.Unconfirmed = accountResult.Unconfirmed - accInfo.Confirmed
	}

	nds := dw.Update("accounts").Set(
		goqu.Record{
			"confirmed":      newAccountData.Confirmed,
			"unconfirmed":    newAccountData.Unconfirmed,
			"locked":         newAccountData.Locked,
			"total_sent":     newAccountData.TotalSent,
			"total_received": newAccountData.TotalReceived,
		}).Where(
		goqu.Ex{
			"account": accountResult.Account,
		},
	)
	nquery, _, err := nds.ToSQL()
	if err != nil {
		return err
	}
	_, err = d.db.Exec(nquery)
	if err != nil {
		return err
	}

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

// ProcessSlot process the block slot and modifies the database.
func (d *Database) ProcessSlot(b *primitives.Block) error {
	return nil
}

// ProcessEpoch process the block information and modifies the epoch information.
func (d *Database) ProcessEpoch(b *primitives.Block, epoch int64) error {
	//
	//epochRandao, err := d.getEpochRandao(epoch)
	//if err != nil {
	//	return err
	//}
	//
	//for i := range epochRandao {
	//	epochRandao[i] ^= b.RandaoSignature[i]
	//}
	//
	//dw := goqu.Dialect(d.driver)
	//ds := dw.Update("epochs").Set(
	//	goqu.Record{
	//		"randao": hex.EncodeToString(epochRandao[:]),
	//	}).Where(
	//	goqu.Ex{
	//		"epoch": int(epoch),
	//	},
	//)
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

func (d *Database) GetValidatorBalance(pubkey string) (int, error) {
	dw := goqu.Dialect(d.driver)

	ds := dw.From("validators").Select("balance").Where(
		goqu.Ex{
			"public_key": pubkey,
		},
	)

	query, _, err := ds.ToSQL()
	if err != nil {
		return 0, err
	}

	row := d.db.QueryRow(query)

	var balance int
	err = row.Scan(&balance)
	if err != nil {
		return 0, err
	}
	return balance, nil
}

func (d *Database) GetActiveValidatorIndices() ([]uint64, error) {
	dw := goqu.Dialect(d.driver)

	ds := dw.From("validators").Select("id").Where(
		goqu.Ex{
			"status": int(primitives.StatusActive),
		},
	)

	query, _, err := ds.ToSQL()
	if err != nil {
		return nil, err
	}

	row, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}

	var indexes []uint64
	for {
		if !row.Next() {
			break
		}
		var index uint64
		err = row.Scan(&index)
		if err != nil {
			return nil, err
		}
		indexes = append(indexes, index)
	}

	return indexes, nil
}

func (d *Database) ModifyValidatorBalance(pubkey string, balance int) error {
	currBalance, err := d.GetValidatorBalance(pubkey)
	if err != nil {
		return err
	}

	dw := goqu.Dialect(d.driver)

	ds := dw.Update("validators").Set(
		goqu.Record{
			"balance": currBalance + balance,
		}).Where(
		goqu.Ex{
			"public_key": pubkey,
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

func (d *Database) ProcessMempoolTransaction(tx *proto.Tx) {

	publicKey, err := hex.DecodeString(tx.FromPublicKey)
	if err != nil {
		d.log.Error(err)
	}
	var pkh [20]byte
	hash := chainhash.HashB(publicKey)
	copy(pkh[:], hash[:])

	senderInfo := &AccountInfo{
		Account:     hex.EncodeToString(pkh[:]),
		Unconfirmed: -1 * int(tx.Amount+tx.Fee),
	}

	err = d.modifyAccountRowForMempoolTxs(senderInfo)
	if err != nil {
		d.log.Error(err)
	}

	receiverInfo := &AccountInfo{
		Account:     tx.To,
		Unconfirmed: int(tx.Amount),
	}

	err = d.modifyAccountRowForMempoolTxs(receiverInfo)
	if err != nil {
		d.log.Error(err)
	}

}

func (d *Database) Initialize() (string, error) {
	init, err := initialization.LoadParams(d.netParams.Name)
	if err != nil {
		return "", err
	}

	// Add genesis
	genesis := primitives.GetGenesisBlock()
	err = d.InsertBlock(&genesis)
	if err != nil {
		return "", err
	}
	genesisHash := genesis.Hash()

	_, premine, err := bech32.Decode(init.PremineAddress)
	if err != nil {
		return "", err
	}

	// This is just for testing purposes TODO: remove for production.
	// Add 400,000 tPOLIS hardcoded in the state
	premineAddr := &AccountInfo{
		Account:       hex.EncodeToString(premine[:]),
		Confirmed:     int(400000 * d.netParams.UnitsPerCoin),
		TotalReceived: int(400000 * d.netParams.UnitsPerCoin),
	}
	err = d.modifyAccountRow(premineAddr)
	if err != nil {
		return "", err
	}

	// Add validators to the validator registry
	for _, v := range init.Validators {
		err = d.insertMockDeposit(genesisHash.String(), v.PublicKey, hex.EncodeToString(premine[:]))
		if err != nil {
			return "", err
		}
		err = d.addValidator(v.PublicKey, hex.EncodeToString(premine[:]), primitives.StatusActive)
		if err != nil {
			return "", err
		}
	}

	// Add first epoch and slots
	err = d.initializeEpoch(1)
	if err != nil {
		return "", err
	}
	return genesisHash.String(), nil
}

// NewDB creates a db client
func NewDB(dbConnString string, log logger.Logger, wg *sync.WaitGroup, driver string, netParams *params.ChainParams) *Database {
	db, err := sql.Open(driver, dbConnString)
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
		driver:    driver,
		netParams: netParams,
	}

	dbclient.log.Info("Database connection established")

	return dbclient
}
