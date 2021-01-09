package commands

import (
	"encoding/hex"
	"encoding/json"
	"github.com/olympus-protocol/ogen/cmd/ogen/initialization"
	"github.com/olympus-protocol/ogen/internal/keystore"
	"github.com/olympus-protocol/ogen/pkg/bip39"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/hdwallet"
	"github.com/olympus-protocol/ogen/pkg/params"
	"github.com/spf13/cobra"
	"io/ioutil"
	"path"
)

var (
	amount      int
	network     string
	genesistime int64
)

var genParamsCmd = &cobra.Command{
	Use:   "genparams",
	Short: "Used to generate parameters for network initialization",
	Long:  `Used to generate parameters for network initialization`,
	Run: func(cmd *cobra.Command, args []string) {

		var netParams *params.ChainParams
		switch network {
		case "testnet":
			netParams = &params.TestNet
		default:
			netParams = &params.MainNet
		}

		bls.Initialize(netParams, "blst")

		entropy, err := bip39.NewEntropy(256)
		if err != nil {
			panic(err)
		}

		mnemonic, err := bip39.NewMnemonic(entropy)
		if err != nil {
			panic(err)
		}

		seed := bip39.NewSeed(mnemonic, "no password")

		premine, err := hdwallet.CreateHDWallet(seed, "m/12381/1997/0/0")
		if err != nil {
			panic(err)
		}

		dirPath := "./cmd/ogen/initialization/"

		ks := keystore.NewKeystore()

		err = ks.CreateKeystore()
		if err != nil {
			panic(err)
		}

		validatorsKeys, err := ks.GenerateNewValidatorKey(uint64(amount))
		if err != nil {
			panic(err)
		}

		validators := make([]initialization.Validators, amount)
		for i, key := range validatorsKeys {
			v := initialization.Validators{
				PublicKey: hex.EncodeToString(key.Secret.PublicKey().Marshal()),
			}
			if network != "mainnet" {
				v.PrivateKey = hex.EncodeToString(key.Secret.Marshal())
			}
			validators[i] = v
		}

		initParams := initialization.NetworkInitialParams{
			Validators:     validators,
			PremineAddress: premine.PublicKey().ToAccount(&netParams.AccountPrefixes),
			GenesisTime:    genesistime,
		}

		if network != "mainnet" {
			initParams.PremineMnemonic = mnemonic
		}

		b, err := json.MarshalIndent(initParams, "", " ")
		if err != nil {
			panic(err)
		}

		err = ioutil.WriteFile(path.Join(dirPath, network+"_params.json"), b, 0700)
		if err != nil {
			panic(err)
		}

		_ = ks.Close()
	},
}

func init() {
	genParamsCmd.Flags().StringVar(&network, "network", "testnet", "network name to generate params to")
	genParamsCmd.Flags().IntVar(&amount, "amount", 8, "amount of validators to generate")
	genParamsCmd.Flags().Int64Var(&genesistime, "genesistime", 0, "genesis time to start the chain")

	rootCmd.AddCommand(genParamsCmd)
}
