package cli

import (
	"encoding/json"
	"fmt"
	"github.com/olympus-protocol/ogen/config"
	"github.com/olympus-protocol/ogen/miner/keystore"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/spf13/cobra"
	"io/ioutil"
	"path"
	"time"
)

var (
	genesisTimeString int64
	connect           []string
	withdrawAddress   string
	outFile string
)

func init() {
	generateChainCmd.Flags().Int64Var(&genesisTimeString, "genesistime", 0, "sets a genesis time for the blockchain (defaults to now)")
	generateChainCmd.Flags().StringSliceVar(&connect, "connect", []string{}, "IP addresses for initial connections for this blockchain")
	generateChainCmd.Flags().StringVar(&withdrawAddress, "withdrawaddress", "1111111111111111111114oLvT2", "withdraw address for validators or unspendable address if not defined")
	generateChainCmd.Flags().StringVar(&outFile, "out", "chain.json", "chain file to save")

	rootCmd.AddCommand(generateChainCmd)
}

var generateChainCmd = &cobra.Command{
	Use:   "chain",
	Short: "Generates chain file from the keys in your wallet",
	Long:  `Generates chain file from keys in your wallet`,
	Run: func(cmd *cobra.Command, args []string) {
		keystorePath := path.Join(DataFolder, "wallet")
		k, err := keystore.NewBadgerKeystore(keystorePath)
		if err != nil {
			fmt.Printf("could not open database: %s\n", err)
			return
		}
		defer k.Close()

		keys, err := k.GetKeys()
		if err != nil {
			fmt.Printf("could not get keys from database: %s\n", err)
			return
		}

		var genesisTime time.Time
		if genesisTimeString == 0 {
			genesisTime = time.Unix(time.Now().Unix(), 0)
		} else {
			genesisTime = time.Unix(genesisTimeString, 0)
		}

		validators := make([]primitives.ValidatorInitialization, len(keys))
		for i := range validators {
			pub := keys[i].DerivePublicKey()
			pubSer := pub.Serialize()
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