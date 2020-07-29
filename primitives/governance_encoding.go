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

	// Field (0) 'Hash'
	dst = append(dst, c.Hash[:]...)

	// Field (1) 'Data'
	if c.Data == nil {
		c.Data = new(CommunityVoteData)
	}
	if dst, err = c.Data.MarshalSSZTo(dst); err != nil {
		return
	}

	return
}

// UnmarshalSSZ ssz unmarshals the CommunityVoteDataInfo object
func (c *CommunityVoteDataInfo) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size != 132 {
		return ssz.ErrSize
	}

	// Field (0) 'Hash'
	copy(c.Hash[:], buf[0:32])

	// Field (1) 'Data'
	if c.Data == nil {
		c.Data = new(CommunityVoteData)
	}
	if err = c.Data.UnmarshalSSZ(buf[32:132]); err != nil {
		return err
	}

	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the CommunityVoteDataInfo object
func (c *CommunityVoteDataInfo) SizeSSZ() (size int) {
	size = 132
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

// MarshalSSZ ssz marshals the GovernanceSerializable object
func (g *GovernanceSerializable) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(g)
}

// MarshalSSZTo ssz marshals the GovernanceSerializable object to a target array
func (g *GovernanceSerializable) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf
	offset := int(8)

	// Offset (0) 'ReplaceVotes'
	dst = ssz.WriteOffset(dst, offset)
	offset += len(g.ReplaceVotes) * 52

	// Offset (1) 'CommunityVotes'
	dst = ssz.WriteOffset(dst, offset)
	offset += len(g.CommunityVotes) * 132

	// Field (0) 'ReplaceVotes'
	if len(g.ReplaceVotes) > 1099511627776 {
		err = ssz.ErrListTooBig
		return
	}
	for ii := 0; ii < len(g.ReplaceVotes); ii++ {
		if dst, err = g.ReplaceVotes[ii].MarshalSSZTo(dst); err != nil {
			return
		}
	}

	// Field (1) 'CommunityVotes'
	if len(g.CommunityVotes) > 1099511627776 {
		err = ssz.ErrListTooBig
		return
	}
	for ii := 0; ii < len(g.CommunityVotes); ii++ {
		if dst, err = g.CommunityVotes[ii].MarshalSSZTo(dst); err != nil {
			return
		}
	}

	return
}

// UnmarshalSSZ ssz unmarshals the GovernanceSerializable object
func (g *GovernanceSerializable) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size < 8 {
		return ssz.ErrSize
	}

	tail := buf
	var o0, o1 uint64

	// Offset (0) 'ReplaceVotes'
	if o0 = ssz.ReadOffset(buf[0:4]); o0 > size {
		return ssz.ErrOffset
	}

	// Offset (1) 'CommunityVotes'
	if o1 = ssz.ReadOffset(buf[4:8]); o1 > size || o0 > o1 {
		return ssz.ErrOffset
	}

	// Field (0) 'ReplaceVotes'
	{
		buf = tail[o0:o1]
		num, err := ssz.DivideInt2(len(buf), 52, 1099511627776)
		if err != nil {
			return err
		}
		g.ReplaceVotes = make([]*ReplacementVotes, num)
		for ii := 0; ii < num; ii++ {
			if g.ReplaceVotes[ii] == nil {
				g.ReplaceVotes[ii] = new(ReplacementVotes)
			}
			if err = g.ReplaceVotes[ii].UnmarshalSSZ(buf[ii*52 : (ii+1)*52]); err != nil {
				return err
			}
		}
	}

	// Field (1) 'CommunityVotes'
	{
		buf = tail[o1:]
		num, err := ssz.DivideInt2(len(buf), 132, 1099511627776)
		if err != nil {
			return err
		}
		g.CommunityVotes = make([]*CommunityVoteDataInfo, num)
		for ii := 0; ii < num; ii++ {
			if g.CommunityVotes[ii] == nil {
				g.CommunityVotes[ii] = new(CommunityVoteDataInfo)
			}
			if err = g.CommunityVotes[ii].UnmarshalSSZ(buf[ii*132 : (ii+1)*132]); err != nil {
				return err
			}
		}
	}
	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the GovernanceSerializable object
func (g *GovernanceSerializable) SizeSSZ() (size int) {
	size = 8

	// Field (0) 'ReplaceVotes'
	size += len(g.ReplaceVotes) * 52

	// Field (1) 'CommunityVotes'
	size += len(g.CommunityVotes) * 132

	return
}

// HashTreeRoot ssz hashes the GovernanceSerializable object
func (g *GovernanceSerializable) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(g)
}

