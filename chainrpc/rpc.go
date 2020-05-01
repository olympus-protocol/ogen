package chainrpc

import (
	"net/http"

	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"

	"github.com/olympus-protocol/ogen/logger"
)

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
