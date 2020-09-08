package main

import (
	"fmt"
	"github.com/olympus-protocol/ogen/cmd/ogen-d/rpcclient"
	"github.com/spf13/cobra"
	"os"
)

var rpcHost string

func init() {
	indexerCmd.Flags().StringVar(&rpcHost, "rpc_host", "127.0.0.1:24127", "IP and port of the RPC Server to connect")
}

var indexerCmd = &cobra.Command{
	Use:   "ogend",
	Short: "Execute the block explorer",
	Long:  `Execute the block explorer to sync with a running instance of ogen`,
	Run: func(cmd *cobra.Command, args []string) {
		rpcclient.Run(rpcHost, args)
	},
}

func main() {
	err := indexerCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
}
