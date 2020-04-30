package cli

import (
	"log"
	"net/rpc"

	"github.com/olympus-protocol/ogen/chainrpc"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(walletCmd)
}

var walletCmd = &cobra.Command{
	Use:   "wallet",
	Short: "Run wallet of Olympus",
	Long:  `Run wallet of Olympus`,
	Run: func(cmd *cobra.Command, args []string) {
		//make connection to rpc server
		client, err := rpc.DialHTTP("tcp", ":1234")
		if err != nil {
			log.Fatalf("Error in dialing. %s", err)
		}
		//make arguments object
		callArgs := &chainrpc.Empty{}
		//this will store returned result
		var result uint64
		//call remote procedure with args
		err = client.Call("RPCServer.TestMethod", callArgs, &result)
		if err != nil {
			log.Fatalf("error in Arith %s", err)
		}
		//we got our result in result
		log.Printf("%d\n", result)
	},
}
