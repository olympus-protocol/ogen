package bls

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/utils/bech32"
	"github.com/olympus-protocol/ogen/utils/bitfield"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

// Multipub represents multiple public keys that can be signed by some subset numNeeded.
type Multipub struct {
	PublicKeys [][]byte `ssz-size:"?,32" ssz-max:"16777216"`
	NumNeeded  uint16
}

// NewMultipub constructs a new multi-pubkey.
func NewMultipub(pubs [][]byte, numNeeded uint16) *Multipub {
	return &Multipub{
		PublicKeys: pubs,
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

func PublicKeyHashesToMultisigHash(pubkeys [][]byte, numNeeded uint16) []byte {
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

	return h20[:]
}

// Hash gets the hash of the multipub.
func (m *Multipub) Hash() []byte {
	pubkeyHashes := make([][]byte, 0, len(m.PublicKeys))

	for i, p := range m.PublicKeys {
		// TODO handle error
		pub, _ := PublicKeyFromBytes(p)
		pubkeyHashes[i] = pub.Hash()
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
	PublicKey  *Multipub
	Signatures [][]byte          `ssz-size:"?,32" ssz-max:"32"`
	KeysSigned bitfield.Bitfield `ssz-max:"32"`
}

// NewMultisig creates a new blank multisig.
func NewMultisig(multipub *Multipub) *Multisig {
	return &Multisig{
		PublicKey:  multipub,
		Signatures: [][]byte{},
		KeysSigned: bitfield.NewBitfield(uint(len(multipub.PublicKeys))),
	}
}

// Sign signs a multisig through a secret key.
func (m *Multisig) Sign(secKey *SecretKey, msg []byte) error {
	pub := secKey.PublicKey().Marshal()

	idx := -1
	for i := range m.PublicKey.PublicKeys {
		if bytes.Equal(m.PublicKey.PublicKeys[i], pub) {
			idx = i
		}
	}
	if idx == -1 {
		return fmt.Errorf("could not find public key %x in multipub", pub)
	}

	if m.KeysSigned.Get(uint(idx)) {
		return nil
	}
	msgI := chainhash.HashH(append(msg, pub...))

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

	aggSig := AggregateSignaturesBytes(m.Signatures)

	activePubsBytes := make([][]byte, 0)
	activePubs := make([]*PublicKey, 0)
	for i := range m.PublicKey.PublicKeys {
		if m.KeysSigned.Get(uint(i)) {
			pub, _ := PublicKeyFromBytes(m.PublicKey.PublicKeys[i])
			activePubs = append(activePubs, pub)
			activePubsBytes = append(activePubsBytes, m.PublicKey.PublicKeys[i])
		}
	}

	if len(m.Signatures) != len(activePubs) {
		return false
	}

	msgs := make([][32]byte, len(m.Signatures))
	for i := range msgs {
		msgs[i] = chainhash.HashH(append(msg, activePubsBytes[i]...))
	}

	return aggSig.AggregateVerify(activePubs, msgs)
}

// Copy copies the signature.
func (m *Multisig) Copy() *Multisig {
	newMultisig := &Multisig{}
	newMultisig.Signatures = make([][]byte, len(m.Signatures))
	for i := range newMultisig.Signatures {
		newMultisig.Signatures[i] = m.Signatures[i]
	}

	pub := m.PublicKey.Copy()
	newMultisig.PublicKey = pub

	newMultisig.KeysSigned = m.KeysSigned.Copy()

	return newMultisig
}
