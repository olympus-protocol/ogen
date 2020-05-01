package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/olympus-protocol/ogen/chainrpc"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(walletCmd)
}

func completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "getbalance", Description: "Get balance of wallet"},
		{Text: "getaddress", Description: "Get current wallet addresses"},
		{Text: "send", Description: "Send money to a user"},
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
				var address string
				address, err := rpcClient.GetAddress()
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Println(address)
			default:
				fmt.Printf("Unknown command: %s\n", words[0])
			}
		}
	},
}
