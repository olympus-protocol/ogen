package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/olympus-protocol/ogen/config"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/wallet"
	"github.com/spf13/cobra"
)

var (
	genesisTimeString int64
	connect           []string
	withdrawAddress   string
	outFile           string
)

func init() {
	generateChainCmd.Flags().Int64Var(&genesisTimeString, "genesistime", 0, "sets a genesis time for the blockchain (defaults to now)")
	generateChainCmd.Flags().StringSliceVar(&connect, "connect", []string{}, "IP addresses for initial connections for this blockchain")
	generateChainCmd.Flags().StringVar(&withdrawAddress, "withdrawaddress", "olpub166swhm8xkmusyu3kz4a9r8x0lc2qsncd9jnke6", "withdraw address for validators or unspendable address if not defined")
	generateChainCmd.Flags().StringVar(&outFile, "out", "chain.json", "chain file to save")

	rootCmd.AddCommand(generateChainCmd)
}

var generateChainCmd = &cobra.Command{
	Use:   "chain",
	Short: "Generates chain file from the keys in your wallet",
	Long:  `Generates chain file from keys in your wallet`,
	Run: func(cmd *cobra.Command, args []string) {
		keystorePath := path.Join(DataFolder, "wallet")

		walletDB, err := badger.Open(badger.DefaultOptions(keystorePath).WithLogger(nil))
		if err != nil {
			panic(err)
		}
		k := wallet.NewValidatorWallet(walletDB)
		defer k.Close()

		keys, err := k.GetValidatorKeys()
		if err != nil {
			fmt.Printf("could not get keys from database: %s\n", err)
			return
		}

		genesisTime := time.Unix(genesisTimeString, 0)

		validators := make([]primitives.ValidatorInitialization, len(keys))
		for i := range validators {
			pub := keys[i].PublicKey()
			pubSer := pub.Marshal()
			validators[i] = primitives.ValidatorInitialization{
				PubKey:       pubSer,
				PayeeAddress: withdrawAddress,
			}
		}

		chainFile := config.ChainFile{
			Validators:         validators,
			GenesisTime:        genesisTime.Unix(),
			InitialConnections: connect,
		}

		chainFileBytes, err := json.Marshal(chainFile)
		if err != nil {
			fmt.Printf("error writing json chain file: %s\n", err)
			return
		}

		if err := ioutil.WriteFile(outFile, chainFileBytes, 0644); err != nil {
			fmt.Printf("error writing json chain file: %s\n", err)
			return
		}
	},
}
