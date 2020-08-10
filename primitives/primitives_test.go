package primitives_test

import (
	fuzz "github.com/google/gofuzz"
	testdata "github.com/olympus-protocol/ogen/test"
	"github.com/olympus-protocol/ogen/utils/bitfield"
	"github.com/olympus-protocol/ogen/utils/chainhash"
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
	var deposits []*primitives.Deposit
	var exits []*primitives.Exit
	f.Fuzz(&deposits)
	f.Fuzz(&exits)

	f.NumElements(1000, 1000)
	var txs []*primitives.Tx
	f.Fuzz(&txs)

	f.NumElements(20, 20)
	var randaoSlash []*primitives.RANDAOSlashing
	f.Fuzz(&randaoSlash)

	f.NumElements(2, 2)
	var proposerSlash []*primitives.ProposerSlashing
	f.Fuzz(&proposerSlash)

	f.NumElements(128, 128)
	var governanceVotes []*primitives.GovernanceVote
	f.Fuzz(&governanceVotes)

	v := primitives.Block{
		Header:            blockheader,
		Votes:             fuzzMultiValidatorVote(32),
		Txs:               txs,
		Deposits:          deposits,
		Exits:             exits,
		VoteSlashings:     fuzzVoteSlashing(10),
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
	v.ParticipationBitfield = bitfield.NewBitlist(8)

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

	var currManagers [][20]byte
	f.NumElements(5, 5)
	f.Fuzz(&currManagers)

	var latesBlockHashes [][32]byte
	f.NumElements(64, 64)
	f.Fuzz(&latesBlockHashes)

	var proposerQueue, prevEpochVoteAssign, currEpochVoteAssign, nextPropQueue []uint64
	f.NumElements(100, 100)
	f.Fuzz(&proposerQueue)
	f.Fuzz(&prevEpochVoteAssign)
	f.Fuzz(&currEpochVoteAssign)
	f.Fuzz(&nextPropQueue)

	bl := bitfield.NewBitlist(5 * 8)

	v := primitives.State{
		CoinsState:                    cs,
		ValidatorRegistry:             valreg,
		LatestValidatorRegistryChange: latestRegistry,
		RANDAO:                        randao,
		NextRANDAO:                    nextrandao,
		Slot:                          slot,
		EpochIndex:                    epoch,
		ProposerQueue:                 proposerQueue,
		PreviousEpochVoteAssignments:  prevEpochVoteAssign,
		CurrentEpochVoteAssignments:   currEpochVoteAssign,
		NextProposerQueue:             nextPropQueue,
		JustificationBitfield:         justifbit,
		FinalizedEpoch:                finalepoch,
		LatestBlockHashes:             latesBlockHashes,
		JustifiedEpoch:                justified,
		JustifiedEpochHash:            justifiedepoch,
		CurrentEpochVotes:             fuzzAcceptedVoteInfo(10),
		PreviousJustifiedEpoch:        previousjustepoch,
		PreviousJustifiedEpochHash:    previousjustified,
		PreviousEpochVotes:            fuzzAcceptedVoteInfo(10),
		CurrentManagers:               currManagers,
		ManagerReplacement:            bl,
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

func Test_StateSerializeForInitialParams(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var genHash [32]byte
	f.Fuzz(&genHash)

	var currManagers [][20]byte
	f.NumElements(5, 5)
	f.Fuzz(&currManagers)

	val := make([]*primitives.Validator, 128)
	for i := range val {
		var pub [48]byte
		var payee [20]byte
		f.Fuzz(&pub)
		f.Fuzz(&payee)
		v := &primitives.Validator{
			Balance:          100,
			PubKey:           pub,
			PayeeAddress:     payee,
			Status:           primitives.StatusActive,
			FirstActiveEpoch: 0,
			LastActiveEpoch:  0,
		}
		val[i] = v
	}

	hash, err := testdata.PremineAddr.PublicKey().Hash()

	assert.NoError(t, err)

	is := &primitives.State{
		CoinsState: primitives.CoinsState{
			Balances: map[[20]byte]uint64{
				hash: 400 * 1000000,
			},
			Nonces: make(map[[20]byte]uint64),
		},
		ValidatorRegistry:             val,
		LatestValidatorRegistryChange: 0,
		RANDAO:                        chainhash.Hash{},
		NextRANDAO:                    chainhash.Hash{},
		Slot:                          0,
		EpochIndex:                    0,
		JustificationBitfield:         0,
		JustifiedEpoch:                0,
		FinalizedEpoch:                0,
		LatestBlockHashes:             make([][32]byte, 0),
		JustifiedEpochHash:            genHash,
		CurrentEpochVotes:             make([]*primitives.AcceptedVoteInfo, 0),
		PreviousJustifiedEpoch:        0,
		PreviousJustifiedEpochHash:    genHash,
		PreviousEpochVotes:            make([]*primitives.AcceptedVoteInfo, 0),
		CurrentManagers:               currManagers,
		VoteEpoch:                     0,
		VoteEpochStartSlot:            0,
		Governance: primitives.Governance{
			ReplaceVotes:   make(map[[20]byte]chainhash.Hash),
			CommunityVotes: make(map[chainhash.Hash]primitives.CommunityVoteData),
		},
		VotingState:        primitives.GovernanceStateActive,
		LastPaidSlot:       0,
		ManagerReplacement: bitfield.NewBitlist(5 * 8),
	}

	activeValidators := is.GetValidatorIndicesActiveAt(0)
	is.ProposerQueue = primitives.DetermineNextProposers(chainhash.Hash{}, activeValidators, &testdata.IntTestParams)
	is.NextProposerQueue = primitives.DetermineNextProposers(chainhash.Hash{}, activeValidators, &testdata.IntTestParams)
	is.CurrentEpochVoteAssignments = primitives.Shuffle(chainhash.Hash{}, activeValidators)
	is.PreviousEpochVoteAssignments = primitives.Shuffle(chainhash.Hash{}, activeValidators)

	ser, err := is.Marshal()
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
			ParticipationBitfield: bitfield.NewBitlist(10),
		}
		votes = append(votes, v)
	}
	return votes
}

func fuzzAcceptedVoteInfo(n int) []*primitives.AcceptedVoteInfo {
	var avInfo []*primitives.AcceptedVoteInfo
	for i := 0; i < n; i++ {
		f := fuzz.New().NilChance(0)
		d := new(primitives.VoteData)
		var sig [96]byte
		f.Fuzz(d)
		f.Fuzz(&sig)
		v := &primitives.AcceptedVoteInfo{
			Data:                  d,
			ParticipationBitfield: bitfield.NewBitlist(10),
			Proposer:              0,
			InclusionDelay:        0,
		}
		avInfo = append(avInfo, v)
	}
	return avInfo
}
