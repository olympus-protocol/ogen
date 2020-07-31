package bloom_test

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/olympus-protocol/ogen/bloom"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

func TestBloomBasic(t *testing.T) {
	b := bloom.NewBloomFilter(1000000)

	b.Add(chainhash.HashH([]byte("hello!")))
	b.Add(chainhash.HashH([]byte("hello2!")))
	b.Add(chainhash.HashH([]byte("hello3!")))
	b.Add(chainhash.HashH([]byte("hello4!")))

	assert.True(t, b.Has(chainhash.HashH([]byte("hello!"))))

	assert.True(t, b.Has(chainhash.HashH([]byte("hello2!"))))

	assert.True(t, b.Has(chainhash.HashH([]byte("hello3!"))))

	assert.True(t, b.Has(chainhash.HashH([]byte("hello4!"))))

	assert.False(t, b.Has(chainhash.HashH([]byte("hello5!"))))

}
