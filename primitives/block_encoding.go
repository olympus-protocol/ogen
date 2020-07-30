// Code generated by fastssz. DO NOT EDIT.
package primitives

import (
	ssz "github.com/ferranbt/fastssz"
)

// MarshalSSZ ssz marshals the Votes object
func (v *Votes) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(v)
}

// MarshalSSZTo ssz marshals the Votes object to a target array
func (v *Votes) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf
	offset := int(4)

	// Offset (0) 'Votes'
	dst = ssz.WriteOffset(dst, offset)
	for ii := 0; ii < len(v.Votes); ii++ {
		offset += 4
		offset += v.Votes[ii].SizeSSZ()
	}

	// Field (0) 'Votes'
	if len(v.Votes) > 32 {
		err = ssz.ErrListTooBig
		return
	}
	{
		offset = 4 * len(v.Votes)
		for ii := 0; ii < len(v.Votes); ii++ {
			dst = ssz.WriteOffset(dst, offset)
			offset += v.Votes[ii].SizeSSZ()
		}
	}
	for ii := 0; ii < len(v.Votes); ii++ {
		if dst, err = v.Votes[ii].MarshalSSZTo(dst); err != nil {
			return
		}
	}

	return
}

// UnmarshalSSZ ssz unmarshals the Votes object
func (v *Votes) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size < 4 {
		return ssz.ErrSize
	}

	tail := buf
	var o0 uint64

	// Offset (0) 'Votes'
	if o0 = ssz.ReadOffset(buf[0:4]); o0 > size {
		return ssz.ErrOffset
	}

	// Field (0) 'Votes'
	{
		buf = tail[o0:]
		num, err := ssz.DecodeDynamicLength(buf, 32)
		if err != nil {
			return err
		}
		v.Votes = make([]*MultiValidatorVote, num)
		err = ssz.UnmarshalDynamic(buf, num, func(indx int, buf []byte) (err error) {
			if v.Votes[indx] == nil {
				v.Votes[indx] = new(MultiValidatorVote)
			}
			if err = v.Votes[indx].UnmarshalSSZ(buf); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the Votes object
func (v *Votes) SizeSSZ() (size int) {
	size = 4

	// Field (0) 'Votes'
	for ii := 0; ii < len(v.Votes); ii++ {
		size += 4
		size += v.Votes[ii].SizeSSZ()
	}

	return
}

// HashTreeRoot ssz hashes the Votes object
func (v *Votes) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(v)
}

// HashTreeRootWith ssz hashes the Votes object with a hasher
func (v *Votes) HashTreeRootWith(hh *ssz.Hasher) (err error) {
	indx := hh.Index()

	// Field (0) 'Votes'
	{
		subIndx := hh.Index()
		num := uint64(len(v.Votes))
		if num > 32 {
			err = ssz.ErrIncorrectListSize
			return
		}
		for i := uint64(0); i < num; i++ {
			if err = v.Votes[i].HashTreeRootWith(hh); err != nil {
				return
			}
		}
		hh.MerkleizeWithMixin(subIndx, num, 32)
	}

	hh.Merkleize(indx)
	return
}

// MarshalSSZ ssz marshals the Txs object
func (t *Txs) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(t)
}

// MarshalSSZTo ssz marshals the Txs object to a target array
func (t *Txs) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf
	offset := int(4)

	// Offset (0) 'Txs'
	dst = ssz.WriteOffset(dst, offset)
	offset += len(t.Txs) * 188

	// Field (0) 'Txs'
	if len(t.Txs) > 1000 {
		err = ssz.ErrListTooBig
		return
	}
	for ii := 0; ii < len(t.Txs); ii++ {
		if dst, err = t.Txs[ii].MarshalSSZTo(dst); err != nil {
			return
		}
	}

	return
}

// UnmarshalSSZ ssz unmarshals the Txs object
func (t *Txs) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size < 4 {
		return ssz.ErrSize
	}

	tail := buf
	var o0 uint64

	// Offset (0) 'Txs'
	if o0 = ssz.ReadOffset(buf[0:4]); o0 > size {
		return ssz.ErrOffset
	}

	// Field (0) 'Txs'
	{
		buf = tail[o0:]
		num, err := ssz.DivideInt2(len(buf), 188, 1000)
		if err != nil {
			return err
		}
		t.Txs = make([]*Tx, num)
		for ii := 0; ii < num; ii++ {
			if t.Txs[ii] == nil {
				t.Txs[ii] = new(Tx)
			}
			if err = t.Txs[ii].UnmarshalSSZ(buf[ii*188 : (ii+1)*188]); err != nil {
				return err
			}
		}
	}
	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the Txs object
