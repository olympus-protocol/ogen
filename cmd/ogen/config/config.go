package config

import (
	"context"
	"github.com/olympus-protocol/ogen/cmd/ogen/initialization"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/params"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"path"
	"time"
)

var (
	// Persistent
	DataPath string
	Debug    bool
	LogFile  bool

	NetworkName  string
	Port         string
	RPCProxy     bool
	RPCProxyPort string
	RPCProxyAddr string
	RPCPort      string
	RPCWallet    bool
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

func Init() {
	setDataPath()

	GlobalFlags = &Flags{
		DataPath:     DataPath,
		NetworkName:  NetworkName,
		Port:         Port,
		RPCProxy:     RPCProxy,
		RPCProxyPort: RPCProxyPort,
		RPCProxyAddr: RPCProxyAddr,
		RPCPort:      RPCPort,
		RPCWallet:    RPCWallet,
		Debug:        Debug,
		LogFile:      LogFile,
	}

	var log logger.Logger

	if GlobalFlags.LogFile {
		logFile, err := os.OpenFile(path.Join(GlobalFlags.DataPath, "logger.log"), os.O_CREATE|os.O_RDWR, 0755)
		if err != nil {
			panic(err)
		}
		log = logger.New(logFile)
	} else {
		log = logger.New(os.Stdin)
	}

	if GlobalFlags.Debug {
		log = log.WithDebug()
	}
	var netParams *params.ChainParams
	switch GlobalFlags.NetworkName {
	case "mainnet":
		netParams = &params.Mainnet
	default:
		netParams = &params.TestNet
	}

	initparams, err := initialization.LoadParams(GlobalFlags.NetworkName)
	if err != nil {
		log.Error("no params specified for that network")
		panic(err)
	}

	initialValidators := make([]initialization.ValidatorInitialization, len(initparams.Validators))
	for i := range initialValidators {
		v := initialization.ValidatorInitialization{
			PubKey:       initparams.Validators[i].PublicKey,
			PayeeAddress: initparams.PremineAddress,
		}
		initialValidators[i] = v
	}

	var genesisTime time.Time
	if initparams.GenesisTime == 0 {
		genesisTime = time.Now()
	} else {
		genesisTime = time.Unix(initparams.GenesisTime, 0)
	}

	ip := &initialization.InitializationParameters{
		GenesisTime:       genesisTime,
		InitialValidators: initialValidators,
		PremineAddress:    initparams.PremineAddress,
	}

	GlobalParams = &Params{
		Logger:     log,
		NetParams:  netParams,
		InitParams: ip,
		Context:    context.Background(),
	}

}

func setDataPath() {
	if DataPath != "" {
		// Use config file from the flag.
		viper.AddConfigPath(DataPath)
		viper.SetConfigName("config")
		if _, err := os.Stat(DataPath); os.IsNotExist(err) {
			err = os.MkdirAll(DataPath, 0744)
			if err != nil {
				panic(err)
			}
		}
	} else {
		configDir, err := os.UserConfigDir()
		if err != nil {
			panic(err)
		}

		ogenDir := path.Join(configDir, "ogen")

		if _, err := os.Stat(ogenDir); os.IsNotExist(err) {
			err = os.MkdirAll(ogenDir, 0744)
			if err != nil {
				panic(err)
			}
		}

		DataPath = ogenDir

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(ogenDir)
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()
}

var shutdownRequestChannel = make(chan struct{})

var interruptSignals = []os.Signal{os.Interrupt}

func InterruptListener() {
	log := GlobalParams.Logger
	_, cancel := context.WithCancel(GlobalParams.Context)
	go func() {
		interruptChannel := make(chan os.Signal, 1)
		signal.Notify(interruptChannel, interruptSignals...)
		select {
		case sig := <-interruptChannel:
			log.Warnf("Received signal (%s).  Shutting down...",
				sig)
		case <-shutdownRequestChannel:
			log.Warn("Shutdown requested.  Shutting down...")
		}
		cancel()
		for {
			select {
			case sig := <-interruptChannel:
				log.Warnf("Received signal (%s).  Already "+
					"shutting down...", sig)

			case <-shutdownRequestChannel:
				log.Warn("Shutdown requested.  Already " +
					"shutting down...")
			}
		}
	}()
}
