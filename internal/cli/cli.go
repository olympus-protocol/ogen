package cli

import (
	"github.com/olympus-protocol/ogen/internal/cli/rpcclient"
	"github.com/spf13/cobra"
)

var rpcHost string

var cliCmd = &cobra.Command{
	Use:   "cli",
	Short: "Starts the integrated RPC command line.",
	Long:  `Starts the integrated RPC command line.`,
	Run: func(cmd *cobra.Command, args []string) {
		rpcclient.Run(rpcHost, args)
	},
}

func init() {

	cliCmd.Flags().StringVar(&rpcHost, "rpc_host", "127.0.0.1:24127", "IP and port of the RPC Server to connect")

	rootCmd.AddCommand(cliCmd)
}
