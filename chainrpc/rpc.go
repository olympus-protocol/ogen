package chainrpc

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"

	"github.com/olympus-protocol/ogen/logger"
	"github.com/olympus-protocol/ogen/wallet"
)

type Wallet struct {
	config *Config
	log    *logger.Logger
	wallet *wallet.Wallet
}

func NewRPCWallet(wallet *wallet.Wallet) *Wallet {
	return &Wallet{
		wallet: wallet,
	}
}

type Empty struct{}

func (r *Wallet) GetAddress(req *http.Request, args *interface{}, reply *string) error {
	*reply = r.wallet.GetAddress()
	return nil
}

type Config struct {
	Network string
	Address string
	Log     *logger.Logger
}

func ServeRPC(w *Wallet, config Config) {
	s := rpc.NewServer()
	s.RegisterCodec(json.NewCodec(), "application/json")
	s.RegisterService(w, "Wallet")
	go http.ListenAndServe(config.Address, s)
}

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
