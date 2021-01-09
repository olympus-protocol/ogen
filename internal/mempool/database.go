package mempool

import (
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/pkg/burnproof"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"go.etcd.io/bbolt"
)

var (
	depositsBkt        = []byte("deposits")
	exitsBkt           = []byte("exits")
	partialExitsBkt    = []byte("partial_exits")
	txsBkt             = []byte("txs")
	governanceVotesBkt = []byte("governance_votes")
	coinProofsBkt      = []byte("coin_proofs")
)

// Load fills up the pool from disk
func (p *pool) Load() {
	datapath := config.GlobalFlags.DataPath
	db, _ := bbolt.Open(datapath+"/mempool.db", 0700, nil)
	defer func() {
		_ = db.Close()
	}()
	_ = db.View(func(tx *bbolt.Tx) error {

		deposits := tx.Bucket(depositsBkt)
		if deposits != nil {
			_ = deposits.ForEach(func(k, v []byte) error {
				var h [32]byte
				copy(h[:], k[:])
				d := new(primitives.Deposit)
				_ = d.Unmarshal(v)
				_ = p.AddDeposit(d)
				return nil
			})
		}

		exits := tx.Bucket(exitsBkt)
		if exits != nil {
			_ = exits.ForEach(func(k, v []byte) error {
				d := new(primitives.Exit)
				_ = d.Unmarshal(v)
				_ = p.AddExit(d)
				return nil
			})
		}

		partialExits := tx.Bucket(partialExitsBkt)
		if partialExits != nil {
			_ = partialExits.ForEach(func(k, v []byte) error {
				d := new(primitives.PartialExit)
				_ = d.Unmarshal(v)
				_ = p.AddPartialExit(d)
				return nil
			})
		}

		txs := tx.Bucket(txsBkt)
		if txs != nil {
			_ = txs.ForEach(func(k, v []byte) error {
				d := new(primitives.Tx)
				_ = d.Unmarshal(v)
				_ = p.AddTx(d)
				return nil
			})
		}

		governanceVotes := tx.Bucket(governanceVotesBkt)
		if governanceVotes != nil {
			_ = governanceVotes.ForEach(func(k, v []byte) error {
				d := new(primitives.GovernanceVote)
				_ = d.Unmarshal(v)
				_ = p.AddGovernanceVote(d)
				return nil
			})
		}

		coinProofs := tx.Bucket(coinProofsBkt)
		if coinProofs != nil {
			_ = coinProofs.ForEach(func(k, v []byte) error {
				d := new(burnproof.CoinsProofSerializable)
				_ = d.Unmarshal(v)
				_ = p.AddCoinProof(d)
				return nil
			})
		}

		return nil
	})

	return
}

func (p *pool) Store() {
	datapath := config.GlobalFlags.DataPath
	db, _ := bbolt.Open(datapath+"/mempool", 0700, nil)
	defer func() {
		_ = db.Close()
	}()

	_ = db.Update(func(tx *bbolt.Tx) error {
		deposits, _ := tx.CreateBucketIfNotExists(depositsBkt)
		p.depositsLock.Lock()
		for k, d := range p.deposits {
			b, _ := d.Marshal()
			_ = deposits.Put(k[:], b)
		}
		p.depositsLock.Unlock()

		exits, _ := tx.CreateBucketIfNotExists(exitsBkt)
		p.exitsLock.Lock()
		for k, d := range p.exits {
			b, _ := d.Marshal()
			_ = exits.Put(k[:], b)
		}
		p.exitsLock.Unlock()

		partialExits, _ := tx.CreateBucketIfNotExists(partialExitsBkt)
		p.partialExitsLock.Lock()
		for k, d := range p.partialExits {
			b, _ := d.Marshal()
			_ = partialExits.Put(k[:], b)
		}
		p.partialExitsLock.Unlock()

		txs, _ := tx.CreateBucketIfNotExists(txsBkt)
		p.txsLock.Lock()
		for _, d := range p.txs {
			for _, tx := range d.transactions {
				h := tx.Hash()
				b, _ := tx.Marshal()
				_ = txs.Put(h[:], b)
			}
		}
		p.txsLock.Unlock()

		governanceVotes, _ := tx.CreateBucketIfNotExists(governanceVotesBkt)
		p.governanceVoteLock.Lock()
		for k, d := range p.governanceVotes {
			b, _ := d.Marshal()
			_ = governanceVotes.Put(k[:], b)
		}
		p.governanceVoteLock.Unlock()

		coinProofs, _ := tx.CreateBucketIfNotExists(coinProofsBkt)
		p.coinProofsLock.Lock()
		for k, d := range p.coinProofs {
			b, _ := d.Marshal()
			_ = coinProofs.Put(k[:], b)
		}
		p.coinProofsLock.Unlock()

		return nil
	})

	return
}