func (t *Txs) SizeSSZ() (size int) {
	size = 4

	// Field (0) 'Txs'
	size += len(t.Txs) * 188

	return
}

// HashTreeRoot ssz hashes the Txs object
func (t *Txs) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(t)
}

// HashTreeRootWith ssz hashes the Txs object with a hasher
func (t *Txs) HashTreeRootWith(hh *ssz.Hasher) (err error) {
	indx := hh.Index()

	// Field (0) 'Txs'
	{
		subIndx := hh.Index()
		num := uint64(len(t.Txs))
		if num > 1000 {
			err = ssz.ErrIncorrectListSize
			return
		}
		for i := uint64(0); i < num; i++ {
			if err = t.Txs[i].HashTreeRootWith(hh); err != nil {
				return
			}
		}
		hh.MerkleizeWithMixin(subIndx, num, 1000)
	}

	hh.Merkleize(indx)
	return
}

// MarshalSSZ ssz marshals the Deposits object
func (d *Deposits) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(d)
}

// MarshalSSZTo ssz marshals the Deposits object to a target array
func (d *Deposits) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf
	offset := int(4)

	// Offset (0) 'Deposits'
	dst = ssz.WriteOffset(dst, offset)
	offset += len(d.Deposits) * 308

	// Field (0) 'Deposits'
	if len(d.Deposits) > 32 {
		err = ssz.ErrListTooBig
		return
	}
	for ii := 0; ii < len(d.Deposits); ii++ {
		if dst, err = d.Deposits[ii].MarshalSSZTo(dst); err != nil {
			return
		}
	}

	return
}

// UnmarshalSSZ ssz unmarshals the Deposits object
func (d *Deposits) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size < 4 {
		return ssz.ErrSize
	}

	tail := buf
	var o0 uint64

	// Offset (0) 'Deposits'
	if o0 = ssz.ReadOffset(buf[0:4]); o0 > size {
		return ssz.ErrOffset
	}

	// Field (0) 'Deposits'
	{
		buf = tail[o0:]
		num, err := ssz.DivideInt2(len(buf), 308, 32)
		if err != nil {
			return err
		}
		d.Deposits = make([]*Deposit, num)
		for ii := 0; ii < num; ii++ {
			if d.Deposits[ii] == nil {
				d.Deposits[ii] = new(Deposit)
			}
			if err = d.Deposits[ii].UnmarshalSSZ(buf[ii*308 : (ii+1)*308]); err != nil {
				return err
			}
		}
	}
	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the Deposits object
func (d *Deposits) SizeSSZ() (size int) {
	size = 4

	// Field (0) 'Deposits'
	size += len(d.Deposits) * 308

	return
}

// HashTreeRoot ssz hashes the Deposits object
func (d *Deposits) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(d)
}

// HashTreeRootWith ssz hashes the Deposits object with a hasher
func (d *Deposits) HashTreeRootWith(hh *ssz.Hasher) (err error) {
	indx := hh.Index()

	// Field (0) 'Deposits'
	{
		subIndx := hh.Index()
		num := uint64(len(d.Deposits))
		if num > 32 {
			err = ssz.ErrIncorrectListSize
			return
		}
		for i := uint64(0); i < num; i++ {
			if err = d.Deposits[i].HashTreeRootWith(hh); err != nil {
				return
			}
		}
		hh.MerkleizeWithMixin(subIndx, num, 32)
	}

	hh.Merkleize(indx)
	return
}

// MarshalSSZ ssz marshals the Exits object
func (e *Exits) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(e)
}

