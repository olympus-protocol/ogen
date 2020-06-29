package csmt

import (
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

// CombineHashes combines two branches of the tree.
func CombineHashes(left *chainhash.Hash, right *chainhash.Hash) chainhash.Hash {
	return chainhash.HashH(append(left[:], right[:]...))
}

// EmptyTrees are empty trees for each level.
var EmptyTrees [256]chainhash.Hash

// EmptyTree is the hash of an empty tree.
var EmptyTree = chainhash.Hash{}

// UpdateWitness allows an executor to securely update the tree root so that only a single key is changed.
type UpdateWitness struct {
	Key             chainhash.Hash
	OldValue        chainhash.Hash
	NewValue        chainhash.Hash
	WitnessBitfield chainhash.Hash
	LastLevel       uint8
	Witnesses       []chainhash.Hash
}

// Copy returns a copy of the update witness.
func (uw *UpdateWitness) Copy() UpdateWitness {
	newUw := *uw

	newUw.Witnesses = make([]chainhash.Hash, len(uw.Witnesses))
	for i := range newUw.Witnesses {
		copy(newUw.Witnesses[i][:], uw.Witnesses[i][:])
	}

	return newUw
}

// VerificationWitness allows an executor to verify a specific node in the tree.
type VerificationWitness struct {
	Key             chainhash.Hash
	Value           chainhash.Hash
	WitnessBitfield chainhash.Hash
	Witnesses       []chainhash.Hash
	LastLevel       uint8
}

// Copy returns a copy of the update witness.
func (vw *VerificationWitness) Copy() VerificationWitness {
	newVw := *vw

	newVw.Witnesses = make([]chainhash.Hash, len(vw.Witnesses))
	for i := range newVw.Witnesses {
		copy(newVw.Witnesses[i][:], vw.Witnesses[i][:])
	}

	return newVw
}

var emptyHash = chainhash.Hash{}

// isRight checks if the key is in the left or right subtree at a certain level. Level 255 is the root level.
func isRight(key chainhash.Hash, level uint8) bool {
	return key[level/8]&(1<<uint(level%8)) != 0
}

// calculateSubtreeHashWithOneLeaf calculates the hash of a subtree with only a single leaf at a certain height.
// atLevel is the height to calculate at.
func calculateSubtreeHashWithOneLeaf(key *chainhash.Hash, value *chainhash.Hash, atLevel uint8) chainhash.Hash {
	h := *value

	for i := uint8(0); i < atLevel; i++ {
		right := isRight(*key, i+1)

		// the key is in the right subtree
		if right {
			h = CombineHashes(&EmptyTrees[i], &h)
		} else {
			h = CombineHashes(&h, &EmptyTrees[i])
		}
	}

	return h
}

func insertIntoTree(t TreeDatabaseTransaction, root *Node, key chainhash.Hash, value chainhash.Hash, level uint8) (*Node, error) {
	right := isRight(key, level)

	if level == 0 {
		if root != nil && !root.Empty() {
			// remove the old node if it exists
			err := t.DeleteNode(root.GetHash())
			if err != nil {
				return nil, err
			}
		}

		// bottom leafs should have no siblings and a value
		return t.NewNode(nil, nil, value)
	}

	// if this tree is empty and we're inserting, we know it's the only key in the subtree, so let's mark it as such and
	// fill in the necessary values
	if root == nil || root.Empty() {
		return t.NewSingleNode(key, value, calculateSubtreeHashWithOneLeaf(&key, &value, level))
	}

	leftHash := root.Left()
	rightHash := root.Right()

	var newLeftBranch *Node
	var newRightBranch *Node

	if leftHash != nil {
		oldLeftBranch, err := t.GetNode(*leftHash)
		if err != nil {
			return nil, err
		}
		newLeftBranch = oldLeftBranch
	}

	if rightHash != nil {
		oldRightBranch, err := t.GetNode(*rightHash)
		if err != nil {
			return nil, err
		}
		newRightBranch = oldRightBranch
	}

	// if there is only one key in this subtree,
	if root.IsSingle() {
		rootKey := root.GetSingleKey()

		// this operation is an update
		if rootKey.IsEqual(&key) {
			// delete the old root
			err := t.DeleteNode(root.GetHash())
			if err != nil {
				return nil, err
			}

			// calculate the new root hash for this subtree
			return t.NewSingleNode(key, value, calculateSubtreeHashWithOneLeaf(&key, &value, level))
		}

		// check if the old key goes in the left or right
		subRight := isRight(rootKey, level)

		// we know this is a single, so the left and right should be nil
		if subRight {
			rightBranchInserted, err := insertIntoTree(t, newRightBranch, rootKey, root.GetSingleValue(), level-1)
			if err != nil {
				return nil, err
			}
			newRightBranch = rightBranchInserted
		} else {
			leftBranchInserted, err := insertIntoTree(t, newLeftBranch, rootKey, root.GetSingleValue(), level-1)
			if err != nil {
				return nil, err
			}
			newLeftBranch = leftBranchInserted
		}
	}

	// delete the old root because it was added to left or right branch
	err := t.DeleteNode(root.GetHash())
	if err != nil {
		return nil, err
	}

	if right {
		rightBranchInserted, err := insertIntoTree(t, newRightBranch, key, value, level-1)
		if err != nil {
			return nil, err
		}
		newRightBranch = rightBranchInserted
	} else {
		leftBranchInserted, err := insertIntoTree(t, newLeftBranch, key, value, level-1)
		if err != nil {
			return nil, err
		}
		newLeftBranch = leftBranchInserted
	}

	lv := EmptyTrees[level-1]
	if newLeftBranch != nil && !newLeftBranch.Empty() {
		lv = newLeftBranch.GetHash()
	}

	rv := EmptyTrees[level-1]
	if newRightBranch != nil && !newRightBranch.Empty() {
		rv = newRightBranch.GetHash()
	}

	newHash := CombineHashes(&lv, &rv)

	return t.NewNode(newLeftBranch, newRightBranch, newHash)
}
