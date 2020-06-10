package bls

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/utils/bech32"
	"github.com/olympus-protocol/ogen/utils/bitfield"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
)

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

// Copy returns a copy of the multipub.
func (m *Multipub) Copy() *Multipub {
	newM := *m
	newM.PublicKeys = make([]*PublicKey, len(m.PublicKeys))
	for i := range newM.PublicKeys {
		newM.PublicKeys[i] = m.PublicKeys[i].Copy()
	}

	return &newM
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

// Type returns the type of the multipub.
func (m *Multipub) Type() FunctionalSignatureType {
	return TypeMulti
}

func PublicKeyHashesToMultisigHash(pubkeys [][20]byte, numNeeded uint16) [20]byte {
	out := make([]byte, 0, 2+20*len(pubkeys))

	numNeededBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(numNeededBytes, numNeeded)

	out = append(out, out...)
	for _, p := range pubkeys {
		out = append(out, p[:]...)
	}

	h := chainhash.HashH(out)
	var h20 [20]byte
	copy(h20[:], h[:20])

	return h20
}

// Hash gets the hash of the multipub.
func (m *Multipub) Hash() [20]byte {
	pubkeyHashes := make([][20]byte, 0, len(m.PublicKeys))

	for i, p := range m.PublicKeys {
		pubkeyHashes[i] = p.Hash()
	}

	return PublicKeyHashesToMultisigHash(pubkeyHashes, m.NumNeeded)
}

// ToBech32 returns the bech32 address.
func (m *Multipub) ToBech32(prefixes params.AddrPrefixes) string {
	pkh := m.Hash()
	return bech32.Encode(prefixes.Multisig, pkh[:])
}

// Multisig represents an m-of-n multisig.
type Multisig struct {
	PublicKey  Multipub
	Signatures []*Signature
	KeysSigned bitfield.Bitfield
}

// NewMultisig creates a new blank multisig.
func NewMultisig(multipub Multipub) *Multisig {
	return &Multisig{
		PublicKey:  multipub,
		Signatures: []*Signature{},
		KeysSigned: bitfield.NewBitfield(uint(len(multipub.PublicKeys))),
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

// Type returns the type of the multisig.
func (m *Multisig) Type() FunctionalSignatureType {
	return TypeMulti
}

// Copy copies the signature.
func (m *Multisig) Copy() FunctionalSignature {
	newMultisig := &Multisig{}
	newMultisig.Signatures = make([]*Signature, len(m.Signatures))
	for i := range newMultisig.Signatures {
		newMultisig.Signatures[i] = m.Signatures[i].Copy()
	}

	pub := m.PublicKey.Copy()
	newMultisig.PublicKey = *pub

	newMultisig.KeysSigned = m.KeysSigned.Copy()

	return newMultisig
}

var _ FunctionalSignature = &Multisig{}
var _ FunctionalPublicKey = &Multipub{}
