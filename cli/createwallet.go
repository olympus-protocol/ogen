package cli

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/utils/bech32"
	"github.com/olympus-protocol/ogen/wallet"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(generateWalletCmd)
}

var generateWalletCmd = &cobra.Command{
	Use:   "wallet <name> <network>",
	Short: "Creates new wallets.",
	Long:  `Creates new wallets.`,
	Args:  cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		if len(args) < 2 {
			fmt.Print("Invalid arguments. Please specify wallet name and network: wallet <name> <network>")
			return
		}
		var net *params.ChainParams
		switch args[1] {
		case "mainnet":
			net = &params.Mainnet
		case "testnet":
			net = &params.TestNet
		default:
			net = &params.Mainnet
		}
		w, err := wallet.NewWallet(context.Background(), nil, DataFolder, net, nil, nil, nil, nil)
		if err != nil {
			panic(err)
		}
		err = w.OpenWallet(args[0])
		if err != nil {
			panic(err)
		}
		key, err := w.GetPublicKey()
		if err != nil {
			panic(err)
		}
		colorPubkey := color.New(color.FgGreen)
		colorPubkey.Printf("Public Account: ")
		fmt.Printf("%s\n", bech32.Encode(net.AddrPrefix.Public, key[:]))
	},
}
