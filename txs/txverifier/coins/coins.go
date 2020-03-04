package coins_txverifier

import (
	"bytes"
	"errors"
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/state"
	"github.com/olympus-protocol/ogen/txs/txpayloads"
	coins_txpayload "github.com/olympus-protocol/ogen/txs/txpayloads/coins"
	"github.com/olympus-protocol/ogen/utils/amount"
	"reflect"
	"sync"
)

var (
	ErrorNoTypeSpecified  = errors.New("coinTx-no-type-rules: the selected action is not specified on the tx verifier scheme")
	ErrorInvalidSignature = errors.New("coinTx-invalid-signature: the signature verification is invalid")
	ErrorMatchDataNoExist = errors.New("coinTx-not-found-match-data: the match data searched doesn't exists")
	ErrorDataNoMatch      = errors.New("coinTx-invalid-match-data: the data used to sign the transaction doesn't match")
	ErrorSpentTooMuch     = errors.New("coinTx-spend-exceed: the outputs sum is greater than utxos values")
	ErrorSpentParse       = errors.New("coinTx-spend-parse: unable to calculate spent amount")
	ErrorGetAggPubKey     = errors.New("coinTx-get-agg-pub-key: unable to get aggregated public key")
	ErrorGetMsg           = errors.New("coinTx-get-sig-msg: unable to get signature message")
	ErrorGetSig           = errors.New("coinTx-get-sig: unable to get signature")
)

type CoinsTxVerifier struct {
	state  *state.State
	params *params.ChainParams
}

func (v CoinsTxVerifier) DeserializePayload(payload []byte, Action p2p.TxAction) (txpayloads.Payload, error) {
	var Payload txpayloads.Payload
	switch Action {
	case p2p.Transfer:
		Payload = new(coins_txpayload.PayloadTransfer)
	case p2p.Generate:
		Payload = new(coins_txpayload.PayloadGenerate)
	case p2p.Pay:
		Payload = new(coins_txpayload.PayloadPay)
	default:
		return nil, ErrorNoTypeSpecified
	}
	buf := bytes.NewBuffer(payload)
	err := Payload.Deserialize(buf)
	if err != nil {
		return nil, err
	}
	return Payload, nil
}

func (v CoinsTxVerifier) SigVerify(payload []byte, Action p2p.TxAction) error {
	switch Action {
	case p2p.Transfer:
		VerPayload, err := v.DeserializePayload(payload, Action)
		if err != nil {
			return err
		}
		aggPubKey, err := VerPayload.GetAggPubKey()
		if err != nil {
			return ErrorGetAggPubKey
		}
		msg, err := VerPayload.GetMessage()
		if err != nil {
			return ErrorGetMsg
		}
		sig, err := VerPayload.GetSignature()
		if err != nil {
			return ErrorGetSig
		}
		valid, err := bls.VerifySig(aggPubKey, msg, sig)
		if err != nil {
			return err
		}
		if !valid {
			return ErrorInvalidSignature
		}
	case p2p.Generate:
	case p2p.Pay:
	}

	return nil
}

type routineRes struct {
	Err error
}

func (v CoinsTxVerifier) SigVerifyBatch(payload [][]byte, Action p2p.TxAction) error {
	var wg sync.WaitGroup
	doneChan := make(chan routineRes, len(payload))
	for _, singlePayload := range payload {
		wg.Add(1)
		go func(wg *sync.WaitGroup, payload []byte) {
			defer wg.Done()
			var resp routineRes
			err := v.SigVerify(payload, Action)
			if err != nil {
				resp.Err = err
			}
			doneChan <- resp
			return
		}(&wg, singlePayload)
	}
	wg.Wait()
	doneRes := <-doneChan
	if doneRes.Err != nil {
		return doneRes.Err
	}
	return nil
}

func (v CoinsTxVerifier) MatchVerify(payload []byte, Action p2p.TxAction) error {
	VerPayload, err := v.DeserializePayload(payload, Action)
	if err != nil {
		return err
	}
	switch Action {
	case p2p.Transfer:
		hashInv, err := VerPayload.GetHashInvForDataMatch()
		if err != nil {
			return err
		}
		var owners []string
		var spendable amount.AmountType
		for _, hash := range hashInv {
			ok := v.state.UtxoState.Have(hash)
			if !ok {
				return ErrorMatchDataNoExist
			}
			utxo := v.state.UtxoState.Get(hash)
			spendable += amount.AmountType(utxo.Amount)
			owners = append(owners, utxo.Owner)
		}
		spent, err := VerPayload.GetSpentAmount()
		if err != nil {
			return ErrorSpentParse
		}
		if spent > spendable {
			return ErrorSpentTooMuch
		}
		pubKeys, err := VerPayload.GetPublicKeys()
		if err != nil {
			return err
		}
		var verifyOwnerAddr []string
		for _, pubKey := range pubKeys {
			pubKeyHash, err := pubKey.ToBech32(v.params.AddressPrefixes, false)
			if err != nil {
				return err
			}
			verifyOwnerAddr = append(verifyOwnerAddr, pubKeyHash)
		}
		equal := reflect.DeepEqual(owners, verifyOwnerAddr)
		if !equal {
			return ErrorDataNoMatch
		}

	case p2p.Generate:
	case p2p.Pay:
	}
	return nil
}

func (v CoinsTxVerifier) MatchVerifyBatch(payload [][]byte, Action p2p.TxAction) error {
	var wg sync.WaitGroup
	doneChan := make(chan routineRes, len(payload))
	for _, singlePayload := range payload {
		wg.Add(1)
		go func(wg *sync.WaitGroup, payload []byte) {
			defer wg.Done()
			var resp routineRes
			err := v.MatchVerify(payload, Action)
			if err != nil {
				resp.Err = err
			}
			doneChan <- resp
			return
		}(&wg, singlePayload)
	}
	wg.Wait()
	doneRes := <-doneChan
	if doneRes.Err != nil {
		return doneRes.Err
	}
	return nil
}

func NewCoinsTxVerifier(state *state.State, params *params.ChainParams) CoinsTxVerifier {
	v := CoinsTxVerifier{
		state:  state,
		params: params,
	}
	return v
}
