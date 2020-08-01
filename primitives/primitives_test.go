package primitives_test

import (
	fuzz "github.com/google/gofuzz"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/prysmaticlabs/go-bitfield"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/olympus-protocol/ogen/primitives"
)

func Test_BlockHeaderSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v primitives.BlockHeader
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.BlockHeader
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_BlockSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)

	blockheader := new(primitives.BlockHeader)
	sig := [96]byte{}
	rsig := [96]byte{}

	f.Fuzz(blockheader)
	f.Fuzz(&sig)
	f.Fuzz(&rsig)

	f.NumElements(32, 32)
	deposits := new(primitives.Deposits)
	exits := new(primitives.Exits)
	f.Fuzz(deposits)
	f.Fuzz(exits)

	f.NumElements(1000, 1000)
	txs := new(primitives.Txs)
	f.Fuzz(txs)

	f.NumElements(20, 20)
	randaoSlash := new(primitives.RANDAOSlashings)
	f.Fuzz(randaoSlash)

	f.NumElements(2, 2)
	proposerSlash := new(primitives.ProposerSlashings)
	f.Fuzz(proposerSlash)

	f.NumElements(128, 128)
	governanceVotes := new(primitives.GovernanceVotes)
	f.Fuzz(governanceVotes)

	v := primitives.Block{
		Header:            blockheader,
		Votes:             &primitives.Votes{Votes: fuzzMultiValidatorVote(32)},
		Txs:               txs,
		Deposits:          deposits,
		Exits:             exits,
		VoteSlashings:     &primitives.VoteSlashings{VoteSlashings: fuzzVoteSlashing(10)},
		RANDAOSlashings:   randaoSlash,
		ProposerSlashings: proposerSlash,
		GovernanceVotes:   governanceVotes,
		Signature:         sig,
		RandaoSignature:   rsig,
	}

	ser, err := v.Marshal()
	assert.NoError(t, err)
	var desc primitives.Block
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_DepositSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v primitives.Deposit
	f.Fuzz(&v)
	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.Deposit
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_DeposiDatatSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v primitives.DepositData
	f.Fuzz(&v)
	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.DepositData
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_ExitSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v primitives.Exit
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.Exit
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_EpochReceiptSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v primitives.EpochReceipt
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.EpochReceipt
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_CommunityVoteDataSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0).NumElements(5, 5)
	var v primitives.CommunityVoteData
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.CommunityVoteData
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_ReplacementVoteSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v primitives.ReplacementVotes
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.ReplacementVotes
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_CommunityVoteSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0).NumElements(5, 5)
	var v primitives.CommunityVoteData
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.CommunityVoteData
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_GovernanceVoteSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v primitives.GovernanceVote
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.GovernanceVote
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_VoteSlashingSerialize(t *testing.T) {
	v := fuzzVoteSlashing(1)
	ser, err := v[0].Marshal()
	assert.NoError(t, err)

	desc := new(primitives.VoteSlashing)
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v[0], desc)
}

func Test_RANDAOSlashingSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v primitives.RANDAOSlashing
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)
	var desc primitives.RANDAOSlashing
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_ProposerSlashingSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v primitives.ProposerSlashing
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.ProposerSlashing
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_ValidatorSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v primitives.Validator
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.Validator
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_AcceptedVoteInfoSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v primitives.AcceptedVoteInfo
	f.Fuzz(&v)
	v.ParticipationBitfield = bitfield.NewBitlist(uint64(2042))

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.AcceptedVoteInfo
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_VoteDataSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v primitives.VoteData
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.VoteData
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_SingleValidatorVoteSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v primitives.SingleValidatorVote
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.SingleValidatorVote
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_MultiValidatorVoteSerialize(t *testing.T) {
	v := fuzzMultiValidatorVote(1)
	ser, err := v[0].Marshal()
	assert.NoError(t, err)

	desc := new(primitives.MultiValidatorVote)
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v[0], desc)
}

func Test_CoinStateSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0).NumElements(10000, 10000)
	balances := map[[20]byte]uint64{}
	nonces := map[[20]byte]uint64{}
	f.Fuzz(&balances)
	f.Fuzz(&nonces)

	v := primitives.CoinsState{
		Balances: balances,
		Nonces:   nonces,
	}

	ser, err := v.Marshal()

	assert.NoError(t, err)

	var desc primitives.CoinsState
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

