// Code generated by fastssz. DO NOT EDIT.
package primitives

import (
	ssz "github.com/ferranbt/fastssz"
)

// MarshalSSZ ssz marshals the CommunityVoteDataInfo object
func (c *CommunityVoteDataInfo) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(c)
}

// MarshalSSZTo ssz marshals the CommunityVoteDataInfo object to a target array
func (c *CommunityVoteDataInfo) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf
	offset := int(36)

	// Field (0) 'Hash'
	dst = append(dst, c.Hash[:]...)

	// Offset (1) 'Data'
	dst = ssz.WriteOffset(dst, offset)
	if c.Data == nil {
		c.Data = new(CommunityVoteData)
	}
	offset += c.Data.SizeSSZ()

	// Field (1) 'Data'
	if dst, err = c.Data.MarshalSSZTo(dst); err != nil {
		return
	}

	return
}

// UnmarshalSSZ ssz unmarshals the CommunityVoteDataInfo object
func (c *CommunityVoteDataInfo) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size < 36 {
		return ssz.ErrSize
	}

	tail := buf
	var o1 uint64

	// Field (0) 'Hash'
	copy(c.Hash[:], buf[0:32])

	// Offset (1) 'Data'
	if o1 = ssz.ReadOffset(buf[32:36]); o1 > size {
		return ssz.ErrOffset
	}

	// Field (1) 'Data'
	{
		buf = tail[o1:]
		if c.Data == nil {
			c.Data = new(CommunityVoteData)
		}
		if err = c.Data.UnmarshalSSZ(buf); err != nil {
			return err
		}
	}
	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the CommunityVoteDataInfo object
func (c *CommunityVoteDataInfo) SizeSSZ() (size int) {
	size = 36

	// Field (1) 'Data'
	if c.Data == nil {
		c.Data = new(CommunityVoteData)
	}
	size += c.Data.SizeSSZ()

	return
}

// HashTreeRoot ssz hashes the CommunityVoteDataInfo object
func (c *CommunityVoteDataInfo) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(c)
}

// HashTreeRootWith ssz hashes the CommunityVoteDataInfo object with a hasher
func (c *CommunityVoteDataInfo) HashTreeRootWith(hh *ssz.Hasher) (err error) {
	indx := hh.Index()

	// Field (0) 'Hash'
	hh.PutBytes(c.Hash[:])

	// Field (1) 'Data'
	if err = c.Data.HashTreeRootWith(hh); err != nil {
		return
	}

	hh.Merkleize(indx)
	return
}

// MarshalSSZ ssz marshals the ReplacementVotes object
func (r *ReplacementVotes) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(r)
}

// MarshalSSZTo ssz marshals the ReplacementVotes object to a target array
func (r *ReplacementVotes) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf

	// Field (0) 'Account'
	dst = append(dst, r.Account[:]...)

	// Field (1) 'Hash'
	dst = append(dst, r.Hash[:]...)

	return
}

// UnmarshalSSZ ssz unmarshals the ReplacementVotes object
func (r *ReplacementVotes) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size != 52 {
		return ssz.ErrSize
	}

	// Field (0) 'Account'
	copy(r.Account[:], buf[0:20])

	// Field (1) 'Hash'
	copy(r.Hash[:], buf[20:52])

	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the ReplacementVotes object
func (r *ReplacementVotes) SizeSSZ() (size int) {
	size = 52
	return
}

// HashTreeRoot ssz hashes the ReplacementVotes object
func (r *ReplacementVotes) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(r)
}

// HashTreeRootWith ssz hashes the ReplacementVotes object with a hasher
func (r *ReplacementVotes) HashTreeRootWith(hh *ssz.Hasher) (err error) {
	indx := hh.Index()

	// Field (0) 'Account'
	hh.PutBytes(r.Account[:])

	// Field (1) 'Hash'
	hh.PutBytes(r.Hash[:])

	hh.Merkleize(indx)
	return
}

// MarshalSSZ ssz marshals the CommunityVoteData object
func (c *CommunityVoteData) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(c)
}

// MarshalSSZTo ssz marshals the CommunityVoteData object to a target array
func (c *CommunityVoteData) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf
	offset := int(4)

	// Offset (0) 'ReplacementCandidates'
	dst = ssz.WriteOffset(dst, offset)
	offset += len(c.ReplacementCandidates) * 20

	// Field (0) 'ReplacementCandidates'
	if len(c.ReplacementCandidates) > 5 {
		err = ssz.ErrListTooBig
		return
	}
	for ii := 0; ii < len(c.ReplacementCandidates); ii++ {
		dst = append(dst, c.ReplacementCandidates[ii][:]...)
	}

	return
}

