package commands

import (
	"fmt"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/internal/blockdb"
	"github.com/olympus-protocol/ogen/internal/server"
	"github.com/olympus-protocol/ogen/pkg/params"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path"
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
	rootCmd.PersistentFlags().BoolVar(&config.Debug, "debug", false, "Displays debug information.")
	rootCmd.PersistentFlags().BoolVar(&config.LogFile, "log_file", false, "Display log information to file.")

	rootCmd.Flags().StringVar(&config.NetworkName, "network", "testnet", "String of the network to connect.")
	rootCmd.Flags().StringVar(&config.Port, "port", "24126", "Default port for p2p connections listener.")

	rootCmd.Flags().BoolVar(&config.RPCProxy, "rpc_proxy", false, "Enable http proxy for RPC server.")
	rootCmd.Flags().StringVar(&config.RPCProxyPort, "rpc_proxy_port", "8080", "Port for the http proxy.")
	rootCmd.Flags().StringVar(&config.RPCPort, "rpc_port", "24127", "RPC server port.")
	rootCmd.Flags().StringVar(&config.RPCProxyAddr, "rpc_proxy_addr", "localhost", "RPC proxy address to serve the http server.")
	rootCmd.Flags().BoolVar(&config.RPCWallet, "rpc_wallet", false, "Enable wallet access through RPC.")

	err := viper.BindPFlags(rootCmd.PersistentFlags())
	if err != nil {
		panic(err)
	}

	config.Init()
}

func initConfig() {
	if viper.GetString("datadir") != "" {
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

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	err := viper.BindPFlags(rootCmd.Flags())
	if err != nil {
		panic(err)
	}
}
