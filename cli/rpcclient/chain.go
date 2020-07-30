package rpcclient

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/olympus-protocol/ogen/proto"
)

// GetChainInfo returns current chain information https://doc.oly.tech/documentation/rpc-interface/commands/chain#getchaininfo
func (c *RPCClient) GetChainInfo() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	res, err := c.chain.GetChainInfo(ctx, &proto.Empty{})
	if err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// GetRawBlock returns the requested block serialized https://doc.oly.tech/documentation/rpc-interface/commands/chain#getrawblock
func (c *RPCClient) GetRawBlock(args []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if len(args) < 1 {
		return "", errors.New("Usage: getrawblock <hash>")
	}
	req := &proto.Hash{
		Hash: args[0],
	}
	res, err := c.chain.GetRawBlock(ctx, req)
	if err != nil {
		return "", err
	}
	return res.GetRawBlock(), nil
}

// GetBlockHash returns current block hash https://doc.oly.tech/documentation/rpc-interface/commands/chain#getblockhash
func (c *RPCClient) GetBlockHash(args []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if len(args) < 1 {
		return "", errors.New("Usage: getblockhash <height>")
	}
	height, err := strconv.Atoi(args[0])
	if err != nil {
		return "", errors.New("unable to parse block height")
	}
	req := &proto.Number{
		Number: uint64(height),
	}
	res, err := c.chain.GetBlockHash(ctx, req)
	if err != nil {
		return "", err
	}
	return res.GetHash(), nil
}

// GetBlock returns current block in a human readable format https://doc.oly.tech/documentation/rpc-interface/commands/chain#getblock
func (c *RPCClient) GetBlock(args []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if len(args) < 1 {
		return "", errors.New("Usage: getblock <hash>")
	}
	req := &proto.Hash{
		Hash: args[0],
	}
	res, err := c.chain.GetBlock(ctx, req)
	if err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// GetAccountInfo returns the specified account information https://doc.oly.tech/documentation/rpc-interface/commands/chain#getaccountinfo
func (c *RPCClient) GetAccountInfo(args []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if len(args) < 1 {
		return "", errors.New("Usage: getaccountinfo <account>")
	}
	req := &proto.Account{
		Account: args[0],
	}
	res, err := c.chain.GetAccountInfo(ctx, req)
	if err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// GetTransaction returns the specified transaction on a human readable format https://doc.oly.tech/documentation/rpc-interface/commands/chain#gettransaction
func (c *RPCClient) GetTransaction(args []string) (string, error) {
	if len(args) < 1 {
		return "", errors.New("Usage: gettransaction <txid>")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	res, err := c.chain.GetTransaction(ctx, &proto.Hash{Hash: args[0]})
	if err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
