package bls_test

import (
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/bls/common"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDisallowZeroSecretKeys(t *testing.T) {
	bls.Initialize(&testdata.TestParams, "herumi")
	_, err := bls.SecretKeyFromBytes(common.ZeroSecretKey[:])
	require.Equal(t, common.ErrSecretUnmarshal, err)

	bls.Initialize(&testdata.TestParams, "blst")
	// Blst does a zero check on the key during deserialization.
	_, err = bls.SecretKeyFromBytes(common.ZeroSecretKey[:])
	require.Equal(t, common.ErrSecretUnmarshal, err)
}

func TestDisallowZeroPublicKeys(t *testing.T) {
	bls.Initialize(&testdata.TestParams, "herumi")
	_, err := bls.PublicKeyFromBytes(common.InfinitePublicKey[:])
	require.Equal(t, common.ErrInfinitePubKey, err)

	bls.Initialize(&testdata.TestParams, "blst")
	_, err = bls.PublicKeyFromBytes(common.InfinitePublicKey[:])
	require.Equal(t, common.ErrInfinitePubKey, err)
}

func TestDisallowZeroPublicKeys_AggregatePubkeys(t *testing.T) {
	bls.Initialize(&testdata.TestParams, "herumi")
	_, err := bls.AggregatePublicKeys([][]byte{common.InfinitePublicKey[:], common.InfinitePublicKey[:]})
	require.Equal(t, common.ErrInfinitePubKey, err)

	bls.Initialize(&testdata.TestParams, "blst")
	_, err = bls.AggregatePublicKeys([][]byte{common.InfinitePublicKey[:], common.InfinitePublicKey[:]})
	require.Equal(t, common.ErrInfinitePubKey, err)
}
