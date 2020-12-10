package commands

import (
	"fmt"
	"github.com/olympus-protocol/ogen/internal/indexer"
	"github.com/olympus-protocol/ogen/pkg/params"
	"github.com/spf13/cobra"
	"os"
)

var (
	rpcEndpoint  string
	dbConnString string
)

func init() {
	indexerCmd.Flags().StringVar(&rpcEndpoint, "rpc_host", "127.0.0.1:24127", "IP and port of the RPC Server to connect")
	indexerCmd.Flags().StringVar(&dbConnString, "dbconn", "", "Database connection string")

	rootCmd.AddCommand(indexerCmd)
}

var indexerCmd = &cobra.Command{
	Use:   "indexer <network>",
	Short: "Execute the and indexer to organize the blockchain information through RPC",
	Long:  `Execute the and indexer to organize the blockchain information through RPC`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) < 1 {
			fmt.Println("indexer <network> [flags]")
			os.Exit(0)
		}

		network := args[0]

		var netParams *params.ChainParams
		switch network {
		case "testnet":
			netParams = &params.TestNet
		case "mainnet":
			netParams = &params.Mainnet
		default:
			fmt.Println("unknown network parameters")
			os.Exit(0)
		}

		if dbConnString == "" {
			fmt.Println("Missing database connection string")
			os.Exit(0)
		}

		idx, err := indexer.NewIndexer(dbConnString, rpcEndpoint, netParams)
		if err != nil {
			os.Exit(0)
		}

		idx.Start()
		<-idx.Context().Done()
		idx.Stop()
	},
}
