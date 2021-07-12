package kilic

import (
	"crypto/rand"
	"errors"
	"github.com/dgraph-io/ristretto"
	"github.com/kilic/bls12-381"
	"github.com/olympus-protocol/ogen/pkg/bls/common"
	"math/big"
)

var dst = []byte("BLS_SIG_BLS12381G2_XMD:SHA-256_SSWU_RO_POP_")

var maxKeys = int64(100000)

var pubkeyCache, _ = ristretto.NewCache(&ristretto.Config{
	NumCounters: maxKeys,
	MaxCost:     1 << 22, // ~4mb is cache max size
	BufferItems: 64,
})

var qBig, _ = new(big.Int).SetString("73eda753299d7d483339d80809a1d80553bda402fffe5bfeffffffff00000001", 16)

var (
	// ErrorSecSize returned when the secrete key size is wrong
	ErrorSecSize = errors.New("secret key size is wrong")
	// ErrorSecUnmarshal returned when the secret key is not valid
	ErrorSecUnmarshal = errors.New("secret key bytes are not on curve")
	// ErrorPubKeySize returned when the pubkey size is wrong
	ErrorPubKeySize = errors.New("pub key size is wrong")
	// ErrorPubKeyUnmarshal returned when the pubkey is not valid
	ErrorPubKeyUnmarshal = errors.New("could not unmarshal bytes into public key")
	// ErrorSigSize returned when the signature bytes doesn't match the length
	ErrorSigSize = errors.New("signature should be 96 bytes")
	// ErrorSigUnmarshal returned when the pubkey is not valid
	ErrorSigUnmarshal = errors.New("could not unmarshal bytes into signature")
)

type Kilic struct {
}

func (k Kilic) SecretKeyFromBytes(privKey []byte) (common.SecretKey, error) {

	if len(privKey) != 32 {
		return nil, ErrorSecSize
	}

	fr := bls12381.NewFr().FromBytes(privKey)

	in := new(big.Int).SetBytes(privKey)
	res := in.Cmp(qBig)

	if res >= 1 {
		return nil, ErrorSecUnmarshal
	}

	fr.FromBytes(privKey)

	return &SecretKey{p: fr}, nil
}

func (k Kilic) PublicKeyFromBytes(pubKey []byte) (common.PublicKey, error) {
	if len(pubKey) != 48 {
		return nil, ErrorPubKeySize
	}
	if cv, ok := pubkeyCache.Get(string(pubKey)); ok {
		return cv.(*PublicKey).Copy(), nil
	}

	p, err := bls12381.NewG1().FromCompressed(pubKey)
	if err != nil {
		return nil, ErrorPubKeyUnmarshal
	}

	obj := &PublicKey{p: p}
	pubkeyCache.Set(string(pubKey), obj.Copy(), 48)

	return obj, nil
}

func (k Kilic) SignatureFromBytes(sig []byte) (common.Signature, error) {
	size := len(sig)
	if size != 96 {
		return nil, ErrorSigSize
	}

	p, err := bls12381.NewG2().FromCompressed(sig)
	if err != nil {
		return nil, ErrorSigUnmarshal
	}

	return &Signature{s: p}, nil
}

func (k Kilic) AggregatePublicKeys(pubs [][]byte) (common.PublicKey, error) {
	if len(pubs) == 0 {
		return &PublicKey{}, nil
	}

	p, err := k.PublicKeyFromBytes(pubs[0])
	if err != nil {
		return nil, err
	}

	for _, pub := range pubs[1:] {
		pubkey, err := k.PublicKeyFromBytes(pub)
		if err != nil {
			return nil, err
		}
		p.Aggregate(pubkey)
	}

	return p, nil
}

func (k Kilic) Aggregate(sigs []common.Signature) common.Signature {
	return k.AggregateSignatures(sigs)
}

func (k Kilic) AggregateSignatures(sigs []common.Signature) common.Signature {
	if len(sigs) == 0 {
		return &Signature{}
	}

	if len(sigs) == 1 {
		return sigs[0]
	}

	g2 := bls12381.NewG2()

	aggregated := &bls12381.PointG2{}

	for _, agSig := range sigs {
		g2.Add(aggregated, aggregated, agSig.(*Signature).s)
	}

	return &Signature{s: aggregated}
}

func (k Kilic) NewAggregateSignature() common.Signature {
	p, err := bls12381.NewG2().HashToCurve([]byte{'m', 'o', 'c', 'k'}, dst)
	if err != nil {
		return nil
	}
	return &Signature{s: p}
}

func (k Kilic) RandKey() (common.SecretKey, error) {
	fr, err := bls12381.NewFr().Rand(rand.Reader)
	if err != nil {
		return nil, err
	}
	return &SecretKey{p: fr}, nil
}

var _ common.Implementation = &Kilic{}

func NewKilicInterface() common.Implementation {
	return &Kilic{}
}
