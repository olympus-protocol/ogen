package cli

import (
	"crypto/rand"
	"fmt"
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/config"
	"github.com/olympus-protocol/ogen/db/blockdb"
	"github.com/olympus-protocol/ogen/logger"
	"github.com/olympus-protocol/ogen/miner"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path"
	"time"
)

const numTestValidators = 128

const (
	version  = "0.1.0"
)

func getTestInitializationParameters() (*primitives.InitializationParameters, []bls.SecretKey) {
	vals := make([]primitives.ValidatorInitialization, numTestValidators)
	keys := make([]bls.SecretKey, numTestValidators)
	for i := range vals {
		k, err := bls.RandSecretKey(rand.Reader)
		if err != nil {
			panic(err)
		}

		keys[i] = *k

		vals[i] = primitives.ValidatorInitialization{
			PubKey:       keys[i].DerivePublicKey().Serialize(),
			PayeeAddress: "",
		}
	}

	return &primitives.InitializationParameters{
		InitialValidators: vals,
		GenesisTime:       time.Now().Add(1 * time.Second),
	}, keys
}

// loadOgen is the main function to run ogen.
func loadOgen(configParams *config.Config, log *logger.Logger) error {
	var currParams params.ChainParams
	switch configParams.NetworkName {
	case "mainnet":
		currParams = params.Mainnet
	default:
		currParams = params.TestNet
	}
	db, err := blockdb.NewBlockDB(configParams.DataFolder, currParams, log)
	if err != nil {
		return err
	}
	ip, keys := getTestInitializationParameters()
	listenChan := config.InterruptListener(log)
	s, err := server.NewServer(configParams, log, currParams, db, false, *ip, miner.NewBasicKeystore(keys))
	if err != nil {
		return err
	}
	go s.Start()
	<-listenChan
	err = s.Stop()
	if err != nil {

	}
	db.Close()
	return nil
}

var (
	DataFolder       string

	rootCmd = &cobra.Command{
		Use:   "ogen",
		Short: "Ogen is a Go Olympus implementation",
		Long: `A Golang implementation of the Olympus protocol.
Next generation blockchain secured by CASPER.`,
		Run: func(cmd *cobra.Command, args []string) {
			c := &config.Config{
				DataFolder:   DataFolder,
				Debug:        viper.GetBool("debug"),
				Listen:       viper.GetBool("listen"),
				NetworkName:  viper.GetString("network"),
				ConnectNodes: viper.GetStringSlice("connect"),
				Port:         int32(viper.GetUint("port")),
				MaxPeers:     int32(viper.GetUint("maxpeers")),
				Mode:         viper.GetString("mode"),
				Wallet:       viper.GetBool("wallet"),
			}
			log := logger.New(os.Stdin)
			if viper.GetBool("debug") {
				log = log.WithDebug()
			}
			log.Infof("Starting Ogen v%v", config.OgenVersion())
			log.Trace("loading log on debug mode")
			err := loadOgen(c, log)
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

		fmt.Println(ogenDir)

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(ogenDir)
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		fmt.Println(err)
	}

	err := viper.BindPFlags(rootCmd.Flags())
	if err != nil {
		panic(err)
	}
}