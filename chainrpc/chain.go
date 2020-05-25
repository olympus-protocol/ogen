package chainrpc

import (
	"bytes"
	"encoding/hex"
	"errors"
	"net/http"

	"github.com/olympus-protocol/ogen/chain"
	"github.com/olympus-protocol/ogen/logger"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

// Chain is the chain RPC.
type Chain struct {
	config *Config
	log    *logger.Logger

	chain *chain.Blockchain
}

// NewRPCChain constructs an RPC chain.
func NewRPCChain(ch *chain.Blockchain) *Chain {
	return &Chain{
		chain: ch,
	}
}

type ChainInfoResponse struct {
	Blocks        uint64 `json: "blocks"`
	LastBlockHash string `json: "last_block_hash"`
	Validators    uint64 `json: "validators"`
}

// GetChainInfo sends money to an address.
func (c *Chain) GetChainInfo(req *http.Request, args *interface{}, reply *ChainInfoResponse) error {
	state := c.chain.State()
	*reply = ChainInfoResponse{
		Blocks:        state.Tip().Height,
		LastBlockHash: state.Tip().Hash.String(),
		Validators:    uint64(len(state.TipState().ValidatorRegistry)),
	}
	return nil
}

// GetBlockHash returns the hash of the specified block height.
func (c *Chain) GetBlockHash(req *http.Request, args *uint64, reply *chainhash.Hash) error {
	blockRow, exists := c.chain.State().Chain().GetNodeByHeight(*args)
	if !exists {
		return errors.New("block not found")
	}
	*reply = blockRow.Hash
	return nil
}

// GetBlock returns the raw block for a specified hash.
func (c *Chain) GetBlock(req *http.Request, args *string, reply *string) error {
	hash, err := chainhash.NewHashFromStr(*args)
	if err != nil {
		return err
	}
	block, err := c.chain.GetBlock(*hash)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer([]byte{})
	block.Encode(buf)
	*reply = hex.EncodeToString(buf.Bytes())
	return err
}
