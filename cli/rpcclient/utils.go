package rpcclient

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/chainrpc/proto"
)

func (c *RPCClient) sendRawTransaction(args []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if len(args) > 1 || len(args) < 1 {
		return "", errors.New("Usage: sendrawtransaction <raw_transaction>")
	}
	req := &proto.RawData{Data: args[0]}
	res, err := c.utils.SendRawTransaction(ctx, req)
	if err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (c *RPCClient) genKeyPair(args []string, raw bool) (string, error) {
	if len(args) > 0 {
		return "", errors.New("Usage: genkeypair || genrawkeypair")
	}
	blsKeyPair := bls.RandKey()
	var res bls.KeyPair
	if raw {
		res = bls.KeyPair{
			Public:  hex.EncodeToString(blsKeyPair.PublicKey().Marshal()),
			Private: hex.EncodeToString(blsKeyPair.Marshal()),
		}
	} else {
		addr, err := blsKeyPair.PublicKey().ToAddress("olpub")
		if err != nil {
			return "", errors.New("unable to encode public key to bech32")
		}
		wif, err := blsKeyPair.ToWIF("olprv")
		if err != nil {
			return "", errors.New("unable to encode private key to bech32")
		}
		res = bls.KeyPair{
			Public:  addr,
			Private: wif,
		}
	}

	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (c *RPCClient) genValidatorKey(args []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if len(args) > 0 {
		return "", errors.New("Usage: genvalidatorkey")
	}
	res, err := c.utils.GenValidatorKey(ctx, &proto.Empty{})
	if err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (c *RPCClient) decodeRawTransaction(args []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if len(args) > 1 || len(args) < 1 {
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

func (c *RPCClient) decodeRawBlock(args []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if len(args) > 1 || len(args) < 1 {
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
