package rpcclient

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/olympus-protocol/ogen/api/proto"
	"github.com/olympus-protocol/ogen/pkg/bls"
)

func (c *RPCClient) SubmitRawData(args []string) (string, error) {
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

func (c *RPCClient) GenKeyPair(raw bool) (string, error) {
	blsKeyPair := bls.RandKey()

	var res *bls.KeyPair
	if raw {
		res = &bls.KeyPair{
			Public:  hex.EncodeToString(blsKeyPair.PublicKey().Marshal()),
			Private: hex.EncodeToString(blsKeyPair.Marshal()),
		}
	} else {
		res = &bls.KeyPair{
			Public:  blsKeyPair.PublicKey().ToAccount(),
			Private: blsKeyPair.ToWIF(),
		}
	}

	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (c *RPCClient) GenValidatorKey(args []string) (out string, err error) {
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
	res, err := c.utils.GenValidatorKey(ctx, req)
	if err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (c *RPCClient) DecodeRawTransaction(args []string) (string, error) {
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

func (c *RPCClient) DecodeRawBlock(args []string) (string, error) {
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
