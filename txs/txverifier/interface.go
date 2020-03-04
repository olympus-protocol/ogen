package txverifier

import (
	"errors"
	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/state"
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
	SigVerify(payload []byte, Action p2p.TxAction) error
	SigVerifyBatch(payload [][]byte, Action p2p.TxAction) error
	MatchVerify(payload []byte, Action p2p.TxAction) error
	MatchVerifyBatch(payload [][]byte, Action p2p.TxAction) error
	DeserializePayload(payload []byte, Action p2p.TxAction) (txpayloads.Payload, error)
}

func (txv *TxVerifier) VerifyTx(tx *p2p.MsgTx) error {
	var verifier Verifier
	switch tx.TxType {
	case p2p.Coins:
		verifier = txv.coins
	case p2p.Governance:
		verifier = txv.gov
	case p2p.Users:
		verifier = txv.users
	case p2p.Votes:
		verifier = txv.votes
	case p2p.Worker:
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

func (txv *TxVerifier) VerifyTxsBatch(txs []*p2p.MsgTx, txTypes p2p.TxType, txAction p2p.TxAction) error {
	var verifier Verifier
	switch txTypes {
	case p2p.Coins:
		verifier = txv.coins
	case p2p.Governance:
		verifier = txv.gov
	case p2p.Users:
		verifier = txv.users
	case p2p.Votes:
		verifier = txv.votes
	case p2p.Worker:
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

func NewTxVerifier(currentState *state.State, params *params.ChainParams) *TxVerifier {
	return &TxVerifier{
		coins:   coins_txverifier.NewCoinsTxVerifier(currentState, params),
		gov:     gov_txverifier.NewGovTxVerifier(currentState, params),
		users:   users_txverifier.NewUsersTxVerifier(currentState, params),
		votes:   votes_txverifier.NewVotesTxVerifier(currentState, params),
		workers: workers_txverifier.NewWorkersTxVerifier(currentState, params),
	}
}
