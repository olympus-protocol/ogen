package testdata

import (
	"github.com/olympus-protocol/ogen/primitives"
)

var VoteSlashing = primitives.VoteSlashing{
	Vote1: MultiValidatorVote,
	Vote2: MultiValidatorVote,
}

var RANDAOSlashing = primitives.RANDAOSlashing{
	RandaoReveal: sig.Marshal(),
	Slot: 1000,
	ValidatorPubkey: randKey.Marshal(),
}

var ProposerSlashing = primitives.ProposerSlashing{
	BlockHeader1: BlockHeader,
	BlockHeader2: BlockHeader,
	Signature1: sig.Marshal(),
	Signature2: sig.Marshal(),
	ValidatorPublicKey: randKey.Marshal(),
}
