package rpcclient

import (
	"fmt"
	"os"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		// Chain methods
		{Text: "getchaininfo", Description: "Get the chain status"},
		{Text: "getrawblock", Description: "Get the serialized block data"},
		{Text: "getblock", Description: "Get the block data"},
		{Text: "getblockhash", Description: "Get the block hash of specified height"},
		{Text: "getaccountinfo", Description: "Get the specified account information"},
		{Text: "gettransaction", Description: "Returns the transaction information"},

		// Validators methods
		{Text: "getvalidatorslist", Description: "Get the network validators list"},
		{Text: "getaccountvalidators", Description: "Get the validators with deposits from an account"},

		// Network methods
		{Text: "getnetworkinfo", Description: "Get current network information"},
		{Text: "getpeersinfo", Description: "Get current connected peers"},
		{Text: "addpeer", Description: "Add a new peer to the connections"},

		// Utils methods
		{Text: "sendrawtransaction", Description: "Broadcasts a serialized transaction to the network"},
		{Text: "genkeypair", Description: "Get a key pair on bech32 encoded format"},
		{Text: "genrawkeypair", Description: "Get a key pair on bls serialized format"},
		{Text: "genvalidatorkey", Description: "Create a new validator key and store the private key on the keychain"},
		{Text: "decoderawtransaction", Description: "Returns a serialized transaction on human readable format"},
		{Text: "decoderawblock", Description: "Returns a serialized block on human readable format"},

		// Wallet methods
		{Text: "listwallets", Description: "Returns a list of available wallets by name"},
		{Text: "createwallet", Description: "Creates a new wallet and returns the public account"},
		{Text: "closewallet", Description: "Closes current open wallet"},
		{Text: "getbalance", Description: "Get the current open wallet balance"},
		{Text: "getaccount", Description: "Returns the public account of the open wallet"},
		{Text: "sendtransaction", Description: "Sends a transaction using the current open wallet"},
		{Text: "startvalidator", Description: "Starts a validator using the current open wallet as the deposit holder"},
		{Text: "exitvalidator", Description: "Exits a validator from the current open wallet"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

// Empty is the empty request.
type Empty struct{}

// CLI is the module that allows operations across multiple services.
type CLI struct {
	rpcClient *RPCClient
}

var ctrlCKeybind = prompt.OptionAddKeyBind(prompt.KeyBind{
	Key: prompt.ControlC,
	Fn:  func(*prompt.Buffer) { os.Exit(0) },
})
var ctrlDKeybind = prompt.OptionAddKeyBind(prompt.KeyBind{
	Key: prompt.ControlD,
	Fn:  func(*prompt.Buffer) { os.Exit(0) },
})

// Run runs the CLI.
func (c *CLI) Run() {
	color.Green("Welcome to the Olympus CLI")
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
		// Chain methods
		case "getchaininfo":
			out, err = c.rpcClient.getChainInfo(args[1:])
		case "getrawblock":
			out, err = c.rpcClient.getRawBlock(args[1:])
		case "getblockhash":
			out, err = c.rpcClient.getBlockHash(args[1:])
		case "getblock":
			out, err = c.rpcClient.getBlock(args[1:])
		case "getaccountinfo":
			out, err = c.rpcClient.getAccountInfo(args[1:])
		case "gettransaction":
			out, err = c.rpcClient.getTransaction(args[1:])

		// Validator methods
		case "getvalidatorslist":
			out, err = c.rpcClient.getValidatorsList(args[1:])
		case "getaccountvalidators":
			out, err = c.rpcClient.getAccountValidators(args[1:])

		// Network methods
		case "getnetworkinfo":
			out, err = c.rpcClient.getNetworkInfo(args[1:])
		case "getpeersinfo":
			out, err = c.rpcClient.getPeersInfo(args[1:])
		case "addpeer":
			out, err = c.rpcClient.addPeer(args[1:])

		// Utils methods
		case "sendrawtransaction":
			out, err = c.rpcClient.sendRawTransaction(args[1:])
		case "genkeypair":
			out, err = c.rpcClient.genKeyPair(args[1:], false)
		case "genrawkeypair":
			out, err = c.rpcClient.genKeyPair(args[1:], true)
		case "genvalidatorkey":
			out, err = c.rpcClient.genValidatorKey(args[1:])
		case "decoderawtransaction":
			out, err = c.rpcClient.decodeRawTransaction(args[1:])
		case "decoderawblock":
			out, err = c.rpcClient.decodeRawBlock(args[1:])

		// Wallet methods
		case "listwallets":
			out, err = c.rpcClient.listWallets(args[1:])
		case "createwallet":
			out, err = c.rpcClient.createWallet(args[1:])
		case "openwallet":
			out, err = c.rpcClient.openWallet(args[1:])
		case "closewallet":
			out, err = c.rpcClient.closeWallet(args[1:])
		case "getbalance":
			out, err = c.rpcClient.getBalance(args[1:])
		case "getaccount":
			out, err = c.rpcClient.getAccount(args[1:])
		case "sendtransaction":
			out, err = c.rpcClient.sendTransaction(args[1:])
		case "startvalidator":
			out, err = c.rpcClient.startValidator(args[1:])
		case "exitvalidator":
			out, err = c.rpcClient.exitValidator(args[1:])
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

func newCli(rpcClient *RPCClient) *CLI {
	return &CLI{
		rpcClient: rpcClient,
	}
}

func Run(cmd *cobra.Command, args []string) {
	rpc, err := cmd.Flags().GetString("rpc")
	if err != nil {
		panic(err)
	}
	rpcClient := NewRPCClient(rpc)
	cli := newCli(rpcClient)
	cli.Run()
}
