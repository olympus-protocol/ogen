package bls

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/utils/bech32"
	"github.com/olympus-protocol/ogen/utils/bitfield"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/prysmaticlabs/go-ssz"
)

// Multipub represents multiple public keys that can be signed by
// some subset numNeeded.
type Multipub struct {
	PublicKeys [][]byte
	NumNeeded  uint16
}

// Marshal encodes the data.
func (m *Multipub) Marshal() []byte {
	b, _ := ssz.Marshal(m)
	return b
}

// Unmarshal decodes the data.
func (m *Multipub) Unmarshal(b []byte) error {
	return ssz.Unmarshal(b, m)
}

// NewMultipub constructs a new multi-pubkey.
func NewMultipub(pubs []*PublicKey, numNeeded uint16) *Multipub {
	pubsB := [][]byte{}
	for _, p := range pubs {
		pubsB = append(pubsB, p.Marshal())
	}
	return &Multipub{
		PublicKeys: pubsB,
		NumNeeded:  numNeeded,
	}
}

// Copy returns a copy of the multipub.
func (m *Multipub) Copy() *Multipub {
	newM := *m
	newM.PublicKeys = make([][]byte, len(m.PublicKeys))
	for i := range newM.PublicKeys {
		newM.PublicKeys[i] = m.PublicKeys[i]
	}

	return &newM
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
func (m *Multipub) Hash() ([20]byte, error) {
	pubkeyHashes := make([][20]byte, 0, len(m.PublicKeys))

	for i, p := range m.PublicKeys {
		pub, err := PublicKeyFromBytes(p)
		if err != nil {
			return [20]byte{}, err
		}
		pubkeyHashes[i], err = pub.Hash()
		if err != nil {
			return [20]byte{}, err
		}
	}

	return PublicKeyHashesToMultisigHash(pubkeyHashes, m.NumNeeded), nil
}

// ToBech32 returns the bech32 address.
func (m *Multipub) ToBech32(prefixes params.AddrPrefixes) string {
	pkh, _ := m.Hash()
	return bech32.Encode(prefixes.Multisig, pkh[:])
}

// Multisig represents an m-of-n multisig.
type Multisig struct {
	PublicKey  Multipub
	Signatures [][]byte
	KeysSigned bitfield.Bitfield
}

// Marshal encodes the data.
func (m *Multisig) Marshal() ([]byte, error) {
	return ssz.Marshal(m)
}

// Unmarshal decodes the data.
func (m *Multisig) Unmarshal(b []byte) error {
	return ssz.Unmarshal(b, m)
}

// NewMultisig creates a new blank multisig.
func NewMultisig(multipub Multipub) *Multisig {
	return &Multisig{
		PublicKey:  multipub,
		Signatures: [][]byte{},
		KeysSigned: bitfield.NewBitfield(uint(len(multipub.PublicKeys))),
	}
}

// GetPublicKey gets the public key included in the signature.
func (m *Multisig) GetPublicKey() (FunctionalPublicKey, error) {
	return &m.PublicKey, nil
}

// Sign signs a multisig through a secret key.
func (m *Multisig) Sign(secKey *SecretKey, msg []byte) error {
	pub := secKey.PublicKey()

	idx := -1
	for i := range m.PublicKey.PublicKeys {
		if bytes.Equal(m.PublicKey.PublicKeys[i], pub.Marshal()) {
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

	m.Signatures = append(m.Signatures, sig.Marshal())
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

	aggSig, err := AggregateSignaturesBytes(m.Signatures)
	if err != nil {
		return false
	}
	activePubs := make([][]byte, 0)
	activePubsKeys := make([]*PublicKey, 0)
	for i := range m.PublicKey.PublicKeys {
		if m.KeysSigned.Get(uint(i)) {
			activePubs = append(activePubs, m.PublicKey.PublicKeys[i])
			pub, err := PublicKeyFromBytes(m.PublicKey.PublicKeys[i])
			if err != nil {
				return false
			}
			activePubsKeys = append(activePubsKeys, pub)
		}
	}

	if len(m.Signatures) != len(activePubs) {
		return false
	}

	msgs := make([][32]byte, len(m.Signatures))
	for i := range msgs {
		msgs[i] = chainhash.HashH(append(msg, activePubs[i]...))
	}

	return aggSig.AggregateVerify(activePubsKeys, msgs)
}

// Type returns the type of the multisig.
func (m *Multisig) Type() FunctionalSignatureType {
	return TypeMulti
}

// Copy copies the signature.
func (m *Multisig) Copy() FunctionalSignature {
	newMultisig := &Multisig{}
	newMultisig.Signatures = make([][]byte, len(m.Signatures))
	for i := range newMultisig.Signatures {
		newMultisig.Signatures[i] = m.Signatures[i]
	}

	pub := m.PublicKey.Copy()
	newMultisig.PublicKey = *pub

	newMultisig.KeysSigned = m.KeysSigned.Copy()

	return newMultisig
}

var _ FunctionalSignature = &Multisig{}
var _ FunctionalPublicKey = &Multipub{}
