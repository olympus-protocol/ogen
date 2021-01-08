package multisig

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/olympus-protocol/ogen/pkg/bitfield"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/bls/common"

	"github.com/olympus-protocol/ogen/pkg/bech32"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
)

const (
	// MaxMultipubSize is the maximum amount of bytes a Multipub key can contain. 32 public keys.
	MaxMultipubSize = (32 * 48) + 8
	// MaxMultisigSize is the maximum amount of bytes a Multisig can contain. 32 public keys and 32 signatures.
	MaxMultisigSize = MaxMultipubSize + (96 * 32) + 5 + 8
)

// Multipub represents multiple public keys that can be signed by some subset numNeeded.
type Multipub struct {
	PublicKeys [][48]byte `ssz-max:"32"`
	NumNeeded  uint64
}

// Marshal encodes the data.
func (m *Multipub) Marshal() ([]byte, error) {
	return m.MarshalSSZ()
}

// Unmarshal decodes the data.
func (m *Multipub) Unmarshal(b []byte) error {
	return m.UnmarshalSSZ(b)
}

// NewMultipub constructs a new multi-pubkey.
func NewMultipub(pubs []common.PublicKey, numNeeded uint64) *Multipub {
	var pubsB [][48]byte
	for _, p := range pubs {
		var pub [48]byte
		copy(pub[:], p.Marshal())
		pubsB = append(pubsB, pub)
	}
	return &Multipub{
		PublicKeys: pubsB,
		NumNeeded:  numNeeded,
	}
}

// Copy returns a copy of the multipub.
func (m *Multipub) Copy() *Multipub {
	newM := &Multipub{}
	newM.PublicKeys = make([][48]byte, len(m.PublicKeys))
	for i := range newM.PublicKeys {
		newM.PublicKeys[i] = m.PublicKeys[i]
	}
	newM.NumNeeded = m.NumNeeded
	return newM
}

// PublicKeyHashesToMultisigHash returns the hash of multiple publickey hashes
func PublicKeyHashesToMultisigHash(pubkeys [][20]byte, numNeeded uint64) [20]byte {
	keys := len(pubkeys)
	out := make([]byte, (20*keys)+8)

	numNeededBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(numNeededBytes, numNeeded)

	for _, p := range pubkeys {
		out = append(out, p[:]...)
	}
	out = append(out, numNeededBytes...)
	h := chainhash.HashH(out)
	var h20 [20]byte
	copy(h20[:], h[:20])

	return h20
}

// Hash gets the hash of the multipub.
func (m *Multipub) Hash() ([20]byte, error) {
	pubkeyHashes := make([][20]byte, len(m.PublicKeys))

	for i, p := range m.PublicKeys {
		pub, err := bls.PublicKeyFromBytes(p[:])
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
func (m *Multipub) ToBech32() (string, error) {
	pkh, err := m.Hash()
	if err != nil {
		return "", err
	}
	return bech32.Encode(bls.Prefix.Public, pkh[:]), nil
}

// Multisig represents an m-of-n multisig.
type Multisig struct {
	PublicKey  *Multipub
	Signatures [][96]byte       `ssz-max:"32"`
	KeysSigned bitfield.Bitlist `ssz:"bitlist" ssz-max:"32"`
}

// Marshal encodes the data.
func (m *Multisig) Marshal() ([]byte, error) {
	return m.MarshalSSZ()
}

// Unmarshal decodes the data.
func (m *Multisig) Unmarshal(b []byte) error {
	return m.UnmarshalSSZ(b)
}

// NewMultisig creates a new blank multisig.
func NewMultisig(multipub *Multipub) *Multisig {
	return &Multisig{
		PublicKey:  multipub,
		Signatures: [][96]byte{},
		KeysSigned: bitfield.NewBitlist(uint64(len(multipub.PublicKeys))),
	}
}

// GetPublicKey gets the public key included in the signature.
func (m *Multisig) GetPublicKey() (*Multipub, error) {
	return m.PublicKey, nil
}

// Sign signs a multisig through a secret key.
func (m *Multisig) Sign(secKey common.SecretKey, msg []byte) error {
	pub := secKey.PublicKey()

	idx := -1
	for i := range m.PublicKey.PublicKeys {
		if bytes.Equal(m.PublicKey.PublicKeys[i][:], pub.Marshal()) {
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
	var s [96]byte
	copy(s[:], sig.Marshal())
	m.Signatures = append(m.Signatures, s)
	m.KeysSigned.Set(uint(idx))

	return nil
}

// Verify verifies a multisig message.
func (m *Multisig) Verify(msg []byte) bool {

	if uint(len(m.PublicKey.PublicKeys)) > uint(len(m.KeysSigned))*8 {
		return false
	}

	if len(m.Signatures) < int(m.PublicKey.NumNeeded) {
		return false
	}

	var sigs []common.Signature

	for _, sigBytes := range m.Signatures {
		sig, err := bls.SignatureFromBytes(sigBytes[:])
		if err != nil {
			return false
		}
		sigs = append(sigs, sig)
	}

	aggSig := bls.AggregateSignatures(sigs)

	activePubs := make([][48]byte, 0)
	activePubsKeys := make([]common.PublicKey, 0)

	for i := range m.PublicKey.PublicKeys {
		if m.KeysSigned.Get(uint(i)) {
			activePubs = append(activePubs, m.PublicKey.PublicKeys[i])
			pub, err := bls.PublicKeyFromBytes(m.PublicKey.PublicKeys[i][:])
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
		msgs[i] = chainhash.HashH(append(msg, activePubs[i][:]...))
	}

	return aggSig.AggregateVerify(activePubsKeys, msgs)
}

// Copy copies the signature.
func (m *Multisig) Copy() *Multisig {
	newMultisig := &Multisig{}
	newMultisig.Signatures = make([][96]byte, len(m.Signatures))
	for i := range newMultisig.Signatures {
		newMultisig.Signatures[i] = m.Signatures[i]
	}

	pub := m.PublicKey.Copy()
	newMultisig.PublicKey = pub

	newMultisig.KeysSigned = bitfield.NewBitlist(m.KeysSigned.Len())
	for i, b := range m.KeysSigned {
		newMultisig.KeysSigned[i] = b
	}

	return newMultisig
}
