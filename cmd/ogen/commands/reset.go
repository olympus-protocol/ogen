package commands

import (
	"os"
	"path"

	"github.com/olympus-protocol/ogen/internal/blockdb"
	"github.com/spf13/cobra"
	"go.etcd.io/bbolt"
)

func init() {
	rootCmd.AddCommand(resetCmd)
}

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Removes all chain data and chain.json",
	Long:  `Removes all chain data and chain.json`,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := bbolt.Open(path.Join(DataFolder, "chain.db"), 0600, nil)
		if err != nil {
			panic(err)
		}
		err = db.Update(func(tx *bbolt.Tx) error {
			tx.DeleteBucket(blockdb.BlockDBBucketKey)
			return nil
		})
		if err != nil {
			panic(err)
		}
		if err != nil {
			panic(err)
		}
		_ = os.Remove(path.Join(DataFolder, "chain.json"))
	},
}
