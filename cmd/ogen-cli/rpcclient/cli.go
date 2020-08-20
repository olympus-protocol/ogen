package rpcclient

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/fatih/color"
	"github.com/spf13/viper"
)

var chainCmd = []prompt.Suggest{
	{Text: "getchaininfo", Description: "Get the chain status"},
	{Text: "getrawblock", Description: "Get the serialized block data"},
	{Text: "getblock", Description: "Get the block data"},
	{Text: "getblockhash", Description: "Get the block hash of specified height"},
	{Text: "getaccountinfo", Description: "Get the specified account information"},
	{Text: "gettransaction", Description: "Returns the transaction information"},
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
	{Text: "startproposer", Description: "Unlocks the keystore and starts the proposer service."},
	{Text: "stopproposer", Description: "Locks the keystore and stops the proposer service."},
	{Text: "submitrawdata", Description: "Broadcasts a serialized transaction to the network"},
	{Text: "genkeypair", Description: "Get a key pair on bech32 encoded format"},
	{Text: "genrawkeypair", Description: "Get a key pair on bls serialized format"},
	{Text: "genvalidatorkey", Description: "Create a new validator key and store the private key on the keychain"},
	{Text: "decoderawtransaction", Description: "Returns a serialized transaction on human readable format"},
	{Text: "decoderawblock", Description: "Returns a serialized block on human readable format"},
}

var walletCmd = []prompt.Suggest{
	{Text: "listwallets", Description: "Returns a list of available wallets by name"},
	{Text: "openwallet", Description: "Open a created wallet"},
	{Text: "createwallet", Description: "Creates a new wallet and returns the public account"},
	{Text: "closewallet", Description: "Closes current open wallet"},
	{Text: "importwallet", Description: "Creates a new wallet based on the wif string private key"},
	{Text: "dumpwallet", Description: "Exports the private key on wif format of the open wallet"},
	{Text: "getbalance", Description: "Get the current open wallet balance"},
	{Text: "getvalidators", Description: "Get validator list for open wallet"},
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
	return prompt.FilterHasPrefix(commands, d.GetWordBeforeCursor(), true)
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
func (c *CLI) Run(optArgs []string) {
	color.Green("Welcome to the Ogen cli")
	for {
		var t string
		if len(optArgs) == 0 {
			t = prompt.Input("> ", completer, prompt.OptionCompletionWordSeparator(" "), ctrlCKeybind, ctrlDKeybind)
		} else {
			t = strings.Join(optArgs, " ")
			optArgs[0] = "exit"
		}

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
		case "help":
			out = "Ogen CLI commands \n\n"

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
			out, err = c.rpcClient.getChainInfo()
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
			out, err = c.rpcClient.getValidatorsList()
		case "getaccountvalidators":
			out, err = c.rpcClient.getAccountValidators(args[1:])

		// Network methods
		case "getnetworkinfo":
			out, err = c.rpcClient.getNetworkInfo()
		case "getpeersinfo":
			out, err = c.rpcClient.getPeersInfo()
		case "addpeer":
			out, err = c.rpcClient.addPeer(args[1:])

		// Utils methods
		case "startproposer":
			out, err = c.rpcClient.startProposer(args[1:])
		case "stopproposer":
			out, err = c.rpcClient.stopProposer(args[1:])
		case "submitrawdata":
			out, err = c.rpcClient.submitRawData(args[1:])
		case "genkeypair":
			out, err = c.rpcClient.genKeyPair(false)
		case "genrawkeypair":
			out, err = c.rpcClient.genKeyPair(true)
		case "genvalidatorkey":
			out, err = c.rpcClient.genValidatorKey(args[1:])
		case "decoderawtransaction":
			out, err = c.rpcClient.decodeRawTransaction(args[1:])
		case "decoderawblock":
			out, err = c.rpcClient.decodeRawBlock(args[1:])

		// Wallet methods
		case "listwallets":
			out, err = c.rpcClient.listWallets()
		case "createwallet":
			out, err = c.rpcClient.createWallet(args[1:])
		case "openwallet":
			out, err = c.rpcClient.openWallet(args[1:])
		case "closewallet":
			out, err = c.rpcClient.closeWallet()
		case "importwallet":
			out, err = c.rpcClient.importWallet(args[1:])
		case "dumpwallet":
			out, err = c.rpcClient.dumpWallet()
		case "getbalance":
			out, err = c.rpcClient.getBalance()
		case "getvalidators":
			out, err = c.rpcClient.getValidators()
		case "getaccount":
			out, err = c.rpcClient.getAccount()
		case "sendtransaction":
			out, err = c.rpcClient.sendTransaction(args[1:])
		case "startvalidator":
			out, err = c.rpcClient.startValidator(args[1:])
		case "exitvalidator":
			out, err = c.rpcClient.exitValidator(args[1:])

		// Misc methods
		case "exit":
			return

		default:
			err = fmt.Errorf("Unknown command: %s", args[0])
		}

		if err != nil {
			color.Red("%s", err.Error())
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

func Run(host string, args []string) {
	DataFolder := viper.GetString("datadir")
	if DataFolder != "" {
		// Use config file from the flag.
		viper.AddConfigPath(DataFolder)
		viper.SetConfigName("config")
	} else {
		configDir, err := os.UserConfigDir()
		if err != nil {
			panic(err)
		}

		ogenDir := path.Join(configDir, "ogen")

		if _, err := os.Stat(ogenDir); os.IsNotExist(err) {
			err = os.Mkdir(ogenDir, 0744)
			if err != nil {
				panic(err)
			}
		}

		DataFolder = ogenDir

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(ogenDir)
		viper.SetConfigName("config")
	}
	rpcClient := NewRPCClient(host, DataFolder)
	cli := newCli(rpcClient)
	cli.Run(args)
}
