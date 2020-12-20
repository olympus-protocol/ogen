package commands

import (
	"fmt"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/cmd/ogen/initialization"
	"github.com/olympus-protocol/ogen/internal/blockdb"
	"github.com/olympus-protocol/ogen/internal/server"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/params"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path"
	"time"
)

var (
	DataPath string
	NetName  string
	Port     string
	Debug    bool
	LogFile  bool

	RPCPort       string
	RPCWallet     bool
	RPCProxy      bool
	RPCProxyPort  string
	RPCPRoxyAddr  string
	Dashboard     bool
	DashboardPort string
)

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&DataPath, "datadir", "", "Directory to store the chain data.")

	rootCmd.Flags().StringVar(&NetName, "network", "testnet", "String of the network to connect.")
	rootCmd.Flags().StringVar(&Port, "port", "24126", "Default port for p2p connections listener.")

	rootCmd.Flags().BoolVar(&RPCProxy, "rpc_proxy", false, "Enable http proxy for RPC server.")
	rootCmd.Flags().StringVar(&RPCProxyPort, "rpc_proxy_port", "8081", "Port for the http proxy.")
	rootCmd.Flags().StringVar(&RPCPort, "rpc_port", "24127", "RPC server port.")
	rootCmd.Flags().StringVar(&RPCPRoxyAddr, "rpc_proxy_addr", "localhost", "RPC proxy address to serve the http server.")
	rootCmd.Flags().BoolVar(&RPCWallet, "rpc_wallet", false, "Enable wallet access through RPC.")

	rootCmd.Flags().StringVar(&DashboardPort, "dashboard_port", "8080", "Port to expose node dashboard.")
	rootCmd.Flags().BoolVar(&Dashboard, "dashboard", false, "Expose node dashboard.")

	rootCmd.PersistentFlags().BoolVar(&Debug, "debug", false, "Displays debug information.")
	rootCmd.PersistentFlags().BoolVar(&LogFile, "logfile", false, "Display log information to file.")

	err := viper.BindPFlags(rootCmd.PersistentFlags())
	if err != nil {
		panic(err)
	}

	err = viper.BindPFlags(rootCmd.Flags())
	if err != nil {
		panic(err)
	}
}

func initConfig() {
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

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	err := viper.BindPFlags(rootCmd.Flags())
	if err != nil {
		panic(err)
	}

	config.GlobalFlags = &config.Flags{
		DataPath:      DataPath,
		NetworkName:   NetName,
		Port:          Port,
		RPCProxy:      RPCProxy,
		RPCProxyPort:  RPCProxyPort,
		RPCProxyAddr:  RPCPRoxyAddr,
		RPCPort:       RPCPort,
		RPCWallet:     RPCWallet,
		Debug:         Debug,
		LogFile:       LogFile,
		DashboardPort: DashboardPort,
		Dashboard:     Dashboard,
	}

	var log logger.Logger

	if config.GlobalFlags.LogFile {
		logFile, err := os.OpenFile(path.Join(DataPath, "logger.log"), os.O_CREATE|os.O_RDWR, 0755)
		if err != nil {
			panic(err)
		}
		log = logger.New(logFile)
	} else {
		log = logger.New(os.Stdin)
	}

	if config.GlobalFlags.Debug {
		log = log.WithDebug()
	}
	var netParams *params.ChainParams
	switch config.GlobalFlags.NetworkName {
	case "mainnet":
		netParams = &params.Mainnet
	case "devnet":
		netParams = &params.DevNet
	default:
		netParams = &params.TestNet
	}

	initparams, err := initialization.LoadParams(config.GlobalFlags.NetworkName)
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

	config.GlobalParams = &config.Params{
		Logger:     log,
		NetParams:  netParams,
		InitParams: ip,
	}
}

var rootCmd = &cobra.Command{
	Use:   "ogen",
	Short: "Ogen is a Go Olympus implementation",
	Long:  "A Golang implementation of the Olympus protocol. Next generation blockchain secured by CASPER.",
	Run: func(cmd *cobra.Command, args []string) {
		log := config.GlobalParams.Logger

		log.Infof("Starting Ogen v%v", params.Version)
		log.Trace("Loading log on debug mode")

		config.InterruptListener()

		db, err := blockdb.NewLevelDB()
		if err != nil {
			log.Fatal(err)
		}

		s, err := server.NewServer(db)
		if err != nil {
			log.Fatal(err)
		}

		go s.Start()

		<-config.GlobalParams.Context.Done()

		err = s.Stop()
		if err != nil {
			log.Fatal(err)
		}
	},
}

func Execute() error {
	return rootCmd.Execute()
}
