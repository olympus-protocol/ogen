package wallet

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/fatih/color"
	"github.com/olympus-protocol/ogen/chainrpc"
	"github.com/olympus-protocol/ogen/wallet"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "getbalance", Description: "Get balance of wallet"},
		{Text: "getaddress", Description: "Get current wallet addresses"},
		{Text: "sendtoaddress", Description: "Send money to a user"},
		{Text: "listvalidators", Description: "List owned and managed validators"},
		{Text: "startvalidator", Description: "Start a validator by submitting a deposit transaction"},
		{Text: "generatevalidatorkey", Description: "Generates a validator key and allows managing it"},
		{Text: "exitvalidator", Description: "Attempts to exit an owned validator"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

// Empty is the empty request.
type Empty struct{}

func askWalletPass() ([]byte, error) {
	fmt.Printf("Password: ")
	return wallet.AskPass()
}

// CLI is the module that allows validator and wallet operation.
type CLI struct {
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

func (wc *CLI) GetAddress() (string, error) {
	address, err := wc.rpcClient.GetAddress()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Wallet address: %s", address), nil
}

func (wc *CLI) GetBalance(args []string) (string, error) {
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

func (wc *CLI) SendToAddress(args []string) (string, error) {
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

const validatorsPerPage = 32

// ListValidators lists the validators owned or managed by the wallet.
func (wc *CLI) ListValidators(args []string) (string, error) {
	validators, err := wc.rpcClient.ListValidators()
	if err != nil {
		return "", fmt.Errorf("could not get validator list: %s", err)
	}

	page := 1
	if len(args) == 1 {
		page, err = strconv.Atoi(args[0])
		if err != nil {
			return "", err
		}
	}

	numVals := 0

	if page > len(validators.Validators)/validatorsPerPage+1 {
		return "", fmt.Errorf("page %d is out of range (1 - %d)", page, len(validators.Validators)/validatorsPerPage)
	}

	if page <= 0 {
		return "", fmt.Errorf("page %d is out of range (1 - %d)", page, len(validators.Validators)/validatorsPerPage)
	}

	color.Magenta(" %-67s | %-20s | %-12s | %8s | %6s\n", "Public Key", "Balance", "Status", "Managed?", "Owned?")
	for _, v := range validators.Validators[(page-1)*validatorsPerPage:] {
		fmt.Printf(" %-67s | %-20f | %-12s | %-8t | %-6t\n", base64.StdEncoding.EncodeToString(v.Pubkey[:]), float64(v.Balance)/1000, v.Status, v.HavePrivateKey, v.HaveWithdrawalKey)
		numVals++
		if numVals == validatorsPerPage {
			break
		}
	}

	return fmt.Sprintf("Page %d/%d, Showing validators %d-%d/%d", page, len(validators.Validators)/validatorsPerPage+1, (page-1)*validatorsPerPage, page*validatorsPerPage, len(validators.Validators)), nil
}

// StartValidator starts a validator given it's private key.
func (wc *CLI) StartValidator(args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("Usage: startvalidator <privkey>")
	}

	privkey, err := base64.StdEncoding.DecodeString(args[0])
	if err != nil {
		return "", errors.Wrap(err, "cannot parse privkey")
	}

	if len(privkey) != 32 {
		return "", fmt.Errorf("expected private key to be 32 bytes long, but got %d", len(privkey))
	}

	var privKeyBytes [32]byte
	copy(privKeyBytes[:], privkey)

	deposit, err := wc.rpcClient.StartValidator(privKeyBytes, askWalletPass)
	if err != nil {
		return "", err
	}

	pubkey := deposit.PublicKey

	return fmt.Sprintf("started validator %s", base64.StdEncoding.EncodeToString(pubkey[:])), nil
}

// ExitValidator exits a validator given it's public key.
func (wc *CLI) ExitValidator(args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("Usage: exitvalidator <pubkey>")
	}

	pubkey, err := base64.StdEncoding.DecodeString(args[0])
	if err != nil {
		return "", errors.Wrap(err, "cannot parse pubkey")
	}

	if len(pubkey) != 48 {
		return "", fmt.Errorf("expected public key to be 32 bytes long, but got %d", len(pubkey))
	}

	var pubKeyBytes [48]byte
	copy(pubKeyBytes[:], pubkey)

	_, err = wc.rpcClient.ExitValidator(pubKeyBytes, askWalletPass)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("exited validator %s", base64.StdEncoding.EncodeToString(pubKeyBytes[:])), nil
}

// GenerateValidatorKey generates a validator key and starts managing it.
func (wc *CLI) GenerateValidatorKey() (string, error) {
	key, err := wc.rpcClient.GenerateValidatorKey()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Validator Private Key: %s", base64.StdEncoding.EncodeToString(key.PrivateKey[:])), nil
}

// Run runs the wallet CLI.
func (wc *CLI) Run() {
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
		case "listvalidators":
			out, err = wc.ListValidators(args[1:])
		case "startvalidator":
			out, err = wc.StartValidator(args[1:])
		case "generatevalidatorkey":
			out, err = wc.GenerateValidatorKey()
		case "exitvalidator":
			out, err = wc.ExitValidator(args[1:])
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

func NewWalletCLI(rpcClient *chainrpc.RPCClient) *CLI {
	return &CLI{
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
