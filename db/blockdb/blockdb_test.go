package blockdb

import (
	"github.com/grupokindynos/ogen/params"
)

var blockDB *BlockDB

func init() {
	bldb, err := NewBlockDB("./", params.Mainnet, nil)
	if err != nil {
		panic(err)
	}
	blockDB = bldb
}
