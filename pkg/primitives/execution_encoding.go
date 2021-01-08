// Code generated by fastssz. DO NOT EDIT.
package primitives

import (
	ssz "github.com/ferranbt/fastssz"
)

// MarshalSSZ ssz marshals the Execution object
func (e *Execution) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(e)
}

// MarshalSSZTo ssz marshals the Execution object to a target array
func (e *Execution) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf
	offset := int(184)

	// Field (0) 'FromPubKey'
	dst = append(dst, e.FromPubKey[:]...)

	// Field (1) 'To'
	dst = append(dst, e.To[:]...)

	// Offset (2) 'Input'
	dst = ssz.WriteOffset(dst, offset)
	offset += len(e.Input)

	// Field (3) 'Signature'
	dst = append(dst, e.Signature[:]...)

	// Field (4) 'Gas'
	dst = ssz.MarshalUint64(dst, e.Gas)

	// Field (5) 'GasLimit'
	dst = ssz.MarshalUint64(dst, e.GasLimit)

	// Field (2) 'Input'
	if len(e.Input) > 7168 {
		err = ssz.ErrBytesLength
		return
	}
	dst = append(dst, e.Input...)

	return
}

// UnmarshalSSZ ssz unmarshals the Execution object
func (e *Execution) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size < 184 {
		return ssz.ErrSize
	}

	tail := buf
	var o2 uint64

	// Field (0) 'FromPubKey'
	copy(e.FromPubKey[:], buf[0:48])

	// Field (1) 'To'
	copy(e.To[:], buf[48:68])

	// Offset (2) 'Input'
	if o2 = ssz.ReadOffset(buf[68:72]); o2 > size {
		return ssz.ErrOffset
	}

	// Field (3) 'Signature'
	copy(e.Signature[:], buf[72:168])

	// Field (4) 'Gas'
	e.Gas = ssz.UnmarshallUint64(buf[168:176])

	// Field (5) 'GasLimit'
	e.GasLimit = ssz.UnmarshallUint64(buf[176:184])

	// Field (2) 'Input'
	{
		buf = tail[o2:]
		if len(buf) > 7168 {
			return ssz.ErrBytesLength
		}
		if cap(e.Input) == 0 {
			e.Input = make([]byte, 0, len(buf))
		}
		e.Input = append(e.Input, buf...)
	}
	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the Execution object
func (e *Execution) SizeSSZ() (size int) {
	size = 184

	// Field (2) 'Input'
	size += len(e.Input)

	return
}

// HashTreeRoot ssz hashes the Execution object
func (e *Execution) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(e)
}

// HashTreeRootWith ssz hashes the Execution object with a hasher
func (e *Execution) HashTreeRootWith(hh *ssz.Hasher) (err error) {
	indx := hh.Index()

	// Field (0) 'FromPubKey'
	hh.PutBytes(e.FromPubKey[:])

	// Field (1) 'To'
	hh.PutBytes(e.To[:])

	// Field (2) 'Input'
	if len(e.Input) > 7168 {
		err = ssz.ErrBytesLength
		return
	}
	hh.PutBytes(e.Input)

	// Field (3) 'Signature'
	hh.PutBytes(e.Signature[:])

	// Field (4) 'Gas'
	hh.PutUint64(e.Gas)

	// Field (5) 'GasLimit'
	hh.PutUint64(e.GasLimit)

	hh.Merkleize(indx)
	return
}
