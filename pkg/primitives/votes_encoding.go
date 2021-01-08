// Code generated by fastssz. DO NOT EDIT.
package primitives

import (
	ssz "github.com/ferranbt/fastssz"
)

// MarshalSSZ ssz marshals the AcceptedVoteInfo object
func (a *AcceptedVoteInfo) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(a)
}

// MarshalSSZTo ssz marshals the AcceptedVoteInfo object to a target array
func (a *AcceptedVoteInfo) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf
	offset := int(148)

	// Field (0) 'Data'
	if a.Data == nil {
		a.Data = new(VoteData)
	}
	if dst, err = a.Data.MarshalSSZTo(dst); err != nil {
		return
	}

	// Offset (1) 'ParticipationBitfield'
	dst = ssz.WriteOffset(dst, offset)
	offset += len(a.ParticipationBitfield)

	// Field (2) 'Proposer'
	dst = ssz.MarshalUint64(dst, a.Proposer)

	// Field (3) 'InclusionDelay'
	dst = ssz.MarshalUint64(dst, a.InclusionDelay)

	// Field (1) 'ParticipationBitfield'
	if len(a.ParticipationBitfield) > 6250 {
		err = ssz.ErrBytesLength
		return
	}
	dst = append(dst, a.ParticipationBitfield...)

	return
}

// UnmarshalSSZ ssz unmarshals the AcceptedVoteInfo object
func (a *AcceptedVoteInfo) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size < 148 {
		return ssz.ErrSize
	}

	tail := buf
	var o1 uint64

	// Field (0) 'Data'
	if a.Data == nil {
		a.Data = new(VoteData)
	}
	if err = a.Data.UnmarshalSSZ(buf[0:128]); err != nil {
		return err
	}

	// Offset (1) 'ParticipationBitfield'
	if o1 = ssz.ReadOffset(buf[128:132]); o1 > size {
		return ssz.ErrOffset
	}

	// Field (2) 'Proposer'
	a.Proposer = ssz.UnmarshallUint64(buf[132:140])

	// Field (3) 'InclusionDelay'
	a.InclusionDelay = ssz.UnmarshallUint64(buf[140:148])

	// Field (1) 'ParticipationBitfield'
	{
		buf = tail[o1:]
		if err = ssz.ValidateBitlist(buf, 6250); err != nil {
			return err
		}
		if cap(a.ParticipationBitfield) == 0 {
			a.ParticipationBitfield = make([]byte, 0, len(buf))
		}
		a.ParticipationBitfield = append(a.ParticipationBitfield, buf...)
	}
	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the AcceptedVoteInfo object
func (a *AcceptedVoteInfo) SizeSSZ() (size int) {
	size = 148

	// Field (1) 'ParticipationBitfield'
	size += len(a.ParticipationBitfield)

	return
}

// HashTreeRoot ssz hashes the AcceptedVoteInfo object
func (a *AcceptedVoteInfo) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(a)
}

// HashTreeRootWith ssz hashes the AcceptedVoteInfo object with a hasher
func (a *AcceptedVoteInfo) HashTreeRootWith(hh *ssz.Hasher) (err error) {
	indx := hh.Index()

	// Field (0) 'Data'
	if err = a.Data.HashTreeRootWith(hh); err != nil {
		return
	}

	// Field (1) 'ParticipationBitfield'
	if len(a.ParticipationBitfield) == 0 {
		err = ssz.ErrEmptyBitlist
		return
	}
	hh.PutBitlist(a.ParticipationBitfield, 6250)

	// Field (2) 'Proposer'
	hh.PutUint64(a.Proposer)

	// Field (3) 'InclusionDelay'
	hh.PutUint64(a.InclusionDelay)

	hh.Merkleize(indx)
	return
}

// MarshalSSZ ssz marshals the VoteData object
func (v *VoteData) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(v)
}

// MarshalSSZTo ssz marshals the VoteData object to a target array
func (v *VoteData) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf

	// Field (0) 'Slot'
	dst = ssz.MarshalUint64(dst, v.Slot)

	// Field (1) 'FromEpoch'
	dst = ssz.MarshalUint64(dst, v.FromEpoch)

	// Field (2) 'FromHash'
	dst = append(dst, v.FromHash[:]...)

	// Field (3) 'ToEpoch'
	dst = ssz.MarshalUint64(dst, v.ToEpoch)

	// Field (4) 'ToHash'
	dst = append(dst, v.ToHash[:]...)

	// Field (5) 'BeaconBlockHash'
	dst = append(dst, v.BeaconBlockHash[:]...)

	// Field (6) 'Nonce'
	dst = ssz.MarshalUint64(dst, v.Nonce)

	return
}

