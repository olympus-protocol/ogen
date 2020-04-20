package chain

import (
	"github.com/olympus-protocol/ogen/chain/index"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

func (s *StateService) initializeDatabase(blockNode *index.BlockRow, state primitives.State) error {
	s.blockChain.SetTip(blockNode)

	s.setFinalizedHead(blockNode.Hash, state)
	s.setJustifiedHead(blockNode.Hash, state)
	if err := s.db.SetBlockRow(blockNode.ToBlockNodeDisk()); err != nil {
		return err
	}

	if err := s.db.SetFinalizedHead(blockNode.Hash); err != nil {
		return err
	}
	if err := s.db.SetJustifiedHead(blockNode.Hash); err != nil {
		return err
	}
	if err := s.db.SetFinalizedState(&state); err != nil {
		return err
	}
	if err := s.db.SetJustifiedState(&state); err != nil {
		return err
	}

	if err := s.db.SetTip(blockNode.Hash); err != nil {
		return err
	}

	return nil
}

func (s *StateService) loadBlockIndex(genesisHash chainhash.Hash) error {
	justifiedHead, err := s.db.GetJustifiedHead()
	if err != nil {
		return err
	}

	queue := []chainhash.Hash{genesisHash}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		rowDisk, err := s.db.GetBlockRow(current)
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

func (s *StateService) loadJustifiedAndFinalizedStates() error {
	finalizedHead, err := s.db.GetFinalizedHead()
	if err != nil {
		return err
	}
	finalizedState, err := s.db.GetFinalizedState()
	if err != nil {
		return err
	}
	justifiedHead, err := s.db.GetJustifiedHead()
	if err != nil {
		return err
	}
	justifiedState, err := s.db.GetJustifiedState()
	if err != nil {
		return err
	}

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

	s.stateMap[hash] = newStateDerivedFromBlock(state)
}

func (s *StateService) loadStateMap() error {
	justifiedNode := s.justifiedHead.node

	justifiedNodeWithChildren, err := s.db.GetBlockRow(justifiedNode.Hash)
	if err != nil {
		return err
	}

	loadQueue := justifiedNodeWithChildren.Children

	justifiedState, err := s.db.GetJustifiedState()
	if err != nil {
		return err
	}

	s.setBlockState(justifiedNode.Hash, justifiedState)

	s.blockChain.SetTip(justifiedNode)

	for len(loadQueue) > 0 {
		toLoad := loadQueue[0]
		loadQueue = loadQueue[1:]

		node, err := s.db.GetBlockRow(toLoad)
		if err != nil {
			return err
		}

		bl, err := s.db.GetRawBlock(node.Hash)
		if err != nil {
			return err
		}

		_, err = s.Add(bl)
		if err != nil {
			return err
		}

		_, err = s.blockIndex.LoadBlockNode(node)
		if err != nil {
			return err
		}

		// TODO: fork choice on importing

		loadQueue = append(loadQueue, node.Children...)
	}

	justifiedHead, err := s.db.GetJustifiedHead()
	if err != nil {
		return err
	}

	justifiedHeadState, err := s.db.GetJustifiedState()
	if err != nil {
		return err
	}

	s.setJustifiedHead(justifiedHead, *justifiedHeadState)

	return nil
}

func (s *StateService) loadBlockchainFromDisk(genesisHash chainhash.Hash) error {
	s.log.Info("loading block index...")
	err := s.loadBlockIndex(genesisHash)
	if err != nil {
		return err
	}
	s.log.Info("loading justified and finalized states...")
	err = s.loadJustifiedAndFinalizedStates()
	if err != nil {
		return err
	}
	s.log.Info("populating state map")
	err = s.loadStateMap()
	if err != nil {
		return err
	}
	return nil
}
