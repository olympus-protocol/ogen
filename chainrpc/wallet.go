package chainrpc

import (
	"net/http"

	"github.com/olympus-protocol/ogen/logger"
	"github.com/olympus-protocol/ogen/utils/chainhash"
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

func (w *Wallet) GetBalance(req *http.Request, args *string, reply *uint64) error {
	bal, err := w.wallet.GetBalance(*args)
	if err != nil {
		return err
	}

	*reply = bal
	return nil
}

type SendToAddressRequest struct {
	Password  []byte
	ToAddress string
	Amount    uint64
}

func (w *Wallet) SendToAddress(req *http.Request, args *SendToAddressRequest, reply *chainhash.Hash) error {
	reply, err := w.wallet.SendToAddress(args.Password, args.ToAddress, args.Amount)
	return err
}
