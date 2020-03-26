package txverifier

import (
	"errors"

	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/txs/txpayloads"
	coins_txverifier "github.com/olympus-protocol/ogen/txs/txverifier/coins"
	gov_txverifier "github.com/olympus-protocol/ogen/txs/txverifier/gov"
	users_txverifier "github.com/olympus-protocol/ogen/txs/txverifier/users"
	votes_txverifier "github.com/olympus-protocol/ogen/txs/txverifier/votes"
	workers_txverifier "github.com/olympus-protocol/ogen/txs/txverifier/workers"
)

var (
	ErrorNoTxSchemeForType = errors.New("tx-verifier: error, no tx verification for this type")
)

type TxVerifier struct {
	coins   coins_txverifier.CoinsTxVerifier
	gov     gov_txverifier.GovTxVerifier
	users   users_txverifier.UsersTxVerifier
	votes   votes_txverifier.VotesTxVerifier
	workers workers_txverifier.WorkersTxVerifier
}

type Verifier interface {
	SigVerify(payload []byte, Action primitives.TxAction) error
	SigVerifyBatch(payload [][]byte, Action primitives.TxAction) error
	MatchVerify(payload []byte, Action primitives.TxAction) error
	MatchVerifyBatch(payload [][]byte, Action primitives.TxAction) error
	DeserializePayload(payload []byte, Action primitives.TxAction) (txpayloads.Payload, error)
}

func (txv *TxVerifier) VerifyTx(tx *p2p.MsgTx) error {
	var verifier Verifier
	switch tx.TxType {
	case primitives.TxCoins:
		verifier = txv.coins
	case primitives.TxGovernance:
		verifier = txv.gov
	case primitives.TxUsers:
		verifier = txv.users
	case primitives.TxVotes:
		verifier = txv.votes
	case primitives.TxWorker:
		verifier = txv.workers
	default:
		// TODO make sure we create a secure way to omit not known tx types
		return ErrorNoTxSchemeForType
	}
	err := verifier.MatchVerify(tx.Payload, tx.TxAction)
	if err != nil {
		return err
	}
	return verifier.SigVerify(tx.Payload, tx.TxAction)
}

func (txv *TxVerifier) VerifyTxsBatch(txs []primitives.Tx, txTypes primitives.TxType, txAction primitives.TxAction) error {
	var verifier Verifier
	switch txTypes {
	case primitives.TxCoins:
		verifier = txv.coins
	case primitives.TxGovernance:
		verifier = txv.gov
	case primitives.TxUsers:
		verifier = txv.users
	case primitives.TxVotes:
		verifier = txv.votes
	case primitives.TxWorker:
		verifier = txv.workers
	default:
		// TODO make sure we create a secure way to omit not known tx types
		return ErrorNoTxSchemeForType
	}
	var payloads [][]byte
	for _, tx := range txs {
		payloads = append(payloads, tx.Payload)
	}
	err := verifier.MatchVerifyBatch(payloads, txAction)
	if err != nil {
		return err
	}
	return verifier.SigVerifyBatch(payloads, txAction)
}

func NewTxVerifier(currentState *primitives.State, params *params.ChainParams) *TxVerifier {
	return &TxVerifier{
		coins:   coins_txverifier.NewCoinsTxVerifier(currentState, params),
		gov:     gov_txverifier.NewGovTxVerifier(currentState, params),
		users:   users_txverifier.NewUsersTxVerifier(currentState, params),
		votes:   votes_txverifier.NewVotesTxVerifier(currentState, params),
		workers: workers_txverifier.NewWorkersTxVerifier(currentState, params),
	}
}
