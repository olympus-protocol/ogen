package config

import (
	"context"
	"encoding/hex"
	"github.com/olympus-protocol/ogen/cmd/ogen/initialization"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/bls/common"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/params"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	testdata "github.com/olympus-protocol/ogen/test"
	"os"
	"os/signal"
	"time"
)

type Flags struct {
	DataPath      string
	NetworkName   string
	Port          string
	RPCProxy      bool
	RPCProxyPort  string
	RPCProxyAddr  string
	RPCPort       string
	RPCWallet     bool
	RPCAuthToken  string
	Debug         bool
	LogFile       bool
	Dashboard     bool
	DashboardPort string
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

func SetTestParams() {
	GlobalParams = &Params{
		Logger:     logger.New(os.Stdin),
		NetParams:  &testdata.TestParams,
		InitParams: testInitParams(&testdata.TestParams),
		Context:    context.Background(),
	}
}

func SetTestFlags() {
	_ = os.MkdirAll("./test_data", 0700)
	GlobalFlags = &Flags{
		DataPath:     "test_data",
		NetworkName:  "test_network",
		Port:         "",
		RPCProxy:     false,
		RPCProxyPort: "",
		RPCProxyAddr: "",
		RPCPort:      "",
		RPCWallet:    false,
		RPCAuthToken: "",
		Debug:        false,
		LogFile:      false,
	}
}

func testInitParams(netParams *params.ChainParams) *initialization.InitializationParameters {
	var validatorsKeys []common.SecretKey
	var validators []*primitives.Validator

	addr := testdata.PremineAddr.PublicKey().ToAccount(&netParams.AccountPrefixes)

	addrHash, _ := testdata.PremineAddr.PublicKey().Hash()
	for i := 0; i < 100; i++ {
		key, _ := bls.RandKey()
		validatorsKeys = append(validatorsKeys, key)
		val := &primitives.Validator{
			Balance:          100 * 1e8,
			PayeeAddress:     addrHash,
			Status:           primitives.StatusActive,
			FirstActiveEpoch: 0,
			LastActiveEpoch:  0,
		}
		copy(val.PubKey[:], key.PublicKey().Marshal())
		validators = append(validators, val)
	}

	var initparams initialization.InitializationParameters

	initparams.GenesisTime = time.Unix(time.Now().Unix(), 0)
	initparams.InitialValidators = []initialization.ValidatorInitialization{}

	// Convert the validators to initialization params.
	for _, vk := range validatorsKeys {
		val := initialization.ValidatorInitialization{
			PubKey:       hex.EncodeToString(vk.PublicKey().Marshal()),
			PayeeAddress: addr,
		}
		initparams.InitialValidators = append(initparams.InitialValidators, val)
	}
	initparams.PremineAddress = addr
	return &initparams
}
