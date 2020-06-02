package rpcclient

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/olympus-protocol/ogen/chainrpc/proto"
)

func (c *RPCClient) getValidatorsList(args []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if len(args) > 0 {
		return "", errors.New("Usage: getvalidatorslist")
	}
	res, err := c.validators.GetValidatorsList(ctx, &proto.Empty{})
	if err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (c *RPCClient) startValidator(args []string) (string, error) {
	return "", nil
}

func (c *RPCClient) exitValidator(args []string) (string, error) {
	return "", nil
}

func (c *RPCClient) getAccountValidators(args []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if len(args) > 1 {
		return "", errors.New("Usage: getaccountvalidators <account>")
	}
	req := &proto.Account{Account: args[0]}
	res, err := c.validators.GetAccountValidators(ctx, req)
	if err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
