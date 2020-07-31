package p2p_test

import (
	"bytes"
	fuzz "github.com/google/gofuzz"
	"github.com/olympus-protocol/ogen/primitives"
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
	// TODO fix weird behaviour, when using 2 or more blocks, it doesn't work.
	v := p2p.MsgBlocks{
		Blocks: fuzzedBlock(1),
	}

	ser, err := v.Marshal()

	assert.NoError(t, err)

	var desc p2p.MsgBlocks

	err = desc.Unmarshal(ser)

	assert.NoError(t, err)

	assert.Equal(t, v, desc)
}

func fuzzedBlock(n int) []*primitives.Block {
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
		votes := new(primitives.Votes)
		deposits := new(primitives.Deposits)
		exits := new(primitives.Exits)
		f.Fuzz(votes)

		f.Fuzz(deposits)
		f.Fuzz(exits)

		f.NumElements(9000, 9000)
		txs := new(primitives.Txs)
		f.Fuzz(txs)

		f.NumElements(10, 10)
		votesSlash := new(primitives.VoteSlashings)
		f.Fuzz(votesSlash)

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
			Votes:             votes,
			Txs:               txs,
			Deposits:          deposits,
			Exits:             exits,
			VoteSlashings:     votesSlash,
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
