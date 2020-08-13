package cli_test

import (
	"github.com/olympus-protocol/ogen/internal/cli/rpcclient"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"testing"
)

var rpcHost string

func NewRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cli_test",
		Short: "Mock RPC Cli",
		Long:  `Simulates the functionality of the cli (single command only)`,
		Run: func(cmd *cobra.Command, args []string) {
			rpcclient.Run(rpcHost, args)
		},
	}
}

func Test_ChainCommands(t *testing.T) {
	cmd := NewRootCmd()
	cmd.Flags().StringVar(&rpcHost, "rpc_host", "127.0.0.1:24127", "IP and port of the RPC Server to connect")

	cmd.SetArgs([]string{"getchaininfo"})
	assert.NotPanics(t, func() {
		err := cmd.Execute()
		assert.Nil(t, err)
	})
	cmd.SetArgs([]string{"getrawblock <hash>"})
	assert.NotPanics(t, func() {
		err := cmd.Execute()
		assert.Nil(t, err)
	})
	cmd.SetArgs([]string{"getblockhash <height>"})
	assert.NotPanics(t, func() {
		err := cmd.Execute()
		assert.Nil(t, err)
	})
	cmd.SetArgs([]string{"getblock <hash>"})
	assert.NotPanics(t, func() {
		err := cmd.Execute()
		assert.Nil(t, err)
	})
	cmd.SetArgs([]string{"getaccountinfo <account>"})
	assert.NotPanics(t, func() {
		err := cmd.Execute()
		assert.Nil(t, err)
	})
	cmd.SetArgs([]string{"gettransaction <txid>"})
	assert.NotPanics(t, func() {
		err := cmd.Execute()
		assert.Nil(t, err)
	})
}

func Test_ValidatorCommands(t *testing.T) {
	cmd := NewRootCmd()

	cmd.SetArgs([]string{"getvalidatorslist"})
	assert.NotPanics(t, func() {
		err := cmd.Execute()
		assert.Nil(t, err)
	})
	cmd.SetArgs([]string{"getaccountvalidators <account>"})
	assert.NotPanics(t, func() {
		err := cmd.Execute()
		assert.Nil(t, err)
	})

}

func Test_NetworkCommands(t *testing.T) {
	cmd := NewRootCmd()

	cmd.SetArgs([]string{"getnetworkinfo"})
	assert.NotPanics(t, func() {
		err := cmd.Execute()
		assert.Nil(t, err)
	})
	cmd.SetArgs([]string{"getpeersinfo <account>"})
	assert.NotPanics(t, func() {
		err := cmd.Execute()
		assert.Nil(t, err)
	})
	cmd.SetArgs([]string{"addpeer <addr>"})
	assert.NotPanics(t, func() {
		err := cmd.Execute()
		assert.Nil(t, err)
	})
}

func Test_UtilCommands(t *testing.T) {
	cmd := NewRootCmd()

	cmd.SetArgs([]string{"startproposer <keystore_pass>"})
	assert.NotPanics(t, func() {
		err := cmd.Execute()
		assert.Nil(t, err)
	})
	cmd.SetArgs([]string{"stopproposer"})
	assert.NotPanics(t, func() {
		err := cmd.Execute()
		assert.Nil(t, err)
	})
	cmd.SetArgs([]string{"submitrawdata <raw_data> <type>"})
	assert.NotPanics(t, func() {
		err := cmd.Execute()
		assert.Nil(t, err)
	})
	cmd.SetArgs([]string{"genkeypair"})
	assert.NotPanics(t, func() {
		err := cmd.Execute()
		assert.Nil(t, err)
	})
	cmd.SetArgs([]string{"genrawkeypair"})
	assert.NotPanics(t, func() {
		err := cmd.Execute()
		assert.Nil(t, err)
	})
	cmd.SetArgs([]string{"genvalidatorkey <keys> <pass>"})
	assert.NotPanics(t, func() {
		err := cmd.Execute()
		assert.Nil(t, err)
	})
	cmd.SetArgs([]string{"decoderawtransaction <raw_tx>"})
	assert.NotPanics(t, func() {
		err := cmd.Execute()
		assert.Nil(t, err)
	})
	cmd.SetArgs([]string{"decoderawblock <raw_block>"})
	assert.NotPanics(t, func() {
		err := cmd.Execute()
		assert.Nil(t, err)
	})
}

func Test_WalletCommands(t *testing.T) {
	cmd := NewRootCmd()

	cmd.SetArgs([]string{"listwallets"})
	assert.NotPanics(t, func() {
		err := cmd.Execute()
		assert.Nil(t, err)
	})
	cmd.SetArgs([]string{"createwallet <name> <pass>"})
	assert.NotPanics(t, func() {
		err := cmd.Execute()
		assert.Nil(t, err)
	})
	cmd.SetArgs([]string{"openwallet <name> <pass>"})
	assert.NotPanics(t, func() {
		err := cmd.Execute()
		assert.Nil(t, err)
	})
	cmd.SetArgs([]string{"closewallet"})
	assert.NotPanics(t, func() {
		err := cmd.Execute()
		assert.Nil(t, err)
	})
	cmd.SetArgs([]string{"importwallet <name> <wif>"})
	assert.NotPanics(t, func() {
		err := cmd.Execute()
		assert.Nil(t, err)
	})
	cmd.SetArgs([]string{"dumpwallet"})
	assert.NotPanics(t, func() {
		err := cmd.Execute()
		assert.Nil(t, err)
	})
	cmd.SetArgs([]string{"getbalance"})
	assert.NotPanics(t, func() {
		err := cmd.Execute()
		assert.Nil(t, err)
	})
	cmd.SetArgs([]string{"getvalidators"})
	assert.NotPanics(t, func() {
		err := cmd.Execute()
		assert.Nil(t, err)
	})
	cmd.SetArgs([]string{"getaccount"})
	assert.NotPanics(t, func() {
		err := cmd.Execute()
		assert.Nil(t, err)
	})
	cmd.SetArgs([]string{"sendtransaction <account> <amount>"})
	assert.NotPanics(t, func() {
		err := cmd.Execute()
		assert.Nil(t, err)
	})
	cmd.SetArgs([]string{"startvalidator <privkey>"})
	assert.NotPanics(t, func() {
		err := cmd.Execute()
		assert.Nil(t, err)
	})
	cmd.SetArgs([]string{"exitvalidator <pubkey>"})
	assert.NotPanics(t, func() {
		err := cmd.Execute()
		assert.Nil(t, err)
	})
}
