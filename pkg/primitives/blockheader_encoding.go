// Code generated by fastssz. DO NOT EDIT.
package primitives

import (
	ssz "github.com/ferranbt/fastssz"
)

// MarshalSSZ ssz marshals the BlockHeader object
func (b *BlockHeader) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(b)
}

// MarshalSSZTo ssz marshals the BlockHeader object to a target array
func (b *BlockHeader) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf

	// Field (0) 'Version'
	dst = ssz.MarshalUint64(dst, b.Version)

	// Field (1) 'Nonce'
	dst = ssz.MarshalUint64(dst, b.Nonce)

	// Field (2) 'TxMerkleRoot'
	dst = append(dst, b.TxMerkleRoot[:]...)

	// Field (3) 'TxMultiMerkleRoot'
	dst = append(dst, b.TxMultiMerkleRoot[:]...)

	// Field (4) 'VoteMerkleRoot'
	dst = append(dst, b.VoteMerkleRoot[:]...)

	// Field (5) 'DepositMerkleRoot'
	dst = append(dst, b.DepositMerkleRoot[:]...)

	// Field (6) 'ExitMerkleRoot'
	dst = append(dst, b.ExitMerkleRoot[:]...)

	// Field (7) 'PartialExitMerkleRoot'
	dst = append(dst, b.PartialExitMerkleRoot[:]...)

	// Field (8) 'VoteSlashingMerkleRoot'
	dst = append(dst, b.VoteSlashingMerkleRoot[:]...)

	// Field (9) 'RANDAOSlashingMerkleRoot'
	dst = append(dst, b.RANDAOSlashingMerkleRoot[:]...)

	// Field (10) 'ProposerSlashingMerkleRoot'
	dst = append(dst, b.ProposerSlashingMerkleRoot[:]...)

	// Field (11) 'GovernanceVotesMerkleRoot'
	dst = append(dst, b.GovernanceVotesMerkleRoot[:]...)

	// Field (12) 'CoinProofsMerkleRoot'
	dst = append(dst, b.CoinProofsMerkleRoot[:]...)

	// Field (13) 'ExecutionsMerkleRoot'
	dst = append(dst, b.ExecutionsMerkleRoot[:]...)

	// Field (14) 'PrevBlockHash'
	dst = append(dst, b.PrevBlockHash[:]...)

	// Field (15) 'Timestamp'
	dst = ssz.MarshalUint64(dst, b.Timestamp)

	// Field (16) 'Slot'
	dst = ssz.MarshalUint64(dst, b.Slot)

	// Field (17) 'StateRoot'
	dst = append(dst, b.StateRoot[:]...)

	// Field (18) 'FeeAddress'
	dst = append(dst, b.FeeAddress[:]...)

	return
}

