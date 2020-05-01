package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/olympus-protocol/ogen/chainrpc"
	"github.com/olympus-protocol/ogen/wallet"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(walletCmd)
}

func completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "getbalance", Description: "Get balance of wallet"},
		{Text: "getaddress", Description: "Get current wallet addresses"},
		{Text: "sendtoaddress", Description: "Send money to a user"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

type Empty struct{}

var walletCmd = &cobra.Command{
	Use:   "wallet",
	Short: "Run wallet of Olympus",
	Long:  `Run wallet of Olympus`,
	Run: func(cmd *cobra.Command, args []string) {
		rpcClient := chainrpc.NewRPCClient("http://localhost:24127")
		fmt.Println("Please select table.")
		for {
			t := prompt.Input("> ", completer, prompt.OptionCompletionWordSeparator(" "), prompt.OptionAddKeyBind(prompt.KeyBind{
				Key: prompt.ControlC,
				Fn:  func(*prompt.Buffer) { os.Exit(0) },
			}))

			words := strings.Split(t, " ")
			if len(words) == 0 {
				continue
			}

			switch words[0] {
			case "getaddress":
				address, err := rpcClient.GetAddress()
				if err != nil {
					fmt.Println(err)
					continue
				}
				fmt.Println(address)
			case "getbalance":
				bal, err := rpcClient.GetBalance()
				if err != nil {
					fmt.Println(err)
					continue
				}
				fmt.Printf("Balance: %d\n", bal)
			case "sendtoaddress":
				if len(words) != 3 {
					fmt.Println("Usage: sendtoaddress <toaddress> <amount>")
					continue
				}
				toAddress := words[1]
				amount, err := strconv.ParseInt(words[2], 10, 64)
				if err != nil {
					fmt.Println("Usage: sendtoaddress <toaddress> <amount>")
					continue
				}
				if amount <= 0 {
					fmt.Println("amount must be positive")
					continue
				}
				txid, err := rpcClient.SendToAddress(toAddress, uint64(amount), nil)

				if err.Error() == "wallet locked, need authentication" {
					fmt.Printf("Password: ")
					pass, err := wallet.AskPass()
					if err != nil {
						fmt.Println(err)
						continue
					}

					txid, err = rpcClient.SendToAddress(toAddress, uint64(amount), pass)
					if err != nil {
						fmt.Println(err)
						continue
					}
				} else if err != nil {
					fmt.Println(err)
					continue
				}
				fmt.Printf("Sent transaction: %s\n", txid)
			default:
				fmt.Printf("Unknown command: %s\n", words[0])
			}
		}
	},
}