// MarshalSSZTo ssz marshals the Exits object to a target array
func (e *Exits) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf
	offset := int(4)

	// Offset (0) 'Exits'
	dst = ssz.WriteOffset(dst, offset)
	offset += len(e.Exits) * 192

	// Field (0) 'Exits'
	if len(e.Exits) > 32 {
		err = ssz.ErrListTooBig
		return
	}
	for ii := 0; ii < len(e.Exits); ii++ {
		if dst, err = e.Exits[ii].MarshalSSZTo(dst); err != nil {
			return
		}
	}

	return
}

// UnmarshalSSZ ssz unmarshals the Exits object
func (e *Exits) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size < 4 {
		return ssz.ErrSize
	}

	tail := buf
	var o0 uint64

	// Offset (0) 'Exits'
	if o0 = ssz.ReadOffset(buf[0:4]); o0 > size {
		return ssz.ErrOffset
	}

	// Field (0) 'Exits'
	{
		buf = tail[o0:]
		num, err := ssz.DivideInt2(len(buf), 192, 32)
		if err != nil {
			return err
		}
		e.Exits = make([]*Exit, num)
		for ii := 0; ii < num; ii++ {
			if e.Exits[ii] == nil {
				e.Exits[ii] = new(Exit)
			}
			if err = e.Exits[ii].UnmarshalSSZ(buf[ii*192 : (ii+1)*192]); err != nil {
				return err
			}
		}
	}
	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the Exits object
func (e *Exits) SizeSSZ() (size int) {
	size = 4

	// Field (0) 'Exits'
	size += len(e.Exits) * 192

	return
}

// HashTreeRoot ssz hashes the Exits object
func (e *Exits) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(e)
}

// HashTreeRootWith ssz hashes the Exits object with a hasher
func (e *Exits) HashTreeRootWith(hh *ssz.Hasher) (err error) {
	indx := hh.Index()

	// Field (0) 'Exits'
	{
		subIndx := hh.Index()
		num := uint64(len(e.Exits))
		if num > 32 {
			err = ssz.ErrIncorrectListSize
			return
		}
		for i := uint64(0); i < num; i++ {
			if err = e.Exits[i].HashTreeRootWith(hh); err != nil {
				return
			}
		}
		hh.MerkleizeWithMixin(subIndx, num, 32)
	}

	hh.Merkleize(indx)
	return
}

// MarshalSSZ ssz marshals the VoteSlashings object
func (v *VoteSlashings) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(v)
}

// MarshalSSZTo ssz marshals the VoteSlashings object to a target array
func (v *VoteSlashings) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf
	offset := int(4)

	// Offset (0) 'VoteSlashings'
	dst = ssz.WriteOffset(dst, offset)
	for ii := 0; ii < len(v.VoteSlashings); ii++ {
		offset += 4
		offset += v.VoteSlashings[ii].SizeSSZ()
	}

	// Field (0) 'VoteSlashings'
	if len(v.VoteSlashings) > 10 {
		err = ssz.ErrListTooBig
		return
	}
	{
		offset = 4 * len(v.VoteSlashings)
		for ii := 0; ii < len(v.VoteSlashings); ii++ {
			dst = ssz.WriteOffset(dst, offset)
			offset += v.VoteSlashings[ii].SizeSSZ()
		}
	}
	for ii := 0; ii < len(v.VoteSlashings); ii++ {
		if dst, err = v.VoteSlashings[ii].MarshalSSZTo(dst); err != nil {
			return
		}
	}

	return
}

