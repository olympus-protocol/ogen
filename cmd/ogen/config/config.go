package config

import (
	"context"
	"github.com/olympus-protocol/ogen/cmd/ogen/initialization"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/params"
	"os"
	"os/signal"
)

var (
	DataPath string
)

type Flags struct {
	DataPath     string
	NetworkName  string
	Port         string
	RPCProxy     bool
	RPCProxyPort string
	RPCProxyAddr string
	RPCPort      string
	RPCWallet    bool
	RPCAuthToken string
	Debug        bool
	LogFile      bool
}

type Params struct {
	Logger     logger.Logger
	NetParams  *params.ChainParams
	InitParams *initialization.InitializationParameters
	Context    context.Context
}

var GlobalFlags *Flags

var GlobalParams *Params

var shutdownRequestChannel = make(chan struct{})

var interruptSignals = []os.Signal{os.Interrupt}

func InterruptListener() {
	ctx, cancel := context.WithCancel(context.Background())
	GlobalParams.Context = ctx
	go func() {
		interruptChannel := make(chan os.Signal, 1)
		signal.Notify(interruptChannel, interruptSignals...)
		select {
		case sig := <-interruptChannel:
			GlobalParams.Logger.Warnf("Received signal (%s).  Shutting down...",
				sig)
		case <-shutdownRequestChannel:
			GlobalParams.Logger.Warn("Shutdown requested.  Shutting down...")
		}
		cancel()
		for {
			select {
			case sig := <-interruptChannel:
				GlobalParams.Logger.Warnf("Received signal (%s).  Already "+
					"shutting down...", sig)

			case <-shutdownRequestChannel:
				GlobalParams.Logger.Warn("Shutdown requested.  Already " +
					"shutting down...")
			}
		}
	}()
}
