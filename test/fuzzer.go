package testdata

import (
	fuzz "github.com/google/gofuzz"
	"github.com/olympus-protocol/ogen/pkg/bitfield"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

// FuzzBlockHeader return a slice with n BlockHeader structs.
func FuzzBlockHeader(n int) []*primitives.BlockHeader {
	var v []*primitives.BlockHeader
	f := fuzz.New().NilChance(0)
	for i := 0; i < n; i++ {
		d := new(primitives.BlockHeader)
		f.Fuzz(&d)
		v = append(v, d)
	}
	return v
}

// FuzzVoteData simply creates a slice with VoteData
func FuzzVoteData(n int) []*primitives.VoteData {
	var v []*primitives.VoteData
	f := fuzz.New().NilChance(0)
	for i := 0; i < n; i++ {
		d := new(primitives.VoteData)
		f.Fuzz(&d)
		v = append(v, d)
	}
	return v
}

// FuzzAcceptedVoteInfo return a slice with n AcceptedVoteInfo structs.
// If correct is true will return correctly serializable structs
// If complete is true will return information with no nil pointers.
func FuzzAcceptedVoteInfo(n int, correct bool, complete bool) []*primitives.AcceptedVoteInfo {
	var v []*primitives.AcceptedVoteInfo
	for i := 0; i < n; i++ {
		i := new(primitives.AcceptedVoteInfo)
		f := fuzz.New().NilChance(0)
		f.Fuzz(i)
		i.ParticipationBitfield = bitfield.NewBitlist(6250)
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

// FuzzMultiValidatorVote creates a slice of MultiValidatorVote
// If correct is true will return correctly serializable structs
// If complete is true will return information with no nil pointers.
func FuzzMultiValidatorVote(n int) []*primitives.MultiValidatorVote {
	var v []*primitives.MultiValidatorVote
	f := fuzz.New().NilChance(0)
	for i := 0; i < n; i++ {
		d := new(primitives.MultiValidatorVote)
		f.Fuzz(&d)
		d.ParticipationBitfield = bitfield.NewBitlist(50000)
		var sig [96]byte
		copy(sig[:], bls.NewAggregateSignature().Marshal())
		d.Sig = sig
		v = append(v, d)
	}
	return v
}

// FuzzValidator creates a slice of Validator
func FuzzValidator(n int) []*primitives.Validator {
	var v []*primitives.Validator
	f := fuzz.New().NilChance(0)
	for i := 0; i < n; i++ {
		d := new(primitives.Validator)
		f.Fuzz(&d)
		v = append(v, d)
	}
	return v
}

// FuzzVoteSlashing creates a slice of VoteSlashing
// If correct is true will return correctly serializable structs
// If complete is true will return information with no nil pointers.
func FuzzVoteSlashing(n int) []*primitives.VoteSlashing {
	var v []*primitives.VoteSlashing
	for i := 0; i < n; i++ {
		d := &primitives.VoteSlashing{
			Vote1: FuzzMultiValidatorVote(1)[0],
			Vote2: FuzzMultiValidatorVote(1)[0],
		}
		v = append(v, d)
	}
	return v
}

// FuzzRANDAOSlashing creates a slice of RANDAOSlashing
func FuzzRANDAOSlashing(n int) []*primitives.RANDAOSlashing {
	f := fuzz.New().NilChance(0)
	var v []*primitives.RANDAOSlashing
	for i := 0; i < n; i++ {
		d := new(primitives.RANDAOSlashing)
		f.Fuzz(d)
		var sig [96]byte
		var pub [48]byte
		k, _ := bls.RandKey()
		copy(sig[:], bls.NewAggregateSignature().Marshal())
		copy(pub[:], k.PublicKey().Marshal())
		d.RandaoReveal = sig
		d.ValidatorPubkey = pub
		v = append(v, d)
	}
	return v
}

// FuzzProposerSlashing creates a slice of ProposerSlashing
// If complete is true will return information with no nil pointers.
func FuzzProposerSlashing(n int, complete bool) []*primitives.ProposerSlashing {
	var v []*primitives.ProposerSlashing
	for i := 0; i < n; i++ {
		d := &primitives.ProposerSlashing{
			BlockHeader1: FuzzBlockHeader(1)[0],
			BlockHeader2: FuzzBlockHeader(1)[0],
		}
		var sig [96]byte
		var pub [48]byte
		k, _ := bls.RandKey()
		copy(sig[:], bls.NewAggregateSignature().Marshal())
		copy(pub[:], k.PublicKey().Marshal())
		d.Signature1 = sig
		d.Signature2 = sig
		d.ValidatorPublicKey = pub
		if !complete {
			d.BlockHeader1 = nil
			d.BlockHeader2 = nil
		}
		v = append(v, d)
	}
	return v
}

// FuzzCoinState returns a CoinState with n balances and nonces
func FuzzCoinState(n int) *primitives.CoinsState {
	f := fuzz.New().NilChance(0).NumElements(n, n)
	balances := map[[20]byte]uint64{}
	nonces := map[[20]byte]uint64{}
	proofs := map[[32]byte]struct{}{}
	f.Fuzz(&balances)
	f.Fuzz(&nonces)
	f.Fuzz(&proofs)
	v := &primitives.CoinsState{
		Balances:       balances,
		Nonces:         nonces,
		ProofsVerified: proofs,
	}
	return v
}

// FuzzCoinStateSerializable returns a CoinState with n balances and nonces
func FuzzCoinStateSerializable(n int) *primitives.CoinsStateSerializable {
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

// FuzzDeposit creates a slice of Deposits.
// If complete is true it will create deposits with not nil pointers
func FuzzDeposit(n int, complete bool) []*primitives.Deposit {
	var v []*primitives.Deposit
	for i := 0; i < n; i++ {
		d := &primitives.Deposit{
			Data: FuzzDepositData(),
		}
		var sig [96]byte
		var pub [48]byte
		k, _ := bls.RandKey()
		copy(sig[:], bls.NewAggregateSignature().Marshal())
		copy(pub[:], k.PublicKey().Marshal())
		d.PublicKey = pub
		d.Signature = sig
		if !complete {
			d.Data = nil
		}
		v = append(v, d)
	}
	return v
}

// FuzzDepositData returns a DepositData struct
func FuzzDepositData() *primitives.DepositData {
	f := fuzz.New().NilChance(0)
	d := new(primitives.DepositData)
	f.Fuzz(d)
	var sig [96]byte
	var pub [48]byte
	k, _ := bls.RandKey()
	copy(sig[:], bls.NewAggregateSignature().Marshal())
	copy(pub[:], k.PublicKey().Marshal())
	d.PublicKey = pub
	d.ProofOfPossession = sig
	return d
}

// FuzzBlock returns a Block slice
// If correct is true will return correctly serializable structs
// If complete is true will return information with no nil pointers.
func FuzzBlock(n int, correct bool, complete bool) []*primitives.Block {
	var v []*primitives.Block
	for i := 0; i < n; i++ {
		b := &primitives.Block{
			Header:            FuzzBlockHeader(1)[0],
			Votes:             FuzzMultiValidatorVote(4),
			Deposits:          FuzzDeposit(10, true),
			Exits:             FuzzExits(10),
			PartialExit:       FuzzPartialExits(10),
			Txs:               FuzzTx(10),
			ProposerSlashings: FuzzProposerSlashing(2, true),
			VoteSlashings:     FuzzVoteSlashing(5),
			RANDAOSlashings:   FuzzRANDAOSlashing(2),
		}

		var sig [96]byte
		copy(sig[:], bls.NewAggregateSignature().Marshal())

		b.Signature = sig
		b.RandaoSignature = sig
		if !correct {
			b.Votes = FuzzMultiValidatorVote(50)
		}
		if !complete {
			b.Header = nil
		}
		v = append(v, b)
	}
	return v
}

// FuzzExits return an slice of Exits
// If correct is true will return correctly serializable structs
// If complete is true will return information with no nil pointers.
func FuzzExits(n int) []*primitives.Exit {
	var v []*primitives.Exit
	for i := 0; i < n; i++ {
		var sig [96]byte
		var pub [48]byte
		k, _ := bls.RandKey()
		copy(sig[:], bls.NewAggregateSignature().Marshal())
		copy(pub[:], k.PublicKey().Marshal())
		d := &primitives.Exit{
			ValidatorPubkey: pub,
			Signature:       sig,
			WithdrawPubkey:  pub,
		}
		v = append(v, d)
	}
	return v
}

// FuzzPartialExits return an slice of PartialExits
func FuzzPartialExits(n int) []*primitives.PartialExit {
	var v []*primitives.PartialExit
	for i := 0; i < n; i++ {
		var sig [96]byte
		var pub [48]byte
		k, _ := bls.RandKey()
		copy(sig[:], bls.NewAggregateSignature().Marshal())
		copy(pub[:], k.PublicKey().Marshal())
		d := &primitives.PartialExit{
			ValidatorPubkey: pub,
			Signature:       sig,
			WithdrawPubkey:  pub,
			Amount:          10 * 1e8,
		}
		v = append(v, d)
	}
	return v
}

// FuzzTx returns a slice of n Tx
func FuzzTx(n int) []*primitives.Tx {
	f := fuzz.New().NilChance(0)
	var v []*primitives.Tx
	for i := 0; i < n; i++ {
		d := new(primitives.Tx)
		f.Fuzz(d)
		k, _ := bls.RandKey()
		pubBytes := k.PublicKey().Marshal()
		copy(d.FromPublicKey[:], pubBytes)
		msg := d.SignatureMessage()
		sig := k.Sign(msg[:])
		copy(d.Signature[:], sig.Marshal())
		v = append(v, d)
	}
	return v
}