// UnmarshalSSZ ssz unmarshals the VoteSlashings object
func (v *VoteSlashings) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size < 4 {
		return ssz.ErrSize
	}

	tail := buf
	var o0 uint64

	// Offset (0) 'VoteSlashings'
	if o0 = ssz.ReadOffset(buf[0:4]); o0 > size {
		return ssz.ErrOffset
	}

	// Field (0) 'VoteSlashings'
	{
		buf = tail[o0:]
		num, err := ssz.DecodeDynamicLength(buf, 10)
		if err != nil {
			return err
		}
		v.VoteSlashings = make([]*VoteSlashing, num)
		err = ssz.UnmarshalDynamic(buf, num, func(indx int, buf []byte) (err error) {
			if v.VoteSlashings[indx] == nil {
				v.VoteSlashings[indx] = new(VoteSlashing)
			}
			if err = v.VoteSlashings[indx].UnmarshalSSZ(buf); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the VoteSlashings object
func (v *VoteSlashings) SizeSSZ() (size int) {
	size = 4

	// Field (0) 'VoteSlashings'
	for ii := 0; ii < len(v.VoteSlashings); ii++ {
		size += 4
		size += v.VoteSlashings[ii].SizeSSZ()
	}

	return
}

// HashTreeRoot ssz hashes the VoteSlashings object
func (v *VoteSlashings) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(v)
}

// HashTreeRootWith ssz hashes the VoteSlashings object with a hasher
func (v *VoteSlashings) HashTreeRootWith(hh *ssz.Hasher) (err error) {
	indx := hh.Index()

	// Field (0) 'VoteSlashings'
	{
		subIndx := hh.Index()
		num := uint64(len(v.VoteSlashings))
		if num > 10 {
			err = ssz.ErrIncorrectListSize
			return
		}
		for i := uint64(0); i < num; i++ {
			if err = v.VoteSlashings[i].HashTreeRootWith(hh); err != nil {
				return
			}
		}
		hh.MerkleizeWithMixin(subIndx, num, 10)
	}

	hh.Merkleize(indx)
	return
}

// MarshalSSZ ssz marshals the RANDAOSlashings object
func (r *RANDAOSlashings) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(r)
}

// MarshalSSZTo ssz marshals the RANDAOSlashings object to a target array
func (r *RANDAOSlashings) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf
	offset := int(4)

	// Offset (0) 'RANDAOSlashings'
	dst = ssz.WriteOffset(dst, offset)
	offset += len(r.RANDAOSlashings) * 152

	// Field (0) 'RANDAOSlashings'
	if len(r.RANDAOSlashings) > 20 {
		err = ssz.ErrListTooBig
		return
	}
	for ii := 0; ii < len(r.RANDAOSlashings); ii++ {
		if dst, err = r.RANDAOSlashings[ii].MarshalSSZTo(dst); err != nil {
			return
		}
	}

	return
}

// UnmarshalSSZ ssz unmarshals the RANDAOSlashings object
func (r *RANDAOSlashings) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size < 4 {
		return ssz.ErrSize
	}

	tail := buf
	var o0 uint64

	// Offset (0) 'RANDAOSlashings'
	if o0 = ssz.ReadOffset(buf[0:4]); o0 > size {
		return ssz.ErrOffset
	}

	// Field (0) 'RANDAOSlashings'
	{
		buf = tail[o0:]
		num, err := ssz.DivideInt2(len(buf), 152, 20)
		if err != nil {
			return err
		}
		r.RANDAOSlashings = make([]*RANDAOSlashing, num)
		for ii := 0; ii < num; ii++ {
			if r.RANDAOSlashings[ii] == nil {
				r.RANDAOSlashings[ii] = new(RANDAOSlashing)
			}
			if err = r.RANDAOSlashings[ii].UnmarshalSSZ(buf[ii*152 : (ii+1)*152]); err != nil {
				return err
			}
		}
	}
	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the RANDAOSlashings object
func (r *RANDAOSlashings) SizeSSZ() (size int) {
	size = 4

	// Field (0) 'RANDAOSlashings'
	size += len(r.RANDAOSlashings) * 152

	return
}

// HashTreeRoot ssz hashes the RANDAOSlashings object
func (r *RANDAOSlashings) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(r)
}

// HashTreeRootWith ssz hashes the RANDAOSlashings object with a hasher
func (r *RANDAOSlashings) HashTreeRootWith(hh *ssz.Hasher) (err error) {
	indx := hh.Index()

	// Field (0) 'RANDAOSlashings'
	{
		subIndx := hh.Index()
		num := uint64(len(r.RANDAOSlashings))
		if num > 20 {
			err = ssz.ErrIncorrectListSize
			return
		}
		for i := uint64(0); i < num; i++ {
			if err = r.RANDAOSlashings[i].HashTreeRootWith(hh); err != nil {
				return
			}
		}
		hh.MerkleizeWithMixin(subIndx, num, 20)
	}

	hh.Merkleize(indx)
	return
}

// MarshalSSZ ssz marshals the ProposerSlashings object
func (p *ProposerSlashings) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(p)
}

