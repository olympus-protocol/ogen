package chainrpc_test

import (
	"testing"

	"github.com/olympus-protocol/ogen/chainrpc"
)

func Test_GenCertificates(t *testing.T) {
	err := chainrpc.GenerateCerts()
	if err != nil {
		t.Fatal(err)
	}
}

func Test_LoadCertificates(t *testing.T) {
	_, err := chainrpc.LoadCerts()
	if err != nil {
		t.Fatal(err)
	}
}
