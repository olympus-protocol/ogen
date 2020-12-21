package primitives

import "github.com/olympus-protocol/ogen/pkg/chainhash"

// Contract is a smart contract in the ogen network
type Contract struct {
	// PublicKey is the public key of the contract.
	PublicKey [48]byte

	// FromPublicKey is the public key of the address interacting with the contract.
	FromPublicKey [48]byte

	// ByteCode is the valid smart contract compiled to bytecode form
	ByteCode *DynamicBytes `ssz-max:"32768"`

	// InputData is the data passed to the VM as a param for executing the contract
	InputData *DynamicBytes `ssz-max:"32768"`

	// Gas the amount of gas to be used when executing the contract
	Gas uint64
}

func (c *Contract) Marshal() ([]byte, error) {
	return c.MarshalSSZ()
}

func (c *Contract) Unmarshal(b []byte) error {
	return c.Unmarshal(b)
}

func (c *Contract) Hash() chainhash.Hash {
	h, _ := c.Marshal()
	return chainhash.HashH(h)
}
