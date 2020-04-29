package primitives

import (
	"io"

	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
)

const (
	maxBlockSize = 1024 * 512 // 512 kilobytes
)

// Block is a block in the blockchain.
type Block struct {
	Header          BlockHeader
	Votes           []MultiValidatorVote
	Txs             []Tx
	Signature       [96]byte
	RandaoSignature [96]byte
}

func (b *Block) Hash() chainhash.Hash {
	return b.Header.Hash()
}

func merkleRootTxs(txs []Tx) chainhash.Hash {
	if len(txs) == 0 {
		return chainhash.Hash{}
	}
	if len(txs) == 1 {
		return txs[0].Hash()
	}
	mid := len(txs) / 2
	h1 := merkleRootTxs(txs[:mid])
	h2 := merkleRootTxs(txs[mid:])

	return chainhash.HashH(append(h1[:], h2[:]...))
}

func (b *Block) TransactionMerkleRoot() chainhash.Hash {
	return merkleRootTxs(b.Txs)
}

func merkleRootVotes(votes []MultiValidatorVote) chainhash.Hash {
	if len(votes) == 0 {
		return chainhash.Hash{}
	}
	if len(votes) == 1 {
		return votes[0].Hash()
	}
	mid := len(votes) / 2
	h1 := merkleRootVotes(votes[:mid])
	h2 := merkleRootVotes(votes[mid:])

	return chainhash.HashH(append(h1[:], h2[:]...))
}

func (b *Block) VotesMerkleRoot() chainhash.Hash {
	return merkleRootVotes(b.Votes)
}

func (m *Block) Encode(w io.Writer) error {
	err := m.Header.Serialize(w)
	if err != nil {
		return err
	}
	err = serializer.WriteVarInt(w, uint64(len(m.Txs)))
	if err != nil {
		return err
	}
	for _, tx := range m.Txs {
		err := tx.Encode(w)
		if err != nil {
			return err
		}
	}
	err = serializer.WriteVarInt(w, uint64(len(m.Votes)))
	if err != nil {
		return err
	}
	for _, vote := range m.Votes {
		err := vote.Serialize(w)
		if err != nil {
			return err
		}
	}
	err = serializer.WriteElements(w, m.Signature, m.RandaoSignature)
	if err != nil {
		return err
	}
	return nil
}

func (m *Block) Decode(r io.Reader) error {
	err := m.Header.Deserialize(r)
	if err != nil {
		return err
	}
	txCount, err := serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	m.Txs = make([]Tx, txCount)
	for i := uint64(0); i < txCount; i++ {
		err := m.Txs[i].Decode(r)
		if err != nil {
			return err
		}
	}
	voteCount, err := serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	m.Votes = make([]MultiValidatorVote, voteCount)
	for i := range m.Votes {
		err := m.Votes[i].Deserialize(r)
		if err != nil {
			return err
		}
	}
	err = serializer.ReadElements(r, &m.Signature, &m.RandaoSignature)
	if err != nil {
		return err
	}
	return nil
}
