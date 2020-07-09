package cli

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/fatih/color"
	"github.com/olympus-protocol/ogen/keystore"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(generateKeysCmd)
}

var generateKeysCmd = &cobra.Command{
	Use:   "generate <numkeys> <password>",
	Short: "Creates validator keys and stores into the keystore",
	Long:  `Creates validator keys and stores into the keystore`,
	Args:  cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		if len(args) < 2 {
			panic("please specificy a number of keys and a keystore password")
		}
		numKeys, err := strconv.Atoi(args[0])
		if err != nil {
			panic("invalid argument: " + args[0] + "\n")
		}
		k, err := keystore.NewKeystore(DataFolder, nil)
		if err != nil {
			panic(err)
		}

		keys, err := k.GenerateNewValidatorKey(uint64(numKeys), args[1])
		if err != nil {
			fmt.Printf("error generating key: %s\n", err)
			return
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
