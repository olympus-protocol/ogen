package bloom

import (
	bitfcheck "github.com/olympus-protocol/ogen/utils/bitfield"
	"math/big"
	"sync"

	"github.com/olympus-protocol/ogen/utils/chainhash"
)

// BloomFilter keeps track of seen hashes like in the case of blocks or transactions.
type Filter struct {
	filter []uint8
	lock   sync.Mutex
}

// NewBloomFilter creates a new empty bloom filter with a certain size.
func NewBloomFilter(sizeBytes int) *Filter {
	return &Filter{
		filter: make([]uint8, sizeBytes),
	}
}

// Has checks if a bloom filter has a certain hash.
func (bf *Filter) Has(h chainhash.Hash) bool {
	bf.lock.Lock()
	defer bf.lock.Unlock()
	vhBig := new(big.Int).SetBytes(h[:])
	vhBig.Mod(vhBig, big.NewInt(int64(len(bf.filter)*8)))
	bloomIdx := vhBig.Uint64()
	return bitfcheck.Get(bf.filter, uint(bloomIdx))
}

// Add adds a hash to the bloom filter.
func (bf *Filter) Add(h chainhash.Hash) {
	bf.lock.Lock()
	defer bf.lock.Unlock()
	vhBig := new(big.Int).SetBytes(h[:])
	vhBig.Mod(vhBig, big.NewInt(int64(len(bf.filter)*8)))
	bloomIdx := vhBig.Uint64()
	bitfcheck.Set(bf.filter, uint(bloomIdx))
}
