package rpcclient

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/olympus-protocol/ogen/proto"
)

func (c *RPCClient) listWallets(args []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if len(args) > 0 {
		return "", errors.New("Usage: listwallets")
	}
	res, err := c.wallet.ListWallets(ctx, &proto.Empty{})
	if err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (c *RPCClient) createWallet(args []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if len(args) > 1 || len(args) < 1 {
		return "", errors.New("Usage: createwallet <name>")
	}
	res, err := c.wallet.CreateWallet(ctx, &proto.Name{Name: args[0]})
	if err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (c *RPCClient) openWallet(args []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if len(args) > 1 || len(args) < 1 {
		return "", errors.New("Usage: openwallet <name>")
	}
	res, err := c.wallet.OpenWallet(ctx, &proto.Name{Name: args[0]})
	if err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (c *RPCClient) closeWallet(args []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if len(args) > 0 {
		return "", errors.New("Usage: closewallet")
	}
	res, err := c.wallet.CloseWallet(ctx, &proto.Empty{})
	if err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (c *RPCClient) getBalance(args []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if len(args) > 0 {
		return "", errors.New("Usage: getbalance")
	}
	res, err := c.wallet.GetBalance(ctx, &proto.Empty{})
	if err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (c *RPCClient) getAccount(args []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if len(args) > 0 {
		return "", errors.New("Usage: getaccount")
	}
	res, err := c.wallet.GetAccount(ctx, &proto.Empty{})
	if err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (c *RPCClient) sendTransaction(args []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if len(args) > 2 || len(args) < 2 {
		return "", errors.New("Usage: sendtransaction <account> <amount>")
	}
	res, err := c.wallet.SendTransaction(ctx, &proto.SendTransactionInfo{Account: args[0], Amount: args[1]})
	if err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (c *RPCClient) exitValidator(args []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if len(args) > 1 || len(args) < 1 {
		return "", errors.New("Usage: exitvalidator <pub_key>")
	}
	res, err := c.wallet.ExitValidator(ctx, &proto.KeyPair{Public: args[0]})
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if len(args) > 1 || len(args) < 1 {
		return "", errors.New("Usage: startvalidator <priv_key>")
	}
	res, err := c.wallet.StartValidator(ctx, &proto.KeyPair{Private: args[0]})
	if err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}