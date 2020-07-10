package wallet_test

import (
	"context"
	"testing"

	"github.com/olympus-protocol/ogen/params"
	"github.com/olympus-protocol/ogen/wallet"
)

var testPass = "test_password"

func Test_NewWallet(t *testing.T) {
	walletMan, err := wallet.NewWallet(context.Background(), nil, "./", &params.Mainnet, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = walletMan.NewWallet("test_wallet", nil, testPass)
	if err != nil {
		t.Fatal(err)
	}
	err = walletMan.CloseWallet()
	if err != nil {
		t.Fatal(err)
	}
}

func Test_OpenWallet(t *testing.T) {
	
}
