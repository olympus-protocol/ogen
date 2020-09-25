package commands

import (
	"encoding/json"
	"fmt"
	"github.com/olympus-protocol/ogen/internal/state"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"path"
	"time"

	"github.com/olympus-protocol/ogen/internal/blockdb"
	"github.com/olympus-protocol/ogen/internal/server"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/params"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
)

var GlobalDataFolder string

// loadOgen is the main function to run ogen.
func loadOgen(ctx context.Context, configParams *server.GlobalConfig, log logger.Logger, currParams *params.ChainParams) error {
	db, err := blockdb.NewBadgerDB(configParams.DataFolder, currParams, log)
	if err != nil {
		return err
	}
	s, err := server.NewServer(ctx, configParams, log, currParams, db, configParams.InitConfig)
	if err != nil {
		return err
	}
	go s.Start()
	<-ctx.Done()
	db.Close()
	err = s.Stop()
	if err != nil {
		return err
	}
	return nil
}

func getChainFile(path string, currParams *params.ChainParams) (*state.ChainFile, error) {
	chainFile := new(state.ChainFile)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		resp, err := http.Get(currParams.ChainFileURL)
		if err != nil {
			return nil, err
		}
		chainFileBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		chainFileBytesHash := chainhash.HashH(chainFileBytes)
		if !chainFileBytesHash.IsEqual(&currParams.ChainFileHash) {
			return nil, fmt.Errorf("chain file hash does not match (expected: %s, got: %s)", currParams.ChainFileHash.String(), chainFileBytesHash)
		}

		err = ioutil.WriteFile(path, chainFileBytes, 0644)
		if err != nil {
			return nil, fmt.Errorf("unable to write chain file")
		}

		err = json.Unmarshal(chainFileBytes, chainFile)
		if err != nil {
			return nil, err
		}
	} else {
		chainFileBytes, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(chainFileBytes, chainFile)
		if err != nil {
			return nil, err
		}
	}
	return chainFile, nil
}

var (
	rootCmd = &cobra.Command{
		Use:   "ogen",
		Short: "Ogen is a Go Olympus implementation",
		Long: `A Golang implementation of the Olympus protocol.
Next generation blockchain secured by CASPER.`,
		Run: func(cmd *cobra.Command, args []string) {
			var log logger.Logger

			if viper.GetBool("log_file") {
				logFile, err := os.OpenFile(path.Join(GlobalDataFolder, "logger.log"), os.O_CREATE|os.O_RDWR, 0755)
				if err != nil {
					panic(err)
				}
				log = logger.New(logFile)
			} else {
				log = logger.New(os.Stdin)
			}

			if viper.GetBool("debug") {
				log = log.WithDebug()
			}

			networkName := viper.GetString("network")

			var currParams *params.ChainParams
			switch networkName {
			case "mainnet":
				currParams = &params.Mainnet
			default:
				currParams = &params.TestNet
			}

			cf, err := getChainFile(path.Join(GlobalDataFolder, "chain.json"), currParams)
			if err != nil {
				log.Fatalf("could not load chainfile: %s", err)
			}

			ip := cf.ToInitializationParameters()
			genesisTime := viper.GetUint64("genesistime")
			if genesisTime != 0 {
				ip.GenesisTime = time.Unix(int64(genesisTime), 0)
			}

			rpcauth := ""
			if viper.GetString("rpc_auth_token") != "" {
				rpcauth = viper.GetString("rpc_auth_token")
			} else {
				rpcauth = randomAuthToken()
			}

			c := &server.GlobalConfig{
				DataFolder: GlobalDataFolder,

				NetworkName: networkName,
				Port:        viper.GetString("port"),

				InitConfig: ip,

				RPCProxy:     viper.GetBool("rpc_proxy"),
				RPCProxyPort: viper.GetString("rpc_proxy_port"),
				RPCProxyAddr: viper.GetString("rpc_proxy_addr"),
				RPCPort:      viper.GetString("rpc_port"),
				RPCWallet:    viper.GetBool("rpc_wallet"),
				RPCAuthToken: rpcauth,

				Debug: viper.GetBool("debug"),

				LogFile: viper.GetBool("log_file"),
				Pprof:   viper.GetBool("pprof"),
			}

			log.Infof("Starting Ogen v%v", params.Version)
			log.Trace("Loading log on debug mode")
			ctx, cancel := context.WithCancel(context.Background())

			InterruptListener(log, cancel)

			err = loadOgen(ctx, c, log, currParams)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&GlobalDataFolder, "datadir", "", "Directory to store the chain data.")

	rootCmd.Flags().String("network", "testnet", "String of the network to connect.")
	rootCmd.Flags().String("port", "24126", "Default port for p2p connections listener.")

	rootCmd.Flags().Bool("rpc_proxy", false, "Enable http proxy for RPC server.")
	rootCmd.Flags().String("rpc_proxy_port", "8080", "Port for the http proxy.")
	rootCmd.Flags().String("rpc_port", "24127", "RPC server port.")
	rootCmd.Flags().String("rpc_proxy_addr", "localhost", "RPC proxy address to serve the http server.")
	rootCmd.Flags().Bool("rpc_wallet", false, "Enable wallet access through RPC.")

	rootCmd.Flags().Uint64("genesistime", 0, "Overrides the genesis time on chain.json")
	rootCmd.PersistentFlags().Bool("debug", false, "Displays debug information.")
	rootCmd.PersistentFlags().Bool("log_file", false, "Display log information to file.")
	rootCmd.PersistentFlags().Bool("pprof", false, "Run ogen with a profiling server attached.")

	err := viper.BindPFlags(rootCmd.PersistentFlags())
	if err != nil {
		panic(err)
	}
}

func initConfig() {
	if GlobalDataFolder != "" {
		// Use config file from the flag.
		viper.AddConfigPath(GlobalDataFolder)
		viper.SetConfigName("config")
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

		GlobalDataFolder = ogenDir

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(ogenDir)
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	err := viper.BindPFlags(rootCmd.Flags())
	if err != nil {
		panic(err)
	}
}

func randomAuthToken() string {
	rand.Seed(time.Now().UnixNano())
	digits := "0123456789"
	specials := "~=+%^*/()[]{}/!@#$?|"
	all := "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		digits + specials
	length := 8
	buf := make([]byte, length)
	buf[0] = digits[rand.Intn(len(digits))]
	buf[1] = specials[rand.Intn(len(specials))]
	for i := 2; i < length; i++ {
		buf[i] = all[rand.Intn(len(all))]
	}
	rand.Shuffle(len(buf), func(i, j int) {
		buf[i], buf[j] = buf[j], buf[i]
	})
	return string(buf)
}

var shutdownRequestChannel = make(chan struct{})

var interruptSignals = []os.Signal{os.Interrupt}

func InterruptListener(log logger.Logger, cancel context.CancelFunc) {
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
