package commands

import (
	"github.com/dgraph-io/badger"
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
		badgerdb, err := badger.Open(badger.DefaultOptions(GlobalDataFolder + "/chain").WithLogger(nil))
		if err != nil {
			panic(err)
		}
		err = badgerdb.DropAll()
		if err != nil {
			panic(err)
		}
		_ = os.Remove(path.Join(GlobalDataFolder, "chain.json"))
	},
}
