package commands

import (
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

// loadOgen is the main function to run ogen.
func loadOgen() error {

	db, err := blockdb.NewBadgerDB()
	if err != nil {
		return err
	}

	s, err := server.NewServer(db)
	if err != nil {
		return err
	}

	go s.Start()

	<-config.GlobalParams.Context.Done()

	db.Close()
	err = s.Stop()
	if err != nil {
		return err
	}
	return nil
}

var (
	rootCmd = &cobra.Command{
		Use:   "ogen",
		Short: "Ogen is a Go Olympus implementation",
		Long: `A Golang implementation of the Olympus protocol.
Next generation blockchain secured by CASPER.`,
		Run: func(cmd *cobra.Command, args []string) {
			log := config.GlobalParams.Logger

			log.Infof("Starting Ogen v%v", params.Version)
			log.Trace("Loading log on debug mode")

			config.InterruptListener()

			err := loadOgen()
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

	rootCmd.PersistentFlags().StringVar(&config.DataPath, "datadir", "", "Directory to store the chain data.")
	rootCmd.PersistentFlags().Bool("debug", false, "Displays debug information.")
	rootCmd.PersistentFlags().Bool("log_file", false, "Display log information to file.")

	rootCmd.Flags().String("network", "testnet", "String of the network to connect.")
	rootCmd.Flags().String("port", "24126", "Default port for p2p connections listener.")

	rootCmd.Flags().Bool("rpc_proxy", false, "Enable http proxy for RPC server.")
	rootCmd.Flags().String("rpc_proxy_port", "8080", "Port for the http proxy.")
	rootCmd.Flags().String("rpc_port", "24127", "RPC server port.")
	rootCmd.Flags().String("rpc_proxy_addr", "localhost", "RPC proxy address to serve the http server.")
	rootCmd.Flags().Bool("rpc_wallet", false, "Enable wallet access through RPC.")

	err := viper.BindPFlags(rootCmd.PersistentFlags())
	if err != nil {
		panic(err)
	}

	config.GlobalFlags = &config.Flags{
		DataPath:     config.DataPath,
		NetworkName:  viper.GetString("network"),
		Port:         viper.GetString("port"),
		RPCProxy:     viper.GetBool("rpc_proxy"),
		RPCProxyPort: viper.GetString("rpc_proxy_port"),
		RPCProxyAddr: viper.GetString("rpc_proxy_addr"),
		RPCPort:      viper.GetString("rpc_port"),
		RPCWallet:    viper.GetBool("rpc_wallet"),
		Debug:        viper.GetBool("debug"),
		LogFile:      viper.GetBool("log_file"),
	}

	var log logger.Logger

	if config.GlobalFlags.LogFile {
		logFile, err := os.OpenFile(path.Join(config.GlobalFlags.DataPath, "logger.log"), os.O_CREATE|os.O_RDWR, 0755)
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

func initConfig() {
	if config.DataPath != "" {
		// Use config file from the flag.
		viper.AddConfigPath(config.DataPath)
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

		config.DataPath = ogenDir

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(ogenDir)
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()

	err := viper.BindPFlags(rootCmd.Flags())
	if err != nil {
		panic(err)
	}
}
