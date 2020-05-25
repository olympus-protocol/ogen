package rpcclient

import (
	"fmt"
	"strconv"
)

// GetBlock returns block information
func (c *CLI) GetChainInfo() (string, error) {
	info, err := c.rpcClient.GetChainInfo()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Blocks: %v, LastBlockHash: %v. Validators: %v", info.Blocks, info.LastBlockHash, info.Validators), nil
}

// GetBlock returns block information
func (c *CLI) GetBlock(args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("Usage: getblock <hash>")
	}
	block, err := c.rpcClient.GetBlock(args[0])
	if err != nil {
		return "", err
	}
	return block, nil
}

// GetBlockHash returns the blockhash at specified height
func (c *CLI) GetBlockHash(args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("Usage: getblockhash <height>")
	}
	n, err := strconv.Atoi(args[0])
	if err != nil {
		return "", err
	}
	blockHash, err := c.rpcClient.GetBlockHash(uint64(n))
	if err != nil {
		return "", err
	}
	return blockHash.String(), nil
}