// MarshalSSZTo ssz marshals the ProposerSlashings object to a target array
func (p *ProposerSlashings) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf
	offset := int(4)

	// Offset (0) 'ProposerSlashings'
	dst = ssz.WriteOffset(dst, offset)
	offset += len(p.ProposerSlashings) * 984

	// Field (0) 'ProposerSlashings'
	if len(p.ProposerSlashings) > 2 {
		err = ssz.ErrListTooBig
		return
	}
	for ii := 0; ii < len(p.ProposerSlashings); ii++ {
		if dst, err = p.ProposerSlashings[ii].MarshalSSZTo(dst); err != nil {
			return
		}
	}

	return
}

// UnmarshalSSZ ssz unmarshals the ProposerSlashings object
func (p *ProposerSlashings) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size < 4 {
		return ssz.ErrSize
	}

	tail := buf
	var o0 uint64

	// Offset (0) 'ProposerSlashings'
	if o0 = ssz.ReadOffset(buf[0:4]); o0 > size {
		return ssz.ErrOffset
	}

	// Field (0) 'ProposerSlashings'
	{
		buf = tail[o0:]
		num, err := ssz.DivideInt2(len(buf), 984, 2)
		if err != nil {
			return err
		}
		p.ProposerSlashings = make([]*ProposerSlashing, num)
		for ii := 0; ii < num; ii++ {
			if p.ProposerSlashings[ii] == nil {
				p.ProposerSlashings[ii] = new(ProposerSlashing)
			}
			if err = p.ProposerSlashings[ii].UnmarshalSSZ(buf[ii*984 : (ii+1)*984]); err != nil {
				return err
			}
		}
	}
	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the ProposerSlashings object
func (p *ProposerSlashings) SizeSSZ() (size int) {
	size = 4

	// Field (0) 'ProposerSlashings'
	size += len(p.ProposerSlashings) * 984

	return
}

// HashTreeRoot ssz hashes the ProposerSlashings object
func (p *ProposerSlashings) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(p)
}

// HashTreeRootWith ssz hashes the ProposerSlashings object with a hasher
func (p *ProposerSlashings) HashTreeRootWith(hh *ssz.Hasher) (err error) {
	indx := hh.Index()

	// Field (0) 'ProposerSlashings'
	{
		subIndx := hh.Index()
		num := uint64(len(p.ProposerSlashings))
		if num > 2 {
			err = ssz.ErrIncorrectListSize
			return
		}
		for i := uint64(0); i < num; i++ {
			if err = p.ProposerSlashings[i].HashTreeRootWith(hh); err != nil {
				return
			}
		}
		hh.MerkleizeWithMixin(subIndx, num, 2)
	}

	hh.Merkleize(indx)
	return
}

// MarshalSSZ ssz marshals the GovernanceVotes object
func (g *GovernanceVotes) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(g)
}

// MarshalSSZTo ssz marshals the GovernanceVotes object to a target array
func (g *GovernanceVotes) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf
	offset := int(4)

	// Offset (0) 'GovernanceVotes'
	dst = ssz.WriteOffset(dst, offset)
	offset += len(g.GovernanceVotes) * 4112

	// Field (0) 'GovernanceVotes'
	if len(g.GovernanceVotes) > 128 {
		err = ssz.ErrListTooBig
		return
	}
	for ii := 0; ii < len(g.GovernanceVotes); ii++ {
		if dst, err = g.GovernanceVotes[ii].MarshalSSZTo(dst); err != nil {
			return
		}
	}

	return
}

// UnmarshalSSZ ssz unmarshals the GovernanceVotes object
func (g *GovernanceVotes) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size < 4 {
		return ssz.ErrSize
	}

	tail := buf
	var o0 uint64

	// Offset (0) 'GovernanceVotes'
	if o0 = ssz.ReadOffset(buf[0:4]); o0 > size {
		return ssz.ErrOffset
	}

	// Field (0) 'GovernanceVotes'
	{
		buf = tail[o0:]
		num, err := ssz.DivideInt2(len(buf), 4112, 128)
		if err != nil {
			return err
		}
		g.GovernanceVotes = make([]*GovernanceVote, num)
		for ii := 0; ii < num; ii++ {
			if g.GovernanceVotes[ii] == nil {
				g.GovernanceVotes[ii] = new(GovernanceVote)
			}
			if err = g.GovernanceVotes[ii].UnmarshalSSZ(buf[ii*4112 : (ii+1)*4112]); err != nil {
				return err
			}
		}
	}
	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the GovernanceVotes object
