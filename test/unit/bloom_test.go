package unit_test

import (
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

	if !b.Has(chainhash.HashH([]byte("hello!"))) {
		t.Fatal("expected hello! to be in bloom filter")
	}
	if !b.Has(chainhash.HashH([]byte("hello2!"))) {
		t.Fatal("expected hello2! to be in bloom filter")
	}
	if !b.Has(chainhash.HashH([]byte("hello3!"))) {
		t.Fatal("expected hello3! to be in bloom filter")
	}
	if !b.Has(chainhash.HashH([]byte("hello4!"))) {
		t.Fatal("expected hello4! to be in bloom filter")
	}
	if b.Has(chainhash.HashH([]byte("hello5!"))) {
		t.Fatal("expected hello5! to not be in bloom filter")
	}
}
