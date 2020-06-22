package primitives_test

import "testing"

func Test_SlashingSerialize(t *testing.T) {
	err := serProposerSlashing()
	err = serVoteSlashing()
	err = serRANDAOSlashing()
}

func serProposerSlashing() error {
	return nil
}

func serVoteSlashing() error {
	return nil
}

func serRANDAOSlashing() error {
	return nil
}
