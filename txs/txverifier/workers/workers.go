package workers_txverifier

import (
	"bytes"
	"errors"
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/state"
	"github.com/olympus-protocol/ogen/txs/txpayloads"
	workers_txpayload "github.com/olympus-protocol/ogen/txs/txpayloads/workers"
	"reflect"
	"sync"
)

var (
	ErrorNoTypeSpecified  = errors.New("workerTx-no-type-rules: the selected action is not specified on the tx verifier scheme")
	ErrorInvalidSignature = errors.New("workerTx-invalid-signature: the signature verification is invalid")
	ErrorMatchDataNoExist = errors.New("workerTx-not-found-match-data: the match data searched doesn't exists")
	ErrorDataNoMatch      = errors.New("workerTx-invalid-match-data: the data used to sign the transaction doesn't match")
)

type WorkersTxVerifier struct {
	params *params.ChainParams
	state  *state.State
}

func (v WorkersTxVerifier) DeserializePayload(payload []byte, Action primitives.TxAction) (txpayloads.Payload, error) {
	var Payload txpayloads.Payload
	switch Action {
	case primitives.Upload:
		Payload = new(workers_txpayload.PayloadUploadAndUpdate)
	case primitives.Update:
		Payload = new(workers_txpayload.PayloadUploadAndUpdate)
	case primitives.Revoke:
		Payload = new(workers_txpayload.PayloadRevoke)
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

func (v WorkersTxVerifier) SigVerify(payload []byte, Action primitives.TxAction) error {
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

func (v WorkersTxVerifier) SigVerifyBatch(payload [][]byte, Action primitives.TxAction) error {
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

func (v WorkersTxVerifier) MatchVerify(payload []byte, Action primitives.TxAction) error {
	VerPayload, err := v.DeserializePayload(payload, Action)
	if err != nil {
		return err
	}
	searchHash, err := VerPayload.GetHashForDataMatch()
	if err != nil {
		return err
	}
	matchPubKey, err := VerPayload.GetPublicKey()
	if err != nil {
		return err
	}
	switch Action {
	case primitives.Upload:
		ok := v.state.UtxoState.Have(searchHash)
		if !ok {
			return ErrorMatchDataNoExist
		}
		data := v.state.UtxoState.Get(searchHash)
		pubKeyHash, err := matchPubKey.ToBech32(v.params.AddressPrefixes, false)
		if err != nil {
			return err
		}
		equal := reflect.DeepEqual(pubKeyHash, data.Owner)
		if !equal {
			return ErrorDataNoMatch
		}
	case primitives.Update:
		ok := v.state.UtxoState.Have(searchHash)
		if !ok {
			return ErrorMatchDataNoExist
		}
		data := v.state.UtxoState.Get(searchHash)
		pubKeyHash, err := matchPubKey.ToBech32(v.params.AddressPrefixes, false)
		if err != nil {
			return err
		}
		equal := reflect.DeepEqual(pubKeyHash, data.Owner)
		if !equal {
			return ErrorDataNoMatch
		}
	case primitives.Revoke:
		ok := v.state.WorkerState.Have(searchHash)
		if !ok {
			return ErrorMatchDataNoExist
		}
		data := v.state.WorkerState.Get(searchHash)
		pubKeyBytes := matchPubKey.Serialize()
		equal := reflect.DeepEqual(pubKeyBytes, data.WorkerData.PubKey)
		if !equal {
			return ErrorDataNoMatch
		}
	}
	return nil
}

func (v WorkersTxVerifier) MatchVerifyBatch(payload [][]byte, Action primitives.TxAction) error {
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

func NewWorkersTxVerifier(state *state.State, params *params.ChainParams) WorkersTxVerifier {
	v := WorkersTxVerifier{
		state:  state,
		params: params,
	}
	return v
}
