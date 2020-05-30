package cli

import (
	"github.com/olympus-protocol/ogen/cli/rpcclient"
	"github.com/spf13/cobra"
)

func init() {
	cliCmd.Flags().String("rpc", "127.0.0.1:24127", "RPC address and port to connect to")

	rootCmd.AddCommand(cliCmd)
}

var cliCmd = &cobra.Command{
	Use:   "cli",
	Short: "Run RPC Client of Olympus",
	Long:  `Run RPC Client of Olympus`,
	Run:   rpcclient.Run,
}
