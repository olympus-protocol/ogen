package commands

import (
	"github.com/dgraph-io/badger"
	"github.com/olympus-protocol/ogen/cmd/ogen/config"
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
		badgerdb, err := badger.Open(badger.DefaultOptions(config.DataPath + "/chain").WithLogger(nil))
		if err != nil {
			panic(err)
		}
		err = badgerdb.DropAll()
		if err != nil {
			panic(err)
		}
		_ = os.Remove(path.Join(config.DataPath, "chain.json"))
		_ = os.RemoveAll(path.Join(config.DataPath, "peerstore"))
	},
}