// UnmarshalSSZ ssz unmarshals the BlockHeader object
func (b *BlockHeader) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size != 500 {
		return ssz.ErrSize
	}

	// Field (0) 'Version'
	b.Version = ssz.UnmarshallUint64(buf[0:8])

	// Field (1) 'Nonce'
	b.Nonce = ssz.UnmarshallUint64(buf[8:16])

	// Field (2) 'TxMerkleRoot'
	copy(b.TxMerkleRoot[:], buf[16:48])

	// Field (3) 'TxMultiMerkleRoot'
	copy(b.TxMultiMerkleRoot[:], buf[48:80])

	// Field (4) 'VoteMerkleRoot'
	copy(b.VoteMerkleRoot[:], buf[80:112])

	// Field (5) 'DepositMerkleRoot'
	copy(b.DepositMerkleRoot[:], buf[112:144])

	// Field (6) 'ExitMerkleRoot'
	copy(b.ExitMerkleRoot[:], buf[144:176])

	// Field (7) 'PartialExitMerkleRoot'
	copy(b.PartialExitMerkleRoot[:], buf[176:208])

	// Field (8) 'VoteSlashingMerkleRoot'
	copy(b.VoteSlashingMerkleRoot[:], buf[208:240])

	// Field (9) 'RANDAOSlashingMerkleRoot'
	copy(b.RANDAOSlashingMerkleRoot[:], buf[240:272])

	// Field (10) 'ProposerSlashingMerkleRoot'
	copy(b.ProposerSlashingMerkleRoot[:], buf[272:304])

	// Field (11) 'GovernanceVotesMerkleRoot'
	copy(b.GovernanceVotesMerkleRoot[:], buf[304:336])

	// Field (12) 'CoinProofsMerkleRoot'
	copy(b.CoinProofsMerkleRoot[:], buf[336:368])

	// Field (13) 'ExecutionsMerkleRoot'
	copy(b.ExecutionsMerkleRoot[:], buf[368:400])

	// Field (14) 'PrevBlockHash'
	copy(b.PrevBlockHash[:], buf[400:432])

	// Field (15) 'Timestamp'
	b.Timestamp = ssz.UnmarshallUint64(buf[432:440])

	// Field (16) 'Slot'
	b.Slot = ssz.UnmarshallUint64(buf[440:448])

	// Field (17) 'StateRoot'
	copy(b.StateRoot[:], buf[448:480])

	// Field (18) 'FeeAddress'
	copy(b.FeeAddress[:], buf[480:500])

	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the BlockHeader object
func (b *BlockHeader) SizeSSZ() (size int) {
	size = 500
	return
}

// HashTreeRoot ssz hashes the BlockHeader object
func (b *BlockHeader) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(b)
}

// HashTreeRootWith ssz hashes the BlockHeader object with a hasher
func (b *BlockHeader) HashTreeRootWith(hh *ssz.Hasher) (err error) {
	indx := hh.Index()

	// Field (0) 'Version'
	hh.PutUint64(b.Version)

	// Field (1) 'Nonce'
	hh.PutUint64(b.Nonce)

	// Field (2) 'TxMerkleRoot'
	hh.PutBytes(b.TxMerkleRoot[:])

	// Field (3) 'TxMultiMerkleRoot'
	hh.PutBytes(b.TxMultiMerkleRoot[:])

	// Field (4) 'VoteMerkleRoot'
	hh.PutBytes(b.VoteMerkleRoot[:])

	// Field (5) 'DepositMerkleRoot'
	hh.PutBytes(b.DepositMerkleRoot[:])

	// Field (6) 'ExitMerkleRoot'
	hh.PutBytes(b.ExitMerkleRoot[:])

	// Field (7) 'PartialExitMerkleRoot'
	hh.PutBytes(b.PartialExitMerkleRoot[:])

	// Field (8) 'VoteSlashingMerkleRoot'
	hh.PutBytes(b.VoteSlashingMerkleRoot[:])

	// Field (9) 'RANDAOSlashingMerkleRoot'
	hh.PutBytes(b.RANDAOSlashingMerkleRoot[:])

	// Field (10) 'ProposerSlashingMerkleRoot'
	hh.PutBytes(b.ProposerSlashingMerkleRoot[:])

	// Field (11) 'GovernanceVotesMerkleRoot'
	hh.PutBytes(b.GovernanceVotesMerkleRoot[:])

	// Field (12) 'CoinProofsMerkleRoot'
	hh.PutBytes(b.CoinProofsMerkleRoot[:])

	// Field (13) 'ExecutionsMerkleRoot'
	hh.PutBytes(b.ExecutionsMerkleRoot[:])

	// Field (14) 'PrevBlockHash'
	hh.PutBytes(b.PrevBlockHash[:])

	// Field (15) 'Timestamp'
	hh.PutUint64(b.Timestamp)

	// Field (16) 'Slot'
	hh.PutUint64(b.Slot)

	// Field (17) 'StateRoot'
	hh.PutBytes(b.StateRoot[:])

	// Field (18) 'FeeAddress'
	hh.PutBytes(b.FeeAddress[:])

	hh.Merkleize(indx)
	return
}
