package chain

import (
	"errors"
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/state"
	"github.com/olympus-protocol/ogen/txs/txverifier"
	"sync"
)

type txSchemes struct {
	Type   primitives.TxType
	Action primitives.TxAction
}

type TxPayloadInv struct {
	txs  map[txSchemes][]primitives.Tx
	lock sync.RWMutex
}

func (txp *TxPayloadInv) Add(scheme txSchemes, tx primitives.Tx, wg *sync.WaitGroup) {
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

func (ch *Blockchain) newTxPayloadInv(txs []primitives.Tx, blocks int) (*TxPayloadInv, error) {
	txPayloads := &TxPayloadInv{
		txs: make(map[txSchemes][]primitives.Tx),
	}
	var wg sync.WaitGroup
	for _, tx := range txs {
		wg.Add(1)
		scheme := txSchemes{
			Type:   tx.TxType,
			Action: tx.TxAction,
		}
		go func(scheme txSchemes, tx primitives.Tx) {
			txPayloads.Add(scheme, tx, &wg)
		}(scheme, tx)
	}
	wg.Wait()
	if len(txPayloads.txs[txSchemes{
		Type:   primitives.Coins,
		Action: primitives.Generate,
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

func (ch *Blockchain) valid(block *primitives.Block) (*state.State, error) {
	// 1. first verify basic block properties

	// a. ensure we have the parent block
	parentBlock, ok := ch.state.View.GetRowByHash(block.Header.PrevBlockHash)
	if !ok {
		return nil, ErrNoParent
	}

	height := parentBlock.Height + 1

	// b. verify block signature
	err := ch.verifyBlockSig(block, uint32(height))
	if err != nil {
		ch.log.Warn(err)
		return nil, err
	}

	// 2. verify block against previous block's state
	currentTip, oldState := ch.state.View.Tip()

	if !currentTip.Hash.IsEqual(&block.Header.PrevBlockHash) {
		return nil, ErrNoParent
	}

	txPayloadInv, err := ch.newTxPayloadInv(block.Txs, 1)
	if err != nil {
		return nil, err
	}

	// a. verify transactions
	ch.log.Debugf("tx inventory created types to verify: %v", len(txPayloadInv.txs))
	err = ch.verifyTx(&oldState, txPayloadInv)
	if err != nil {
		return nil, err
	}
	ch.log.Debugf("tx verification finished successfully")
	return &oldState, nil
}

func (ch *Blockchain) verifyBlockSig(block *primitives.Block, height uint32) error {
	if height < ch.params.LastPreWorkersBlock {
		sig, err := block.MinerSig()
		if err != nil {
			return err
		}
		pubKey, err := block.MinerPubKey()
		if err != nil {
			return err
		}
		blockHash := block.Hash()
		valid, err := bls.VerifySig(pubKey, blockHash[:], sig)
		if err != nil {
			return err
		}
		if !valid {
			return ErrorInvalidBlockSig
		}
		//pubKeyHash, err := pubKey.ToBech32(ch.params.AddressPrefixes, false)
		//if err != nil {
		//	return err
		//}
		// TODO: ensure block pubkey matches expected worker
		//equal := reflect.DeepEqual(pubKeyHash, ch.params.PreWorkersPubKeyHash)
		//if !equal {
		//	return ErrorPubKeyNoMatch
		//}
		ch.log.Infof("Block signature verified: pre-workers phase.")
	} else {
		// TODO use worker lists
		ch.log.Infof("Block signature verified: Worker rewarded: ")
	}
	return nil
}

func (ch *Blockchain) verifyTx(prevState *state.State, inv *TxPayloadInv) error {

	for scheme, txs := range inv.txs {
		txVerifier := txverifier.NewTxVerifier(&*prevState, &ch.params)
		err := txVerifier.VerifyTxsBatch(txs, scheme.Type, scheme.Action)
		if err != nil {
			return err
		}
	}
	return nil
}
