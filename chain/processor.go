package chain

import (
	"errors"
	"fmt"
	"time"

	"github.com/olympus-protocol/ogen/bdb"
	"github.com/olympus-protocol/ogen/chain/index"
	"github.com/olympus-protocol/ogen/primitives"
)

type blockRowAndValidator struct {
	row       *index.BlockRow
	validator uint32
}

// UpdateChainHead updates the blockchain head if needed
func (ch *Blockchain) UpdateChainHead(txn bdb.DBUpdateTransaction) error {
	_, justifiedState := ch.state.GetJustifiedHead()

	activeValidatorIndices := justifiedState.GetValidatorIndicesActiveAt(int64(justifiedState.EpochIndex))
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

			ch.log.Infof("setting head to %s", head.Hash)

			err := txn.SetTip(head.Hash)
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

func (ch *Blockchain) getLatestAttestationTarget(validator uint32) (row *index.BlockRow, err error) {
	var att *primitives.MultiValidatorVote
	att, ok := ch.state.GetLatestVote(validator)
	if !ok {
		return nil, fmt.Errorf("attestation target not found")
	}

	row, ok = ch.state.blockIndex.Get(att.Data.BeaconBlockHash)
	if !ok {
		return nil, errors.New("couldn't find block attested to by validator in index")
	}
	return row, nil
}

// ProcessBlock processes an incoming block from a peer or the miner.
func (ch *Blockchain) ProcessBlock(block *primitives.Block) error {
	// 1. first verify basic block properties
	// b. get parent block
	blockTime := ch.genesisTime.Add(time.Second * time.Duration(ch.params.SlotDuration*block.Header.Slot))

	if time.Now().Add(time.Second * 2).Before(blockTime) {
		return fmt.Errorf("block %d processed at %s, but should wait until %s", block.Header.Slot, time.Now(), blockTime)
	}

	// 2. verify block against previous block's state

	newState, receipts, err := ch.State().Add(block)
	if err != nil {
		ch.log.Warn(err)
		return err
	}

	if len(receipts) > 0 {
		msg := "\nEpoch Receipts\n----------\n"
		receiptTypes := make(map[primitives.ReceiptType]int64)

		for _, r := range receipts {
			if _, ok := receiptTypes[r.Type]; !ok {
				receiptTypes[r.Type] = r.Amount
			} else {
				receiptTypes[r.Type] += r.Amount
			}
		}

		for rt, amount := range receiptTypes {
			if amount > 0 {
				msg += fmt.Sprintf("rewarded %d for %s\n", amount, rt)
			} else if amount < 0 {
				msg += fmt.Sprintf("penalized %d for %s\n", -amount, rt)
			} else {
				msg += fmt.Sprintf("neutral increments for %s\n", rt)
			}
		}

		ch.log.Debugf(msg)
	}

	return ch.db.Update(func(txn bdb.DBUpdateTransaction) error {
		err = txn.AddRawBlock(block)
		if err != nil {
			return err
		}

		row, err := ch.state.blockIndex.Add(*block)
		if err != nil {
			return err
		}

		// set current block row in database
		if err := txn.SetBlockRow(row.ToBlockNodeDisk()); err != nil {
			return err
		}

		// update parent to point at current
		if err := txn.SetBlockRow(row.Parent.ToBlockNodeDisk()); err != nil {
			return err
		}

		for _, a := range block.Votes {
			validators, err := newState.GetVoteCommittee(a.Data.Slot, &ch.params)
			if err != nil {
				return err
			}

			ch.state.SetLatestVotesIfNeeded(validators, &a)
		}

		rowHash := row.Hash
		ch.state.setBlockState(rowHash, newState)

		// TODO: remove when we have fork choice
		if err := ch.UpdateChainHead(txn); err != nil {
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
		if err := txn.SetFinalizedHead(finalizedHash); err != nil {
			return err
		}
		if err := txn.SetFinalizedState(finalizedState); err != nil {
			return err
		}

		ch.state.RemoveBeforeSlot(newState.FinalizedEpoch * ch.params.EpochLength)

		justifiedState, found := ch.state.GetStateForHash(newState.JustifiedEpochHash)
		if !found {
			return fmt.Errorf("could not find justified state with hash %s in state map", newState.JustifiedEpochHash)
		}
		if err := txn.SetJustifiedHead(newState.JustifiedEpochHash); err != nil {
			return err
		}
		if err := txn.SetJustifiedState(justifiedState); err != nil {
			return err
		}

		// TODO: delete state before finalized

		ch.log.Infof("new block at slot: %d with %d finalized and %d justified", block.Header.Slot, newState.FinalizedEpoch, newState.JustifiedEpoch)

		ch.notifeeLock.RLock()
		for i := range ch.notifees {
			i.NewTip(row, block)
		}
		ch.notifeeLock.RUnlock()
		return nil
	})
}