// Is not possible to test against equal states because of slice ordering. TODO find a solution
func Test_GovernanceSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0).NumElements(5, 5)

	replace := map[[20]byte]chainhash.Hash{}
	community := map[chainhash.Hash]primitives.CommunityVoteData{}
	f.Fuzz(&replace)
	f.Fuzz(&community)

	v := primitives.Governance{
		ReplaceVotes:   replace,
		CommunityVotes: community,
	}

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.Governance
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)
}

// Is not possible to test against equal states because of slice ordering. TODO find a solution
func Test_StateSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0).NumElements(100, 100)
	balances := map[[20]byte]uint64{}
	nonces := map[[20]byte]uint64{}
	f.Fuzz(&balances)
	f.Fuzz(&nonces)

	cs := primitives.CoinsState{
		Balances: balances,
		Nonces:   nonces,
	}

	f.NilChance(0).NumElements(5, 5)

	replace := map[[20]byte]chainhash.Hash{}
	community := map[chainhash.Hash]primitives.CommunityVoteData{}
	f.Fuzz(&replace)
	f.Fuzz(&community)

	gs := primitives.Governance{
		ReplaceVotes:   replace,
		CommunityVotes: community,
	}
	var randao, nextrandao, justifiedepoch, previousjustified, governance chainhash.Hash
	f.Fuzz(&randao)
	f.Fuzz(&nextrandao)
	f.Fuzz(&justifiedepoch)
	f.Fuzz(&previousjustified)
	f.Fuzz(&governance)

	var latestRegistry, slot, epoch, justifbit, finalepoch, justified uint64
	var previousjustepoch, voteepoch, votestartslot, votestate, lastpayedslot uint64

	f.Fuzz(&latestRegistry)
	f.Fuzz(&slot)
	f.Fuzz(&epoch)
	f.Fuzz(&justifbit)
	f.Fuzz(&finalepoch)
	f.Fuzz(&justifiedepoch)
	f.Fuzz(&previousjustepoch)
	f.Fuzz(&voteepoch)
	f.Fuzz(&votestartslot)
	f.Fuzz(&votestate)
	f.Fuzz(&lastpayedslot)
	f.Fuzz(&justified)

	var valreg []*primitives.Validator
	f.NumElements(100, 100)
	f.Fuzz(&valreg)

	v := primitives.State{
		CoinsState:                    cs,
		ValidatorRegistry:             valreg,
		LatestValidatorRegistryChange: latestRegistry,
		RANDAO:                        randao,
		NextRANDAO:                    nextrandao,
		Slot:                          slot,
		EpochIndex:                    epoch,
		ProposerQueue:                 nil,
		PreviousEpochVoteAssignments:  nil,
		CurrentEpochVoteAssignments:   nil,
		NextProposerQueue:             nil,
		JustificationBitfield:         justifbit,
		FinalizedEpoch:                finalepoch,
		LatestBlockHashes:             nil,
		JustifiedEpoch:                justified,
		JustifiedEpochHash:            justifiedepoch,
		CurrentEpochVotes:             nil,
		PreviousJustifiedEpoch:        previousjustepoch,
		PreviousJustifiedEpochHash:    previousjustified,
		PreviousEpochVotes:            nil,
		CurrentManagers:               nil,
		ManagerReplacement:            nil,
		Governance:                    gs,
		VoteEpoch:                     voteepoch,
		VoteEpochStartSlot:            votestartslot,
		VotingState:                   votestate,
		LastPaidSlot:                  lastpayedslot,
	}

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.State
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)
}

func Test_TransactionSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v primitives.Tx
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc primitives.Tx
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func fuzzVoteSlashing(n int) []*primitives.VoteSlashing {
	var votes []*primitives.VoteSlashing
	for i := 0; i < n; i++ {
		v := &primitives.VoteSlashing{
			Vote1: fuzzMultiValidatorVote(1)[0],
			Vote2: fuzzMultiValidatorVote(1)[0],
		}
		votes = append(votes, v)
	}
	return votes
}

func fuzzMultiValidatorVote(n int) []*primitives.MultiValidatorVote {
	var votes []*primitives.MultiValidatorVote
	for i := 0; i < n; i++ {
		f := fuzz.New().NilChance(0)
		d := new(primitives.VoteData)
		var sig [96]byte
		f.Fuzz(d)
		f.Fuzz(&sig)
		v := &primitives.MultiValidatorVote{
			Data:                  d,
			Sig:                   sig,
			ParticipationBitfield: bitfield.NewBitlist(uint64(2042)),
		}
		votes = append(votes, v)
	}
	return votes
}
