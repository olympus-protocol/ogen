package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
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

	mnet "github.com/multiformats/go-multiaddr-net"
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

func tcpAddressStringToMultiaddr(addrString string) (multiaddr.Multiaddr, error) {
	netAddr, err := net.ResolveTCPAddr("tcp", addrString)
	if err != nil {
		return nil, err
	}

	return mnet.FromNetAddr(netAddr)
}

func tcpAddressesStringToMultiaddr(addrStrings []string) ([]multiaddr.Multiaddr, error) {
	out := make([]multiaddr.Multiaddr, len(addrStrings))
	for i := range out {
		o, err := tcpAddressStringToMultiaddr(addrStrings[i])
		if err != nil {
			return nil, err
		}
		out[i] = o
	}

	return out, nil
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

			listenAddr, err := tcpAddressStringToMultiaddr(viper.GetString("listen"))
			if err != nil {
				log.Fatalf("error parsing listen address: %s", err)
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
				DataFolder:    DataFolder,
				InitConfig:    ip,
				Debug:         viper.GetBool("debug"),
				Listen:        []multiaddr.Multiaddr{listenAddr},
				NetworkName:   networkName,
				AddNodes:      addNodes,
				Port:          int32(viper.GetUint("port")),
				MaxPeers:      int32(viper.GetUint("maxpeers")),
				MiningEnabled: viper.GetBool("enablemining"),
				RPCAddress:    viper.GetString("rpcaddress"),
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

	rootCmd.Flags().String("listen", "0.0.0.0:24126", "listen for new connections")
	rootCmd.Flags().String("network", "testnet", "network name to use (testnet or mainnet)")
	rootCmd.Flags().Uint16("maxpeers", 9, "maximum peers to connect to or allow connections for")
	rootCmd.Flags().StringSlice("connect", []string{}, "IP addresses of nodes to connect to initially")
	rootCmd.Flags().StringSlice("add", []string{}, "IP addresses of nodes to add")
	rootCmd.Flags().String("chainfile", "chain.json", "Chain file to use for blockchain initialization")
	rootCmd.Flags().Bool("enablemining", false, "should mining be enabled")
	rootCmd.Flags().Uint64("genesistime", 0, "genesis time override")
	rootCmd.Flags().String("rpcaddress", "127.0.0.1:24127", "RPC listen address")

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
