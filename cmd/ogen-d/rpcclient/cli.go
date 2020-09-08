package rpcclient

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"path"
)

// Empty is the empty request.
type Empty struct{}

// CLI is the module that allows operations across multiple services.
type CLI struct {
	rpcClient *RPCClient
}

// Run runs the CLI.
func (c *CLI) Run(optArgs []string) {
	fmt.Println(c.rpcClient.address)
	//Here Runs the RPC
	/* add functions in go files for the rpcClient:
	func (c *RPCClient) getNetworkInfo() (string, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		res, err := c.network.GetNetworkInfo(ctx, &proto.Empty{})
		if err != nil {
			return "", err
		}
		b, err := json.MarshalIndent(res, "", "  ")
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	after that , in this function we can call them like this: c.rpcClient.getnetworkInfo
	*/
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
