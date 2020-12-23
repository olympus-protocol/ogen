package main

import (
	"context"
	"fmt"
	"github.com/leaanthony/mewn"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
	"github.com/olympus-protocol/ogen/cmd/ogen/initialization"
	"github.com/olympus-protocol/ogen/internal/blockdb"
	"github.com/olympus-protocol/ogen/internal/server"
	"github.com/olympus-protocol/ogen/pkg/logger"
	"github.com/olympus-protocol/ogen/pkg/params"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wailsapp/wails"
	"os"
	"path"
	"time"
)

var (
	DataPath string
	NetName  string
	Port     string
)

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&DataPath, "datadir", "", "Directory to store the chain data.")

	rootCmd.Flags().StringVar(&NetName, "network", "testnet", "String of the network to connect.")
	rootCmd.Flags().StringVar(&Port, "port", "24126", "Default port for p2p connections listener.")

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
		RPCProxy:      false,
		RPCProxyPort:  "",
		RPCProxyAddr:  "",
		RPCPort:       "",
		RPCWallet:     false,
		RPCAuthToken:  "",
		Debug:         false,
		LogFile:       true,
		Dashboard:     false,
		DashboardPort: "",
	}

	logFile, err := os.OpenFile(path.Join(DataPath, "logger.log"), os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		panic(err)
	}
	log := logger.New(logFile)

	var netParams *params.ChainParams
	switch config.GlobalFlags.NetworkName {
	case "devnet":
		netParams = &params.DevNet
	case "testnet":
		netParams = &params.TestNet
	default:
		netParams = &params.MainNet
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
		Context:    context.Background(),
	}
}

var rootCmd = &cobra.Command{
	Use:   "ogen-gui",
	Short: "Ogen GUI is binded GUI for the ogen binary",
	Long:  "A Golang implementation of the Olympus protocol. Next generation blockchain secured by CASPER.",
	Run: func(cmd *cobra.Command, args []string) {

		db, err := blockdb.NewLevelDB()
		if err != nil {
			panic(err)
		}

		s, err := server.NewServer(db)
		if err != nil {
			panic(err)
		}

		go s.Start()

		js := mewn.String("./frontend/build/static/js/main.js")
		css := mewn.String("./frontend/build/static/css/main.css")

		app := wails.CreateApp(&wails.AppConfig{
			Width:  1024,
			Height: 768,
			Title:  "Olympus",
			JS:     js,
			CSS:    css,
			Colour: "#131313",
		})

		app.Bind(s.Wallet())

		err = app.Run()
		if err != nil {
			panic(err)
		}

		err = s.Stop()
		if err != nil {
			panic(err)
		}
	},
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}
