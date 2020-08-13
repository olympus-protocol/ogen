package config

import (
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

const (
	OgenVersion = "0.0.1"
)

type Config struct {
	DataFolder string

	NetworkName  string
	InitialNodes []peer.AddrInfo
	Port         string

	InitConfig primitives.InitializationParameters

	RPCProxy     bool
	RPCProxyPort string
	RPCProxyAddr string
	RPCPort      string
	RPCWallet    bool
	RPCAuthToken string

	Debug   bool
	LogFile bool
	Pprof   bool
}
