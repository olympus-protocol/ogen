// Code generated by fastssz. DO NOT EDIT.
package primitives

import (
	ssz "github.com/ferranbt/fastssz"
)

// MarshalSSZ ssz marshals the AccountInfo object
func (a *AccountInfo) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(a)
}

// MarshalSSZTo ssz marshals the AccountInfo object to a target array
func (a *AccountInfo) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf

	// Field (0) 'Account'
	dst = append(dst, a.Account[:]...)

	// Field (1) 'Info'
	dst = ssz.MarshalUint64(dst, a.Info)

	return
}

// UnmarshalSSZ ssz unmarshals the AccountInfo object
func (a *AccountInfo) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size != 28 {
		return ssz.ErrSize
	}

	// Field (0) 'Account'
	copy(a.Account[:], buf[0:20])

	// Field (1) 'Info'
	a.Info = ssz.UnmarshallUint64(buf[20:28])

	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the AccountInfo object
func (a *AccountInfo) SizeSSZ() (size int) {
	size = 28
	return
}

// HashTreeRoot ssz hashes the AccountInfo object
func (a *AccountInfo) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(a)
}

// HashTreeRootWith ssz hashes the AccountInfo object with a hasher
func (a *AccountInfo) HashTreeRootWith(hh *ssz.Hasher) (err error) {
	indx := hh.Index()

	// Field (0) 'Account'
	hh.PutBytes(a.Account[:])

	// Field (1) 'Info'
	hh.PutUint64(a.Info)

	hh.Merkleize(indx)
	return
}

// MarshalSSZ ssz marshals the CoinsStateSerializable object
func (c *CoinsStateSerializable) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(c)
}

// MarshalSSZTo ssz marshals the CoinsStateSerializable object to a target array
func (c *CoinsStateSerializable) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf
	offset := int(8)

	// Offset (0) 'Balances'
	dst = ssz.WriteOffset(dst, offset)
	offset += len(c.Balances) * 28

	// Offset (1) 'Nonces'
	dst = ssz.WriteOffset(dst, offset)
	offset += len(c.Nonces) * 28

	// Field (0) 'Balances'
	if len(c.Balances) > 310995116277762 {
		err = ssz.ErrListTooBig
		return
	}
	for ii := 0; ii < len(c.Balances); ii++ {
		if dst, err = c.Balances[ii].MarshalSSZTo(dst); err != nil {
			return
		}
	}

	// Field (1) 'Nonces'
	if len(c.Nonces) > 310995116277762 {
		err = ssz.ErrListTooBig
		return
	}
	for ii := 0; ii < len(c.Nonces); ii++ {
		if dst, err = c.Nonces[ii].MarshalSSZTo(dst); err != nil {
			return
		}
	}

	return
}

// UnmarshalSSZ ssz unmarshals the CoinsStateSerializable object
func (c *CoinsStateSerializable) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size < 8 {
		return ssz.ErrSize
	}

	tail := buf
	var o0, o1 uint64

	// Offset (0) 'Balances'
	if o0 = ssz.ReadOffset(buf[0:4]); o0 > size {
		return ssz.ErrOffset
	}

	// Offset (1) 'Nonces'
	if o1 = ssz.ReadOffset(buf[4:8]); o1 > size || o0 > o1 {
		return ssz.ErrOffset
	}

	// Field (0) 'Balances'
	{
		buf = tail[o0:o1]
		num, err := ssz.DivideInt2(len(buf), 28, 310995116277762)
		if err != nil {
			return err
		}
		c.Balances = make([]*AccountInfo, num)
		for ii := 0; ii < num; ii++ {
			if c.Balances[ii] == nil {
				c.Balances[ii] = new(AccountInfo)
			}
			if err = c.Balances[ii].UnmarshalSSZ(buf[ii*28 : (ii+1)*28]); err != nil {
				return err
			}
		}
	}

	// Field (1) 'Nonces'
	{
		buf = tail[o1:]
		num, err := ssz.DivideInt2(len(buf), 28, 310995116277762)
		if err != nil {
			return err
		}
		c.Nonces = make([]*AccountInfo, num)
		for ii := 0; ii < num; ii++ {
			if c.Nonces[ii] == nil {
				c.Nonces[ii] = new(AccountInfo)
			}
			if err = c.Nonces[ii].UnmarshalSSZ(buf[ii*28 : (ii+1)*28]); err != nil {
				return err
			}
		}
	}
	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the CoinsStateSerializable object
func (c *CoinsStateSerializable) SizeSSZ() (size int) {
	size = 8

	// Field (0) 'Balances'
	size += len(c.Balances) * 28

	// Field (1) 'Nonces'
	size += len(c.Nonces) * 28

	return
}

// HashTreeRoot ssz hashes the CoinsStateSerializable object
func (c *CoinsStateSerializable) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(c)
}

// HashTreeRootWith ssz hashes the CoinsStateSerializable object with a hasher
func (c *CoinsStateSerializable) HashTreeRootWith(hh *ssz.Hasher) (err error) {
	indx := hh.Index()

	// Field (0) 'Balances'
	{
		subIndx := hh.Index()
		num := uint64(len(c.Balances))
		if num > 310995116277762 {
			err = ssz.ErrIncorrectListSize
			return
		}
		for i := uint64(0); i < num; i++ {
			if err = c.Balances[i].HashTreeRootWith(hh); err != nil {
				return
			}
		}
		hh.MerkleizeWithMixin(subIndx, num, 310995116277762)
	}

	// Field (1) 'Nonces'
	{
		subIndx := hh.Index()
		num := uint64(len(c.Nonces))
		if num > 310995116277762 {
			err = ssz.ErrIncorrectListSize
			return
		}
		for i := uint64(0); i < num; i++ {
			if err = c.Nonces[i].HashTreeRootWith(hh); err != nil {
				return
			}
		}
		hh.MerkleizeWithMixin(subIndx, num, 310995116277762)
	}

	hh.Merkleize(indx)
	return
}
