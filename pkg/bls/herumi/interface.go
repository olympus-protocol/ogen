package herumi
/*
import (
	"github.com/herumi/bls-eth-go-binary/bls"
	"github.com/olympus-protocol/ogen/pkg/bls/common"
)

func init() {
	if err := bls.Init(bls.BLS12_381); err != nil {
		panic(err)
	}
	if err := bls.SetETHmode(bls.EthModeDraft07); err != nil {
		panic(err)
	}
	// Check subgroup order for pubkeys and signatures.
	bls.VerifyPublicKeyOrder(true)
	bls.VerifySignatureOrder(true)
}

type Herumi struct{}

var _ common.Implementation = &Herumi{}

func NewHerumiInterface() common.Implementation {
	return &Herumi{}
}
*/