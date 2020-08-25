package bls_kilic

import (
	"github.com/olympus-protocol/ogen/pkg/bech32"
	bls_interface "github.com/olympus-protocol/ogen/pkg/bls/interface"
	"github.com/olympus-protocol/ogen/pkg/chainhash"

	"github.com/dgraph-io/ristretto"
	bls12 "github.com/kilic/bls12-381"
	"github.com/pkg/errors"
)

var maxKeys = int64(100000)
var pubkeyCache, _ = ristretto.NewCache(&ristretto.Config{
	NumCounters: maxKeys,
	MaxCost:     1 << 22, // ~4mb is cache max size
	BufferItems: 64,
})

// PublicKey used in the BLS signature scheme.
type PublicKey struct {
	p *bls12.PointG1
}

// PublicKeyFromBytes creates a BLS public key from a  BigEndian byte slice.
func (b KilicImplementation) PublicKeyFromBytes(pubKey []byte) (bls_interface.PublicKey, error) {
	if cv, ok := pubkeyCache.Get(string(pubKey)); ok {
		return cv.(*PublicKey).Copy(), nil
	}
	p, err := bls12.NewG1().FromCompressed(pubKey)
	if err != nil {
		return nil, errors.New("could not unmarshal bytes into public key")
	}
	pubKeyObj := &PublicKey{p: p}
	copiedKey := pubKeyObj.Copy()
	pubkeyCache.Set(string(pubKey), copiedKey, 48)
	return pubKeyObj, nil
}

// AggregatePublicKeys aggregates the provided raw public keys into a single key.
func (b KilicImplementation) AggregatePublicKeys(pubs [][]byte) (bls_interface.PublicKey, error) {
	//agg := new(blst.P1Aggregate)
	//mulP1 := make([]*blst.P1Affine, 0, len(pubs))
	//for _, pubkey := range pubs {
	//	if cv, ok := pubkeyCache.Get(string(pubkey)); ok {
	//		mulP1 = append(mulP1, cv.(*PublicKey).Copy().(*PublicKey).p)
	//		continue
	//	}
	//	p := new(blstPublicKey).Uncompress(pubkey)
	//	if p == nil {
	//		return nil, errors.New("could not unmarshal bytes into public key")
	//	}
	//	pubKeyObj := &PublicKey{p: p}
	//	pubkeyCache.Set(string(pubkey), pubKeyObj.Copy(), 48)
	//	mulP1 = append(mulP1, p)
	//}
	//agg.Aggregate(mulP1)
	//return &PublicKey{p: agg.ToAffine()}, nil
	return nil, nil
}

// Marshal a public key into a LittleEndian byte slice.
func (p *PublicKey) Marshal() []byte {
	return bls12.NewG1().ToCompressed(p.p)
}

// Copy the public key to a new pointer reference.
func (p *PublicKey) Copy() bls_interface.PublicKey {
	np := *p.p
	return &PublicKey{p: &np}
}

// Aggregate two public keys.
func (p *PublicKey) Aggregate(p2 bls_interface.PublicKey) bls_interface.PublicKey {
	return &PublicKey{p: bls12.NewG1().Add(p.p, p.p, p2.(*PublicKey).p)}
}

// Hash calculates the hash of the public key.
func (p *PublicKey) Hash() ([20]byte, error) {
	pkS := bls12.NewG1().ToCompressed(p.p)
	h := chainhash.HashH(pkS[:])
	var hBytes [20]byte
	copy(hBytes[:], h[:])
	return hBytes, nil
}

// ToAccount converts the public key to a Bech32 address.
func (p *PublicKey) ToAccount() string {
	out := make([]byte, 20)
	h := chainhash.HashH(p.Marshal())
	copy(out[:], h[:20])
	return bech32.Encode(bls_interface.Prefix.Public, out)
}
