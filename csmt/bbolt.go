package csmt

import (
	"bytes"

	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
)

var errorNoData = errors.New("no data")

// BoltTreeDB is a tree database implemented on top of bbolt.
type BoltTreeDB struct {
	db      *bbolt.DB
	bktname string
}

// Hash gets the hash of the root.
func (b *BoltTreeDB) Hash() (*chainhash.Hash, error) {
	out := primitives.EmptyTree
	err := b.db.View(func(transaction TreeDatabaseTransaction) error {
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
	if len(nodeItem) <= 0 {
		return nil, errorNoData
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
	if len(valItem) <= 0 {
		return nil, errorNoData
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
	if len(i) <= 0 {
		return nil, errorNoData
	}

	if bytes.Equal(i, primitives.EmptyTree[:]) {
		return nil, nil
	}

	nodeKey := b.bkt.Get(getTreeKey(i))
	if len(nodeKey) <= 0 {
		return nil, errorNoData
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
	if len(i) <= 0 {
		return nil, errorNoData
	}
	return chainhash.NewHash(i)
}

// Update updates the database.
func (b *BoltTreeTransaction) Update(callback func(TreeDatabaseTransaction) error) error {
	bkt := b.bkt.CreateBucketIfNotExists()

	err := callback(&BoltTreeTransaction{badgerTx})
	if err != nil {
		badgerTx.Discard()
		return err
	}
	return badgerTx.Commit()
}

// View creates a read-only transaction for the database.
func (b *BoltTreeTransaction) View(callback func(TreeDatabaseTransaction) error) error {
	badgerTx := b.db.NewTransaction(false)

	err := callback(&BadgerTreeTransaction{badgerTx})
	if err != nil {
		badgerTx.Discard()
		return err
	}
	return badgerTx.Commit()
}

// NewBoltTreeDB creates a new badger tree database from a badger database.
func NewBoltTreeDB(db *bbolt.DB) *BoltTreeDB {
	return &BoltTreeDB{
		db: db,
	}
}

var _ TreeDatabase = &BoltTreeDB{}
