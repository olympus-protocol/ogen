package cli

import (
	"github.com/olympus-protocol/ogen/cli/wallet"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(walletCmd)
}

var walletCmd = &cobra.Command{
	Use:   "wallet",
	Short: "Run wallet of Olympus",
	Long:  `Run wallet of Olympus`,
	Run:   wallet.RunWallet,
}