// HashTreeRootWith ssz hashes the GovernanceSerializable object with a hasher
func (g *GovernanceSerializable) HashTreeRootWith(hh *ssz.Hasher) (err error) {
	indx := hh.Index()

	// Field (0) 'ReplaceVotes'
	{
		subIndx := hh.Index()
		num := uint64(len(g.ReplaceVotes))
		if num > 1099511627776 {
			err = ssz.ErrIncorrectListSize
			return
		}
		for i := uint64(0); i < num; i++ {
			if err = g.ReplaceVotes[i].HashTreeRootWith(hh); err != nil {
				return
			}
		}
		hh.MerkleizeWithMixin(subIndx, num, 1099511627776)
	}

	// Field (1) 'CommunityVotes'
	{
		subIndx := hh.Index()
		num := uint64(len(g.CommunityVotes))
		if num > 1099511627776 {
			err = ssz.ErrIncorrectListSize
			return
		}
		for i := uint64(0); i < num; i++ {
			if err = g.CommunityVotes[i].HashTreeRootWith(hh); err != nil {
				return
			}
		}
		hh.MerkleizeWithMixin(subIndx, num, 1099511627776)
	}

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

	// Field (0) 'ReplacementCandidates'
	for ii := 0; ii < 5; ii++ {
		dst = append(dst, c.ReplacementCandidates[ii][:]...)
	}

	return
}

// UnmarshalSSZ ssz unmarshals the CommunityVoteData object
func (c *CommunityVoteData) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size != 100 {
		return ssz.ErrSize
	}

	// Field (0) 'ReplacementCandidates'

	for ii := 0; ii < 5; ii++ {
		copy(c.ReplacementCandidates[ii][:], buf[0:100][ii*20:(ii+1)*20])
	}

	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the CommunityVoteData object
func (c *CommunityVoteData) SizeSSZ() (size int) {
	size = 100
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
		subIndx := hh.Index()
		for _, i := range c.ReplacementCandidates {
			hh.Append(i[:])
		}
		hh.Merkleize(subIndx)
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
	offset := int(24)

	// Field (0) 'Type'
	dst = ssz.MarshalUint64(dst, g.Type)

	// Offset (1) 'Data'
	dst = ssz.WriteOffset(dst, offset)
	offset += len(g.Data)

	// Offset (2) 'FunctionalSig'
	dst = ssz.WriteOffset(dst, offset)
	offset += len(g.FunctionalSig)

	// Field (3) 'VoteEpoch'
	dst = ssz.MarshalUint64(dst, g.VoteEpoch)

	// Field (1) 'Data'
	if len(g.Data) != 0 {
		err = ssz.ErrBytesLength
		return
	}
	dst = append(dst, g.Data...)

	// Field (2) 'FunctionalSig'
	if len(g.FunctionalSig) != 0 {
		err = ssz.ErrBytesLength
		return
	}
	dst = append(dst, g.FunctionalSig...)

	return
}

// UnmarshalSSZ ssz unmarshals the GovernanceVote object
func (g *GovernanceVote) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size < 24 {
		return ssz.ErrSize
	}

	tail := buf
	var o1, o2 uint64

	// Field (0) 'Type'
	g.Type = ssz.UnmarshallUint64(buf[0:8])

	// Offset (1) 'Data'
	if o1 = ssz.ReadOffset(buf[8:12]); o1 > size {
		return ssz.ErrOffset
	}

	// Offset (2) 'FunctionalSig'
	if o2 = ssz.ReadOffset(buf[12:16]); o2 > size || o1 > o2 {
		return ssz.ErrOffset
	}

	// Field (3) 'VoteEpoch'
	g.VoteEpoch = ssz.UnmarshallUint64(buf[16:24])

	// Field (1) 'Data'
	{
		buf = tail[o1:o2]
		g.Data = append(g.Data, buf...)
	}

	// Field (2) 'FunctionalSig'
	{
		buf = tail[o2:]
		g.FunctionalSig = append(g.FunctionalSig, buf...)
	}
	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the GovernanceVote object
func (g *GovernanceVote) SizeSSZ() (size int) {
	size = 24

	// Field (1) 'Data'
	size += len(g.Data)

	// Field (2) 'FunctionalSig'
	size += len(g.FunctionalSig)

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
	if len(g.Data) != 0 {
		err = ssz.ErrBytesLength
		return
	}
	hh.PutBytes(g.Data)

	// Field (2) 'FunctionalSig'
	if len(g.FunctionalSig) != 0 {
		err = ssz.ErrBytesLength
		return
	}
	hh.PutBytes(g.FunctionalSig)

	// Field (3) 'VoteEpoch'
	hh.PutUint64(g.VoteEpoch)

	hh.Merkleize(indx)
	return
}
