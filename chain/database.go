package chain

import (
	"github.com/olympus-protocol/ogen/bdb"
	"github.com/olympus-protocol/ogen/chain/index"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

func (s *StateService) initializeDatabase(txn bdb.DBUpdateTransaction, blockNode *index.BlockRow, state primitives.State) error {
	s.blockChain.SetTip(blockNode)

	s.setFinalizedHead(blockNode.Hash, state)
	s.setJustifiedHead(blockNode.Hash, state)
	if err := txn.SetBlockRow(blockNode.ToBlockNodeDisk()); err != nil {
		return err
	}

	if err := txn.SetFinalizedHead(blockNode.Hash); err != nil {
		return err
	}
	if err := txn.SetJustifiedHead(blockNode.Hash); err != nil {
		return err
	}
	if err := txn.SetFinalizedState(&state); err != nil {
		return err
	}
	if err := txn.SetJustifiedState(&state); err != nil {
		return err
	}

	if err := txn.SetTip(blockNode.Hash); err != nil {
		return err
	}

	return nil
}

func (s *StateService) loadBlockIndex(txn bdb.DBViewTransaction, genesisHash chainhash.Hash) error {
	justifiedHead, err := txn.GetJustifiedHead()
	if err != nil {
		return err
	}

	queue := []chainhash.Hash{genesisHash}

	for len(queue) > 0 {
		current := queue[0]

		s.log.Debugf("loading block node %s", current)

		queue = queue[1:]

		rowDisk, err := txn.GetBlockRow(current)
		if err != nil {
			return err
		}

		_, err = s.blockIndex.LoadBlockNode(rowDisk)
		if err != nil {
			return err
		}

		if current.IsEqual(&justifiedHead) {
			continue
		}

		queue = append(queue, rowDisk.Children...)
	}

	return nil
}

func (s *StateService) loadJustifiedAndFinalizedStates(txn bdb.DBViewTransaction) error {
	finalizedHead, err := txn.GetFinalizedHead()
	if err != nil {
		return err
	}
	finalizedState, err := txn.GetFinalizedState()
	if err != nil {
		return err
	}
	justifiedHead, err := txn.GetJustifiedHead()
	if err != nil {
		return err
	}
	justifiedState, err := txn.GetJustifiedState()
	if err != nil {
		return err
	}

	s.log.Infof("loaded justified head: %s and finalized head %s", justifiedHead, finalizedHead)

	if err := s.setFinalizedHead(finalizedHead, *finalizedState); err != nil {
		return err
	}
	if err := s.setJustifiedHead(justifiedHead, *justifiedState); err != nil {
		return err
	}

	return nil
}

func (s *StateService) setBlockState(hash chainhash.Hash, state *primitives.State) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.log.Debugf("setting block state for %s", hash)

	s.stateMap[hash] = newStateDerivedFromBlock(state)
}

func (s *StateService) loadStateMap(txn bdb.DBViewTransaction) error {
	finalizedNode := s.finalizedHead.node

	finalizedNodeWithChildren, err := txn.GetBlockRow(finalizedNode.Hash)
	if err != nil {
		return err
	}

	loadQueue := finalizedNodeWithChildren.Children

	justifiedState, err := txn.GetFinalizedState()
	if err != nil {
		return err
	}

	s.setBlockState(finalizedNode.Hash, justifiedState)

	s.blockChain.SetTip(finalizedNode)

	for len(loadQueue) > 0 {
		toLoad := loadQueue[0]
		loadQueue = loadQueue[1:]

		node, err := txn.GetBlockRow(toLoad)
		if err != nil {
			return err
		}

		s.log.Debugf("calculating block state for %s with previous %s", node.Hash, node.Parent)

		bl, err := txn.GetBlock(node.Hash)
		if err != nil {
			return err
		}

		_, _, err = s.Add(bl)
		if err != nil {
			return err
		}

		_, err = s.blockIndex.LoadBlockNode(node)
		if err != nil {
			return err
		}

		loadQueue = append(loadQueue, node.Children...)
	}

	justifiedHead, err := txn.GetJustifiedHead()
	if err != nil {
		return err
	}

	justifiedHeadState, err := txn.GetJustifiedState()
	if err != nil {
		return err
	}

	s.setJustifiedHead(justifiedHead, *justifiedHeadState)

	return nil
}

func (s *StateService) loadBlockchainFromDisk(txn bdb.DBViewTransaction, genesisHash chainhash.Hash) error {
	s.log.Info("loading block index...")
	err := s.loadBlockIndex(txn, genesisHash)
	if err != nil {
		return err
	}
	s.log.Info("loading justified and finalized states...")
	err = s.loadJustifiedAndFinalizedStates(txn)
	if err != nil {
		return err
	}
	s.log.Info("populating state map")
	err = s.loadStateMap(txn)
	if err != nil {
		return err
	}
	return nil
}
