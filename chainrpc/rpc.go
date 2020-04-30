package chainrpc

import (
	"net"
	"net/http"
	"net/rpc"

	"github.com/olympus-protocol/ogen/logger"
)

type RPCServer struct {
	config *Config
	log    *logger.Logger
}

func NewRPCServer(config *Config) *RPCServer {
	return &RPCServer{
		config: config,
		log:    config.Log,
	}
}

type Empty struct{}

func (r *RPCServer) TestMethod(e *Empty, out *uint64) error {
	*out = 9
	return nil
}

type Config struct {
	Network string
	Address string
	Log     *logger.Logger
}

func ServeRPC(r *RPCServer) error {
	s := rpc.NewServer()
	if err := s.Register(r); err != nil {
		return err
	}

	l, err := net.Listen(r.config.Network, r.config.Address)
	if err != nil {
		return err
	}

	go http.Serve(l, s)

	return nil
}
