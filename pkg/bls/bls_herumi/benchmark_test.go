package bls_herumi

import (
	"github.com/olympus-protocol/bls-go/bls"
	bls_interface "github.com/olympus-protocol/ogen/pkg/bls/interface"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/stretchr/testify/require"
	"testing"
)

func BenchmarkPairing(b *testing.B) {
	require.NoError(b, bls.Init(bls.BLS12_381))
	if err := bls.SetETHmode(bls.EthModeDraft07); err != nil {
		panic(err)
	}
	newGt := &bls.GT{}
	newG1 := &bls.G1{}
	newG2 := &bls.G2{}

	newGt.SetInt64(10)
	hash := chainhash.HashH([]byte{})
	require.NoError(b, newG1.HashAndMapTo(hash[:]))
	require.NoError(b, newG2.HashAndMapTo(hash[:]))
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		bls.Pairing(newGt, newG1, newG2)
	}

}
func BenchmarkSignature_Verify(b *testing.B) {
	impl := HerumiImplementation{}
	sk := impl.RandKey()

	msg := []byte("Some msg")
	sig := sk.Sign(msg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		require.Equal(b, true, sig.Verify(sk.PublicKey(), msg))
	}
}

func BenchmarkSignature_AggregateVerify(b *testing.B) {
	sigN := 128 // MAX_ATTESTATIONS per block.
	impl := HerumiImplementation{}

	var pks []bls_interface.PublicKey
	var sigs []bls_interface.Signature
	var msgs [][32]byte
	for i := 0; i < sigN; i++ {
		msg := [32]byte{'s', 'i', 'g', 'n', 'e', 'd', byte(i)}
		sk := impl.RandKey()
		sig := sk.Sign(msg[:])
		pks = append(pks, sk.PublicKey())
		sigs = append(sigs, sig)
		msgs = append(msgs, msg)
	}
	aggregated := impl.Aggregate(sigs)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		require.Equal(b, true, aggregated.AggregateVerify(pks, msgs))
	}
}

func BenchmarkSecretKey_Marshal(b *testing.B) {
	impl := HerumiImplementation{}

	key := impl.RandKey()
	d := key.Marshal()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := impl.SecretKeyFromBytes(d)
		_ = err
	}
}
