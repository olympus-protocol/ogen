// Code generated by fastssz. DO NOT EDIT.
package primitives

import (
	ssz "github.com/ferranbt/fastssz"
)

// MarshalSSZ ssz marshals the SerializableState object
func (s *SerializableState) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(s)
}

// MarshalSSZTo ssz marshals the SerializableState object to a target array
func (s *SerializableState) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf
	offset := int(264)

	// Offset (0) 'CoinsState'
	dst = ssz.WriteOffset(dst, offset)
	if s.CoinsState == nil {
		s.CoinsState = new(CoinsStateSerializable)
	}
	offset += s.CoinsState.SizeSSZ()

	// Offset (1) 'ValidatorRegistry'
	dst = ssz.WriteOffset(dst, offset)
	offset += len(s.ValidatorRegistry) * 100

	// Field (2) 'LatestValidatorRegistryChange'
	dst = ssz.MarshalUint64(dst, s.LatestValidatorRegistryChange)

	// Field (3) 'RANDAO'
	dst = append(dst, s.RANDAO[:]...)

	// Field (4) 'NextRANDAO'
	dst = append(dst, s.NextRANDAO[:]...)

	// Field (5) 'Slot'
	dst = ssz.MarshalUint64(dst, s.Slot)

	// Field (6) 'EpochIndex'
	dst = ssz.MarshalUint64(dst, s.EpochIndex)

	// Offset (7) 'ProposerQueue'
	dst = ssz.WriteOffset(dst, offset)
	offset += len(s.ProposerQueue) * 8

	// Offset (8) 'PreviousEpochVoteAssignments'
	dst = ssz.WriteOffset(dst, offset)
	offset += len(s.PreviousEpochVoteAssignments) * 8

	// Offset (9) 'CurrentEpochVoteAssignments'
	dst = ssz.WriteOffset(dst, offset)
	offset += len(s.CurrentEpochVoteAssignments) * 8

	// Offset (10) 'NextProposerQueue'
	dst = ssz.WriteOffset(dst, offset)
	offset += len(s.NextProposerQueue) * 8

	// Field (11) 'JustificationBitfield'
	dst = ssz.MarshalUint64(dst, s.JustificationBitfield)

	// Field (12) 'FinalizedEpoch'
	dst = ssz.MarshalUint64(dst, s.FinalizedEpoch)

	// Offset (13) 'LatestBlockHashes'
	dst = ssz.WriteOffset(dst, offset)
	offset += len(s.LatestBlockHashes) * 32

	// Field (14) 'JustifiedEpoch'
	dst = ssz.MarshalUint64(dst, s.JustifiedEpoch)

	// Field (15) 'JustifiedEpochHash'
	dst = append(dst, s.JustifiedEpochHash[:]...)

	// Offset (16) 'CurrentEpochVotes'
	dst = ssz.WriteOffset(dst, offset)
	for ii := 0; ii < len(s.CurrentEpochVotes); ii++ {
		offset += 4
		offset += s.CurrentEpochVotes[ii].SizeSSZ()
	}

	// Field (17) 'PreviousJustifiedEpoch'
	dst = ssz.MarshalUint64(dst, s.PreviousJustifiedEpoch)

	// Field (18) 'PreviousJustifiedEpochHash'
	dst = append(dst, s.PreviousJustifiedEpochHash[:]...)

	// Offset (19) 'PreviousEpochVotes'
	dst = ssz.WriteOffset(dst, offset)
	for ii := 0; ii < len(s.PreviousEpochVotes); ii++ {
		offset += 4
		offset += s.PreviousEpochVotes[ii].SizeSSZ()
	}

	// Offset (20) 'CurrentManagers'
	dst = ssz.WriteOffset(dst, offset)
	offset += len(s.CurrentManagers) * 20

	// Offset (21) 'ManagerReplacement'
	dst = ssz.WriteOffset(dst, offset)
	offset += len(s.ManagerReplacement) * 1

	// Offset (22) 'Governance'
	dst = ssz.WriteOffset(dst, offset)
	if s.Governance == nil {
		s.Governance = new(GovernanceSerializable)
	}
	offset += s.Governance.SizeSSZ()

	// Field (23) 'VoteEpoch'
	dst = ssz.MarshalUint64(dst, s.VoteEpoch)

	// Field (24) 'VoteEpochStartSlot'
	dst = ssz.MarshalUint64(dst, s.VoteEpochStartSlot)

	// Field (25) 'VotingState'
	dst = ssz.MarshalUint64(dst, s.VotingState)

	// Field (26) 'LastPaidSlot'
	dst = ssz.MarshalUint64(dst, s.LastPaidSlot)

	// Field (0) 'CoinsState'
	if dst, err = s.CoinsState.MarshalSSZTo(dst); err != nil {
		return
	}

	// Field (1) 'ValidatorRegistry'
	if len(s.ValidatorRegistry) > 1099511627776 {
		err = ssz.ErrListTooBig
		return
	}
	for ii := 0; ii < len(s.ValidatorRegistry); ii++ {
		if dst, err = s.ValidatorRegistry[ii].MarshalSSZTo(dst); err != nil {
			return
		}
	}

	// Field (7) 'ProposerQueue'
	if len(s.ProposerQueue) > 1099511627776 {
		err = ssz.ErrListTooBig
		return
	}
	for ii := 0; ii < len(s.ProposerQueue); ii++ {
		dst = ssz.MarshalUint64(dst, s.ProposerQueue[ii])
	}

	// Field (8) 'PreviousEpochVoteAssignments'
	if len(s.PreviousEpochVoteAssignments) > 1099511627776 {
		err = ssz.ErrListTooBig
		return
	}
	for ii := 0; ii < len(s.PreviousEpochVoteAssignments); ii++ {
		dst = ssz.MarshalUint64(dst, s.PreviousEpochVoteAssignments[ii])
	}

	// Field (9) 'CurrentEpochVoteAssignments'
	if len(s.CurrentEpochVoteAssignments) > 1099511627776 {
		err = ssz.ErrListTooBig
		return
	}
	for ii := 0; ii < len(s.CurrentEpochVoteAssignments); ii++ {
		dst = ssz.MarshalUint64(dst, s.CurrentEpochVoteAssignments[ii])
	}

	// Field (10) 'NextProposerQueue'
	if len(s.NextProposerQueue) > 1099511627776 {
		err = ssz.ErrListTooBig
		return
	}
	for ii := 0; ii < len(s.NextProposerQueue); ii++ {
		dst = ssz.MarshalUint64(dst, s.NextProposerQueue[ii])
	}

	// Field (13) 'LatestBlockHashes'
	if len(s.LatestBlockHashes) > 64 {
		err = ssz.ErrListTooBig
		return
	}
	for ii := 0; ii < len(s.LatestBlockHashes); ii++ {
		dst = append(dst, s.LatestBlockHashes[ii][:]...)
	}

	// Field (16) 'CurrentEpochVotes'
	if len(s.CurrentEpochVotes) > 1099511627776 {
		err = ssz.ErrListTooBig
		return
	}
	{
		offset = 4 * len(s.CurrentEpochVotes)
		for ii := 0; ii < len(s.CurrentEpochVotes); ii++ {
			dst = ssz.WriteOffset(dst, offset)
			offset += s.CurrentEpochVotes[ii].SizeSSZ()
		}
	}
	for ii := 0; ii < len(s.CurrentEpochVotes); ii++ {
		if dst, err = s.CurrentEpochVotes[ii].MarshalSSZTo(dst); err != nil {
			return
		}
	}

	// Field (19) 'PreviousEpochVotes'
	if len(s.PreviousEpochVotes) > 1099511627776 {
		err = ssz.ErrListTooBig
		return
	}
	{
		offset = 4 * len(s.PreviousEpochVotes)
		for ii := 0; ii < len(s.PreviousEpochVotes); ii++ {
			dst = ssz.WriteOffset(dst, offset)
			offset += s.PreviousEpochVotes[ii].SizeSSZ()
		}
	}
	for ii := 0; ii < len(s.PreviousEpochVotes); ii++ {
		if dst, err = s.PreviousEpochVotes[ii].MarshalSSZTo(dst); err != nil {
			return
		}
	}

	// Field (20) 'CurrentManagers'
	if len(s.CurrentManagers) > 5 {
		err = ssz.ErrListTooBig
		return
	}
	for ii := 0; ii < len(s.CurrentManagers); ii++ {
		dst = append(dst, s.CurrentManagers[ii][:]...)
	}

	// Field (21) 'ManagerReplacement'
	if len(s.ManagerReplacement) > 2048 {
		err = ssz.ErrListTooBig
		return
	}
	for ii := 0; ii < len(s.ManagerReplacement); ii++ {
		dst = ssz.MarshalUint8(dst, s.ManagerReplacement[ii])
	}

	// Field (22) 'Governance'
	if dst, err = s.Governance.MarshalSSZTo(dst); err != nil {
		return
	}

	return
}

