package p2p

import (
	ssz "github.com/ferranbt/fastssz"
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

// MarshalSSZ ssz marshals the MsgNewState object
func (m *MsgNewState) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(m)
}

// MarshalSSZTo ssz marshals the MsgNewState object to a target array
func (m *MsgNewState) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf
	offset := int(4)

	// Offset (0) 'Data'
	dst = ssz.WriteOffset(dst, offset)
	if m.Data == nil {
		m.Data = new(primitives.SerializableState)
	}
	offset += m.Data.SizeSSZ()

	// Field (0) 'Data'
	if dst, err = m.Data.MarshalSSZTo(dst); err != nil {
		return
	}

	return
}

// UnmarshalSSZ ssz unmarshals the MsgGetBlocks object
func (m *MsgNewState) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size < 4 {
		return ssz.ErrSize
	}

	tail := buf
	var o0 uint64

	// Offset (0) 'Data'
	if o0 = ssz.ReadOffset(buf[0:4]); o0 > size {
		return ssz.ErrOffset
	}

	// Field (0) 'Data'
	{
		buf = tail[o0:]
		if m.Data == nil {
			m.Data = new(primitives.SerializableState)
		}
		if err = m.Data.UnmarshalSSZ(buf); err != nil {
			return err
		}
	}
	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the MsgGet object
func (m *MsgNewState) SizeSSZ() (size int) {
	size = 32
	return
}

// HashTreeRoot ssz hashes the MsgNewState object
func (m *MsgNewState) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(m)
}

// HashTreeRootWith ssz hashes the MsgNewState object with a hasher
func (m *MsgNewState) HashTreeRootWith(hh *ssz.Hasher) (err error) {
	indx := hh.Index()

	// Field (0) 'Data'
	if err = m.Data.HashTreeRootWith(hh); err != nil {
		return
	}

	hh.Merkleize(indx)
	return
}