// UnmarshalSSZ ssz unmarshals the VoteData object
func (v *VoteData) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size != 128 {
		return ssz.ErrSize
	}

	// Field (0) 'Slot'
	v.Slot = ssz.UnmarshallUint64(buf[0:8])

	// Field (1) 'FromEpoch'
	v.FromEpoch = ssz.UnmarshallUint64(buf[8:16])

	// Field (2) 'FromHash'
	copy(v.FromHash[:], buf[16:48])

	// Field (3) 'ToEpoch'
	v.ToEpoch = ssz.UnmarshallUint64(buf[48:56])

	// Field (4) 'ToHash'
	copy(v.ToHash[:], buf[56:88])

	// Field (5) 'BeaconBlockHash'
	copy(v.BeaconBlockHash[:], buf[88:120])

	// Field (6) 'Nonce'
	v.Nonce = ssz.UnmarshallUint64(buf[120:128])

	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the VoteData object
func (v *VoteData) SizeSSZ() (size int) {
	size = 128
	return
}

// HashTreeRoot ssz hashes the VoteData object
func (v *VoteData) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(v)
}

// HashTreeRootWith ssz hashes the VoteData object with a hasher
func (v *VoteData) HashTreeRootWith(hh *ssz.Hasher) (err error) {
	indx := hh.Index()

	// Field (0) 'Slot'
	hh.PutUint64(v.Slot)

	// Field (1) 'FromEpoch'
	hh.PutUint64(v.FromEpoch)

	// Field (2) 'FromHash'
	hh.PutBytes(v.FromHash[:])

	// Field (3) 'ToEpoch'
	hh.PutUint64(v.ToEpoch)

	// Field (4) 'ToHash'
	hh.PutBytes(v.ToHash[:])

	// Field (5) 'BeaconBlockHash'
	hh.PutBytes(v.BeaconBlockHash[:])

	// Field (6) 'Nonce'
	hh.PutUint64(v.Nonce)

	hh.Merkleize(indx)
	return
}

// MarshalSSZ ssz marshals the MultiValidatorVote object
func (m *MultiValidatorVote) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(m)
}

// MarshalSSZTo ssz marshals the MultiValidatorVote object to a target array
func (m *MultiValidatorVote) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf
	offset := int(228)

	// Field (0) 'Data'
	if m.Data == nil {
		m.Data = new(VoteData)
	}
	if dst, err = m.Data.MarshalSSZTo(dst); err != nil {
		return
	}

	// Field (1) 'Sig'
	dst = append(dst, m.Sig[:]...)

	// Offset (2) 'ParticipationBitfield'
	dst = ssz.WriteOffset(dst, offset)
	offset += len(m.ParticipationBitfield)

	// Field (2) 'ParticipationBitfield'
	if len(m.ParticipationBitfield) > 50000 {
		err = ssz.ErrBytesLength
		return
	}
	dst = append(dst, m.ParticipationBitfield...)

	return
}

// UnmarshalSSZ ssz unmarshals the MultiValidatorVote object
func (m *MultiValidatorVote) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size < 228 {
		return ssz.ErrSize
	}

	tail := buf
	var o2 uint64

	// Field (0) 'Data'
	if m.Data == nil {
		m.Data = new(VoteData)
	}
	if err = m.Data.UnmarshalSSZ(buf[0:128]); err != nil {
		return err
	}

	// Field (1) 'Sig'
	copy(m.Sig[:], buf[128:224])

	// Offset (2) 'ParticipationBitfield'
	if o2 = ssz.ReadOffset(buf[224:228]); o2 > size {
		return ssz.ErrOffset
	}

	// Field (2) 'ParticipationBitfield'
	{
		buf = tail[o2:]
		if err = ssz.ValidateBitlist(buf, 50000); err != nil {
			return err
		}
		if cap(m.ParticipationBitfield) == 0 {
			m.ParticipationBitfield = make([]byte, 0, len(buf))
		}
		m.ParticipationBitfield = append(m.ParticipationBitfield, buf...)
	}
	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the MultiValidatorVote object
func (m *MultiValidatorVote) SizeSSZ() (size int) {
	size = 228

	// Field (2) 'ParticipationBitfield'
	size += len(m.ParticipationBitfield)

	return
}

// HashTreeRoot ssz hashes the MultiValidatorVote object
func (m *MultiValidatorVote) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(m)
}

// HashTreeRootWith ssz hashes the MultiValidatorVote object with a hasher
func (m *MultiValidatorVote) HashTreeRootWith(hh *ssz.Hasher) (err error) {
	indx := hh.Index()

	// Field (0) 'Data'
	if err = m.Data.HashTreeRootWith(hh); err != nil {
		return
	}

	// Field (1) 'Sig'
	hh.PutBytes(m.Sig[:])

	// Field (2) 'ParticipationBitfield'
	if len(m.ParticipationBitfield) == 0 {
		err = ssz.ErrEmptyBitlist
		return
	}
	hh.PutBitlist(m.ParticipationBitfield, 50000)

	hh.Merkleize(indx)
	return
}
