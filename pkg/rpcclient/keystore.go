package rpcclient

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/olympus-protocol/ogen/api/proto"
)

func (c *Client) GenerateKeys(args []string) (out string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	if len(args) < 1 {
		return "", errors.New("Usage: genvalidatorkey <amount>")
	}

	amount, err := strconv.Atoi(args[0])
	if err != nil {
		return out, err
	}

	req := &proto.Number{
		Number: uint64(amount),
	}

	res, err := c.keystore.GenerateKeys(ctx, req)
	if err != nil {
		return "", err
	}

	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (c *Client) GetMnemonic(args []string) (out string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	if len(args) > 0 {
		return "", errors.New("Usage: getmnemonic")
	}

	res, err := c.keystore.GetMnemonic(ctx, &proto.Empty{})
	if err != nil {
		return "", err
	}

	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (c *Client) GetKey(args []string) (out string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	if len(args) < 1 {
		return "", errors.New("Usage: getkeystorekey <public_key>")
	}

	in := &proto.PublicKey{Key: args[1]}

	res, err := c.keystore.GetKey(ctx, in)
	if err != nil {
		return "", err
	}

	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (c *Client) GetKeys(args []string) (out string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	if len(args) > 0 {
		return "", errors.New("Usage: getkeystorekeys")
	}

	res, err := c.keystore.GetKeys(ctx, &proto.Empty{})
	if err != nil {
		return "", err
	}

	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (c *Client) ToggleKeys(args []string) (out string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	if len(args) < 2 {
		return "", errors.New("Usage: togglekey <public_key> <true/false>")
	}

	enabled, err := strconv.ParseBool(args[2])
	if err != nil {
		return "", err
	}

	in := &proto.ToggleKeyMsg{PublicKey: args[1], Enabled: enabled}

	res, err := c.keystore.ToggleKey(ctx, in)
	if err != nil {
		return "", err
	}

	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", err
	}

	return string(b), nil
}
