package cli

import (
	"fmt"
	"github.com/olympus-protocol/ogen/miner/keystore"
	"github.com/spf13/cobra"
	"path"
	"strconv"
)

func init() {
	rootCmd.AddCommand(generateKeysCmd)
}

var generateKeysCmd = &cobra.Command{
	Use:   "generate <numkeys>",
	Short: "Generates validator keys and saves them to your key store",
	Long:  `Generates validator keys and saves them to your key store`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		numKeys, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Printf("invalid argument: %s\n", args[0])
			return
		}

		keystorePath := path.Join(DataFolder, "wallet")
		k, err := keystore.NewBadgerKeystore(keystorePath)
		if err != nil {
			panic(err)
		}

		for i := 0; i < numKeys; i++ {
			_, err := k.GenerateNewKey()
			if err != nil {
				fmt.Printf("error generating key: %s\n", err)
				return
			}
		}

		err = k.Close()
		if err != nil {
			fmt.Printf("error closing wallet: %s\n", err)
		}
	},
}