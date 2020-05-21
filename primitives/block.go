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
	Deposits        []Deposit
	Exits           []Exit
	Signature       []byte
	RandaoSignature []byte
}

// Hash calculates the hash of the block.
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

// ExitMerkleRoot calculates the merkle root of the exits in the block.
func (b *Block) ExitMerkleRoot() chainhash.Hash {
	return merkleRootDeposits(b.Deposits)
}

func merkleRootExits(exits []Exit) chainhash.Hash {
	if len(exits) == 0 {
		return chainhash.Hash{}
	}
	if len(exits) == 1 {
		return exits[0].Hash()
	}
	mid := len(exits) / 2
	h1 := merkleRootExits(exits[:mid])
	h2 := merkleRootExits(exits[mid:])

	return chainhash.HashH(append(h1[:], h2[:]...))
}

// DepositMerkleRoot calculates the merkle root of the deposits in the block.
func (b *Block) DepositMerkleRoot() chainhash.Hash {
	return merkleRootDeposits(b.Deposits)
}

func merkleRootDeposits(deposits []Deposit) chainhash.Hash {
	if len(deposits) == 0 {
		return chainhash.Hash{}
	}
	if len(deposits) == 1 {
		return deposits[0].Hash()
	}
	mid := len(deposits) / 2
	h1 := merkleRootDeposits(deposits[:mid])
	h2 := merkleRootDeposits(deposits[mid:])

	return chainhash.HashH(append(h1[:], h2[:]...))
}

// TransactionMerkleRoot calculates the merkle root of the transactions in the block.
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

// VotesMerkleRoot calculates the merkle root of the transactions in the block.
func (b *Block) VotesMerkleRoot() chainhash.Hash {
	return merkleRootVotes(b.Votes)
}

// Encode encodes the block to the given writer.
func (b *Block) Encode(w io.Writer) error {
	err := b.Header.Serialize(w)
	if err != nil {
		return err
	}
	err = serializer.WriteVarInt(w, uint64(len(b.Txs)))
	if err != nil {
		return err
	}
	for _, tx := range b.Txs {
		err := tx.Encode(w)
		if err != nil {
			return err
		}
	}
	err = serializer.WriteVarInt(w, uint64(len(b.Votes)))
	if err != nil {
		return err
	}
	for _, vote := range b.Votes {
		err := vote.Serialize(w)
		if err != nil {
			return err
		}
	}

	err = serializer.WriteVarInt(w, uint64(len(b.Deposits)))
	if err != nil {
		return err
	}
	for _, deposit := range b.Deposits {
		err := deposit.Encode(w)
		if err != nil {
			return err
		}
	}
	err = serializer.WriteVarInt(w, uint64(len(b.Exits)))
	if err != nil {
		return err
	}
	for _, exit := range b.Exits {
		err := exit.Encode(w)
		if err != nil {
			return err
		}
	}
	if err := serializer.WriteVarBytes(w, b.Signature); err != nil {
		return err
	}
	if err := serializer.WriteVarBytes(w, b.RandaoSignature); err != nil {
		return err
	}
	return nil
}

// Decode decodes the block from the given reader.
func (b *Block) Decode(r io.Reader) error {
	err := b.Header.Deserialize(r)
	if err != nil {
		return err
	}
	txCount, err := serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	b.Txs = make([]Tx, txCount)
	for i := uint64(0); i < txCount; i++ {
		err := b.Txs[i].Decode(r)
		if err != nil {
			return err
		}
	}
	voteCount, err := serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	b.Votes = make([]MultiValidatorVote, voteCount)
	for i := range b.Votes {
		err := b.Votes[i].Deserialize(r)
		if err != nil {
			return err
		}
	}
	depositCount, err := serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	b.Deposits = make([]Deposit, depositCount)
	for i := range b.Deposits {
		err := b.Deposits[i].Decode(r)
		if err != nil {
			return err
		}
	}
	exitCount, err := serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	b.Exits = make([]Exit, exitCount)
	for i := range b.Exits {
		err := b.Exits[i].Decode(r)
		if err != nil {
			return err
		}
	}
	sig, err := serializer.ReadVarBytes(r)
	if err != nil {
		return err
	}
	randaoSig, err := serializer.ReadVarBytes(r)
	if err != nil {
		return err
	}
	b.Signature = sig
	b.RandaoSignature = randaoSig
	return nil
}
