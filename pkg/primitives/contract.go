package primitives

import (
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
)

// Contract is a smart contract in the ogen network
type Contract struct {
	// PublicKey is the public key of the contract.
	PublicKey [48]byte

	// FromPublicKey is the public key of the address interacting with the contract.
	FromPublicKey [48]byte

	// ByteCode is the valid smart contract compiled to bytecode form
	ByteCode []byte

	// InputData is the data passed to the VM as a param for executing the contract
	InputData []byte

	// Gas the amount of gas to be used when executing the contract
	Gas int64
}

// GetFromPublicKey returns the bls public key of the account executing the contract
func (c *Contract) GetFromPublicKey() (*bls.PublicKey, error) {
	return bls.PublicKeyFromBytes(c.FromPublicKey[:])
}

// GetContractPublicKey returns the bls public key of the contract.
func (c *Contract) GetContractPublicKey() (*bls.PublicKey, error) {
	return bls.PublicKeyFromBytes(c.PublicKey[:])
}

// GetFromAccountAddress returns the address of the account interacting with the contract
func (c *Contract) GetFromAccountAddress() ([20]byte, error) {

	pubKey, _ := c.GetFromPublicKey()
	return pubKey.Hash()
}

// GetContractAddress returns the address of the contract
func (c *Contract) GetContractAddress() ([20]byte, error) {

	pubKey, _ := c.GetContractPublicKey()
	return pubKey.Hash()
}

// GetContractGas returns the gas set by the from account for executing the contract
func (c *Contract) GetContractGas() int64 {
	return c.Gas
}

// GetContractByteCode returns the byte code for the contract to be executed
func (c *Contract) GetContractByteCode() []byte {
	return c.ByteCode
}

// GetContractInputData returns the input data for the contract to be executed
func (c *Contract) GetContractInputData() []byte {
	return c.InputData
}

// Marshal encodes the data.
func (c *Contract) Marshal() ([]byte, error) {
	return c.MarshalSSZ()
}

// Unmarshal decodes the data.
func (c *Contract) Unmarshal(b []byte) error {
	return c.UnmarshalSSZ(b)
}

// Hash calculates the hash of the contract
func (c *Contract) Hash() chainhash.Hash {
	b, _ := c.Marshal()
	return chainhash.HashH(b)
}
