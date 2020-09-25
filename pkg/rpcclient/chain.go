package rpcclient

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/olympus-protocol/ogen/api/proto"
)

func (c *Client) GetChainInfo() (string, error) {
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

func (c *Client) GetRawBlock(args []string) (string, error) {
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

func (c *Client) GetBlockHash(args []string) (string, error) {
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

func (c *Client) GetBlock(args []string) (string, error) {
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

func (c *Client) GetAccountInfo(args []string) (string, error) {
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
