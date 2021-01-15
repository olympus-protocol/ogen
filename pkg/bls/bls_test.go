package bls_test

import (
	"testing"
)

func TestDisallowZeroSecretKeys(t *testing.T) {
	/*	bls.Initialize(&testdata.TestParams, "herumi")
		_, err := bls.SecretKeyFromBytes(common.ZeroSecretKey[:])
		require.Equal(t, common.ErrZeroKey, err)*/

	/*bls.Initialize(&testdata.TestParams, "kilic")
	// Blst does a zero check on the key during deserialization.
	_, err := bls.SecretKeyFromBytes(common.ZeroSecretKey[:])
	require.Equal(t, common.ErrSecretUnmarshal, err)*/
}

func TestDisallowZeroPublicKeys(t *testing.T) {
	/*bls.Initialize(&testdata.TestParams, "herumi")
	_, err := bls.PublicKeyFromBytes(common.InfinitePublicKey[:])
	require.Equal(t, common.ErrInfinitePubKey, err)

	bls.Initialize(&testdata.TestParams, "blst")
	_, err = bls.PublicKeyFromBytes(common.InfinitePublicKey[:])
	require.Equal(t, common.ErrInfinitePubKey, err)*/
}

func TestDisallowZeroPublicKeys_AggregatePubkeys(t *testing.T) {
	/*bls.Initialize(&testdata.TestParams, "herumi")
	_, err := bls.AggregatePublicKeys([][]byte{common.InfinitePublicKey[:], common.InfinitePublicKey[:]})
	require.Equal(t, common.ErrInfinitePubKey, err)

	bls.Initialize(&testdata.TestParams, "blst")
	_, err = bls.AggregatePublicKeys([][]byte{common.InfinitePublicKey[:], common.InfinitePublicKey[:]})
	require.Equal(t, common.ErrInfinitePubKey, err)*/
}
