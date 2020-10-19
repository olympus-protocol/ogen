package commands

import (
	"fmt"
	"github.com/olympus-protocol/ogen/cmd/ogen/indexer"
	"github.com/spf13/cobra"
)

var (
	rpcEndpoint string
	dbConnString string
)

func init() {
	indexerCmd.Flags().StringVar(&rpcEndpoint, "rpc_host", "127.0.0.1:24127", "IP and port of the RPC Server to connect")
	indexerCmd.Flags().StringVar(&dbConnString, "database", "", "Database connection string")

	rootCmd.AddCommand(indexerCmd)
}


var indexerCmd = &cobra.Command{
	Use:   "indexer",
	Short: "Execute the and indexer to organize the blockchain information through RPC",
	Long:  `Execute the and indexer to organize the blockchain information through RPC`,
	Run: func(cmd *cobra.Command, args []string) {

		if dbConnString == "" {
			fmt.Println("Missing database connection string")
		}

		idx := indexer.NewIndexer(dbConnString, rpcEndpoint)

		<-idx.Context().Done()
		idx.Close()
	},
}
