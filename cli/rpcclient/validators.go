package rpcclient

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/olympus-protocol/ogen/chainrpc/proto"
)

func (c *RPCClient) GenerateValidatorKey() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	res, err := c.validators.GenerateValidatorKey(ctx, &proto.Empty{})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("PrivateKey: %v", res.GetKey()), nil
}

func (c *RPCClient) StartValidator(args []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if len(args) != 1 {
		return "", fmt.Errorf("Usage: startvalidator <privkey>")
	}
	privKey := args[0]
	if len(privKey) != 32 {
		return "", fmt.Errorf("expected private key to be 32 bytes long, but got %d", len(privKey))
	}
	in := &proto.StartValidatorInfo{
		PrivateKey: privKey,
	}
	res, err := c.validators.StartValidator(ctx, in)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("PublicKey: %v", res.GetPublicKey()), nil
}

func (c *RPCClient) ExitValidator(args []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if len(args) != 1 {
		return "", fmt.Errorf("Usage: exitvalidator <pubkey>")
	}
	pubKey := args[0]
	if len(pubKey) != 48 {
		return "", fmt.Errorf("expected public key to be 32 bytes long, but got %d", len(pubKey))
	}
	in := &proto.ExitValidatorInfo{
		PublicKey: pubKey,
	}
	res, err := c.validators.ExitValidator(ctx, in)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Success: %v", res.Success), nil
}

func (c *RPCClient) GetValidatorsList(args []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	res, err := c.validators.GetValidatorsList(ctx, &proto.Empty{})
	page := 1
	if len(args) == 1 {
		page, err = strconv.Atoi(args[0])
		if err != nil {
			return "", err
		}
	}

	numVals := 0
	validatorsPerPage := 32
	if page > len(res.Validators)/validatorsPerPage+1 {
		return "", fmt.Errorf("page %d is out of range (1 - %d)", page, len(res.Validators)/validatorsPerPage)
	}

	if page <= 0 {
		return "", fmt.Errorf("page %d is out of range (1 - %d)", page, len(res.Validators)/validatorsPerPage)
	}

	color.Magenta(" %-67s | %-10s | %-12s \n", "Public Key", "Balance", "Status")
	for _, v := range res.Validators[(page-1)*validatorsPerPage:] {
		fmt.Printf(" %-67s | %-10s | %-12s \n", v.PublicKey, v.Balance, v.Status)
		numVals++
		if numVals == validatorsPerPage {
			break
		}
	}
	return fmt.Sprintf("Page %d/%d, Showing validators %d-%d/%d", page, len(res.Validators)/validatorsPerPage+1, (page-1)*validatorsPerPage, page*validatorsPerPage, len(res.Validators)), nil
}
