package commands

import (
	"fmt"
	"github.com/olympus-protocol/ogen/cmd/ogen/indexer"
	"github.com/spf13/cobra"
	"os"
)

var (
	rpcEndpoint  string
	dbConnString string
	dbDriver     string
)

func init() {
	indexerCmd.Flags().StringVar(&rpcEndpoint, "rpc_host", "127.0.0.1:24127", "IP and port of the RPC Server to connect")
	indexerCmd.Flags().StringVar(&dbConnString, "dbconn", "", "Database connection string")
	indexerCmd.Flags().StringVar(&dbDriver, "driver", "mysql", "Database driver to connect the database")

	rootCmd.AddCommand(indexerCmd)
}

var indexerCmd = &cobra.Command{
	Use:   "indexer",
	Short: "Execute the and indexer to organize the blockchain information through RPC",
	Long:  `Execute the and indexer to organize the blockchain information through RPC`,
	Run: func(cmd *cobra.Command, args []string) {

		if dbConnString == "" || dbDriver == "" {
			fmt.Println("Missing database connection string or driver")
			os.Exit(0)
		}

		idx := indexer.NewIndexer(dbConnString, rpcEndpoint, dbDriver)

		idx.Start()
		<-idx.Context().Done()
		idx.Close()
	},
}
