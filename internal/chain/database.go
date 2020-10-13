package chain

import (
	"bytes"
	"encoding/hex"
	"github.com/olympus-protocol/ogen/internal/blockdb"
	"github.com/olympus-protocol/ogen/internal/chainindex"
	"github.com/olympus-protocol/ogen/internal/state"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
)

func (s *stateService) initializeDatabase(db blockdb.Database, blockNode *chainindex.BlockRow, state state.State) error {
	s.chain.SetTip(blockNode)

	err := s.SetFinalizedHead(blockNode.Hash, state)
	if err != nil {
		return err
	}
	err = s.SetJustifiedHead(blockNode.Hash, state)
	if err != nil {
		return err
	}

	if err := db.SetBlockRow(blockNode.ToBlockNodeDisk()); err != nil {
		return err
	}

	if err := db.SetFinalizedHead(blockNode.Hash); err != nil {
		return err
	}
	if err := db.SetJustifiedHead(blockNode.Hash); err != nil {
		return err
	}
	if err := db.SetFinalizedState(state); err != nil {
		return err
	}
	if err := db.SetJustifiedState(state); err != nil {
		return err
	}

	if err := db.SetTip(blockNode.Hash); err != nil {
		return err
	}

	return nil
}

func (s *stateService) loadBlockIndex(txn blockdb.Database, genesisHash chainhash.Hash) error {
	tip, err := txn.GetJustifiedHead()
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

		_, err = s.index.LoadBlockNode(rowDisk)
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
	s.stateMapLock.Lock()
	defer s.stateMapLock.Unlock()

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

	s.chain.SetTip(finalizedNode)

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

		_, err = s.index.LoadBlockNode(node)
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
