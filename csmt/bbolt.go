package csmt

import (
	"bytes"

	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
)

// BoltTreeDB is a tree database implemented on top of bbolt.
type BoltTreeDB struct {
	db      *bbolt.DB
	bktname []byte
}

// Hash gets the hash of the root.
func (b *BoltTreeDB) Hash() (*chainhash.Hash, error) {
	out := EmptyTree
	err := b.View(func(transaction TreeDatabaseTransaction) error {
		h, err := transaction.Hash()
		if err != nil {
			return err
		}
		out = *h
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &out, nil
}

// Update updates the database.
func (b *BoltTreeDB) Update(callback func(TreeDatabaseTransaction) error) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists([]byte(b.bktname))
		if err != nil {
			return err
		}
		trtx := &BoltTreeTransaction{
			bkt: bkt,
		}
		return callback(trtx)
	})
}

// View creates a read-only transaction for the database.
func (b *BoltTreeDB) View(callback func(TreeDatabaseTransaction) error) error {
	return b.db.View(func(tx *bbolt.Tx) error {
		trtx := &BoltTreeTransaction{
			bkt: tx.Bucket([]byte(b.bktname)),
		}
		return callback(trtx)
	})
}

var treeDBPrefix = []byte("tree-")
var treeKVPrefix = []byte("kv-")

func getTreeKey(key []byte) []byte {
	return append(treeDBPrefix, key...)
}

func getKVKey(key []byte) []byte {
	return append(treeKVPrefix, key...)
}

// SetRoot sets the root of the database.
func (b *BoltTreeTransaction) SetRoot(n *Node) error {
	nodeHash := n.GetHash()

	return b.bkt.Put([]byte("root"), nodeHash[:])
}

// NewNode creates a new node with the given left and right children and adds it to the database.
func (b *BoltTreeTransaction) NewNode(left *Node, right *Node, subtreeHash chainhash.Hash) (*Node, error) {
	var leftHash *chainhash.Hash
	var rightHash *chainhash.Hash

	if left != nil {
		lh := left.GetHash()
		leftHash = &lh
	}

	if right != nil {
		rh := right.GetHash()
		rightHash = &rh
	}

	newNode := &Node{
		value: subtreeHash,
		left:  leftHash,
		right: rightHash,
	}

	return newNode, b.SetNode(newNode)
}

// NewSingleNode creates a new single node and adds it to the database.
func (b *BoltTreeTransaction) NewSingleNode(key chainhash.Hash, value chainhash.Hash, subtreeHash chainhash.Hash) (*Node, error) {
	n := &Node{
		one:      true,
		oneKey:   &key,
		oneValue: &value,
		value:    subtreeHash,
	}

	return n, b.SetNode(n)
}

// GetNode gets a node from the database.
func (b *BoltTreeTransaction) GetNode(nodeHash chainhash.Hash) (*Node, error) {
	nodeKey := getTreeKey(nodeHash[:])

	nodeItem := b.bkt.Get(nodeKey)
	if nodeItem == nil {
		return nil, errors.New("Unabel to get node")
	}

	return DeserializeNode(nodeItem)
}

// SetNode sets a node in the database.
func (b *BoltTreeTransaction) SetNode(n *Node) error {
	nodeSer := n.Serialize()
	nodeHash := n.GetHash()
	nodeKey := getTreeKey(nodeHash[:])

	return b.bkt.Put(nodeKey, nodeSer)
}

// DeleteNode deletes a node from the database.
func (b *BoltTreeTransaction) DeleteNode(key chainhash.Hash) error {
	return b.bkt.Delete(getTreeKey(key[:]))
}

// Get gets a value from the key-value store.
func (b *BoltTreeTransaction) Get(key chainhash.Hash) (*chainhash.Hash, error) {
	var val chainhash.Hash

	valItem := b.bkt.Get(getKVKey(key[:]))
	if valItem == nil {
		return nil, errors.New("Get Value")
	}

	copy(val[:], valItem)
	return &val, nil
}

// Set sets a value in the key-value store.
func (b *BoltTreeTransaction) Set(key chainhash.Hash, value chainhash.Hash) error {
	return b.bkt.Put(getKVKey(key[:]), value[:])
}

// Root gets the root node.
func (b *BoltTreeTransaction) Root() (*Node, error) {
	i := b.bkt.Get([]byte("root"))
	if i == nil {
		return nil, nil
	}
	if bytes.Equal(i, EmptyTree[:]) {
		return nil, nil
	}

	nodeKey := b.bkt.Get(getTreeKey(i))
	if nodeKey == nil {
		return nil, errors.New("Unable to get Node Key")
	}

	return DeserializeNode(nodeKey)
}

// BoltTreeTransaction represents a bbolt transaction.
type BoltTreeTransaction struct {
	bkt *bbolt.Bucket
}

// Hash gets the hash of the root.
func (b *BoltTreeTransaction) Hash() (*chainhash.Hash, error) {
	i := b.bkt.Get([]byte("root"))
	if i == nil {
		return &EmptyTree, nil
	}
	return chainhash.NewHash(i)
}

// NewBoltTreeDB creates a new bbolt tree database from a bbolt database.
func NewBoltTreeDB(db *bbolt.DB, bkt []byte) *BoltTreeDB {
	return &BoltTreeDB{
		db:      db,
		bktname: bkt,
	}
}

var _ TreeDatabase = &BoltTreeDB{}