func (g *GovernanceVotes) SizeSSZ() (size int) {
	size = 4

	// Field (0) 'GovernanceVotes'
	size += len(g.GovernanceVotes) * 4112

	return
}

// HashTreeRoot ssz hashes the GovernanceVotes object
func (g *GovernanceVotes) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(g)
}

// HashTreeRootWith ssz hashes the GovernanceVotes object with a hasher
func (g *GovernanceVotes) HashTreeRootWith(hh *ssz.Hasher) (err error) {
	indx := hh.Index()

	// Field (0) 'GovernanceVotes'
	{
		subIndx := hh.Index()
		num := uint64(len(g.GovernanceVotes))
		if num > 128 {
			err = ssz.ErrIncorrectListSize
			return
		}
		for i := uint64(0); i < num; i++ {
			if err = g.GovernanceVotes[i].HashTreeRootWith(hh); err != nil {
				return
			}
		}
		hh.MerkleizeWithMixin(subIndx, num, 128)
	}

	hh.Merkleize(indx)
	return
}

// MarshalSSZ ssz marshals the Block object
func (b *Block) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(b)
}

// MarshalSSZTo ssz marshals the Block object to a target array
func (b *Block) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf
	offset := int(596)

	// Field (0) 'Header'
	if b.Header == nil {
		b.Header = new(BlockHeader)
	}
	if dst, err = b.Header.MarshalSSZTo(dst); err != nil {
		return
	}

	// Offset (1) 'Votes'
	dst = ssz.WriteOffset(dst, offset)
	if b.Votes == nil {
		b.Votes = new(Votes)
	}
	offset += b.Votes.SizeSSZ()

	// Offset (2) 'Txs'
	dst = ssz.WriteOffset(dst, offset)
	if b.Txs == nil {
		b.Txs = new(Txs)
	}
	offset += b.Txs.SizeSSZ()

	// Offset (3) 'Deposits'
	dst = ssz.WriteOffset(dst, offset)
	if b.Deposits == nil {
		b.Deposits = new(Deposits)
	}
	offset += b.Deposits.SizeSSZ()

	// Offset (4) 'Exits'
	dst = ssz.WriteOffset(dst, offset)
	if b.Exits == nil {
		b.Exits = new(Exits)
	}
	offset += b.Exits.SizeSSZ()

	// Offset (5) 'VoteSlashings'
	dst = ssz.WriteOffset(dst, offset)
	if b.VoteSlashings == nil {
		b.VoteSlashings = new(VoteSlashings)
	}
	offset += b.VoteSlashings.SizeSSZ()

	// Offset (6) 'RANDAOSlashings'
	dst = ssz.WriteOffset(dst, offset)
	if b.RANDAOSlashings == nil {
		b.RANDAOSlashings = new(RANDAOSlashings)
	}
	offset += b.RANDAOSlashings.SizeSSZ()

	// Offset (7) 'ProposerSlashings'
	dst = ssz.WriteOffset(dst, offset)
	if b.ProposerSlashings == nil {
		b.ProposerSlashings = new(ProposerSlashings)
	}
	offset += b.ProposerSlashings.SizeSSZ()

	// Offset (8) 'GovernanceVotes'
	dst = ssz.WriteOffset(dst, offset)
	if b.GovernanceVotes == nil {
		b.GovernanceVotes = new(GovernanceVotes)
	}
	offset += b.GovernanceVotes.SizeSSZ()

	// Field (9) 'Signature'
	dst = append(dst, b.Signature[:]...)

	// Field (10) 'RandaoSignature'
	dst = append(dst, b.RandaoSignature[:]...)

	// Field (1) 'Votes'
	if dst, err = b.Votes.MarshalSSZTo(dst); err != nil {
		return
	}

	// Field (2) 'Txs'
	if dst, err = b.Txs.MarshalSSZTo(dst); err != nil {
		return
	}

	// Field (3) 'Deposits'
	if dst, err = b.Deposits.MarshalSSZTo(dst); err != nil {
		return
	}

	// Field (4) 'Exits'
	if dst, err = b.Exits.MarshalSSZTo(dst); err != nil {
		return
	}

	// Field (5) 'VoteSlashings'
	if dst, err = b.VoteSlashings.MarshalSSZTo(dst); err != nil {
		return
	}

	// Field (6) 'RANDAOSlashings'
	if dst, err = b.RANDAOSlashings.MarshalSSZTo(dst); err != nil {
		return
	}

	// Field (7) 'ProposerSlashings'
	if dst, err = b.ProposerSlashings.MarshalSSZTo(dst); err != nil {
		return
	}

	// Field (8) 'GovernanceVotes'
	if dst, err = b.GovernanceVotes.MarshalSSZTo(dst); err != nil {
		return
	}

	return
}

