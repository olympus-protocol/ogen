package commands

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/fatih/color"
	"github.com/olympus-protocol/ogen/pkg/rpcclient"
	"github.com/spf13/cobra"
	strings "strings"
)

var chainCmd = []prompt.Suggest{
	{Text: "getchaininfo", Description: "Get the chain status"},
	{Text: "getrawblock", Description: "Get the serialized block data"},
	{Text: "getblock", Description: "Get the block data"},
	{Text: "getblockhash", Description: "Get the block hash of specified height"},
	{Text: "getaccountinfo", Description: "Get the specified account information"},
}

var validatorsCmd = []prompt.Suggest{
	{Text: "getvalidatorslist", Description: "Get the network validators list"},
	{Text: "getaccountvalidators", Description: "Get the validators with deposits from an account"},
}

var netCmd = []prompt.Suggest{
	{Text: "getnetworkinfo", Description: "Get current network information"},
	{Text: "getpeersinfo", Description: "Get current connected hostnode"},
	{Text: "addpeer", Description: "Add a new peer to the connections"},
}

var utilsCmd = []prompt.Suggest{
	{Text: "submitrawdata", Description: "Broadcasts a serialized transaction to the network"},
	{Text: "genkeypair", Description: "Get a key pair on bech32 encoded format"},
	{Text: "genrawkeypair", Description: "Get a key pair on bls serialized format"},
	{Text: "decoderawtransaction", Description: "Returns a serialized transaction on human readable format"},
	{Text: "decoderawblock", Description: "Returns a serialized block on human readable format"},
}

var keystoreCmd = []prompt.Suggest{
	{Text: "genvalidatorkey", Description: "Create a new validator key and store the private key on the keychain"},
}

var walletCmd = []prompt.Suggest{
	{Text: "listwallets", Description: "Returns a list of available wallets by name"},
	{Text: "openwallet", Description: "Open a created wallet"},
	{Text: "createwallet", Description: "Creates a new wallet and returns the public account"},
	{Text: "closewallet", Description: "Closes current open wallet"},
	{Text: "importwallet", Description: "Creates a new wallet based on the wif string private key"},
	{Text: "dumpwallet", Description: "Exports the mnemonic string of a wallet"},
	{Text: "getbalance", Description: "Get the current open wallet balance"},
	{Text: "getvalidators", Description: "Get validator list for open wallet"},
	{Text: "getvalidatorscount", Description: "Get validator numbers for current wallet"},
	{Text: "getaccount", Description: "Returns the public account of the open wallet"},
	{Text: "sendtransaction", Description: "Sends a transaction using the current open wallet"},
	{Text: "startvalidator", Description: "Starts a validator using the current open wallet as the deposit holder"},
	{Text: "exitvalidator", Description: "Exits a validator from the current open wallet"},
}

func completer(d prompt.Document) []prompt.Suggest {
	var commands []prompt.Suggest
	commands = append(commands, chainCmd...)
	commands = append(commands, validatorsCmd...)
	commands = append(commands, netCmd...)
	commands = append(commands, utilsCmd...)
	commands = append(commands, walletCmd...)
	commands = append(commands, keystoreCmd...)
	return prompt.FilterHasPrefix(commands, d.GetWordBeforeCursor(), true)
}

// Empty is the empty request.
type Empty struct{}

// CLI is the module that allows operations across multiple services.
type CLI struct {
	rpcClient *rpcclient.Client
}

var rpcHost string

var cliCmd = &cobra.Command{
	Use:   "console",
	Short: "Starts the integrated RPC command line.",
	Long:  `Starts the integrated RPC command line.`,
	Run: func(cmd *cobra.Command, args []string) {
		StartConsole(rpcHost, args)
	},
}

func init() {
	cliCmd.Flags().StringVar(&rpcHost, "rpc_host", "127.0.0.1:24127", "IP and port of the RPC Server to connect")

	rootCmd.AddCommand(cliCmd)
}

// Run runs the CLI.
func (c *CLI) Run(optArgs []string) {
	color.Green("Welcome to the Ogen console")
	var t string
	if len(optArgs) == 0 {
		prompt.New(c.executor, completer, prompt.OptionSetExitCheckerOnInput(c.exit)).Run()
	} else {
		t = strings.Join(optArgs, " ")
		optArgs[0] = "exit"
		c.executor(t)
	}
}

