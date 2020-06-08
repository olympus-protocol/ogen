package bls

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/utils/bech32"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
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
	PublicKeys []*PublicKey
	NumNeeded  uint16
}

// NewMultipub constructs a new multi-pubkey.
func NewMultipub(pubs []*PublicKey, numNeeded uint16) *Multipub {
	return &Multipub{
		PublicKeys: pubs,
		NumNeeded:  numNeeded,
	}
}

// Encode encodes the public key to the given writer.
func (m *Multipub) Encode(w io.Writer) error {
	if err := serializer.WriteVarInt(w, uint64(len(m.PublicKeys))); err != nil {
		return err
	}

	for _, p := range m.PublicKeys {
		if _, err := w.Write(p.Marshal()); err != nil {
			return err
		}
	}

	if err := serializer.WriteElement(w, m.NumNeeded); err != nil {
		return err
	}

	return nil
}

// Decode decodes the multipub to bytes.
func (m *Multipub) Decode(r io.Reader) error {
	numPubkeys, err := serializer.ReadVarInt(r)
	if err != nil {
		return err
	}

	m.PublicKeys = make([]*PublicKey, numPubkeys)
	for i := range m.PublicKeys {
		pkb := make([]byte, 48)
		if _, err := io.ReadFull(r, pkb); err != nil {
			return err
		}

		m.PublicKeys[i], err = PublicKeyFromBytes(pkb)
		if err != nil {
			return err
		}
	}

	if err := serializer.ReadElement(r, &m.NumNeeded); err != nil {
		return err
	}

	return nil
}

// ToHash gets the hash of the multipub.
func (m *Multipub) ToHash() []byte {
	numNeeded := make([]byte, 2)
	binary.BigEndian.PutUint16(numNeeded, m.NumNeeded)
	out := make([]byte, 0, 2+48*len(m.PublicKeys))
	out = append(out, numNeeded...)
	for _, p := range m.PublicKeys {
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
	PublicKey  Multipub
	Signatures []*Signature
	KeysSigned Bitfield
}

// NewMultisig creates a new blank multisig.
func NewMultisig(multipub Multipub) *Multisig {
	return &Multisig{
		PublicKey:  multipub,
		Signatures: []*Signature{},
		KeysSigned: NewBitfield(uint(len(multipub.PublicKeys))),
	}
}

// GetPublicKey gets the public key included in the signature.
func (m *Multisig) GetPublicKey() FunctionalPublicKey {
	return &m.PublicKey
}

// Encode encodes the multisig to the given writer.
func (m *Multisig) Encode(w io.Writer) error {
	if err := m.PublicKey.Encode(w); err != nil {
		return err
	}

	if err := serializer.WriteVarInt(w, uint64(len(m.Signatures))); err != nil {
		return err
	}

	for _, s := range m.Signatures {
		bs := s.Marshal()
		if _, err := w.Write(bs); err != nil {
			return err
		}
	}

	if err := serializer.WriteVarBytes(w, m.KeysSigned); err != nil {
		return err
	}

	return nil
}

// Decode decodes the multisig from the given reader.
func (m *Multisig) Decode(r io.Reader) error {
	if err := m.PublicKey.Decode(r); err != nil {
		return err
	}

	numSigs, err := serializer.ReadVarInt(r)
	if err != nil {
		return err
	}

	m.Signatures = make([]*Signature, numSigs)

	for i := range m.Signatures {
		sigBytes := make([]byte, 96)
		_, err := io.ReadFull(r, sigBytes)
		if err != nil {
			return err
		}

		sig, err := SignatureFromBytes(sigBytes)
		if err != nil {
			return err
		}

		m.Signatures[i] = sig
	}

	bitfield, err := serializer.ReadVarBytes(r)
	if err != nil {
		return err
	}

	m.KeysSigned = bitfield

	return nil
}

// Sign signs a multisig through a secret key.
func (m *Multisig) Sign(secKey *SecretKey, msg []byte) error {
	pub := secKey.PublicKey()

	idx := -1
	for i := range m.PublicKey.PublicKeys {
		if m.PublicKey.PublicKeys[i].Equals(pub) {
			idx = i
		}
	}

	if idx == -1 {
		return fmt.Errorf("could not find public key %x in multipub", pub.Marshal())
	}

	if m.KeysSigned.Get(uint(idx)) {
		return nil
	}

	msgI := chainhash.HashH(append(msg, pub.Marshal()...))

	sig := secKey.Sign(msgI[:])

	m.Signatures = append(m.Signatures, sig)
	m.KeysSigned.Set(uint(idx))

	return nil
}

// Verify verifies a multisig message.
func (m *Multisig) Verify(msg []byte) bool {
	if uint(len(m.PublicKey.PublicKeys)) > m.KeysSigned.MaxLength() {
		return false
	}

	if len(m.Signatures) < int(m.PublicKey.NumNeeded) {
		return false
	}

	aggSig := AggregateSignatures(m.Signatures)

	activePubs := make([]*PublicKey, 0)
	for i := range m.PublicKey.PublicKeys {
		if m.KeysSigned.Get(uint(i)) {
			activePubs = append(activePubs, m.PublicKey.PublicKeys[i])
		}
	}

	if len(m.Signatures) != len(activePubs) {
		return false
	}

	msgs := make([][32]byte, len(m.Signatures))
	for i := range msgs {
		msgs[i] = chainhash.HashH(append(msg, activePubs[i].Marshal()...))
	}

	return aggSig.AggregateVerify(activePubs, msgs)
}

var _ FunctionalSignature = &Multisig{}
var _ FunctionalPublicKey = &Multipub{}
