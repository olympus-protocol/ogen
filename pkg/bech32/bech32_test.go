package bech32_test

import (
	"github.com/olympus-protocol/ogen/pkg/bech32"
	"github.com/stretchr/testify/assert"
	"testing"
)

type validTestAddress struct {
	address string
	data    []byte
}

var (
	validChecksum = []string{
		"A12UEL5L",
		"an83characterlonghumanreadablepartthatcontainsthenumber1andtheexcludedcharactersbio1tt5tgs",
		"abcdef1qpzry9x8gf2tvdw0s3jn54khce6mua7lmqqqxw",
		"11qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqc8247j",
		"split1checkupstagehandshakeupstreamerranterredcaperred2y9e3w",
	}

	invalidAddress = []string{
		"bc1qw508d6qejxtdg4y5r3zarvary0c5xw7kg3g4ty",
		"bc1qw508d6qejxtdg4y5r3zarvary0c5xw7kv8f3t5",
		"BC13W508D6QEJXTDG4Y5R3ZARVARY0C5XW7KN40WF2",
		"bc1rw5uspcuh",
		"bc10w508d6qejxtdg4y5r3zarvary0c5xw7kw508d6qejxtdg4y5r3zarvary0c5xw7kw5rljs90",
		"BC1QR508D6QEJXTDG4Y5R3ZARVARYV98GJ9P",
		"tb1qrp33g0q5c5txsp9arysrx4k6zdkfs4nce4xj0gdcccefvpysxf3q0sL5k7",
		"tb1qrp33g0q5c5txsp9arysrx4k6zdkfs4nce4xj0gdcccefvpysxf3pjxtptv",
	}
)

func TestBech32(t *testing.T) {

	for _, addr := range validChecksum {
		_, _, err := bech32.Decode(addr)
		assert.NoError(t, err)
	}

	for _, addr := range invalidAddress {
		_, _, err := bech32.Decode(addr)
		assert.NotNil(t, err)
	}
}