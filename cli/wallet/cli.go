package wallet

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/fatih/color"
	"github.com/olympus-protocol/ogen/chainrpc"
	"github.com/olympus-protocol/ogen/wallet"
	"github.com/spf13/cobra"
)

func completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "getbalance", Description: "Get balance of wallet"},
		{Text: "getaddress", Description: "Get current wallet addresses"},
		{Text: "sendtoaddress", Description: "Send money to a user"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

type Empty struct{}

func askWalletPass() ([]byte, error) {
	fmt.Printf("Password: ")
	return wallet.AskPass()
}

type WalletCLI struct {
	rpcClient *chainrpc.RPCClient
}

var ctrlCKeybind = prompt.OptionAddKeyBind(prompt.KeyBind{
	Key: prompt.ControlC,
	Fn:  func(*prompt.Buffer) { os.Exit(0) },
})
var ctrlDKeybind = prompt.OptionAddKeyBind(prompt.KeyBind{
	Key: prompt.ControlD,
	Fn:  func(*prompt.Buffer) { os.Exit(0) },
})

func (wc *WalletCLI) GetAddress() (string, error) {
	address, err := wc.rpcClient.GetAddress()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Wallet address: %s", address), nil
}

func (wc *WalletCLI) GetBalance() (string, error) {
	bal, err := wc.rpcClient.GetBalance()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Wallet balance: %d\n", bal), nil
}

func (wc *WalletCLI) SendToAddress(args []string) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("Usage: sendtoaddress <toaddress> <amount>")
	}
	toAddress := args[0]
	amount, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return "", fmt.Errorf("Usage: sendtoaddress <toaddress> <amount>")
	}
	if amount <= 0 {
		return "", fmt.Errorf("amount must be positive")
	}
	txid, err := wc.rpcClient.SendToAddress(toAddress, uint64(amount), askWalletPass)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Sent transaction: %s\n", txid), nil
}

func (wc *WalletCLI) Run() {
	color.Green("Welcome to the Olympus Wallet CLI")
	for {
		t := prompt.Input("> ", completer, prompt.OptionCompletionWordSeparator(" "), ctrlCKeybind, ctrlDKeybind)

		args := strings.Split(t, " ")
		if len(args) == 0 {
			continue
		}

		if args[0] == "" {
			continue
		}

		var out string
		var err error

		switch args[0] {
		case "getaddress":
			out, err = wc.GetAddress()
		case "getbalance":
			out, err = wc.GetBalance()
		case "sendtoaddress":
			out, err = wc.SendToAddress(args[1:])
		default:
			err = fmt.Errorf("Unknown command: %s", args[0])
		}

		if err != nil {
			color.Red("%s", err)
		} else {
			color.Green("%s", out)
		}
	}
}

func NewWalletCLI(rpcClient *chainrpc.RPCClient) *WalletCLI {
	return &WalletCLI{
		rpcClient: rpcClient,
	}
}

func RunWallet(cmd *cobra.Command, args []string) {
	rpcClient := chainrpc.NewRPCClient("http://localhost:24127")
	walletCLI := NewWalletCLI(rpcClient)
	walletCLI.Run()
}
