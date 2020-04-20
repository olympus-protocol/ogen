package chain

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/olympus-protocol/ogen/chain/index"
	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/txs/txverifier"
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
		Type:   primitives.TxCoins,
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

type blockRowAndValidator struct {
	row       *index.BlockRow
	validator uint32
}

// UpdateChainHead updates the blockchain head if needed
func (ch *Blockchain) UpdateChainHead() error {
	_, justifiedState := ch.state.GetJustifiedHead()

	activeValidatorIndices := justifiedState.GetActiveValidatorIndices()
	var targets []blockRowAndValidator
	for _, i := range activeValidatorIndices {
		bl, err := ch.getLatestAttestationTarget(i)
		if err != nil {
			continue
		}
		targets = append(targets, blockRowAndValidator{
			row:       bl,
			validator: i})
	}

	getVoteCount := func(block *index.BlockRow) uint64 {
		votes := uint64(0)
		for _, target := range targets {
			node := target.row.GetAncestorAtSlot(block.Slot)
			if node == nil {
				return 0
			}
			if node.Hash.IsEqual(&block.Hash) {
				votes += justifiedState.GetEffectiveBalance(target.validator, &ch.params) / 1e8
			}
		}
		return votes
	}

	head, _ := ch.state.GetJustifiedHead()

	// this may seem weird, but it occurs when importing when the justified block is not
	// imported, but the finalized head is. It should never occur other than that
	if head == nil {
		head, _ = ch.state.GetFinalizedHead()
	}

	for {
		children := head.Children
		if len(children) == 0 {
			ch.state.blockChain.SetTip(head)

			err := ch.db.SetTip(head.Hash)
			if err != nil {
				return err
			}

			return nil
		}
		bestVoteCountChild := children[0]
		bestVotes := getVoteCount(bestVoteCountChild)
		for _, c := range children[1:] {
			vc := getVoteCount(c)
			if vc > bestVotes {
				bestVoteCountChild = c
				bestVotes = vc
			}
		}
		head = bestVoteCountChild
	}
}

func (ch *Blockchain) getLatestAttestationTarget(validator uint32) (*index.BlockRow, error) {
	att, err := ch.db.GetLatestVote(validator)
	if err != nil {
		return nil, err
	}

	node, ok := ch.state.blockIndex.Get(att.Data.BeaconBlockHash)
	if !ok {
		return nil, errors.New("couldn't find block attested to by validator in index")
	}
	return node, nil
}

// ProcessBlock processes an incoming block from a peer or the miner.
func (ch *Blockchain) ProcessBlock(block *primitives.Block) error {
	// 1. first verify basic block properties
	// b. get parent block
	blockTime := ch.genesisTime.Add(time.Second * time.Duration(ch.params.SlotDuration*block.Header.Slot))

	if time.Now().Before(blockTime) {
		return fmt.Errorf("block %d processed at %s, but should wait until %s", block.Header.Slot, time.Now(), blockTime)
	}

	// 2. verify block against previous block's state
	oldState, found := ch.state.GetStateForHash(block.Header.PrevBlockHash)
	if !found {
		return fmt.Errorf("missing parent block state: %s", block.Header.PrevBlockHash)
	}

	txPayloadInv, err := ch.newTxPayloadInv(block.Txs, 1)
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
	// TODO: better fork choice here
	newState, err := ch.State().Add(block)
	if err != nil {
		ch.log.Warn(err)
		return err
	}

	err = ch.db.AddRawBlock(block)
	if err != nil {
		return err
	}

	row, err := ch.state.blockIndex.Add(*block)
	if err != nil {
		return err
	}

	// set current block row in database
	if err := ch.db.SetBlockRow(row.ToBlockNodeDisk()); err != nil {
		return err
	}

	// update parent to point at current
	if err := ch.db.SetBlockRow(row.Parent.ToBlockNodeDisk()); err != nil {
		return err
	}

	for _, a := range block.Votes {
		min, max := oldState.GetVoteCommittee(a.Data.Slot, &ch.params)

		validators := make([]uint32, 0, max-min)

		for i := range a.ParticipationBitfield {
			for j := 0; j < 8; j++ {
				if a.ParticipationBitfield[i]&(1<<uint(j)) != 0 {
					validator := uint32(i*8+j) + min
					validators = append(validators, validator)
				}
			}
		}

		if err := ch.db.SetLatestVoteIfNeeded(validators, &a); err != nil {
			return err
		}
	}

	rowHash := row.Hash
	ch.state.setBlockState(rowHash, newState)

	// TODO: remove when we have fork choice
	if err := ch.UpdateChainHead(); err != nil {
		return err
	}

	view, err := ch.State().GetSubView(block.Header.PrevBlockHash)
	if err != nil {
		return err
	}

	finalizedSlot := newState.FinalizedEpoch * ch.params.EpochLength
	finalizedHash, err := view.GetHashBySlot(finalizedSlot)
	if err != nil {
		return err
	}
	finalizedState, found := ch.state.GetStateForHash(finalizedHash)
	if !found {
		return fmt.Errorf("could not find finalized state with hash %s in state map", finalizedHash)
	}
	if err := ch.db.SetFinalizedHead(finalizedHash); err != nil {
		return err
	}
	if err := ch.db.SetFinalizedState(finalizedState); err != nil {
		return err
	}

	justifiedState, found := ch.state.GetStateForHash(newState.JustifiedEpochHash)
	if !found {
		return fmt.Errorf("could not find justified state with hash %s in state map", newState.JustifiedEpochHash)
	}
	if err := ch.db.SetJustifiedHead(newState.JustifiedEpochHash); err != nil {
		return err
	}
	if err := ch.db.SetJustifiedState(justifiedState); err != nil {
		return err
	}

	// TODO: delete state before finalized

	ch.log.Infof("New block accepted Hash: %v, Slot: %d", block.Hash(), block.Header.Slot)

	ch.notifeeLock.RLock()
	for i := range ch.notifees {
		i.NewTip(row, block)
	}
	ch.notifeeLock.RUnlock()

	return nil
}

func (ch *Blockchain) verifyTx(prevState *primitives.State, inv *TxPayloadInv) error {

	for scheme, txs := range inv.txs {
		txVerifier := txverifier.NewTxVerifier(&*prevState, &ch.params)
		err := txVerifier.VerifyTxsBatch(txs, scheme.Type, scheme.Action)
		if err != nil {
			return err
		}
	}
	return nil
}
