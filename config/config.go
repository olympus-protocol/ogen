package config

import (
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/olympus-protocol/ogen/primitives"
)

const (
	version = "0.0.1"
)

type Config struct {
	DataFolder string

	NetworkName string
	AddNodes    []peer.AddrInfo
	MaxPeers    int32
	Port        string

	MiningEnabled bool

	InitConfig primitives.InitializationParameters

	RPCProxy     bool
	RPCProxyPort string
	RPCPort      string
	RPCWallet    bool

	Debug bool
}

func OgenVersion() string {
	return version
}
