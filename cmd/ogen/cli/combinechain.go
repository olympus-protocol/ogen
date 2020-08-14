package cli

import (
	"encoding/json"
	"github.com/olympus-protocol/ogen/internal/state"
	"io/ioutil"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(combineChainCmd)
}

var combineChainCmd = &cobra.Command{
	Use:   "combine",
	Short: "Combines two chain files,",
	Long:  `Combines two chain files.`,
	Args:  cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		chainFilenames := args[:len(args)-1]
		output := args[len(args)-1]

		chainFiles := make([]state.ChainFile, len(chainFilenames))
		for i := range chainFiles {
			f, err := os.Open(chainFilenames[i])
			if err != nil {
				panic(err)
			}
			chainFileBytes, err := ioutil.ReadAll(f)
			if err != nil {
				panic(err)
			}
			if err := json.Unmarshal(chainFileBytes, &chainFiles[i]); err != nil {
				panic(err)
			}
		}

		for _, c := range chainFiles {
			if c.GenesisTime != chainFiles[0].GenesisTime {
				panic("Chain files must have identical genesis time")
			}
		}

		newValidators := make([]state.ValidatorInitialization, 0)
		newInitialPeers := make([]string, 0)
		for _, c := range chainFiles {
			color.Yellow("combining chain file with %d validators and %d initial connections", len(c.Validators), len(c.InitialConnections))
			newValidators = append(newValidators, c.Validators...)
			newInitialPeers = append(newInitialPeers, c.InitialConnections...)
		}

		color.Green("new chain file has %d validators and %d initial connections", len(newValidators), len(newInitialPeers))

		newChainFile := state.ChainFile{
			Validators:         newValidators,
			InitialConnections: newInitialPeers,
			GenesisTime:        chainFiles[0].GenesisTime,
		}

		newChainFileBytes, err := json.Marshal(newChainFile)
		if err != nil {
			panic(err)
		}

		if err := ioutil.WriteFile(output, newChainFileBytes, 0666); err != nil {
			panic(err)
		}
	},
}
