package chainrpc

import (
	"net/http"

	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"

	"github.com/olympus-protocol/ogen/utils/logger"
)

type Config struct {
	Network string
	Address string
	Log     *logger.Logger
}

type RPCServices struct {
	Wallet *Wallet
	Chain  *Chain
}

func ServeRPC(services *RPCServices, config Config) {
	s := rpc.NewServer()
	s.RegisterCodec(json.NewCodec(), "application/json")
	s.RegisterService(services.Wallet, "Wallet")
	s.RegisterService(services.Chain, "Chain")
	go http.ListenAndServe(config.Address, s)
}