// UnmarshalSSZ ssz unmarshals the SerializableState object
func (s *SerializableState) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size < 264 {
		return ssz.ErrSize
	}

	tail := buf
	var o0, o1, o7, o8, o9, o10, o13, o16, o19, o20, o21, o22 uint64

	// Offset (0) 'CoinsState'
	if o0 = ssz.ReadOffset(buf[0:4]); o0 > size {
		return ssz.ErrOffset
	}

	// Offset (1) 'ValidatorRegistry'
	if o1 = ssz.ReadOffset(buf[4:8]); o1 > size || o0 > o1 {
		return ssz.ErrOffset
	}

	// Field (2) 'LatestValidatorRegistryChange'
	s.LatestValidatorRegistryChange = ssz.UnmarshallUint64(buf[8:16])

	// Field (3) 'RANDAO'
	copy(s.RANDAO[:], buf[16:48])

	// Field (4) 'NextRANDAO'
	copy(s.NextRANDAO[:], buf[48:80])

	// Field (5) 'Slot'
	s.Slot = ssz.UnmarshallUint64(buf[80:88])

	// Field (6) 'EpochIndex'
	s.EpochIndex = ssz.UnmarshallUint64(buf[88:96])

	// Offset (7) 'ProposerQueue'
	if o7 = ssz.ReadOffset(buf[96:100]); o7 > size || o1 > o7 {
		return ssz.ErrOffset
	}

	// Offset (8) 'PreviousEpochVoteAssignments'
	if o8 = ssz.ReadOffset(buf[100:104]); o8 > size || o7 > o8 {
		return ssz.ErrOffset
	}

	// Offset (9) 'CurrentEpochVoteAssignments'
	if o9 = ssz.ReadOffset(buf[104:108]); o9 > size || o8 > o9 {
		return ssz.ErrOffset
	}

	// Offset (10) 'NextProposerQueue'
	if o10 = ssz.ReadOffset(buf[108:112]); o10 > size || o9 > o10 {
		return ssz.ErrOffset
	}

	// Field (11) 'JustificationBitfield'
	s.JustificationBitfield = ssz.UnmarshallUint64(buf[112:120])

	// Field (12) 'FinalizedEpoch'
	s.FinalizedEpoch = ssz.UnmarshallUint64(buf[120:128])

	// Offset (13) 'LatestBlockHashes'
	if o13 = ssz.ReadOffset(buf[128:132]); o13 > size || o10 > o13 {
		return ssz.ErrOffset
	}

	// Field (14) 'JustifiedEpoch'
	s.JustifiedEpoch = ssz.UnmarshallUint64(buf[132:140])

	// Field (15) 'JustifiedEpochHash'
	copy(s.JustifiedEpochHash[:], buf[140:172])

	// Offset (16) 'CurrentEpochVotes'
	if o16 = ssz.ReadOffset(buf[172:176]); o16 > size || o13 > o16 {
		return ssz.ErrOffset
	}

	// Field (17) 'PreviousJustifiedEpoch'
	s.PreviousJustifiedEpoch = ssz.UnmarshallUint64(buf[176:184])

	// Field (18) 'PreviousJustifiedEpochHash'
	copy(s.PreviousJustifiedEpochHash[:], buf[184:216])

	// Offset (19) 'PreviousEpochVotes'
	if o19 = ssz.ReadOffset(buf[216:220]); o19 > size || o16 > o19 {
		return ssz.ErrOffset
	}

	// Offset (20) 'CurrentManagers'
	if o20 = ssz.ReadOffset(buf[220:224]); o20 > size || o19 > o20 {
		return ssz.ErrOffset
	}

	// Offset (21) 'ManagerReplacement'
	if o21 = ssz.ReadOffset(buf[224:228]); o21 > size || o20 > o21 {
		return ssz.ErrOffset
	}

	// Offset (22) 'Governance'
	if o22 = ssz.ReadOffset(buf[228:232]); o22 > size || o21 > o22 {
		return ssz.ErrOffset
	}

	// Field (23) 'VoteEpoch'
	s.VoteEpoch = ssz.UnmarshallUint64(buf[232:240])

	// Field (24) 'VoteEpochStartSlot'
	s.VoteEpochStartSlot = ssz.UnmarshallUint64(buf[240:248])

	// Field (25) 'VotingState'
	s.VotingState = ssz.UnmarshallUint64(buf[248:256])

	// Field (26) 'LastPaidSlot'
	s.LastPaidSlot = ssz.UnmarshallUint64(buf[256:264])

	// Field (0) 'CoinsState'
	{
		buf = tail[o0:o1]
		if s.CoinsState == nil {
			s.CoinsState = new(CoinsStateSerializable)
		}
		if err = s.CoinsState.UnmarshalSSZ(buf); err != nil {
			return err
		}
	}

	// Field (1) 'ValidatorRegistry'
	{
		buf = tail[o1:o7]
		num, err := ssz.DivideInt2(len(buf), 100, 1099511627776)
		if err != nil {
			return err
		}
		s.ValidatorRegistry = make([]*Validator, num)
		for ii := 0; ii < num; ii++ {
			if s.ValidatorRegistry[ii] == nil {
				s.ValidatorRegistry[ii] = new(Validator)
			}
			if err = s.ValidatorRegistry[ii].UnmarshalSSZ(buf[ii*100 : (ii+1)*100]); err != nil {
				return err
			}
		}
	}

	// Field (7) 'ProposerQueue'
	{
		buf = tail[o7:o8]
		num, err := ssz.DivideInt2(len(buf), 8, 1099511627776)
		if err != nil {
			return err
		}
		s.ProposerQueue = ssz.ExtendUint64(s.ProposerQueue, num)
		for ii := 0; ii < num; ii++ {
			s.ProposerQueue[ii] = ssz.UnmarshallUint64(buf[ii*8 : (ii+1)*8])
		}
	}

	// Field (8) 'PreviousEpochVoteAssignments'
	{
		buf = tail[o8:o9]
		num, err := ssz.DivideInt2(len(buf), 8, 1099511627776)
		if err != nil {
			return err
		}
		s.PreviousEpochVoteAssignments = ssz.ExtendUint64(s.PreviousEpochVoteAssignments, num)
		for ii := 0; ii < num; ii++ {
			s.PreviousEpochVoteAssignments[ii] = ssz.UnmarshallUint64(buf[ii*8 : (ii+1)*8])
		}
	}

	// Field (9) 'CurrentEpochVoteAssignments'
	{
		buf = tail[o9:o10]
		num, err := ssz.DivideInt2(len(buf), 8, 1099511627776)
		if err != nil {
			return err
		}
		s.CurrentEpochVoteAssignments = ssz.ExtendUint64(s.CurrentEpochVoteAssignments, num)
		for ii := 0; ii < num; ii++ {
			s.CurrentEpochVoteAssignments[ii] = ssz.UnmarshallUint64(buf[ii*8 : (ii+1)*8])
		}
	}

	// Field (10) 'NextProposerQueue'
	{
		buf = tail[o10:o13]
		num, err := ssz.DivideInt2(len(buf), 8, 1099511627776)
		if err != nil {
			return err
		}
		s.NextProposerQueue = ssz.ExtendUint64(s.NextProposerQueue, num)
		for ii := 0; ii < num; ii++ {
			s.NextProposerQueue[ii] = ssz.UnmarshallUint64(buf[ii*8 : (ii+1)*8])
		}
	}

	// Field (13) 'LatestBlockHashes'
	{
		buf = tail[o13:o16]
		num, err := ssz.DivideInt2(len(buf), 32, 64)
		if err != nil {
			return err
		}
		s.LatestBlockHashes = make([][32]byte, num)
		for ii := 0; ii < num; ii++ {
			copy(s.LatestBlockHashes[ii][:], buf[ii*32:(ii+1)*32])
		}
	}

	// Field (16) 'CurrentEpochVotes'
	{
		buf = tail[o16:o19]
		num, err := ssz.DecodeDynamicLength(buf, 1099511627776)
		if err != nil {
			return err
		}
		s.CurrentEpochVotes = make([]*AcceptedVoteInfo, num)
		err = ssz.UnmarshalDynamic(buf, num, func(indx int, buf []byte) (err error) {
			if s.CurrentEpochVotes[indx] == nil {
				s.CurrentEpochVotes[indx] = new(AcceptedVoteInfo)
			}
			if err = s.CurrentEpochVotes[indx].UnmarshalSSZ(buf); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	// Field (19) 'PreviousEpochVotes'
	{
		buf = tail[o19:o20]
		num, err := ssz.DecodeDynamicLength(buf, 1099511627776)
		if err != nil {
			return err
		}
		s.PreviousEpochVotes = make([]*AcceptedVoteInfo, num)
		err = ssz.UnmarshalDynamic(buf, num, func(indx int, buf []byte) (err error) {
			if s.PreviousEpochVotes[indx] == nil {
				s.PreviousEpochVotes[indx] = new(AcceptedVoteInfo)
			}
			if err = s.PreviousEpochVotes[indx].UnmarshalSSZ(buf); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	// Field (20) 'CurrentManagers'
	{
		buf = tail[o20:o21]
		num, err := ssz.DivideInt2(len(buf), 20, 5)
		if err != nil {
			return err
		}
		s.CurrentManagers = make([][20]byte, num)
		for ii := 0; ii < num; ii++ {
			copy(s.CurrentManagers[ii][:], buf[ii*20:(ii+1)*20])
		}
	}

	// Field (21) 'ManagerReplacement'
	{
		buf = tail[o21:o22]
		num, err := ssz.DivideInt2(len(buf), 1, 2048)
		if err != nil {
			return err
		}
		s.ManagerReplacement = ssz.ExtendUint8(s.ManagerReplacement, num)
		for ii := 0; ii < num; ii++ {
			s.ManagerReplacement[ii] = ssz.UnmarshallUint8(buf[ii*1 : (ii+1)*1])
		}
	}

	// Field (22) 'Governance'
	{
		buf = tail[o22:]
		if s.Governance == nil {
			s.Governance = new(GovernanceSerializable)
		}
		if err = s.Governance.UnmarshalSSZ(buf); err != nil {
			return err
		}
	}
	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the SerializableState object
func (s *SerializableState) SizeSSZ() (size int) {
	size = 264

	// Field (0) 'CoinsState'
	if s.CoinsState == nil {
		s.CoinsState = new(CoinsStateSerializable)
	}
	size += s.CoinsState.SizeSSZ()

	// Field (1) 'ValidatorRegistry'
	size += len(s.ValidatorRegistry) * 100

	// Field (7) 'ProposerQueue'
	size += len(s.ProposerQueue) * 8

	// Field (8) 'PreviousEpochVoteAssignments'
	size += len(s.PreviousEpochVoteAssignments) * 8

	// Field (9) 'CurrentEpochVoteAssignments'
	size += len(s.CurrentEpochVoteAssignments) * 8

	// Field (10) 'NextProposerQueue'
	size += len(s.NextProposerQueue) * 8

	// Field (13) 'LatestBlockHashes'
	size += len(s.LatestBlockHashes) * 32

	// Field (16) 'CurrentEpochVotes'
	for ii := 0; ii < len(s.CurrentEpochVotes); ii++ {
		size += 4
		size += s.CurrentEpochVotes[ii].SizeSSZ()
	}

	// Field (19) 'PreviousEpochVotes'
	for ii := 0; ii < len(s.PreviousEpochVotes); ii++ {
		size += 4
		size += s.PreviousEpochVotes[ii].SizeSSZ()
	}

	// Field (20) 'CurrentManagers'
	size += len(s.CurrentManagers) * 20

	// Field (21) 'ManagerReplacement'
	size += len(s.ManagerReplacement) * 1

	// Field (22) 'Governance'
	if s.Governance == nil {
		s.Governance = new(GovernanceSerializable)
	}
	size += s.Governance.SizeSSZ()

	return
}

// HashTreeRoot ssz hashes the SerializableState object
func (s *SerializableState) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(s)
}

// HashTreeRootWith ssz hashes the SerializableState object with a hasher
func (s *SerializableState) HashTreeRootWith(hh *ssz.Hasher) (err error) {
	indx := hh.Index()

	// Field (0) 'CoinsState'
	if err = s.CoinsState.HashTreeRootWith(hh); err != nil {
		return
	}

	// Field (1) 'ValidatorRegistry'
	{
		subIndx := hh.Index()
		num := uint64(len(s.ValidatorRegistry))
		if num > 1099511627776 {
			err = ssz.ErrIncorrectListSize
			return
		}
		for i := uint64(0); i < num; i++ {
			if err = s.ValidatorRegistry[i].HashTreeRootWith(hh); err != nil {
				return
			}
		}
		hh.MerkleizeWithMixin(subIndx, num, 1099511627776)
	}

	// Field (2) 'LatestValidatorRegistryChange'
	hh.PutUint64(s.LatestValidatorRegistryChange)

	// Field (3) 'RANDAO'
	hh.PutBytes(s.RANDAO[:])

	// Field (4) 'NextRANDAO'
	hh.PutBytes(s.NextRANDAO[:])

	// Field (5) 'Slot'
	hh.PutUint64(s.Slot)

	// Field (6) 'EpochIndex'
	hh.PutUint64(s.EpochIndex)

	// Field (7) 'ProposerQueue'
	{
		if len(s.ProposerQueue) > 1099511627776 {
			err = ssz.ErrListTooBig
			return
		}
		subIndx := hh.Index()
		for _, i := range s.ProposerQueue {
			hh.AppendUint64(i)
		}
		hh.FillUpTo32()
		numItems := uint64(len(s.ProposerQueue))
		hh.MerkleizeWithMixin(subIndx, numItems, ssz.CalculateLimit(1099511627776, numItems, 8))
	}

	// Field (8) 'PreviousEpochVoteAssignments'
	{
		if len(s.PreviousEpochVoteAssignments) > 1099511627776 {
			err = ssz.ErrListTooBig
			return
		}
		subIndx := hh.Index()
		for _, i := range s.PreviousEpochVoteAssignments {
			hh.AppendUint64(i)
		}
		hh.FillUpTo32()
		numItems := uint64(len(s.PreviousEpochVoteAssignments))
		hh.MerkleizeWithMixin(subIndx, numItems, ssz.CalculateLimit(1099511627776, numItems, 8))
	}

	// Field (9) 'CurrentEpochVoteAssignments'
	{
		if len(s.CurrentEpochVoteAssignments) > 1099511627776 {
			err = ssz.ErrListTooBig
			return
		}
		subIndx := hh.Index()
		for _, i := range s.CurrentEpochVoteAssignments {
			hh.AppendUint64(i)
		}
		hh.FillUpTo32()
		numItems := uint64(len(s.CurrentEpochVoteAssignments))
		hh.MerkleizeWithMixin(subIndx, numItems, ssz.CalculateLimit(1099511627776, numItems, 8))
	}

	// Field (10) 'NextProposerQueue'
	{
		if len(s.NextProposerQueue) > 1099511627776 {
			err = ssz.ErrListTooBig
			return
		}
		subIndx := hh.Index()
		for _, i := range s.NextProposerQueue {
			hh.AppendUint64(i)
		}
		hh.FillUpTo32()
		numItems := uint64(len(s.NextProposerQueue))
		hh.MerkleizeWithMixin(subIndx, numItems, ssz.CalculateLimit(1099511627776, numItems, 8))
	}

	// Field (11) 'JustificationBitfield'
	hh.PutUint64(s.JustificationBitfield)

	// Field (12) 'FinalizedEpoch'
	hh.PutUint64(s.FinalizedEpoch)

	// Field (13) 'LatestBlockHashes'
	{
		if len(s.LatestBlockHashes) > 64 {
			err = ssz.ErrListTooBig
			return
		}
		subIndx := hh.Index()
		for _, i := range s.LatestBlockHashes {
			hh.Append(i[:])
		}
		numItems := uint64(len(s.LatestBlockHashes))
		hh.MerkleizeWithMixin(subIndx, numItems, ssz.CalculateLimit(64, numItems, 32))
	}

	// Field (14) 'JustifiedEpoch'
	hh.PutUint64(s.JustifiedEpoch)

	// Field (15) 'JustifiedEpochHash'
	hh.PutBytes(s.JustifiedEpochHash[:])

	// Field (16) 'CurrentEpochVotes'
	{
		subIndx := hh.Index()
		num := uint64(len(s.CurrentEpochVotes))
		if num > 1099511627776 {
			err = ssz.ErrIncorrectListSize
			return
		}
		for i := uint64(0); i < num; i++ {
			if err = s.CurrentEpochVotes[i].HashTreeRootWith(hh); err != nil {
				return
			}
		}
		hh.MerkleizeWithMixin(subIndx, num, 1099511627776)
	}

	// Field (17) 'PreviousJustifiedEpoch'
	hh.PutUint64(s.PreviousJustifiedEpoch)

	// Field (18) 'PreviousJustifiedEpochHash'
	hh.PutBytes(s.PreviousJustifiedEpochHash[:])

	// Field (19) 'PreviousEpochVotes'
	{
		subIndx := hh.Index()
		num := uint64(len(s.PreviousEpochVotes))
		if num > 1099511627776 {
			err = ssz.ErrIncorrectListSize
			return
		}
		for i := uint64(0); i < num; i++ {
			if err = s.PreviousEpochVotes[i].HashTreeRootWith(hh); err != nil {
				return
			}
		}
		hh.MerkleizeWithMixin(subIndx, num, 1099511627776)
	}

	// Field (20) 'CurrentManagers'
	{
		if len(s.CurrentManagers) > 5 {
			err = ssz.ErrListTooBig
			return
		}
		subIndx := hh.Index()
		for _, i := range s.CurrentManagers {
			hh.Append(i[:])
		}
		numItems := uint64(len(s.CurrentManagers))
		hh.MerkleizeWithMixin(subIndx, numItems, ssz.CalculateLimit(5, numItems, 32))
	}

	// Field (21) 'ManagerReplacement'
	{
		if len(s.ManagerReplacement) > 2048 {
			err = ssz.ErrListTooBig
			return
		}
		subIndx := hh.Index()
		for _, i := range s.ManagerReplacement {
			hh.AppendUint64(i)
		}
		hh.FillUpTo32()
		numItems := uint64(len(s.ManagerReplacement))
		hh.MerkleizeWithMixin(subIndx, numItems, ssz.CalculateLimit(2048, numItems, 8))
	}

	// Field (22) 'Governance'
	if err = s.Governance.HashTreeRootWith(hh); err != nil {
		return
	}

	// Field (23) 'VoteEpoch'
	hh.PutUint64(s.VoteEpoch)

	// Field (24) 'VoteEpochStartSlot'
	hh.PutUint64(s.VoteEpochStartSlot)

	// Field (25) 'VotingState'
	hh.PutUint64(s.VotingState)

	// Field (26) 'LastPaidSlot'
	hh.PutUint64(s.LastPaidSlot)

	hh.Merkleize(indx)
	return
}
