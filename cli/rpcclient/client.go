package rpcclient

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/rpc/json"
	"github.com/olympus-protocol/ogen/chainrpc"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

// RPCClient represents an RPC connection to a server.
type RPCClient struct {
	address string
}

// NewRPCClient creates a new RPC client.
func NewRPCClient(addr string) *RPCClient {
	return &RPCClient{
		address: addr,
	}
}

// Call calls a method.
func (c *RPCClient) Call(method string, args interface{}, res interface{}) error {
	message, err := json.EncodeClientRequest(method, args)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", c.address, bytes.NewBuffer(message))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf(string(body))
	}

	err = json.DecodeClientResponse(resp.Body, &res)
	if err != nil {
		return err
	}

	return nil
}

// GetInfo gets the blockhash at current height
func (c *RPCClient) GetChainInfo() (*chainrpc.ChainInfoResponse, error) {
	var info chainrpc.ChainInfoResponse
	err := c.Call("Chain.GetChainInfo", Empty{}, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

// GetBlockHash gets the blockhash at current height
func (c *RPCClient) GetBlock(hash string) (string, error) {
	var block string
	err := c.Call("Chain.GetBlock", hash, &block)
	if err != nil {
		return "", err
	}
	return block, nil
}

// GetBlockHash gets the blockhash at current height
func (c *RPCClient) GetBlockHash(height uint64) (*chainhash.Hash, error) {
	var hash chainhash.Hash
	err := c.Call("Chain.GetBlockHash", height, &hash)
	if err != nil {
		return nil, err
	}
	return &hash, nil
}

// GetAddress gets the address of the wallet.
func (c *RPCClient) GetAddress() (string, error) {
	var addr string
	return addr, c.Call("Wallet.GetAddress", nil, &addr)
}

// GetBalance gets the balance of an address or the wallet address.
func (c *RPCClient) GetBalance(address string) (uint64, error) {
	var bal uint64
	return bal, c.Call("Wallet.GetBalance", &address, &bal)
}

// SendToAddress sends a transfer request to the RPC server.
func (c *RPCClient) SendToAddress(to string, amount uint64, askpass func() ([]byte, error)) (*chainhash.Hash, error) {
	var out chainhash.Hash
	err := c.Call("Wallet.SendToAddress", &chainrpc.SendToAddressRequest{
		ToAddress: to,
		Amount:    amount,
		Password:  []byte{},
	}, &out)
	if err.Error() == "wallet locked, need authentication" {
		pass, err := askpass()
		if err != nil {
			return nil, err
		}
		err = c.Call("Wallet.SendToAddress", &chainrpc.SendToAddressRequest{
			ToAddress: to,
			Amount:    amount,
			Password:  pass,
		}, &out)
		if err != nil {
			return nil, err
		}

		return &out, nil
	} else if err != nil {
		return nil, err
	}

	return &out, nil
}

// ListValidators lists validators managed by or owner by this wallet.
func (c *RPCClient) ListValidators() (*chainrpc.ValidatorListReponse, error) {
	var out chainrpc.ValidatorListReponse

	err := c.Call("Wallet.ListValidators", &Empty{}, &out)
	if err != nil {
		return nil, err
	}

	return &out, nil
}

// GenerateValidatorKey generates a validator key.
func (c *RPCClient) GenerateValidatorKey() (*chainrpc.ValidatorKeyResponse, error) {
	var out chainrpc.ValidatorKeyResponse

	err := c.Call("Wallet.GenerateValidatorKey", &Empty{}, &out)
	if err != nil {
		return nil, err
	}

	return &out, nil
}

// StartValidator starts a validator by signing a deposit.
func (c *RPCClient) StartValidator(privkey [32]byte, askpass func() ([]byte, error)) (*chainrpc.StartValidatorResponse, error) {
	deposit := new(chainrpc.StartValidatorResponse)

	err := c.Call("Wallet.StartValidator", &chainrpc.StartValidatorRequest{
		PrivateKey: privkey,
		Password:   []byte{},
	}, deposit)
	if err.Error() == "wallet locked, need authentication" {
		pass, err := askpass()
		if err != nil {
			return nil, err
		}
		err = c.Call("Wallet.StartValidator", &chainrpc.StartValidatorRequest{
			PrivateKey: privkey,
			Password:   pass,
		}, deposit)

		return deposit, err
	}

	return deposit, err
}

// ExitValidator exits a validator by signing a exit.
func (c *RPCClient) ExitValidator(pubkey [48]byte, askpass func() ([]byte, error)) (*chainrpc.ExitValidatorResponse, error) {
	exit := new(chainrpc.ExitValidatorResponse)

	err := c.Call("Wallet.ExitValidator", &chainrpc.ExitValidatorRequest{
		ValidatorPubKey: pubkey,
		Password:        []byte{},
	}, exit)
	if err.Error() == "wallet locked, need authentication" {
		pass, err := askpass()
		if err != nil {
			return nil, err
		}
		err = c.Call("Wallet.ExitValidator", &chainrpc.ExitValidatorRequest{
			ValidatorPubKey: pubkey,
			Password:        pass,
		}, exit)

		return exit, err
	}

	return exit, err
}
