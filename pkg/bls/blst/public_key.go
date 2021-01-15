package blst

/*
import (
	"fmt"
	"github.com/olympus-protocol/ogen/pkg/bech32"
	"github.com/olympus-protocol/ogen/pkg/bls/common"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/params"

	"github.com/dgraph-io/ristretto"
	"github.com/pkg/errors"
)

var maxKeys = int64(1000000)
var pubkeyCache, _ = ristretto.NewCache(&ristretto.Config{
	NumCounters: maxKeys,
	MaxCost:     1 << 26, // ~64mb is cache max size
	BufferItems: 64,
})

// PublicKey used in the BLS signature scheme.
type PublicKey struct {
	p *blstPublicKey
}

// PublicKeyFromBytes creates a BLS public key from a  BigEndian byte slice.
func (b *BLST) PublicKeyFromBytes(pubKey []byte) (common.PublicKey, error) {
	if len(pubKey) != 48 {
		return nil, fmt.Errorf("public key must be %d bytes", 48)
	}
	if cv, ok := pubkeyCache.Get(string(pubKey)); ok {
		return cv.(*PublicKey).Copy(), nil
	}
	// Subgroup check NOT done when decompressing pubkey.
	p := new(blstPublicKey).Uncompress(pubKey)
	if p == nil {
		return nil, errors.New("could not unmarshal bytes into public key")
	}
	// Subgroup and infinity check
	if !p.KeyValidate() {
		// NOTE: the error is not quite accurate since it includes group check
		return nil, common.ErrInfinitePubKey
	}
	pubKeyObj := &PublicKey{p: p}
	copiedKey := pubKeyObj.Copy()
	pubkeyCache.Set(string(pubKey), copiedKey, 48)
	return pubKeyObj, nil
}

// AggregatePublicKeys aggregates the provided raw public keys into a single key.
func (b *BLST) AggregatePublicKeys(pubs [][]byte) (common.PublicKey, error) {
	agg := new(blstAggregatePublicKey)
	mulP1 := make([]*blstPublicKey, 0, len(pubs))
	for _, pubkey := range pubs {
		pubKeyObj, err := b.PublicKeyFromBytes(pubkey)
		if err != nil {
			return nil, err
		}
		mulP1 = append(mulP1, pubKeyObj.(*PublicKey).p)
	}
	// No group check needed here since it is done in PublicKeyFromBytes
	// Note the checks could be moved from PublicKeyFromBytes into Aggregate
	// and take advantage of multi-threading.
	agg.Aggregate(mulP1, false)
	return &PublicKey{p: agg.ToAffine()}, nil
}

// Marshal a public key into a LittleEndian byte slice.
func (p *PublicKey) Marshal() []byte {
	return p.p.Compress()
}

// Copy the public key to a new pointer reference.
func (p *PublicKey) Copy() common.PublicKey {
	np := *p.p
	return &PublicKey{p: &np}
}

// IsInfinite checks if the public key is infinite.
func (p *PublicKey) IsInfinite() bool {
	zeroKey := new(blstPublicKey)
	return p.p.Equals(zeroKey)
}

// Aggregate two public keys.
func (p *PublicKey) Aggregate(p2 common.PublicKey) common.PublicKey {
	agg := new(blstAggregatePublicKey)
	// No group check here since it is checked at decompression time
	agg.Add(p.p, false)
	agg.Add(p2.(*PublicKey).p, false)
	p.p = agg.ToAffine()

	return p
}

// Hash calculates the hash of the public key.
func (p *PublicKey) Hash() ([20]byte, error) {
	pkS := p.Marshal()
	h := chainhash.HashH(pkS[:])
	var hBytes [20]byte
	copy(hBytes[:], h[:])
	return hBytes, nil
}

// ToAccount converts the public key to a Bech32 address.
func (p *PublicKey) ToAccount(prefix *params.AccountPrefixes) string {
	hash, _ := p.Hash()
	return bech32.Encode(prefix.Public, hash[:])
}
*/
