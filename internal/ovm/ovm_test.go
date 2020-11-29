package ogenvm

import (
	"bytes"
	"testing"

	"github.com/ethereum/evmc/v7/bindings/go/evmc"
)

func TestLoad(t *testing.T) {
	i, err := evmc.Load(modulePath)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer i.Destroy()
	if i.Name() != "evmone" {
		t.Fatalf("name is %s", i.Name())
	}
	if i.Version()[0] < '0' || i.Version()[0] > '9' {
		t.Fatalf("version number is weird: %s", i.Version())
	}
}

func TestLoadConfigure(t *testing.T) {
	i, err := evmc.LoadAndConfigure(modulePath)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer i.Destroy()
	if i.Name() != "evmone" {
		t.Fatalf("name is %s", i.Name())
	}
	if i.Version()[0] < '0' || i.Version()[0] > '9' {
		t.Fatalf("version number is weird: %s", i.Version())
	}
}

func TestExecuteEmptyCode(t *testing.T) {
	vm, _ := evmc.Load(modulePath)
	defer vm.Destroy()

	addr := evmc.Address{}
	h := evmc.Hash{}
	output, gasLeft, err := vm.Execute(nil, evmc.Byzantium, evmc.Call, false, 1, 999, addr, addr, nil, h, nil, h)

	if bytes.Compare(output, []byte("")) != 0 {
		t.Errorf("execution unexpected output: %x", output)
	}
	if gasLeft != 999 {
		t.Errorf("execution gas left is incorrect: %d", gasLeft)
	}
	if err != nil {
		t.Errorf("execution returned unexpected error: %v", err)
	}
}
