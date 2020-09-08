package main

import (
	"fmt"
	"github.com/olympus-protocol/ogen/cmd/ogen-d/rpc"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path"
)

var rpcHost string
var DataFolder string

func init() {
	cobra.OnInitialize(initConfig)
	indexerCmd.Flags().StringVar(&rpcHost, "rpc_host", "127.0.0.1:24127", "IP and port of the RPC Server to connect")
	indexerCmd.PersistentFlags().StringVar(&DataFolder, "datadir", "", "Directory to store the chain data.")

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

		ogendDir := path.Join(configDir, "ogen-d")

		if _, err := os.Stat(ogendDir); os.IsNotExist(err) {
			err = os.Mkdir(ogendDir, 0744)
			if err != nil {
				panic(err)
			}
		}

		DataFolder = ogendDir

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(ogendDir)
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	err := viper.BindPFlags(indexerCmd.Flags())
	if err != nil {
		panic(err)
	}
}

var indexerCmd = &cobra.Command{
	Use:   "ogend",
	Short: "Execute the block explorer",
	Long:  `Execute the block explorer to sync with a running instance of ogen`,
	Run: func(cmd *cobra.Command, args []string) {
		configDir := viper.GetString("config")
		fmt.Println(configDir)
		rpc.Run(rpcHost, args)
	},
}

func main() {
	err := indexerCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
}
