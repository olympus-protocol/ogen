package testdata

import (
	"github.com/olympus-protocol/ogen/primitives"
)

var VoteSlashing = primitives.VoteSlashing{
	Vote1: &MultiValidatorVote,
	Vote2: &MultiValidatorVote,
}

var RANDAOSlashing = primitives.RANDAOSlashing{
	RandaoReveal:    sigB,
	Slot:            1000,
	ValidatorPubkey: pubB,
}

var ProposerSlashing = primitives.ProposerSlashing{
	BlockHeader1:       &BlockHeader,
	BlockHeader2:       &BlockHeader,
	Signature1:         sigB,
	Signature2:         sigB,
	ValidatorPublicKey: pubB,
}
