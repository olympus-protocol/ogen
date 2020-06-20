package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/olympus-protocol/ogen/bdb"
	"github.com/olympus-protocol/ogen/config"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/server"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
)

// loadOgen is the main function to run ogen.
func loadOgen(ctx context.Context, configParams *config.Config, log *logger.Logger, currParams params.ChainParams) error {
	db, err := bdb.NewBlockDB(configParams.DataFolder, currParams, log)
	if err != nil {
		return err
	}
	s, err := server.NewServer(ctx, configParams, log, currParams, db, false, configParams.InitConfig)
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

		ioutil.WriteFile(path, chainFileBytes, 0644)

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

			cf, err := getChainFile(DataFolder+"/chain.json", currParams)
			if err != nil {
				log.Fatalf("could not load chainfile: %s", err)
			}

			ip := cf.ToInitializationParameters()
			genesisTime := viper.GetUint64("genesistime")
			if genesisTime != 0 {
				ip.GenesisTime = time.Unix(int64(genesisTime), 0)
			}

			addNodesStrs := viper.GetStringSlice("add")
			addNodesStrs = append(addNodesStrs, cf.InitialConnections...)
			addNodes := make([]peer.AddrInfo, len(addNodesStrs))
			for i := range addNodes {
				maddr, err := multiaddr.NewMultiaddr(addNodesStrs[i])
				if err != nil {
					log.Fatalf("error parsing add node %s: %s", addNodesStrs[i], err)
				}
				pinfo, err := peer.AddrInfoFromP2pAddr(maddr)
				if err != nil {
					log.Fatalf("error parsing add node %s: %s", maddr, pinfo)
				}

				addNodes[i] = *pinfo
			}
			c := &config.Config{
				DataFolder: DataFolder,

				NetworkName: networkName,
				AddNodes:    addNodes,
				MaxPeers:    int32(viper.GetUint("maxpeers")),
				Port:        viper.GetString("port"),

				MiningEnabled: viper.GetBool("enablemining"),

				InitConfig: ip,

				RPCProxy:     viper.GetBool("rpc_proxy"),
				RPCProxyPort: viper.GetString("rpc_proxy_port"),
				RPCPort:      viper.GetString("rpc_port"),
				RPCWallet:    viper.GetBool("rpc_wallet"),

				Debug: viper.GetBool("debug"),
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

	rootCmd.Flags().String("network", "testnet", "network name to use (testnet or mainnet)")
	rootCmd.Flags().StringSlice("add", []string{}, "IP addresses of nodes to add")
	rootCmd.Flags().Uint16("maxpeers", 9, "maximum peers to connect to or allow connections for")
	rootCmd.Flags().String("port", "24126", "port to listen to p2p connections")

	rootCmd.Flags().Bool("enablemining", false, "should mining be enabled")

	rootCmd.Flags().Bool("rpc_proxy", false, "enable http proxy for rpc")
	rootCmd.Flags().String("rpc_proxy_port", "8080", "port to listen for the http proxy")
	rootCmd.Flags().String("rpc_port", "24127", "host/port to listen on for rpc")
	rootCmd.Flags().Bool("rpc_wallet", false, "enable wallet access through rpc")

	rootCmd.Flags().Uint64("genesistime", 0, "genesis time override")
	rootCmd.PersistentFlags().Bool("debug", false, "log debugging info")

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