// UnmarshalSSZ ssz unmarshals the Block object
func (b *Block) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size < 596 {
		return ssz.ErrSize
	}

	tail := buf
	var o1, o2, o3, o4, o5, o6, o7, o8 uint64

	// Field (0) 'Header'
	if b.Header == nil {
		b.Header = new(BlockHeader)
	}
	if err = b.Header.UnmarshalSSZ(buf[0:372]); err != nil {
		return err
	}

	// Offset (1) 'Votes'
	if o1 = ssz.ReadOffset(buf[372:376]); o1 > size {
		return ssz.ErrOffset
	}

	// Offset (2) 'Txs'
	if o2 = ssz.ReadOffset(buf[376:380]); o2 > size || o1 > o2 {
		return ssz.ErrOffset
	}

	// Offset (3) 'Deposits'
	if o3 = ssz.ReadOffset(buf[380:384]); o3 > size || o2 > o3 {
		return ssz.ErrOffset
	}

	// Offset (4) 'Exits'
	if o4 = ssz.ReadOffset(buf[384:388]); o4 > size || o3 > o4 {
		return ssz.ErrOffset
	}

	// Offset (5) 'VoteSlashings'
	if o5 = ssz.ReadOffset(buf[388:392]); o5 > size || o4 > o5 {
		return ssz.ErrOffset
	}

	// Offset (6) 'RANDAOSlashings'
	if o6 = ssz.ReadOffset(buf[392:396]); o6 > size || o5 > o6 {
		return ssz.ErrOffset
	}

	// Offset (7) 'ProposerSlashings'
	if o7 = ssz.ReadOffset(buf[396:400]); o7 > size || o6 > o7 {
		return ssz.ErrOffset
	}

	// Offset (8) 'GovernanceVotes'
	if o8 = ssz.ReadOffset(buf[400:404]); o8 > size || o7 > o8 {
		return ssz.ErrOffset
	}

	// Field (9) 'Signature'
	copy(b.Signature[:], buf[404:500])

	// Field (10) 'RandaoSignature'
	copy(b.RandaoSignature[:], buf[500:596])

	// Field (1) 'Votes'
	{
		buf = tail[o1:o2]
		if b.Votes == nil {
			b.Votes = new(Votes)
		}
		if err = b.Votes.UnmarshalSSZ(buf); err != nil {
			return err
		}
	}

	// Field (2) 'Txs'
	{
		buf = tail[o2:o3]
		if b.Txs == nil {
			b.Txs = new(Txs)
		}
		if err = b.Txs.UnmarshalSSZ(buf); err != nil {
			return err
		}
	}

	// Field (3) 'Deposits'
	{
		buf = tail[o3:o4]
		if b.Deposits == nil {
			b.Deposits = new(Deposits)
		}
		if err = b.Deposits.UnmarshalSSZ(buf); err != nil {
			return err
		}
	}

	// Field (4) 'Exits'
	{
		buf = tail[o4:o5]
		if b.Exits == nil {
			b.Exits = new(Exits)
		}
		if err = b.Exits.UnmarshalSSZ(buf); err != nil {
			return err
		}
	}

	// Field (5) 'VoteSlashings'
	{
		buf = tail[o5:o6]
		if b.VoteSlashings == nil {
			b.VoteSlashings = new(VoteSlashings)
		}
		if err = b.VoteSlashings.UnmarshalSSZ(buf); err != nil {
			return err
		}
	}

	// Field (6) 'RANDAOSlashings'
	{
		buf = tail[o6:o7]
		if b.RANDAOSlashings == nil {
			b.RANDAOSlashings = new(RANDAOSlashings)
		}
		if err = b.RANDAOSlashings.UnmarshalSSZ(buf); err != nil {
			return err
		}
	}

	// Field (7) 'ProposerSlashings'
	{
		buf = tail[o7:o8]
		if b.ProposerSlashings == nil {
			b.ProposerSlashings = new(ProposerSlashings)
		}
		if err = b.ProposerSlashings.UnmarshalSSZ(buf); err != nil {
			return err
		}
	}

	// Field (8) 'GovernanceVotes'
	{
		buf = tail[o8:]
		if b.GovernanceVotes == nil {
			b.GovernanceVotes = new(GovernanceVotes)
		}
		if err = b.GovernanceVotes.UnmarshalSSZ(buf); err != nil {
			return err
		}
	}
	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the Block object
