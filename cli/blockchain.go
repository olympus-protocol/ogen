package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/olympus-protocol/ogen/config"
	"github.com/olympus-protocol/ogen/db/blockdb"
	"github.com/olympus-protocol/ogen/logger"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/server"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
)

const (
	version = "0.1.0"
)

// loadOgen is the main function to run ogen.
func loadOgen(ctx context.Context, configParams *config.Config, log *logger.Logger, currParams params.ChainParams) error {
	db, err := blockdb.NewBlockDB(configParams.DataFolder, currParams, log)
	if err != nil {
		return err
	}
	s, err := server.NewServer(configParams, log, currParams, db, false, configParams.InitConfig)
	if err != nil {
		return err
	}
	go s.Start()
	<-ctx.Done()
	db.Close()
	err = s.Stop()
	if err != nil {
		panic(err)
	}
	return nil
}

func getChainFile(path string, currParams params.ChainParams) (*config.ChainFile, error) {
	chainFile := new(config.ChainFile)
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
			return nil, fmt.Errorf("chain file hash does not match (expected: %s, got: %s)", currParams.ChainFileHash, chainFileBytesHash)
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
	DataFolder string

	rootCmd = &cobra.Command{
		Use:   "ogen",
		Short: "Ogen is a Go Olympus implementation",
		Long: `A Golang implementation of the Olympus protocol.
Next generation blockchain secured by CASPER.`,
		Run: func(cmd *cobra.Command, args []string) {
			log := logger.New(os.Stdin)
			if viper.GetBool("debug") {
				log = log.WithDebug()
			}

			networkName := viper.GetString("network")

			var currParams params.ChainParams
			switch networkName {
			case "mainnet":
				currParams = params.Mainnet
			default:
				currParams = params.TestNet
			}

			cf, err := getChainFile(viper.GetString("chainfile"), currParams)
			if err != nil {
				log.Fatalf("could not load chainfile: %s", err)
			}

			ip := cf.ToInitializationParameters()
			genesisTime := viper.GetUint64("genesistime")
			if genesisTime != 0 {
				ip.GenesisTime = time.Unix(int64(genesisTime), 0)
			}

			c := &config.Config{
				DataFolder:    DataFolder,
				InitConfig:    ip,
				Debug:         viper.GetBool("debug"),
				Listen:        viper.GetBool("listen"),
				NetworkName:   networkName,
				ConnectNodes:  viper.GetStringSlice("connect"),
				AddNodes:      viper.GetStringSlice("add"),
				Port:          int32(viper.GetUint("port")),
				MaxPeers:      int32(viper.GetUint("maxpeers")),
				Mode:          viper.GetString("mode"),
				Wallet:        viper.GetBool("wallet"),
				MiningEnabled: viper.GetBool("enablemining"),
				RPCAddress:    viper.GetString("rpclisten"),
			}

			log.Infof("Starting Ogen v%v", config.OgenVersion())
			log.Trace("loading log on debug mode")
			ctx, cancel := context.WithCancel(context.Background())

			config.InterruptListener(log, cancel)

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

	rootCmd.PersistentFlags().StringVar(&DataFolder, "datadir", "", "data directory to store Ogen data")
	rootCmd.PersistentFlags().Bool("debug", false, "log debugging info")

	rootCmd.Flags().Bool("listen", true, "listen for new connections")
	rootCmd.Flags().String("network", "testnet", "network name to use (testnet or mainnet)")
	rootCmd.Flags().Uint16("port", 24126, "port to listen on for P2P connections")
	rootCmd.Flags().Uint16("maxpeers", 9, "maximum peers to connect to or allow connections for")
	rootCmd.Flags().String("mode", "node", "type of node to run")
	rootCmd.Flags().Bool("wallet", true, "enable wallet")
	rootCmd.Flags().StringSlice("connect", []string{}, "IP addresses of nodes to connect to initially")
	rootCmd.Flags().StringSlice("add", []string{}, "IP addresses of nodes to add")
	rootCmd.Flags().String("chainfile", "chain.json", "Chain file to use for blockchain initialization")
	rootCmd.Flags().Bool("enablemining", true, "should mining be enabled")
	rootCmd.Flags().Uint64("genesistime", 0, "genesis time override")
	rootCmd.Flags().String("rpclisten", ":24127", "RPC listen address")

	err := viper.BindPFlags(rootCmd.PersistentFlags())
	if err != nil {
		panic(err)
	}
}

func er(msg interface{}) {
	fmt.Println("Error:", msg)
	os.Exit(1)
}

func initConfig() {
	if DataFolder != "" {
		// Use config file from the flag.
		viper.AddConfigPath(DataFolder)
		viper.SetConfigName("config")
	} else {
		configDir, err := os.UserConfigDir()
		if err != nil {
			panic(err)
		}

		ogenDir := path.Join(configDir, "ogen")

		if _, err := os.Stat(ogenDir); os.IsNotExist(err) {
			err = os.Mkdir(ogenDir, 0744)
			if err != nil {
				panic(err)
			}
		}

		DataFolder = ogenDir

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
