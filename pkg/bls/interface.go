package bls

import (
	"crypto/rand"
	"errors"
	"github.com/dgraph-io/ristretto"
	"github.com/kilic/bls12-381"
	"github.com/olympus-protocol/ogen/pkg/params"
	"math/big"
)

var dst = []byte("BLS_SIG_BLS12381G2_XMD:SHA-256_SSWU_RO_POP_")

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

var maxKeys = int64(100000)

var pubkeyCache, _ = ristretto.NewCache(&ristretto.Config{
	NumCounters: maxKeys,
	MaxCost:     1 << 22, // ~4mb is cache max size
	BufferItems: 64,
})

// KeyPair is an interface struct to serve keypairs
type KeyPair struct {
	Public  string `json:"public"`
	Private string `json:"private"`
}

var Prefix = params.MainNet.AccountPrefixes

func Initialize(c *params.ChainParams) {
	Prefix = c.AccountPrefixes
}

// SecretKeyFromBytes creates a BLS private key from a BigEndian byte slice.
func SecretKeyFromBytes(privKey []byte) (*SecretKey, error) {
	if len(privKey) != 32 {
		return nil, ErrorSecSize
	}

	fr := bls12381.NewFr()

	in := new(big.Int).SetBytes(privKey)
	res := in.Cmp(qBig)

	if res > 1 {
		return nil, ErrorSecUnmarshal
	}

	fr.FromBytes(privKey)

	return &SecretKey{p: fr}, nil
}

// RandKey creates a new private key using a random input.
func RandKey() (*SecretKey, error) {
	fr, err := bls12381.NewFr().Rand(rand.Reader)
	if err != nil {
		return nil, err
	}
	return &SecretKey{p: fr}, nil
}

// PublicKeyFromBytes creates a BLS public key from a  BigEndian byte slice.
func PublicKeyFromBytes(pubKey []byte) (*PublicKey, error) {
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

// AggregatePublicKeys aggregates the provided raw public keys into a single key.
func AggregatePublicKeys(pubs [][]byte) (*PublicKey, error) {
	if len(pubs) == 0 {
		return &PublicKey{}, nil
	}

	p, err := PublicKeyFromBytes(pubs[0])
	if err != nil {
		return nil, err
	}

	for _, pub := range pubs[1:] {
		pubkey, err := PublicKeyFromBytes(pub)
		if err != nil {
			return nil, err
		}
		p.Aggregate(pubkey)
	}

	return p, nil
}

// SignatureFromBytes creates a BLS signature from a LittleEndian byte slice.
func SignatureFromBytes(sig []byte) (*Signature, error) {
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

// NewAggregateSignature creates a blank aggregate signature.
func NewAggregateSignature() *Signature {
	p, err := bls12381.NewG2().HashToCurve([]byte{'m', 'o', 'c', 'k'}, dst)
	if err != nil {
		return nil
	}
	return &Signature{s: p}
}

// AggregateSignatures converts a list of signatures into a single, aggregated sig.
func AggregateSignatures(sigs []*Signature) *Signature {

	if len(sigs) == 0 {
		return &Signature{}
	}

	if len(sigs) == 1 {
		return sigs[0]
	}

	g2 := bls12381.NewG2()

	aggregated := &bls12381.PointG2{}

	for _, agSig := range sigs {
		g2.Add(aggregated, aggregated, agSig.s)
	}

	return &Signature{s: aggregated}
}

/*// VerifyMultipleSignatures verifies a non-singular set of signatures and its respective pubkeys and messages.
// This method provides a safe way to verify multiple signatures at once. We pick a number randomly from 1 to max
// uint64 and then multiply the signature by it. We continue doing this for all signatures and its respective pubkeys.
// S* = S_1 * r_1 + S_2 * r_2 + ... + S_n * r_n
// P'_{i,j} = P_{i,j} * r_i
// e(S*, G) = \prod_{i=1}^n \prod_{j=1}^{m_i} e(P'_{i,j}, M_{i,j})
// Using this we can verify multiple signatures safely.
func VerifyMultipleSignatures(sigs []*Signature, msgs [][32]byte, pubKeys []*PublicKey) (bool, error) {
	if len(sigs) == 0 || len(pubKeys) == 0 {
		return false, nil
	}
	length := len(sigs)
	if length != len(pubKeys) || length != len(msgs) {
		return false, errors.New("provided signatures, pubkeys and messages have differing lengths")
	}
	// Use a secure source of RNG.
	newGen := NewGenerator()
	randNums := make([]bls12.Fr, length)
	signatures := make([]bls12.G2, length)
	msgSlices := make([]byte, 0, 32*len(msgs))
	for i := 0; i < len(sigs); i++ {
		rNum := newGen.Uint64()
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, rNum)
		if err := randNums[i].SetLittleEndian(b); err != nil {
			return false, err
		}
		// Cast signature to a G2 value
		signatures[i] = *bls12.CastFromSign(sigs[i].s)

		// Flatten message to single byte slice to make it compatible with herumi.
		msgSlices = append(msgSlices, msgs[i][:]...)
	}
	// Perform multi scalar multiplication on all the relevant G2 points
	// with our generated random numbers.
	finalSig := new(bls12.G2)
	bls12.G2MulVec(finalSig, signatures, randNums)

	multiKeys := make([]bls12.PublicKey, length)
	for i := 0; i < len(pubKeys); i++ {
		if pubKeys[i] == nil {
			return false, errors.New("nil public key")
		}
		// Perform scalar multiplication for the corresponding g1 points.
		g1 := new(bls12.G1)
		bls12.G1Mul(g1, bls12.CastFromPublicKey(pubKeys[i].p), &randNums[i])
		multiKeys[i] = *bls12.CastToPublicKey(g1)
	}
	aggSig := bls12.CastToSign(finalSig)

	return aggSig.AggregateVerifyNoCheck(multiKeys, msgSlices), nil
}*/
