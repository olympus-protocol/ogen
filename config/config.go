package config

import (
	"errors"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/olympus-protocol/ogen/primitives"
)

var (
	ErrorPathDontExist   = errors.New("the specified path for datadir doesn't exists")
	ErrorConfigDontExist = errors.New("unable to load config.toml from datadir")
)

const (
	version  = "1.0.0"
)

type Config struct {
	DataFolder    string
	RPCAddress    string
	Debug         bool
	Listen        []multiaddr.Multiaddr
	NetworkName   string
	AddNodes      []peer.AddrInfo
	Port          int32
	MaxPeers      int32
	MiningEnabled bool
	InitConfig    primitives.InitializationParameters
}

func OgenVersion() string {
	return version
}
