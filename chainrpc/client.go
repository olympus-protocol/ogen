package chainrpc

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/rpc/json"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

type RPCClient struct {
	address string
}

func NewRPCClient(addr string) *RPCClient {
	return &RPCClient{
		address: addr,
	}
}

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

func (c *RPCClient) GetAddress() (string, error) {
	var addr string
	return addr, c.Call("Wallet.GetAddress", nil, &addr)
}

func (c *RPCClient) GetBalance() (uint64, error) {
	var bal uint64
	return bal, c.Call("Wallet.GetBalance", nil, &bal)
}

func (c *RPCClient) SendToAddress(to string, amount uint64, askpass func() ([]byte, error)) (*chainhash.Hash, error) {
	var out chainhash.Hash
	err := c.Call("Wallet.SendToAddress", &SendToAddressRequest{
		ToAddress: to,
		Amount:    amount,
		Password:  []byte{},
	}, &out)
	if err.Error() == "wallet locked, need authentication" {
		pass, err := askpass()
		if err != nil {
			return nil, err
		}
		err = c.Call("Wallet.SendToAddress", &SendToAddressRequest{
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
