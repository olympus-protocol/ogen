package burnproof

import (
	ssz "github.com/ferranbt/fastssz"
)

func (c *CoinsProof) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(c)
}

func (c *CoinsProof) MarshalSSZTo(buf []byte) (dst []byte, err error) {

	return
}

func (c *CoinsProof) UnmarshalSSZ(buf []byte) error {
	return nil
}

func (c *CoinsProof) SizeSSZ() (size int) {
	size = 24

	size += len(c.MerkleBranch) * 32

	size += len(c.PkScript)

	size += c.Transaction.SerializeSize()

	return
}

func (c *CoinsProof) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(c)
}

func (c *CoinsProof) HashTreeRootWith(hh *ssz.Hasher) error {
	return nil
}
