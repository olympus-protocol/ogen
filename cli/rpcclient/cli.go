package rpcclient

import (
	"fmt"
	"os"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/fatih/color"
	"github.com/olympus-protocol/ogen/chainrpc"
	"github.com/spf13/cobra"
)

func completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		// Chain methods
		{Text: "getchaininfo", Description: "Get the chain status"},
		{Text: "getblock", Description: "Get the the block data"},
		{Text: "getblockhash", Description: "Get the block hash of specified height"},

		// Wallet methods
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

// CLI is the module that allows validator and wallet operation.
type CLI struct {
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

// Run runs the wallet CLI.
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
			out, err = c.GetChainInfo()
		case "getblock":
			out, err = c.GetBlock(args[1:])
		case "getblockhash":
			out, err = c.GetBlockHash(args[1:])
		// Wallet methods
		case "getaddress":
			out, err = c.GetAddress()
		case "getbalance":
			out, err = c.GetBalance(args[1:])
		case "sendtoaddress":
			out, err = c.SendToAddress(args[1:])
		case "listvalidators":
			out, err = c.ListValidators(args[1:])
		case "startvalidator":
			out, err = c.StartValidator(args[1:])
		case "generatevalidatorkey":
			out, err = c.GenerateValidatorKey()
		case "exitvalidator":
			out, err = c.ExitValidator(args[1:])
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

func NewCLI(rpcClient *chainrpc.RPCClient) *CLI {
	return &CLI{
		rpcClient: rpcClient,
	}
}

func Run(cmd *cobra.Command, args []string) {
	rpc, err := cmd.Flags().GetString("rpc")
	if err != nil {
		panic(err)
	}
	rpcClient := chainrpc.NewRPCClient(rpc)
	cli := NewCLI(rpcClient)
	cli.Run()
}
