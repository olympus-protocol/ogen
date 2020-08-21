package chainrpc_test

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"

	"github.com/olympus-protocol/ogen/internal/chainrpc"
)

func Test_GenCertificates(t *testing.T) {
	err := chainrpc.GenerateCerts("./")
	assert.NoError(t, err)
	rmv()
}

func Test_LoadCertificatesCreating(t *testing.T) {
	_, err := chainrpc.LoadCerts("./")
	assert.NoError(t, err)
}

func Test_LoadCertificates(t *testing.T) {
	_, err := chainrpc.LoadCerts("./")
	assert.NoError(t, err)
	rmv()
}

func rmv() {
	_ = os.RemoveAll("./cert")
}
