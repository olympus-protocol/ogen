// Code generated by fastssz. DO NOT EDIT.
package p2p

import (
	ssz "github.com/ferranbt/fastssz"
)

// MarshalSSZ ssz marshals the MsgFinalization object
func (m *MsgFinalization) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(m)
}

// MarshalSSZTo ssz marshals the MsgFinalization object to a target array
func (m *MsgFinalization) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf

	// Field (0) 'Tip'
	dst = ssz.MarshalUint64(dst, m.Tip)

	// Field (1) 'TipSlot'
	dst = ssz.MarshalUint64(dst, m.TipSlot)

	// Field (2) 'TipHash'
	dst = append(dst, m.TipHash[:]...)

	// Field (3) 'JustifiedSlot'
	dst = ssz.MarshalUint64(dst, m.JustifiedSlot)

	// Field (4) 'JustifiedHeight'
	dst = ssz.MarshalUint64(dst, m.JustifiedHeight)

	// Field (5) 'JustifiedHash'
	dst = append(dst, m.JustifiedHash[:]...)

	// Field (6) 'FinalizedSlot'
	dst = ssz.MarshalUint64(dst, m.FinalizedSlot)

	// Field (7) 'FinalizedHeight'
	dst = ssz.MarshalUint64(dst, m.FinalizedHeight)

	// Field (8) 'FinalizedHash'
	dst = append(dst, m.FinalizedHash[:]...)

	return
}

// UnmarshalSSZ ssz unmarshals the MsgFinalization object
func (m *MsgFinalization) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size != 144 {
		return ssz.ErrSize
	}

	// Field (0) 'Tip'
	m.Tip = ssz.UnmarshallUint64(buf[0:8])

	// Field (1) 'TipSlot'
	m.TipSlot = ssz.UnmarshallUint64(buf[8:16])

	// Field (2) 'TipHash'
	copy(m.TipHash[:], buf[16:48])

	// Field (3) 'JustifiedSlot'
	m.JustifiedSlot = ssz.UnmarshallUint64(buf[48:56])

	// Field (4) 'JustifiedHeight'
	m.JustifiedHeight = ssz.UnmarshallUint64(buf[56:64])

	// Field (5) 'JustifiedHash'
	copy(m.JustifiedHash[:], buf[64:96])

	// Field (6) 'FinalizedSlot'
	m.FinalizedSlot = ssz.UnmarshallUint64(buf[96:104])

	// Field (7) 'FinalizedHeight'
	m.FinalizedHeight = ssz.UnmarshallUint64(buf[104:112])

	// Field (8) 'FinalizedHash'
	copy(m.FinalizedHash[:], buf[112:144])

	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the MsgFinalization object
func (m *MsgFinalization) SizeSSZ() (size int) {
	size = 144
	return
}

// HashTreeRoot ssz hashes the MsgFinalization object
func (m *MsgFinalization) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(m)
}

// HashTreeRootWith ssz hashes the MsgFinalization object with a hasher
func (m *MsgFinalization) HashTreeRootWith(hh *ssz.Hasher) (err error) {
	indx := hh.Index()

	// Field (0) 'Tip'
	hh.PutUint64(m.Tip)

	// Field (1) 'TipSlot'
	hh.PutUint64(m.TipSlot)

	// Field (2) 'TipHash'
	hh.PutBytes(m.TipHash[:])

	// Field (3) 'JustifiedSlot'
	hh.PutUint64(m.JustifiedSlot)

	// Field (4) 'JustifiedHeight'
	hh.PutUint64(m.JustifiedHeight)

	// Field (5) 'JustifiedHash'
	hh.PutBytes(m.JustifiedHash[:])

	// Field (6) 'FinalizedSlot'
	hh.PutUint64(m.FinalizedSlot)

	// Field (7) 'FinalizedHeight'
	hh.PutUint64(m.FinalizedHeight)

	// Field (8) 'FinalizedHash'
	hh.PutBytes(m.FinalizedHash[:])

	hh.Merkleize(indx)
	return
}
