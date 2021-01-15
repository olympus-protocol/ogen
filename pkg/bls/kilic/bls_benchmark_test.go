package kilic_test

import (
	"github.com/olympus-protocol/ogen/pkg/bls/common"
	"github.com/olympus-protocol/ogen/pkg/bls/kilic"
	"github.com/stretchr/testify/require"
	"testing"
)

var benchImpl = kilic.NewKilicInterface()

func BenchmarkSignature_Verify(b *testing.B) {
	sk, err := benchImpl.RandKey()
	require.NoError(b, err)

	msg := []byte("Some msg")
	sig := sk.Sign(msg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		require.Equal(b, true, sig.Verify(sk.PublicKey(), msg))
	}
}

func BenchmarkSignature_AggregateVerify(b *testing.B) {
	sigN := 128 // MAX_ATTESTATIONS per block.

	var pks []common.PublicKey
	var sigs []common.Signature
	var msgs [][32]byte
	for i := 0; i < sigN; i++ {
		msg := [32]byte{'s', 'i', 'g', 'n', 'e', 'd', byte(i)}
		sk, err := benchImpl.RandKey()
		require.NoError(b, err)
		sig := sk.Sign(msg[:])
		pks = append(pks, sk.PublicKey())
		sigs = append(sigs, sig)
		msgs = append(msgs, msg)
	}
	aggregated := benchImpl.Aggregate(sigs)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		require.Equal(b, true, aggregated.AggregateVerify(pks, msgs))
	}
}

func BenchmarkSecretKey_Marshal(b *testing.B) {
	key, err := benchImpl.RandKey()
	require.NoError(b, err)
	d := key.Marshal()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := benchImpl.SecretKeyFromBytes(d)
		_ = err
	}
}
