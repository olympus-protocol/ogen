package p2p_test

import (
	"bytes"
	fuzz "github.com/google/gofuzz"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/prysmaticlabs/go-bitfield"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/olympus-protocol/ogen/p2p"
)

func Test_MessageHeaderSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v p2p.MessageHeader
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc p2p.MessageHeader
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_MsgGetAddrSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v p2p.MsgGetAddr
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc p2p.MsgGetAddr
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_MsgAddrSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0).NumElements(32, 32)
	var v p2p.MsgAddr
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc p2p.MsgAddr
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_MsgGetBlocksSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0).NumElements(64, 64)
	var v p2p.MsgGetBlocks
	f.Fuzz(&v)
	ser, err := v.Marshal()
	assert.NoError(t, err)
	var desc p2p.MsgGetBlocks
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_MsgVersionSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v p2p.MsgVersion
	f.Fuzz(&v)

	ser, err := v.Marshal()
	assert.NoError(t, err)

	var desc p2p.MsgVersion
	err = desc.Unmarshal(ser)
	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_MsgBlocksSerialize(t *testing.T) {
	v := p2p.MsgBlocks{
		Blocks: fuzzBlock(32),
	}

	ser, err := v.Marshal()

	assert.NoError(t, err)

	var desc p2p.MsgBlocks

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func Test_MsgWithHeaderSerialize(t *testing.T) {
	f := fuzz.New().NilChance(0)
	var v p2p.MsgVersion
	f.Fuzz(&v)

	buf := bytes.NewBuffer([]byte{})
	err := p2p.WriteMessage(buf, &v, 333)
	assert.NoError(t, err)

	msg, err := p2p.ReadMessage(buf, 333)
	assert.NoError(t, err)

	assert.Equal(t, msg.(*p2p.MsgVersion), &v)
}

func fuzzBlock(n int) []*primitives.Block {
	var blocks []*primitives.Block
	for i := 0; i < n; i++ {
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

		f.NumElements(100, 100)
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

		v := &primitives.Block{
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
		blocks = append(blocks, v)
	}
	return blocks
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
