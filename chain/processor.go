package chain

import (
	"errors"
	"fmt"
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/state"
	"github.com/olympus-protocol/ogen/txs/txverifier"
	"reflect"
	"sync"
)

type txSchemes struct {
	Type   p2p.TxType
	Action p2p.TxAction
}

type TxPayloadInv struct {
	txs  map[txSchemes][]*p2p.MsgTx
	lock sync.RWMutex
}

func (txp *TxPayloadInv) Add(scheme txSchemes, tx *p2p.MsgTx, wg *sync.WaitGroup) {
	defer wg.Done()
	txp.lock.Lock()
	txp.txs[scheme] = append(txp.txs[scheme], tx)
	txp.lock.Unlock()
	return
}

var (
	ErrorTooManyGenerateTx = errors.New("chainProcessor-too-many-generate: the block contains more generate tx than expected")
	ErrorInvalidBlockSig   = errors.New("chainProcessor-block-sig-verify: the block signature verification failed")
	ErrorPubKeyNoMatch     = errors.New("chainProcessor-invalid-signer: the block signer is not valid")
)

func (ch *Blockchain) newTxPayloadInv(txs []*p2p.MsgTx, blocks int) (*TxPayloadInv, error) {
	txPayloads := &TxPayloadInv{
		txs: make(map[txSchemes][]*p2p.MsgTx),
	}
	var wg sync.WaitGroup
	for _, tx := range txs {
		wg.Add(1)
		scheme := txSchemes{
			Type:   tx.TxType,
			Action: tx.TxAction,
		}
		go func(scheme txSchemes, tx *p2p.MsgTx) {
			txPayloads.Add(scheme, tx, &wg)
		}(scheme, tx)
	}
	wg.Wait()
	if len(txPayloads.txs[txSchemes{
		Type:   p2p.Coins,
		Action: p2p.Generate,
	}]) > blocks {
		return nil, ErrorTooManyGenerateTx
	}
	return txPayloads, nil
}

func (ch *Blockchain) ProcessBlockInv(blockInv p2p.MsgBlockInv) error {
	// TODO: this is disabled for now because we don't have transaction execution done.
	// if we have a block that spends an input, we need to update our state representation
	// for that block before we try to verify other blocks.

	//txs := blockInv.GetTxs()
	//txPayloadInv, err := ch.newTxPayloadInv(txs, len(blockInv.GetBlocks()))
	//if err != nil {
	//	return err
	//}
	//err = ch.verifyTx(txPayloadInv)
	//if err != nil {
	//	return err
	//}
	return nil
}

func (ch *Blockchain) ProcessBlock(block *primitives.Block) error {
	// 1. first verify basic block properties

	// a. ensure block signature is valid
	err := ch.verifyBlockSig(block)
	if err != nil {
		ch.log.Warn(err)
		return err
	}

	// b. ensure we have the parent block
	if !ch.state.View.Has(block.Header().PrevBlockHash) {
		return fmt.Errorf("missing parent block: %s", block.Header().PrevBlockHash)
	}

	// 2. verify block against previous block's state
	oldState, found := ch.state.GetStateForHash(block.Header().PrevBlockHash)
	if !found {
		return fmt.Errorf("missing parent block state: %s", block.Header().PrevBlockHash)
	}

	txPayloadInv, err := ch.newTxPayloadInv(block.MsgBlock.Txs, 1)
	if err != nil {
		ch.log.Warn(err)
		return err
	}

	// a. verify transactions
	ch.log.Debugf("tx inventory created types to verify: %v", len(txPayloadInv.txs))
	err = ch.verifyTx(oldState, txPayloadInv)
	if err != nil {
		ch.log.Warn(err)
		return err
	}
	ch.log.Debugf("tx verification finished successfully")

	// b. apply block transition to state
	ch.log.Debugf("attempting to apply block to state")
	newState, err := oldState.TransitionBlock(block)
	if err != nil {
		ch.log.Warn(err)
		return err
	}
	ch.log.Infof("New block accepted Hash: %v", block.Hash)

	// 3. write block to database
	blocator, err := ch.db.AddRawBlock(block)
	if err != nil {
		ch.log.Warn(err)
		return err
	}

	// 4. add block to chain and set new state
	// TODO: better fork choice
	err = ch.state.Add(block, blocator, true, &newState)
	if err != nil {
		ch.log.Warn(err)
		return err
	}
	return nil
}

func (ch *Blockchain) verifyBlockSig(block *primitives.Block) error {
	if block.Height < ch.params.LastPreWorkersBlock {
		sig, err := block.MinerSig()
		if err != nil {
			return err
		}
		pubKey, err := block.MinerPubKey()
		if err != nil {
			return err
		}
		valid, err := bls.VerifySig(pubKey, block.Hash.CloneBytes(), sig)
		if err != nil {
			return err
		}
		if !valid {
			return ErrorInvalidBlockSig
		}
		pubKeyHash, err := pubKey.ToBech32(ch.params.AddressPrefixes, false)
		if err != nil {
			return err
		}
		equal := reflect.DeepEqual(pubKeyHash, ch.params.PreWorkersPubKeyHash)
		if !equal {
			return ErrorPubKeyNoMatch
		}
		ch.log.Infof("Block signature verified: pre-workers phase.")
	} else {
		// TODO use worker lists
		ch.log.Infof("Block signature verified: Worker rewarded: ")
	}
	return nil
}

type routineResp struct {
	Err error
}

func (ch *Blockchain) verifyTx(prevState *state.State, inv *TxPayloadInv) error {
	var wg sync.WaitGroup
	doneChan := make(chan routineResp, len(inv.txs))

	for scheme, txs := range inv.txs {
		wg.Add(1)
		txState := *prevState
		go func(wg *sync.WaitGroup, scheme txSchemes, txs []*p2p.MsgTx) {
			defer wg.Done()
			var resp routineResp
			txVerifier := txverifier.NewTxVerifier(&txState, &ch.params)
			err := txVerifier.VerifyTxsBatch(txs, scheme.Type, scheme.Action)
			if err != nil {
				resp.Err = err
			}
			doneChan <- resp
		}(&wg, scheme, txs)
	}
	wg.Wait()
	doneRes := <-doneChan
	if doneRes.Err != nil {
		return doneRes.Err
	}
	return nil
}
