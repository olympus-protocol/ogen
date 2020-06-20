package cli

import (
	"os"
	"path"

	"github.com/dgraph-io/badger"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(resetCmd)
}

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset the chain of Olympus",
	Long:  `Reset the chain of Olympus`,
	Run: func(cmd *cobra.Command, args []string) {
		bdb, err := badger.Open(badger.DefaultOptions(path.Join(DataFolder, "chain")))
		if err != nil {
			panic(err)
		}

		err = bdb.DropAll()
		if err != nil {
			panic(err)
		}
		_ = os.Remove(path.Join(DataFolder, "chain.json"))
	},
}
