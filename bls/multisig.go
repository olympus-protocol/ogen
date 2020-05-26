package bls

import (
	"encoding/binary"
	"fmt"

	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/utils/bech32"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

// Bitfield is a bitfield of a certain length.
type Bitfield []byte

// NewBitfield constructs a new bitfield containing a certain length.
func NewBitfield(l uint) Bitfield {
	return make([]byte, (l+7)/8)
}

// Set sets bit i
func (b Bitfield) Set(i uint) {
	b[i/8] |= (1 << (i % 8))
}

// Get gets bit i
func (b Bitfield) Get(i uint) bool {
	return b[i/8]&(1<<(i%8)) != 0
}

// MaxLength is the maximum number of elements the bitfield can hold.
func (b Bitfield) MaxLength() uint {
	return uint(len(b)) * 8
}

// Multipub represents multiple public keys that can be signed by
// some subset numNeeded.
type Multipub struct {
	pubkeys   []*PublicKey
	numNeeded uint16
}

// NewMultipub constructs a new multi-pubkey.
func NewMultipub(pubs []*PublicKey, numNeeded uint16) *Multipub {
	return &Multipub{
		pubkeys:   pubs,
		numNeeded: numNeeded,
	}
}

// ToHash gets the hash of the multipub.
func (m *Multipub) ToHash() []byte {
	numNeeded := make([]byte, 2)
	binary.BigEndian.PutUint16(numNeeded, m.numNeeded)
	out := make([]byte, 0, 2+48*len(m.pubkeys))
	out = append(out, numNeeded...)
	for _, p := range m.pubkeys {
		out = append(out, p.Marshal()...)
	}

	h := chainhash.HashH(out)
	return h[:20]
}

// ToBech32 returns the bech32 address.
func (m *Multipub) ToBech32(prefixes params.AddrPrefixes) string {
	return bech32.Encode(prefixes.Multisig, m.ToHash())
}

// Multisig represents an m-of-n multisig.
type Multisig struct {
	pub        *Multipub
	signatures []*Signature
	keysSigned Bitfield
	msg        []byte
}

// NewMultisig creates a new blank multisig.
func NewMultisig(multipub *Multipub, msg []byte) *Multisig {
	return &Multisig{
		pub:        multipub,
		signatures: []*Signature{},
		keysSigned: NewBitfield(uint(len(multipub.pubkeys))),
		msg:        msg,
	}
}

// Sign signs a multisig through a secret key.
func (m *Multisig) Sign(secKey *SecretKey) error {
	pub := secKey.PublicKey()

	idx := -1
	for i := range m.pub.pubkeys {
		if m.pub.pubkeys[i].Equals(pub) {
			idx = i
		}
	}

	if idx == -1 {
		return fmt.Errorf("could not find public key %x in multipub", pub.Marshal())
	}

	if m.keysSigned.Get(uint(idx)) {
		return nil
	}

	msgI := chainhash.HashH(append(m.msg, pub.Marshal()...))

	sig := secKey.Sign(msgI[:])

	m.signatures = append(m.signatures, sig)
	m.keysSigned.Set(uint(idx))

	return nil
}

// Verify verifies a multisig message.
func (m *Multisig) Verify() bool {
	if uint(len(m.pub.pubkeys)) > m.keysSigned.MaxLength() {
		return false
	}

	if len(m.signatures) < int(m.pub.numNeeded) {
		return false
	}

	aggSig := AggregateSignatures(m.signatures)

	activePubs := make([]*PublicKey, 0)
	for i := range m.pub.pubkeys {
		if m.keysSigned.Get(uint(i)) {
			activePubs = append(activePubs, m.pub.pubkeys[i])
		}
	}

	if len(m.signatures) != len(activePubs) {
		return false
	}

	msgs := make([][32]byte, len(m.signatures))
	for i := range msgs {
		msgs[i] = chainhash.HashH(append(m.msg, activePubs[i].Marshal()...))
	}

	return aggSig.AggregateVerify(activePubs, msgs)
}
