package gov_txverifier

import (
	"bytes"
	"errors"
	"github.com/grupokindynos/ogen/bls"
	"github.com/grupokindynos/ogen/chain/index"
	"github.com/grupokindynos/ogen/p2p"
	"github.com/grupokindynos/ogen/params"
	"github.com/grupokindynos/ogen/txs/txpayloads"
	gov_txpayload "github.com/grupokindynos/ogen/txs/txpayloads/gov"
	"github.com/grupokindynos/ogen/utils/amount"
	"reflect"
	"sync"
)

var (
	ErrorNoTypeSpecified    = errors.New("govTx-no-type-rules: the selected action is not specified on the tx verifier scheme")
	ErrorInvalidSignature   = errors.New("govTx-invalid-signature: the signature verification is invalid")
	ErrorMatchDataNoExist   = errors.New("govTx-not-found-match-data: the match data searched doesn't exists")
	ErrorDataNoMatch        = errors.New("govTx-invalid-match-data: the data used to sign the transaction doesn't match")
	ErrorOwnerShouldBeEmpty = errors.New("govTx-invalid-payed-utxo: the utxo should be payed to network")
	ErrorUtxoAmountToLow    = errors.New("govTx-invalid-utxo-amount: the utxo fee is too low")
)

type GovTxVerifier struct {
	GovIndex  *index.GovIndex
	UtxoIndex *index.UtxosIndex
	params    *params.ChainParams
}

func (v GovTxVerifier) DeserializePayload(payload []byte, Action p2p.TxAction) (txpayloads.Payload, error) {
	var Payload txpayloads.Payload
	switch Action {
	case p2p.Upload:
		Payload = new(gov_txpayload.PayloadUpload)
	case p2p.Revoke:
		Payload = new(gov_txpayload.PayloadRevoke)
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

func (v GovTxVerifier) SigVerify(payload []byte, Action p2p.TxAction) error {
	VerPayload, err := v.DeserializePayload(payload, Action)
	if err != nil {
		return err
	}
	pubKey, err := VerPayload.GetPublicKey()
	if err != nil {
		return err
	}
	msg, err := VerPayload.GetMessage()
	if err != nil {
		return err
	}
	sig, err := VerPayload.GetSignature()
	if err != nil {
		return err
	}
	valid, err := bls.VerifySig(pubKey, msg, sig)
	if err != nil {
		return err
	}
	if !valid {
		return ErrorInvalidSignature
	}
	return nil
}

type routineRes struct {
	Err error
}

func (v GovTxVerifier) SigVerifyBatch(payload [][]byte, Action p2p.TxAction) error {
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

func (v GovTxVerifier) MatchVerify(payload []byte, Action p2p.TxAction) error {
	VerPayload, err := v.DeserializePayload(payload, Action)
	if err != nil {
		return err
	}
	searchHash, err := VerPayload.GetHashForDataMatch()
	if err != nil {
		return err
	}
	switch Action {
	case p2p.Upload:
		ok := v.UtxoIndex.Have(searchHash)
		if !ok {
			return ErrorMatchDataNoExist
		}
		data := v.UtxoIndex.Get(searchHash)
		if data.Owner != "" {
			return ErrorOwnerShouldBeEmpty
		}
		if amount.AmountType(data.Amount) != v.params.GovernanceProposalFee {
			return ErrorUtxoAmountToLow
		}
		matchPubKey, err := VerPayload.GetPublicKey()
		if err != nil {
			return err
		}
		aggPubKey := bls.NewAggregatePublicKey()
		for _, prevInputPubKeys := range data.PrevInputsPubKeys {
			pubKey, err := bls.DeserializePublicKey(prevInputPubKeys)
			if err != nil {
				return err
			}
			aggPubKey.AggregatePubKey(pubKey)
		}
		equal := reflect.DeepEqual(aggPubKey, matchPubKey)
		if !equal {
			return ErrorDataNoMatch
		}
	case p2p.Revoke:
		ok := v.GovIndex.Have(searchHash)
		if !ok {
			return ErrorMatchDataNoExist
		}
		data := v.GovIndex.Get(searchHash)
		utxoHash, err := data.GovData.BurnedUtxo.Hash()
		if err != nil {
			return err
		}
		ok = v.UtxoIndex.Have(utxoHash)
		if !ok {
			return ErrorMatchDataNoExist
		}
		utxoData := v.UtxoIndex.Get(utxoHash)
		if utxoData.Owner != "" {
			return ErrorOwnerShouldBeEmpty
		}
		matchPubKey, err := VerPayload.GetPublicKey()
		if err != nil {
			return err
		}
		aggPubKey := bls.NewAggregatePublicKey()
		for _, prevInputPubKeys := range utxoData.PrevInputsPubKeys {
			pubKey, err := bls.DeserializePublicKey(prevInputPubKeys)
			if err != nil {
				return err
			}
			aggPubKey.AggregatePubKey(pubKey)
		}
		equal := reflect.DeepEqual(aggPubKey, matchPubKey)
		if !equal {
			return ErrorDataNoMatch
		}
	}
	return nil
}

func (v GovTxVerifier) MatchVerifyBatch(payload [][]byte, Action p2p.TxAction) error {
	var wg sync.WaitGroup
	doneChan := make(chan routineRes, len(payload))
	for _, singlePayload := range payload {
		wg.Add(1)
		go func(wg *sync.WaitGroup, payload []byte) {
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

func NewGovTxVerifier(govIndex *index.GovIndex, utxoIndex *index.UtxosIndex, params *params.ChainParams) GovTxVerifier {
	v := GovTxVerifier{
		GovIndex:  govIndex,
		UtxoIndex: utxoIndex,
		params:    params,
	}
	return v
}
