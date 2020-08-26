package bls

import (
	"crypto/rand"
	"errors"
	"github.com/dgraph-io/ristretto"
	"github.com/kilic/bls12-381"
	"github.com/olympus-protocol/ogen/pkg/params"
	"math/big"
)

var (
	// ErrorSecSize returned when the secret bytes doesn't match the length
	ErrorSecSize = errors.New("secret key should be 32 bytes")
	// ErrorSecUnmarshal returned when the secret key is not valid
	ErrorSecUnmarshal = errors.New("secret key bytes are not on curve")
	// ErrorPubSize returned when the public key bytes doesn't match the length
	ErrorPubSize = errors.New("public key should be 48 bytes")
	// ErrorPubKeyUnmarshal returned when the pubkey is not valid
	ErrorPubKeyUnmarshal = errors.New("could not unmarshal bytes into public key")
	// ErrorSigSize returned when the signature bytes doesn't match the length
	ErrorSigSize = errors.New("signature should be 96 bytes")
	// ErrorSigUnmarshal returned when the pubkey is not valid
	ErrorSigUnmarshal = errors.New("could not unmarshal bytes into signature")
)

// KeyPair is an interface struct to serve keypairs
type KeyPair struct {
	Public  string `json:"public"`
	Private string `json:"private"`
}

var Prefix params.AccountPrefixes = params.Mainnet.AccountPrefixes

func Initialize(c params.ChainParams) {
	Prefix = c.AccountPrefixes
}

var maxKeys = int64(100000)

var pubkeyCache, _ = ristretto.NewCache(&ristretto.Config{
	NumCounters: maxKeys,
	MaxCost:     1 << 22, // ~4mb is cache max size
	BufferItems: 64,
})

// SecretKeyFromBytes creates a BLS private key from a BigEndian byte slice.
func SecretKeyFromBytes(privKey []byte) (*SecretKey, error) {
	if len(privKey) != 32 {
		return nil, ErrorSecSize
	}
	p := new(big.Int)
	p.SetBytes(privKey)
	if p.Cmp(curveOrder) != -1 {
		return nil, ErrorSecUnmarshal
	}
	return &SecretKey{p: p}, nil
}

// PublicKeyFromBytes creates a BLS public key from a  BigEndian byte slice.
func PublicKeyFromBytes(pubKey []byte) (*PublicKey, error) {
	if len(pubKey) != 48 {
		return nil, ErrorPubSize
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

// AggregateSignatures converts a list of signatures into a single, aggregated sig.
func AggregateSignatures(sigs []*Signature) *Signature {
	if len(sigs) == 0 {
		return nil
	}
	g2 := bls12381.NewG2()
	sig := &bls12381.PointG2{}
	for i := 0; i < len(sigs); i++ {
		g2.Add(sig, sig, sigs[i].s)
	}
	return &Signature{s: sig}
}

// VerifyMultipleSignatures verifies multiple signatures for distinct messages securely.
func VerifyMultipleSignatures(sigs []*Signature, msgs [][32]byte, pubKeys []*PublicKey) (bool, error) {
	return false, nil
}

// NewAggregateSignature creates a blank aggregate signature.
func NewAggregateSignature() *Signature {
	p, _ := bls12381.NewG2().HashToCurve([]byte{'m', 'o', 'c', 'k'}, dst)
	return &Signature{s: p}
}

// RandKey creates a new private key using a random input.
func RandKey() *SecretKey {
	k, _ := rand.Int(rand.Reader, curveOrder)
	return &SecretKey{p: k}
}