// UnmarshalSSZ ssz unmarshals the CommunityVoteData object
func (c *CommunityVoteData) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size < 4 {
		return ssz.ErrSize
	}

	tail := buf
	var o0 uint64

	// Offset (0) 'ReplacementCandidates'
	if o0 = ssz.ReadOffset(buf[0:4]); o0 > size {
		return ssz.ErrOffset
	}

	// Field (0) 'ReplacementCandidates'
	{
		buf = tail[o0:]
		num, err := ssz.DivideInt2(len(buf), 20, 5)
		if err != nil {
			return err
		}
		c.ReplacementCandidates = make([][20]byte, num)
		for ii := 0; ii < num; ii++ {
			copy(c.ReplacementCandidates[ii][:], buf[ii*20:(ii+1)*20])
		}
	}
	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the CommunityVoteData object
func (c *CommunityVoteData) SizeSSZ() (size int) {
	size = 4

	// Field (0) 'ReplacementCandidates'
	size += len(c.ReplacementCandidates) * 20

	return
}

// HashTreeRoot ssz hashes the CommunityVoteData object
func (c *CommunityVoteData) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(c)
}

// HashTreeRootWith ssz hashes the CommunityVoteData object with a hasher
func (c *CommunityVoteData) HashTreeRootWith(hh *ssz.Hasher) (err error) {
	indx := hh.Index()

	// Field (0) 'ReplacementCandidates'
	{
		if len(c.ReplacementCandidates) > 5 {
			err = ssz.ErrListTooBig
			return
		}
		subIndx := hh.Index()
		for _, i := range c.ReplacementCandidates {
			hh.Append(i[:])
		}
		numItems := uint64(len(c.ReplacementCandidates))
		hh.MerkleizeWithMixin(subIndx, numItems, ssz.CalculateLimit(5, numItems, 32))
	}

	hh.Merkleize(indx)
	return
}

// MarshalSSZ ssz marshals the GovernanceVote object
func (g *GovernanceVote) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(g)
}

// MarshalSSZTo ssz marshals the GovernanceVote object to a target array
func (g *GovernanceVote) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf
	offset := int(164)

	// Field (0) 'Type'
	dst = ssz.MarshalUint64(dst, g.Type)

	// Offset (1) 'Data'
	dst = ssz.WriteOffset(dst, offset)
	offset += len(g.Data)

	// Field (2) 'VoteEpoch'
	dst = ssz.MarshalUint64(dst, g.VoteEpoch)

	// Field (3) 'PublicKey'
	dst = append(dst, g.PublicKey[:]...)

	// Field (4) 'Signature'
	dst = append(dst, g.Signature[:]...)

	// Field (1) 'Data'
	if len(g.Data) > 100 {
		err = ssz.ErrBytesLength
		return
	}
	dst = append(dst, g.Data...)

	return
}

// UnmarshalSSZ ssz unmarshals the GovernanceVote object
func (g *GovernanceVote) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size < 164 {
		return ssz.ErrSize
	}

	tail := buf
	var o1 uint64

	// Field (0) 'Type'
	g.Type = ssz.UnmarshallUint64(buf[0:8])

	// Offset (1) 'Data'
	if o1 = ssz.ReadOffset(buf[8:12]); o1 > size {
		return ssz.ErrOffset
	}

	// Field (2) 'VoteEpoch'
	g.VoteEpoch = ssz.UnmarshallUint64(buf[12:20])

	// Field (3) 'PublicKey'
	copy(g.PublicKey[:], buf[20:68])

	// Field (4) 'Signature'
	copy(g.Signature[:], buf[68:164])

	// Field (1) 'Data'
	{
		buf = tail[o1:]
		if len(buf) > 100 {
			return ssz.ErrBytesLength
		}
		if cap(g.Data) == 0 {
			g.Data = make([]byte, 0, len(buf))
		}
		g.Data = append(g.Data, buf...)
	}
	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the GovernanceVote object
func (g *GovernanceVote) SizeSSZ() (size int) {
	size = 164

	// Field (1) 'Data'
	size += len(g.Data)

	return
}

// HashTreeRoot ssz hashes the GovernanceVote object
func (g *GovernanceVote) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(g)
}

// HashTreeRootWith ssz hashes the GovernanceVote object with a hasher
func (g *GovernanceVote) HashTreeRootWith(hh *ssz.Hasher) (err error) {
	indx := hh.Index()

	// Field (0) 'Type'
	hh.PutUint64(g.Type)

	// Field (1) 'Data'
	if len(g.Data) > 100 {
		err = ssz.ErrBytesLength
		return
	}
	hh.PutBytes(g.Data)

	// Field (2) 'VoteEpoch'
	hh.PutUint64(g.VoteEpoch)

	// Field (3) 'PublicKey'
	hh.PutBytes(g.PublicKey[:])

	// Field (4) 'Signature'
	hh.PutBytes(g.Signature[:])

	hh.Merkleize(indx)
	return
}
