package primitives

import (
	"github.com/go-test/deep"
	"testing"
)

func TestInitializationPubkeyJSON(t *testing.T) {
	ip := InitializationPubkey{1, 2, 3}

	j, err := ip.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	var ip2 InitializationPubkey
	err = ip2.UnmarshalJSON(j)
	if err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal(ip, ip2); diff != nil {
		t.Fatal(diff)
	}
}
