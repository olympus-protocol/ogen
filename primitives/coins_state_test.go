package primitives_test

import (
	"testing"

	"github.com/olympus-protocol/ogen/primitives"
	"github.com/prysmaticlabs/go-ssz"
)

var coinState = primitives.CoinsState{
	Balances: map[[20]byte]uint64{
		{0x0, 0x1, 0x5}:  10000,
		{0x0, 0x1, 0x6}:  50000,
		{0x0, 0x1, 0x99}: 100000,
	},
	Nonces: map[[20]byte]uint64{
		{0x0, 0x1, 0x5}:  567,
		{0x0, 0x1, 0x6}:  622,
		{0x0, 0x1, 0x99}: 9789,
	},
}

func Test_CoinStateSerialize(t *testing.T) {
	ser, err := coinState.Marshal()
	if err != nil {
		t.Error(err)
	}
	var desc primitives.CoinsState
	err = desc.Unmarshal(ser)
	if err != nil {
		t.Error(err)
	}
	equal := ssz.DeepEqual(coinState, desc)
	if !equal {
		t.Error("marshal/unmarshal failed for CoinsState")
	}
}
