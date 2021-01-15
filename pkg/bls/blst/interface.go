package blst

/*
import (
	"github.com/olympus-protocol/ogen/pkg/bls/common"
	blst "github.com/supranational/blst/bindings/go"
	"runtime"
)

type BLST struct{}

var _ common.Implementation = &BLST{}

func init() {
	// Reserve 1 core for general application work
	maxProcs := runtime.GOMAXPROCS(0) - 1
	if maxProcs <= 0 {
		maxProcs = 1
	}
	blst.SetMaxProcs(maxProcs)
}

func NewBLSTInterface() common.Implementation {
	return &BLST{}
}
*/
