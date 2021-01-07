package herumi
/*
import (
	"fmt"
	"github.com/olympus-protocol/ogen/pkg/bech32"
	"github.com/olympus-protocol/ogen/pkg/bls/common"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/params"

	"github.com/dgraph-io/ristretto"
	bls12 "github.com/herumi/bls-eth-go-binary/bls"
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
	p *bls12.PublicKey
}

// PublicKeyFromBytes creates a BLS public key from a  BigEndian byte slice.
func (h *Herumi) PublicKeyFromBytes(pubKey []byte) (common.PublicKey, error) {
	if len(pubKey) != 48 {
		return nil, fmt.Errorf("public key must be %d bytes", 48)
	}
	if cv, ok := pubkeyCache.Get(string(pubKey)); ok {
		return cv.(*PublicKey).Copy(), nil
	}
	p := &bls12.PublicKey{}
	err := p.Deserialize(pubKey)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal bytes into public key")
	}
	pubKeyObj := &PublicKey{p: p}
	if pubKeyObj.IsInfinite() {
		return nil, common.ErrInfinitePubKey
	}
	pubkeyCache.Set(string(pubKey), pubKeyObj.Copy(), 48)
	return pubKeyObj, nil
}

// AggregatePublicKeys aggregates the provided raw public keys into a single key.
func (h *Herumi) AggregatePublicKeys(pubs [][]byte) (common.PublicKey, error) {
	if len(pubs) == 0 {
		return &PublicKey{}, nil
	}
	p, err := h.PublicKeyFromBytes(pubs[0])
	if err != nil {
		return nil, err
	}
	for _, k := range pubs[1:] {
		pubkey, err := h.PublicKeyFromBytes(k)
		if err != nil {
			return nil, err
		}
		p.Aggregate(pubkey)
	}
	return p, nil
}

// Marshal a public key into a LittleEndian byte slice.
func (p *PublicKey) Marshal() []byte {
	return p.p.Serialize()
}

// Copy the public key to a new pointer reference.
func (p *PublicKey) Copy() common.PublicKey {
	np := *p.p
	return &PublicKey{p: &np}
}

// IsInfinite checks if the public key is infinite.
func (p *PublicKey) IsInfinite() bool {
	return p.p.IsZero()
}

// Aggregate two public keys.
func (p *PublicKey) Aggregate(p2 common.PublicKey) common.PublicKey {
	p.p.Add(p2.(*PublicKey).p)
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