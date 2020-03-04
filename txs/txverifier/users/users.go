package users_txverifier

import (
	"bytes"
	"errors"
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/state"
	"github.com/olympus-protocol/ogen/txs/txpayloads"
	users_txpayload "github.com/olympus-protocol/ogen/txs/txpayloads/users"
	"reflect"
	"sync"
)

var (
	ErrorNoTypeSpecified  = errors.New("userTx-no-type-rules: the selected action is not specified on the tx verifier scheme")
	ErrorInvalidSignature = errors.New("userTx-invalid-signature: the signature verification is invalid")
	ErrorMatchDataNoExist = errors.New("userTx-not-found-match-data: the match data searched doesn't exists")
	ErrorDataNoMatch      = errors.New("userTx-invalid-match-data: the data used to sign the transaction doesn't match")
)

type UsersTxVerifier struct {
	params *params.ChainParams
	state  *state.State
}

func (v UsersTxVerifier) DeserializePayload(payload []byte, Action p2p.TxAction) (txpayloads.Payload, error) {
	var Payload txpayloads.Payload
	switch Action {
	case p2p.Upload:
		Payload = new(users_txpayload.PayloadUpload)
	case p2p.Update:
		Payload = new(users_txpayload.PayloadUpdate)
	case p2p.Revoke:
		Payload = new(users_txpayload.PayloadRevoke)
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

func (v UsersTxVerifier) SigVerify(payload []byte, Action p2p.TxAction) error {
	VerPayload, err := v.DeserializePayload(payload, Action)
	if err != nil {
		return err
	}
	switch Action {
	case p2p.Upload:
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
	case p2p.Update:
		aggPubKey, err := VerPayload.GetAggPubKey()
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
		valid, err := bls.VerifySig(aggPubKey, msg, sig)
		if err != nil {
			return err
		}
		if !valid {
			return ErrorInvalidSignature
		}
	case p2p.Revoke:
		hash, err := VerPayload.GetHashForDataMatch()
		if err != nil {
			return err
		}
		userData := v.state.UserState.Get(hash)
		pubKeyBytes := userData.UserData.PubKey
		pubKey, err := bls.DeserializePublicKey(pubKeyBytes)
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
	}
	return nil
}

type routineRes struct {
	Err error
}

func (v UsersTxVerifier) SigVerifyBatch(payload [][]byte, Action p2p.TxAction) error {
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

func (v UsersTxVerifier) MatchVerify(payload []byte, Action p2p.TxAction) error {
	switch Action {
	case p2p.Upload:
		return nil
	case p2p.Update:
		VerPayload, err := v.DeserializePayload(payload, Action)
		if err != nil {
			return err
		}
		searchHash, err := VerPayload.GetHashForDataMatch()
		if err != nil {
			return err
		}
		ok := v.state.UserState.Have(searchHash)
		if !ok {
			return ErrorMatchDataNoExist
		}
		data := v.state.UserState.Get(searchHash)
		matchPubKey, err := VerPayload.GetPublicKey()
		pubKey, err := bls.DeserializePublicKey(data.UserData.PubKey)
		if err != nil {
			return err
		}
		equal := reflect.DeepEqual(matchPubKey, pubKey)
		if !equal {
			return ErrorDataNoMatch
		}
	case p2p.Revoke:
		VerPayload, err := v.DeserializePayload(payload, Action)
		if err != nil {
			return err
		}
		searchHash, err := VerPayload.GetHashForDataMatch()
		if err != nil {
			return err
		}
		ok := v.state.UserState.Have(searchHash)
		if !ok {
			return ErrorMatchDataNoExist
		}
	}
	return nil
}

func (v UsersTxVerifier) MatchVerifyBatch(payload [][]byte, Action p2p.TxAction) error {
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

func NewUsersTxVerifier(state *state.State, params *params.ChainParams) UsersTxVerifier {
	v := UsersTxVerifier{
		state:  state,
		params: params,
	}
	return v
}
