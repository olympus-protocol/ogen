package keystore_test

import (
	"github.com/olympus-protocol/ogen/internal/keystore"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_KeySerialize(t *testing.T) {
	for i := int64(0); i < 100; i++ {
		r, err := bls.RandKey()
		assert.NoError(t, err)

		k := &keystore.Key{
			Secret: r,
			Enable: true,
			Path:   i,
		}

		b, err := k.Marshal()
		assert.NoError(t, err)

		nk := new(keystore.Key)
		err = nk.Unmarshal(b)
		assert.NoError(t, err)

		assert.Equal(t, k, nk)
	}
}
