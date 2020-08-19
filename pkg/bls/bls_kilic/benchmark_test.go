package bls_kilic

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func BenchmarkSignature_Verify(b *testing.B) {
	impl := KilicImplementation{}
	sk := impl.RandKey()

	msg := []byte("Some msg")
	sig := sk.Sign(msg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		require.Equal(b, true, sig.Verify(sk.PublicKey(), msg))
	}
}

func BenchmarkSecretKey_Marshal(b *testing.B) {
	impl := KilicImplementation{}

	key := impl.RandKey()
	d := key.Marshal()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := impl.SecretKeyFromBytes(d)
		_ = err
	}
}
