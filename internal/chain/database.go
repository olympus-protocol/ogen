package chain

import (
	"bytes"
	"encoding/hex"
	"github.com/olympus-protocol/ogen/internal/blockdb"
	"github.com/olympus-protocol/ogen/internal/chainindex"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
)

func (s *stateService) initializeDatabase(txn blockdb.Database, blockNode *chainindex.BlockRow, state state.State) error {
	s.blockChain.SetTip(blockNode)

	err := s.SetFinalizedHead(blockNode.Hash, state)
	if err != nil {
		return err
	}
	err = s.SetJustifiedHead(blockNode.Hash, state)
	if err != nil {
		return err
	}

	if err := txn.SetBlockRow(blockNode.ToBlockNodeDisk()); err != nil {
		return err
	}

	if err := txn.SetFinalizedHead(blockNode.Hash); err != nil {
		return err
	}
	if err := txn.SetJustifiedHead(blockNode.Hash); err != nil {
		return err
	}
	if err := txn.SetFinalizedState(state); err != nil {
		return err
	}
	if err := txn.SetJustifiedState(state); err != nil {
		return err
	}

	if err := txn.SetTip(blockNode.Hash); err != nil {
		return err
	}

	return nil
}

func (s *stateService) loadBlockIndex(txn blockdb.Database, genesisHash chainhash.Hash) error {
	tip, err := txn.GetTip()
	if err != nil {
		return err
	}

	queue := [][32]byte{genesisHash}

	for len(queue) > 0 {
		current := queue[0]

		s.log.Debugf("Loading block node %s", hex.EncodeToString(current[:]))

		queue = queue[1:]

		rowDisk, err := txn.GetBlockRow(current)
		if err != nil {
			return err
		}

		_, err = s.blockIndex.LoadBlockNode(rowDisk)
		if err != nil {
			return err
		}
		if bytes.Equal(current[:], tip[:]) {
			continue
		}

		queue = append(queue, rowDisk.Children...)
	}

	return nil
}

func (s *stateService) loadJustifiedAndFinalizedStates(txn blockdb.Database) error {
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

	if err := s.SetFinalizedHead(finalizedHead, finalizedState); err != nil {
		return err
	}
	if err := s.SetJustifiedHead(justifiedHead, justifiedState); err != nil {
		return err
	}

	return nil
}

func (s *stateService) setBlockState(hash chainhash.Hash, state state.State) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.log.Debugf("setting block state for %s", hash)

	s.stateMap[hash] = newStateDerivedFromBlock(state)
}

func (s *stateService) loadStateMap(txn blockdb.Database) error {
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

		s.log.Debugf("calculating block state for %s with previous %s", hex.EncodeToString(node.Hash[:]), hex.EncodeToString(node.Parent[:]))

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

	err = s.SetJustifiedHead(justifiedHead, justifiedHeadState)
	if err != nil {
		return err
	}

	return nil
}

func (s *stateService) loadBlockchainFromDisk(txn blockdb.Database, genesisHash chainhash.Hash) error {
	s.log.Info("Loading block chainindex...")
	err := s.loadBlockIndex(txn, genesisHash)
	if err != nil {
		return err
	}
	s.log.Info("Loading justified and finalized states...")
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
