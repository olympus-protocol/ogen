package commands

import (
	"context"
	"github.com/olympus-protocol/ogen/internal/hostnode"
	"github.com/olympus-protocol/ogen/internal/logger"
	"github.com/olympus-protocol/ogen/pkg/params"
	"github.com/spf13/cobra"
	"os"
)

var (
	debug   bool
	port    string
	network string
)

var (
	relayerCmd = &cobra.Command{
		Use:   "relayer",
		Short: "Starts a relayer module to work as a DHT address relayer",
		Long:  `Starts a relayer module to work as a DHT address relayer`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			log := logger.New(os.Stdin)
			if debug {
				log.WithDebug()
			}
			config := hostnode.Config{
				Log:  log,
				Port: port,
				Path: DataFolder,
			}
			var currParams params.ChainParams
			if network == "testnet" {
				currParams = params.TestNet
			} else {
				currParams = params.Mainnet
			}

			hn, err := hostnode.NewHostNode(ctx, config, nil, currParams.NetMagic, true)
			if err != nil {
				log.Fatal(err)
			}

			go func() {
				err = hn.Start()
				if err != nil {
					log.Fatal(err)
				}
			}()

			<-hn.GetContext().Done()
		},
	}
)

func init() {
	relayerCmd.Flags().BoolVar(&debug, "debug", false, "start the relayer with debug logging")
	relayerCmd.Flags().StringVar(&port, "port", "24126", "the port on which the relayer will listen")
	relayerCmd.Flags().StringVar(&network, "network", "testnet", "the network to relay")

	rootCmd.AddCommand(relayerCmd)
}