func (b *Block) SizeSSZ() (size int) {
	size = 596

	// Field (1) 'Votes'
	if b.Votes == nil {
		b.Votes = new(Votes)
	}
	size += b.Votes.SizeSSZ()

	// Field (2) 'Txs'
	if b.Txs == nil {
		b.Txs = new(Txs)
	}
	size += b.Txs.SizeSSZ()

	// Field (3) 'Deposits'
	if b.Deposits == nil {
		b.Deposits = new(Deposits)
	}
	size += b.Deposits.SizeSSZ()

	// Field (4) 'Exits'
	if b.Exits == nil {
		b.Exits = new(Exits)
	}
	size += b.Exits.SizeSSZ()

	// Field (5) 'VoteSlashings'
	if b.VoteSlashings == nil {
		b.VoteSlashings = new(VoteSlashings)
	}
	size += b.VoteSlashings.SizeSSZ()

	// Field (6) 'RANDAOSlashings'
	if b.RANDAOSlashings == nil {
		b.RANDAOSlashings = new(RANDAOSlashings)
	}
	size += b.RANDAOSlashings.SizeSSZ()

	// Field (7) 'ProposerSlashings'
	if b.ProposerSlashings == nil {
		b.ProposerSlashings = new(ProposerSlashings)
	}
	size += b.ProposerSlashings.SizeSSZ()

	// Field (8) 'GovernanceVotes'
	if b.GovernanceVotes == nil {
		b.GovernanceVotes = new(GovernanceVotes)
	}
	size += b.GovernanceVotes.SizeSSZ()

	return
}

// HashTreeRoot ssz hashes the Block object
func (b *Block) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(b)
}

// HashTreeRootWith ssz hashes the Block object with a hasher
func (b *Block) HashTreeRootWith(hh *ssz.Hasher) (err error) {
	indx := hh.Index()

	// Field (0) 'Header'
	if err = b.Header.HashTreeRootWith(hh); err != nil {
		return
	}

	// Field (1) 'Votes'
	if err = b.Votes.HashTreeRootWith(hh); err != nil {
		return
	}

	// Field (2) 'Txs'
	if err = b.Txs.HashTreeRootWith(hh); err != nil {
		return
	}

	// Field (3) 'Deposits'
	if err = b.Deposits.HashTreeRootWith(hh); err != nil {
		return
	}

	// Field (4) 'Exits'
	if err = b.Exits.HashTreeRootWith(hh); err != nil {
		return
	}

	// Field (5) 'VoteSlashings'
	if err = b.VoteSlashings.HashTreeRootWith(hh); err != nil {
		return
	}

	// Field (6) 'RANDAOSlashings'
	if err = b.RANDAOSlashings.HashTreeRootWith(hh); err != nil {
		return
	}

	// Field (7) 'ProposerSlashings'
	if err = b.ProposerSlashings.HashTreeRootWith(hh); err != nil {
		return
	}

	// Field (8) 'GovernanceVotes'
	if err = b.GovernanceVotes.HashTreeRootWith(hh); err != nil {
		return
	}

	// Field (9) 'Signature'
	hh.PutBytes(b.Signature[:])

	// Field (10) 'RandaoSignature'
	hh.PutBytes(b.RandaoSignature[:])

	hh.Merkleize(indx)
	return
}
