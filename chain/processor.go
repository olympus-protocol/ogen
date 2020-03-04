package chain

import (
	"errors"
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/primitives"
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
	txs := blockInv.GetTxs()
	txPayloadInv, err := ch.newTxPayloadInv(txs, len(blockInv.GetBlocks()))
	if err != nil {
		return err
	}
	err = ch.verifyTx(txPayloadInv)
	if err != nil {
		return err
	}
	return nil
}

func (ch *Blockchain) ProcessBlock(block *primitives.Block) error {
	err := ch.verifyBlockSig(block)
	if err != nil {
		ch.log.Warn(err)
		return err
	}
	txPayloadInv, err := ch.newTxPayloadInv(block.MsgBlock.Txs, 1)
	if err != nil {
		ch.log.Warn(err)
		return err
	}
	ch.log.Debugf("tx inventory created types to verify: %v", len(txPayloadInv.txs))
	err = ch.verifyTx(txPayloadInv)
	if err != nil {
		ch.log.Warn(err)
		return err
	}
	ch.log.Debugf("tx verification finished successfully")
	ch.log.Infof("New block accepted Hash: %v", block.Hash)
	blocator, err := ch.db.AddRawBlock(block)
	if err != nil {
		ch.log.Warn(err)
		return err
	}
	row, err := ch.state.View.Add(*block.Header(), blocator)
	if err != nil {
		ch.log.Warn(err)
		return err
	}
	rowHash, err := row.Header.Hash()
	if err != nil {
		ch.log.Warn(err)
		return err
	}
	// TODO: better fork choice
	ch.state.View.SetTip(rowHash)
	err = ch.UpdateState(block, 0, 0, 0, true)
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

func (ch *Blockchain) verifyTx(inv *TxPayloadInv) error {
	var wg sync.WaitGroup
	doneChan := make(chan routineResp, len(inv.txs))
	state := ch.state.TipState()

	for scheme, txs := range inv.txs {
		wg.Add(1)
		txState := *state
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
