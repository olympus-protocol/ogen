package cli

import (
	"github.com/olympus-protocol/ogen/cli/rpcclient"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cliCmd)
}

var cliCmd = &cobra.Command{
	Use:   "cli",
	Short: "Starts the integrated RPC command line.",
	Long:  `Starts the integrated RPC command line.`,
	Run:   rpcclient.Run,
}
