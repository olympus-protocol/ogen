package cli

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/fatih/color"
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/keystore"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(generateKeysCmd)
}

var generateKeysCmd = &cobra.Command{
	Use:   "generate <numkeys>",
	Short: "Generates validator keys and saves them to your key store",
	Long:  `Generates validator keys and saves them to your key store`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		numKeys := 1
		if len(args) > 0 {
			numKeys, err = strconv.Atoi(args[0])
			if err != nil {
				fmt.Printf("invalid argument: %s\n", args[0])
				return
			}
		}
		k, err := keystore.NewKeystore(DataFolder, nil)
		if err != nil {
			panic(err)
		}

		keys := make([]*bls.SecretKey, numKeys)
		for i := 0; i < numKeys; i++ {
			key, err := k.GenerateNewValidatorKey()
			if err != nil {
				fmt.Printf("error generating key: %s\n", err)
				return
			}
			keys[i] = key
		}

		err = k.Close()
		if err != nil {
			fmt.Printf("error closing keychain: %s\n", err)
		}

		colorHeader := color.New(color.FgCyan, color.Bold)
		colorSecret := color.New(color.FgRed)
		colorPubkey := color.New(color.FgGreen)
		//colorNormal := color.New(color.Fg)
		for i, k := range keys {
			colorHeader.Printf("Validator #%d\n", i)
			kBytes := k.Marshal()
			pkBytes := k.PublicKey().Marshal()
			keyb := hex.EncodeToString(kBytes[:])
			pkb := hex.EncodeToString(pkBytes[:])
			colorSecret.Printf("Secret Key: ")
			fmt.Printf("%s\n", keyb)
			colorPubkey.Printf("Public Key: ")
			fmt.Printf("%s\n", pkb)
		}
	},
}
