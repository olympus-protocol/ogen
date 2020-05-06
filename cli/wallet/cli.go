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

func amountStringToAmount(a string) (uint64, error) {
	if strings.Contains(".", a) {
		parts := strings.Split(".", a)
		whole, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return 0, err
		}
		fractional, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return 0, err
		}

		return uint64(whole*1000 + fractional), nil
	}
	whole, err := strconv.ParseInt(a, 10, 64)
	if err != nil {
		return 0, err
	}

	return uint64(whole) * 1000, nil
}

func amountToAmountString(amount uint64) string {
	whole := amount / 1000
	fractional := amount % 1000

	return fmt.Sprintf("%d.%.03d", whole, fractional)
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

func (wc *WalletCLI) GetBalance(args []string) (string, error) {
	addr := ""
	if len(args) > 0 {
		addr = args[0]
	}
	bal, err := wc.rpcClient.GetBalance(addr)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Wallet balance: %s", amountToAmountString(bal)), nil
}

func (wc *WalletCLI) SendToAddress(args []string) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("Usage: sendtoaddress <toaddress> <amount>")
	}
	toAddress := args[0]
	amount, err := amountStringToAmount(args[1])
	if err != nil {
		return "", fmt.Errorf("Usage: sendtoaddress <toaddress> <amount>")
	}
	if amount <= 0 {
		return "", fmt.Errorf("amount must be positive")
	}
	_, err = wc.rpcClient.SendToAddress(toAddress, uint64(amount), askWalletPass)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Sent transaction"), nil
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
			out, err = wc.GetBalance(args[1:])
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
	rpc, err := cmd.Flags().GetString("rpc")
	if err != nil {
		panic(err)
	}
	rpcClient := chainrpc.NewRPCClient(rpc)
	walletCLI := NewWalletCLI(rpcClient)
	walletCLI.Run()
}
