package rpcclient

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/olympus-protocol/ogen/chainrpc/proto"
)

func (c *RPCClient) GetChainInfo() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	res, err := c.chain.GetChainInfo(ctx, &proto.Empty{})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("LastBlock: %v, LastBlockHash: %v, Validators: %v", res.GetBlockHeight(), res.GetBlockHash(), res.GetValidators()), nil
}

func (c *RPCClient) GetBlockHash(args []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if len(args) > 1 {
		return "", errors.New("too many arguments")
	}
	height, err := strconv.Atoi(args[0])
	if err != nil {
		return "", errors.New("unable to parse block height")
	}
	req := &proto.GetBlockHashInfo{
		BlockHeigth: uint64(height),
	}
	res, err := c.chain.GetBlockHash(ctx, req)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Hash: %v", res.GetBlockHash()), nil
}

func (c *RPCClient) GetRawBlock(args []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if len(args) > 1 {
		return "", errors.New("too many arguments")
	}
	req := &proto.GetBlockInfo{
		BlockHash: args[0],
	}
	res, err := c.chain.GetRawBlock(ctx, req)
	if err != nil {
		return "", err
	}
	return res.GetRawBlock(), nil
}

func (c *RPCClient) GetBlock(args []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if len(args) > 1 {
		return "", errors.New("too many arguments")
	}
	req := &proto.GetBlockInfo{
		BlockHash: args[0],
	}
	res, err := c.chain.GetRawBlock(ctx, req)
	if err != nil {
		return "", err
	}
	return res.GetRawBlock(), nil
}
