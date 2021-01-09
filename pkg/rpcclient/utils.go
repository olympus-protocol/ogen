package rpcclient

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/olympus-protocol/ogen/pkg/params"
	"strconv"
	"time"

	"github.com/olympus-protocol/ogen/api/proto"
	"github.com/olympus-protocol/ogen/pkg/bls"
)

func (c *Client) SubmitRawData(args []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if len(args) < 2 {
		return "", errors.New("Usage: submitrawdata <raw_data> <type>")
	}
	req := &proto.RawData{Data: args[0], Type: args[1]}
	res, err := c.utils.SubmitRawData(ctx, req)
	if err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (c *Client) GenKeyPair(args []string, raw bool) (string, error) {
	blsKeyPair, err := bls.RandKey()
	if err != nil {
		return "", err
	}

	netName := args[0]
	var netParams *params.ChainParams
	switch netName {
	case "testnet":
		netParams = &params.TestNet
	case "mainnet":
		netParams = &params.MainNet
	default:
		return "", errors.New("no params for " + netName)
	}

	if !raw {
		if len(args) < 1 {
			return "", errors.New("Usage: genkeypair <network>")
		}

		bls.Initialize(netParams, "blst")
	}

	var res *bls.KeyPair
	if raw {
		res = &bls.KeyPair{
			Public:  hex.EncodeToString(blsKeyPair.PublicKey().Marshal()),
			Private: hex.EncodeToString(blsKeyPair.Marshal()),
		}
	} else {
		res = &bls.KeyPair{
			Public:  blsKeyPair.PublicKey().ToAccount(&netParams.AccountPrefixes),
			Private: blsKeyPair.ToWIF(&netParams.AccountPrefixes),
		}
	}

	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (c *Client) GenValidatorKey(args []string) (out string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	amount := 0
	if len(args) < 1 {
		return "", errors.New("Usage: genvalidatorkey <keys>")
	}
	amount, err = strconv.Atoi(args[0])
	if err != nil {
		return out, err
	}
	req := &proto.GenValidatorKeys{
		Keys: uint64(amount),
	}
	res, err := c.keystore.GenValidatorKey(ctx, req)
	if err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (c *Client) DecodeRawTransaction(args []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if len(args) < 1 {
		return "", errors.New("Usage: decoderawtransaction <raw_transaction>")
	}
	req := &proto.RawData{Data: args[0]}
	res, err := c.utils.DecodeRawTransaction(ctx, req)
	if err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (c *Client) DecodeRawBlock(args []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if len(args) < 1 {
		return "", errors.New("Usage: decoderawblock <raw_block>")
	}
	req := &proto.RawData{Data: args[0]}
	res, err := c.utils.DecodeRawBlock(ctx, req)
	if err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
