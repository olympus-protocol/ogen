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

		// Validators methods
		{Text: "listvalidators", Description: "List owned and managed validators"},
		{Text: "generatevalidatorkey", Description: "Generates a validator key and allows managing it"},
		{Text: "exitvalidator", Description: "Attempts to exit an owned validator"},
		{Text: "startvalidator", Description: "Start a validator by submitting a deposit transaction"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

// Empty is the empty request.
type Empty struct{}

// CLI is the module that allows validator and wallet operation.
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
			out, err = c.rpcClient.GetChainInfo()
		case "getblockhash":
			out, err = c.rpcClient.GetBlockHash(args[1:])
		case "getrawblock":
			out, err = c.rpcClient.GetRawBlock(args[1:])
		case "getblock":
			out, err = c.rpcClient.GetBlock(args[1:])

		// Validator methods
		//case "startvalidator":
		//	out, err = c.rpcClient.StartValidator(args[1:])
		//case "exitvalidator":
		//	out, err = c.rpcClient.ExitValidator(args[1:])
		case "listvalidators":
			out, err = c.rpcClient.GetValidatorsList(args[1:])
		case "generatevalidatorkey":
			out, err = c.rpcClient.GenerateValidatorKey()

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

func NewCLI(rpcClient *RPCClient) *CLI {
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
	cli := NewCLI(rpcClient)
	cli.Run()
}
