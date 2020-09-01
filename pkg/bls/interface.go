package bls

import (
	"encoding/binary"
	"errors"
	bls12 "github.com/herumi/bls-eth-go-binary/bls"
	"github.com/olympus-protocol/ogen/pkg/params"
)

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

// KeyPair is an interface struct to serve keypairs
type KeyPair struct {
	Public  string `json:"public"`
	Private string `json:"private"`
}

var Prefix params.AccountPrefixes = params.Mainnet.AccountPrefixes

func Initialize(c params.ChainParams) {
	Prefix = c.AccountPrefixes
}

func init() {
	if err := bls12.Init(bls12.BLS12_381); err != nil {
		panic(err)
	}
	if err := bls12.SetETHmode(bls12.EthModeDraft07); err != nil {
		panic(err)
	}
}

// PublicKeyFromBytes creates a BLS public key from a  BigEndian byte slice.
func PublicKeyFromBytes(pubKey []byte) (*PublicKey, error) {
	if len(pubKey) != 48 {
		return nil, ErrorPubKeySize
	}
	if cv, ok := pubkeyCache.Get(string(pubKey)); ok {
		return cv.(*PublicKey).Copy(), nil
	}
	p := &bls12.PublicKey{}
	err := p.Deserialize(pubKey)
	if err != nil {
		return nil, ErrorPubKeyUnmarshal
	}
	pubKeyObj := &PublicKey{p: p}
	pubkeyCache.Set(string(pubKey), pubKeyObj.Copy(), 48)
	return pubKeyObj, nil
}

// SignatureFromBytes creates a BLS signature from a LittleEndian byte slice.
func SignatureFromBytes(sig []byte) (*Signature, error) {
	if len(sig) != 96 {
		return nil, ErrorSigSize
	}
	signature := &bls12.Sign{}
	err := signature.Deserialize(sig)
	if err != nil {
		return nil, ErrorSigUnmarshal
	}
	return &Signature{s: signature}, nil
}

// NewAggregateSignature creates a blank aggregate signature.
func NewAggregateSignature() *Signature {
	return &Signature{s: bls12.HashAndMapToSignature([]byte{'m', 'o', 'c', 'k'})}
}

// AggregateSignatures converts a list of signatures into a single, aggregated sig.
func AggregateSignatures(sigs []*Signature) *Signature {
	if len(sigs) == 0 {
		return nil
	}
	signature := *sigs[0].Copy().s
	for i := 1; i < len(sigs); i++ {
		signature.Add(sigs[i].s)
	}
	return &Signature{s: &signature}
}

// Aggregate is an alias for AggregateSignatures, defined to conform to BLS specification.
//
// In IETF draft BLS specification:
// Aggregate(signature_1, ..., signature_n) -> signature: an
//      aggregation algorithm that compresses a collection of signatures
//      into a single signature.
//
// In ETH2.0 specification:
// def Aggregate(signatures: Sequence[BLSSignature]) -> BLSSignature
//
// Deprecated: Use AggregateSignatures.
func Aggregate(sigs []*Signature) *Signature {
	return AggregateSignatures(sigs)
}

// VerifyMultipleSignatures verifies a non-singular set of signatures and its respective pubkeys and messages.
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
}

// SecretKeyFromBytes creates a BLS private key from a BigEndian byte slice.
func SecretKeyFromBytes(privKey []byte) (*SecretKey, error) {
	if len(privKey) != 32 {
		return nil, ErrorSecSize
	}
	secKey := &bls12.SecretKey{}
	err := secKey.Deserialize(privKey)
	if err != nil {
		return nil, ErrorSecUnmarshal
	}
	return &SecretKey{p: secKey}, err
}
