package txpayloads

import (
	"errors"
	"github.com/grupokindynos/ogen/bls"
	"github.com/grupokindynos/ogen/utils/amount"
	"github.com/grupokindynos/ogen/utils/chainhash"
	"io"
)

var (
	ErrorNoMethodForPayload = errors.New("this method doesn't match current payload specifics")
)

type Payload interface {
	Serialize(w io.Writer) error
	Deserialize(r io.Reader) error
	GetAggPubKey() (*bls.PublicKey, error)
	GetPublicKeys() ([]*bls.PublicKey, error)
	GetPublicKey() (*bls.PublicKey, error)
	GetSignature() (*bls.Signature, error)
	GetMessage() ([]byte, error)
	GetHashForDataMatch() (chainhash.Hash, error)
	GetHashInvForDataMatch() ([]chainhash.Hash, error)
	GetSpentAmount() (amount.AmountType, error)
}
