package commands

import (
	"os"
	"path"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(resetCmd)
}

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Removes all chain data and chain.json",
	Long:  `Removes all chain data and chain.json`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = os.RemoveAll(path.Join(DataPath, "peerstore"))
		_ = os.RemoveAll(path.Join(DataPath, "chain"))
	},
}
