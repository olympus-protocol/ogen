package primitives_test

import (
	fuzz "github.com/google/gofuzz"
	"github.com/olympus-protocol/ogen/pkg/bitfield"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

// fuzzBlockHeader return a slice with n BlockHeader structs.
func fuzzBlockHeader(n int) []*primitives.BlockHeader {
	var v []*primitives.BlockHeader
	f := fuzz.New().NilChance(0)
	for i := 0; i < n; i++ {
		d := new(primitives.BlockHeader)
		f.Fuzz(&d)
		v = append(v, d)
	}
	return v
}

// fuzzVoteData simply creates a slice with VoteData
func fuzzVoteData(n int) []*primitives.VoteData {
	var v []*primitives.VoteData
	f := fuzz.New().NilChance(0)
	for i := 0; i < n; i++ {
		d := new(primitives.VoteData)
		f.Fuzz(&d)
		v = append(v, d)
	}
	return v
}

// fuzzAcceptedVoteInfo return a slice with n AcceptedVoteInfo structs.
// If correct is true will return correctly serializable structs
// If complete is true will return information with no nil pointers.
func fuzzAcceptedVoteInfo(n int, correct bool, complete bool) []*primitives.AcceptedVoteInfo {
	var v []*primitives.AcceptedVoteInfo
	for i := 0; i < n; i++ {
		i := new(primitives.AcceptedVoteInfo)
		f := fuzz.New().NilChance(0)
		f.Fuzz(i)
		i.ParticipationBitfield = bitfield.NewBitlist(6242)
		if !correct {
			i.ParticipationBitfield = bitfield.NewBitlist(50000)
		}
		if !complete {
			i.Data = nil
		}
		v = append(v, i)
	}
	return v
}

// fuzzMultiValidatorVote creates a slice of MultiValidatorVote
// If correct is true will return correctly serializable structs
// If complete is true will return information with no nil pointers.
func fuzzMultiValidatorVote(n int, correct bool, complete bool) []*primitives.MultiValidatorVote {
	var v []*primitives.MultiValidatorVote
	f := fuzz.New().NilChance(0)
	for i := 0; i < n; i++ {
		d := new(primitives.MultiValidatorVote)
		f.Fuzz(&d)
		d.ParticipationBitfield = bitfield.NewBitlist(6242)
		var sig [96]byte
		copy(sig[:], bls.CurrImplementation.NewAggregateSignature().Marshal())
		d.Sig = sig
		if !correct {
			d.ParticipationBitfield = bitfield.NewBitlist(50000)
		}
		if !complete {
			d.Data = nil
		}
		v = append(v, d)
	}
	return v
}

// fuzzValidator creates a slice of Validator
func fuzzValidator(n int) []*primitives.Validator {
	var v []*primitives.Validator
	f := fuzz.New().NilChance(0)
	for i := 0; i < n; i++ {
		d := new(primitives.Validator)
		f.Fuzz(&d)
		v = append(v, d)
	}
	return v
}

// fuzzVoteSlashing creates a slice of VoteSlashing
// If correct is true will return correctly serializable structs
// If complete is true will return information with no nil pointers.
func fuzzVoteSlashing(n int, correct bool, complete bool) []*primitives.VoteSlashing {
	var v []*primitives.VoteSlashing
	for i := 0; i < n; i++ {
		d := &primitives.VoteSlashing{
			Vote1: fuzzMultiValidatorVote(1, correct, complete)[0],
			Vote2: fuzzMultiValidatorVote(1, correct, complete)[0],
		}
		v = append(v, d)
	}
	return v
}

// fuzzRANDAOSlashing creates a slice of RANDAOSlashing
func fuzzRANDAOSlashing(n int) []*primitives.RANDAOSlashing {
	f := fuzz.New().NilChance(0)
	var v []*primitives.RANDAOSlashing
	for i := 0; i < n; i++ {
		d := new(primitives.RANDAOSlashing)
		f.Fuzz(d)
		var sig [96]byte
		var pub [48]byte
		copy(sig[:], bls.CurrImplementation.NewAggregateSignature().Marshal())
		copy(pub[:], bls.CurrImplementation.RandKey().PublicKey().Marshal())
		d.RandaoReveal = sig
		d.ValidatorPubkey = pub
		v = append(v, d)
	}
	return v
}

// fuzzProposerSlashing creates a slice of ProposerSlashing
// If complete is true will return information with no nil pointers.
func fuzzProposerSlashing(n int, complete bool) []*primitives.ProposerSlashing {
	f := fuzz.New().NilChance(0)
	var v []*primitives.ProposerSlashing
	for i := 0; i < n; i++ {
		d := new(primitives.ProposerSlashing)
		f.Fuzz(d)
		var sig [96]byte
		var pub [48]byte
		copy(sig[:], bls.CurrImplementation.NewAggregateSignature().Marshal())
		copy(pub[:], bls.CurrImplementation.RandKey().PublicKey().Marshal())
		d.Signature1 = sig
		d.Signature2 = sig
		if !complete {
			d.BlockHeader1 = nil
			d.BlockHeader2 = nil
		}
	}
	return v
}

// fuzzCoinState returns a CoinState with n balances and nonces
func fuzzCoinState(n int) *primitives.CoinsState {
	f := fuzz.New().NilChance(0).NumElements(n, n)
	balances := map[[20]byte]uint64{}
	nonces := map[[20]byte]uint64{}
	f.Fuzz(&balances)
	f.Fuzz(&nonces)

	v := &primitives.CoinsState{
		Balances: balances,
		Nonces:   nonces,
	}
	return v
}

// fuzzCoinState returns a CoinState with n balances and nonces
func fuzzCoinStateSerializable(n int) *primitives.CoinsStateSerializable {
	f := fuzz.New().NilChance(0).NumElements(n, n)

	var balances []*primitives.AccountInfo
	var nonces []*primitives.AccountInfo

	f.Fuzz(&balances)
	f.Fuzz(&nonces)

	scs := &primitives.CoinsStateSerializable{
		Balances: balances,
		Nonces:   nonces,
	}

	return scs
}

func fuzzDeposit(n int, complete bool) []*primitives.Deposit {
	var v []*primitives.Deposit
	for i := 0; i < n; i++ {
		d := &primitives.Deposit{
			Data: fuzzDepositData(),
		}
		var sig [96]byte
		var pub [48]byte
		copy(sig[:], bls.CurrImplementation.NewAggregateSignature().Marshal())
		copy(pub[:], bls.CurrImplementation.RandKey().PublicKey().Marshal())
		d.PublicKey = pub
		d.Signature = sig
		if !complete {
			d.Data = nil
		}
		v = append(v, d)
	}
	return v
}

// fuzzDepositData returns a DepositData struct
func fuzzDepositData() *primitives.DepositData {
	f := fuzz.New().NilChance(0)
	d := new(primitives.DepositData)
	f.Fuzz(d)
	var sig [96]byte
	var pub [48]byte
	copy(sig[:], bls.CurrImplementation.NewAggregateSignature().Marshal())
	copy(pub[:], bls.CurrImplementation.RandKey().PublicKey().Marshal())
	d.PublicKey = pub
	d.ProofOfPossession = sig
	return d
}

// fuzzBlock returns a Block slice
// If correct is true will return correctly serializable structs
// If complete is true will return information with no nil pointers.
func fuzzBlock(n int, correct bool, complete bool) []*primitives.Block {
	var v []*primitives.Block
	for i := 0; i < n; i++ {
		b := &primitives.Block{
			Header:            fuzzBlockHeader(1)[0],
			Votes:             fuzzMultiValidatorVote(32, true, true),
			Txs:               nil,
			TxsMulti:          nil,
			Deposits:          fuzzDeposit(128, true),
			Exits:             nil,
			VoteSlashings:     fuzzVoteSlashing(10, true, true),
			RANDAOSlashings:   fuzzRANDAOSlashing(20),
			ProposerSlashings: fuzzProposerSlashing(2, true),
			GovernanceVotes:   nil,
		}

		var sig [96]byte
		copy(sig[:], bls.CurrImplementation.NewAggregateSignature().Marshal())

		b.Signature = sig
		b.RandaoSignature = sig
		if !correct {
			b.Votes = fuzzMultiValidatorVote(50, true, true)
		}
		if !complete {
			b.Header = nil
		}
		v = append(v, b)
	}
	return v
}
