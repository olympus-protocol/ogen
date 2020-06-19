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
	version = "0.0.1"
)

type Config struct {
	DataFolder    string
	Debug         bool
	Listen        []multiaddr.Multiaddr
	NetworkName   string
	AddNodes      []peer.AddrInfo
	Port          int32
	MaxPeers      int32
	MiningEnabled bool
	InitConfig    primitives.InitializationParameters
	Wallet        bool
	RPCProxy      bool
	RPCPort       string
}

func OgenVersion() string {
	return version
}
