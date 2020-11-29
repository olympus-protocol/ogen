package ogenvm

import (
	vm "github.com/ethereum/evmc/v7/bindings/go/evmc"
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

// The path to the evm lib
var modulePath = "./lib/libevmone.so"

// The global var for the initialized vm
var ovm *vm.VM

// loadVM will load the VM into memory ready for execution when the node boots up
func loadVM() error {

	i, err := vm.Load(modulePath)

	if err != nil {
		return err
	}
	ovm = i

	return nil
}

// getVersion returns the version of the virtual machine
func getVersion() string {
	return ovm.Version()
}

// closeVm destroys the VM after excution
func closeVM() {
	ovm.Destroy()
}

// executeContract is the function called to execute the VM
func executeContract(c *primitives.Contract) ([]byte, int64, error) {

	contractAddress, _ := c.GetContractAddress()
	fromAddress, _ := c.GetFromAccountAddress()
	bytecode := c.GetContractByteCode()
	gas := c.GetContractGas()
	inputData := c.GetContractInputData()
	contractHash := c.Hash()

	to := vm.Address(contractAddress)
	from := vm.Address(fromAddress)
	cHash := vm.Hash(contractHash)

	return execute(&to, &from, inputData, cHash, bytecode, cHash, gas)
}

//execute executes smart contract bytecode and returns the output
func execute(to *vm.Address, from *vm.Address, input []byte, value vm.Hash,
	code []byte, create2Salt vm.Hash, gas int64) ([]byte, int64, error) {

	output, gasLeft, err := ovm.Execute(nil, vm.Byzantium, vm.Call, false, 1, gas, *to, *from, nil, value, nil, create2Salt)

	return output, gasLeft, err

}
