package blockdb

import (
	"github.com/olympus-protocol/ogen/params"
)

var blockDB *BlockDB

func init() {
	bldb, err := NewBlockDB("./", params.Mainnet, nil)
	if err != nil {
		panic(err)
	}
	blockDB = bldb
}