func (c *CLI) executor(str string) {
	args := strings.Split(str, " ")

	if len(args) == 0 {
		return
	}

	if args[0] == "" {
		return
	}

	var out string
	var err error

	switch args[0] {
	case "help":
		out = "Ogen CLI command \n\n"

		out += "Chain \n\n"
		for _, c := range chainCmd {
			out += fmt.Sprintf("%-25s %s \n", c.Text, c.Description)
		}
		out += "\n"

		out += "Validators\n\n"
		for _, c := range validatorsCmd {
			out += fmt.Sprintf("%-25s %s \n", c.Text, c.Description)
		}
		out += "\n"

		out += "Network\n\n"
		for _, c := range netCmd {
			out += fmt.Sprintf("%-25s %s \n", c.Text, c.Description)
		}
		out += "\n"

		out += "Utils\n\n"
		for _, c := range utilsCmd {
			out += fmt.Sprintf("%-25s %s \n", c.Text, c.Description)
		}
		out += "\n"

		out += "Wallet\n\n"
		for _, c := range walletCmd {
			out += fmt.Sprintf("%-25s %s \n", c.Text, c.Description)
		}
		out += "\n"
	// Chain methods
	case "getchaininfo":
		out, err = c.rpcClient.GetChainInfo()
	case "getrawblock":
		out, err = c.rpcClient.GetRawBlock(args[1:])
	case "getblockhash":
		out, err = c.rpcClient.GetBlockHash(args[1:])
	case "getblock":
		out, err = c.rpcClient.GetBlock(args[1:])
	case "getaccountinfo":
		out, err = c.rpcClient.GetAccountInfo(args[1:])

	// Validator methods
	case "getvalidatorslist":
		out, err = c.rpcClient.GetValidatorsList()
	case "getaccountvalidators":
		out, err = c.rpcClient.GetAccountValidators(args[1:])

	// Network methods
	case "getnetworkinfo":
		out, err = c.rpcClient.GetNetworkInfo()
	case "getpeersinfo":
		out, err = c.rpcClient.GetPeersInfo()
	case "addpeer":
		out, err = c.rpcClient.AddPeer(args[1:])

	// Utils methods
	case "submitrawdata":
		out, err = c.rpcClient.SubmitRawData(args[1:])
	case "genkeypair":
		out, err = c.rpcClient.GenKeyPair(args[1:], false)
	case "genrawkeypair":
		out, err = c.rpcClient.GenKeyPair(args[1:], true)
	case "genvalidatorkey":
		out, err = c.rpcClient.GenValidatorKey(args[1:])
	case "decoderawtransaction":
		out, err = c.rpcClient.DecodeRawTransaction(args[1:])
	case "decoderawblock":
		out, err = c.rpcClient.DecodeRawBlock(args[1:])

	// Wallet methods
	case "listwallets":
		out, err = c.rpcClient.ListWallets()
	case "createwallet":
		out, err = c.rpcClient.CreateWallet(args[1:])
	case "openwallet":
		out, err = c.rpcClient.OpenWallet(args[1:])
	case "closewallet":
		out, err = c.rpcClient.CloseWallet()
	case "importwallet":
		out, err = c.rpcClient.ImportWallet(args[1:])
	case "dumpwallet":
		out, err = c.rpcClient.DumpWallet()
	case "getbalance":
		out, err = c.rpcClient.GetBalance()
	case "getvalidators":
		out, err = c.rpcClient.GetValidators()
	case "getvalidatorscount":
		out, err = c.rpcClient.GetValidatorsCount()
	case "getaccount":
		out, err = c.rpcClient.GetAccount()
	case "sendtransaction":
		out, err = c.rpcClient.SendTransaction(args[1:])
	case "startvalidator":
		out, err = c.rpcClient.StartValidator(args[1:])
	case "exitvalidator":
		out, err = c.rpcClient.ExitValidator(args[1:])

	// Misc methods
	case "exit":
		return

	default:
		err = fmt.Errorf("unknown command: %s", args[0])
	}

	if err != nil {
		color.Red("%s", err.Error())
	} else {
		color.Green("%s", out)
	}
}

func (c *CLI) exit(in string, breakline bool) bool {
	if in == "exit" {
		breakline = true
		return true
	}
	return false
}

func newCli(rpcClient *rpcclient.Client) *CLI {
	return &CLI{
		rpcClient: rpcClient,
	}
}

func StartConsole(host string, args []string) {
	rpcClient := rpcclient.NewRPCClient(host, false)
	cli := newCli(rpcClient)
	cli.Run(args)
}
