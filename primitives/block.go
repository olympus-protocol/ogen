package primitives

import (
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

const (
	maxBlockSize = 1024 * 512 // 512 kilobytes
)

// Block is a block on the blockchain
type Block struct {
	Header            *BlockHeader
	Votes             []*MultiValidatorVote `ssz-max:"1099511627776"`
	Txs               []*Tx                 `ssz-max:"1099511627776"`
	Deposits          []*Deposit            `ssz-max:"1099511627776"`
	Exits             []*Exit               `ssz-max:"1099511627776"`
	VoteSlashings     []*VoteSlashing       `ssz-max:"1099511627776"`
	RANDAOSlashings   []*RANDAOSlashing     `ssz-max:"1099511627776"`
	ProposerSlashings []*ProposerSlashing   `ssz-max:"1099511627776"`
	GovernanceVotes   []*GovernanceVote     `ssz-max:"1099511627776"`
	Signature         []byte                `ssz-size:"96"`
	RandaoSignature   []byte                `ssz-size:"96"`
}

// Hash calculates the hash of the block.
func (b *Block) Hash() chainhash.Hash {
	return b.Header.Hash()
}

// GovernanceVoteMerkleRoot calculates the merkle root of the governance votes in the
// block.
func (b *Block) GovernanceVoteMerkleRoot() chainhash.Hash {
	return merkleRootGovernanceVotes(b.GovernanceVotes)
}

func merkleRootGovernanceVotes(votes []*GovernanceVote) chainhash.Hash {
	if len(votes) == 0 {
		return chainhash.Hash{}
	}
	if len(votes) == 1 {
		return votes[0].Hash()
	}
	mid := len(votes) / 2
	h1 := merkleRootGovernanceVotes(votes[:mid])
	h2 := merkleRootGovernanceVotes(votes[mid:])

	return chainhash.HashH(append(h1[:], h2[:]...))
}

// TransactionMerkleRoot calculates the merkle root of the transactions in the block.
func (b *Block) TransactionMerkleRoot() chainhash.Hash {
	return merkleRootTxs(b.Txs)
}

func merkleRootTxs(txs []*Tx) chainhash.Hash {
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

func merkleRootExits(exits []*Exit) chainhash.Hash {
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

func merkleRootDeposits(deposits []*Deposit) chainhash.Hash {
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

func merkleRootVotes(votes []*MultiValidatorVote) chainhash.Hash {
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

func merkleRootProposerSlashings(txs []*ProposerSlashing) chainhash.Hash {
	if len(txs) == 0 {
		return chainhash.Hash{}
	}
	if len(txs) == 1 {
		return txs[0].Hash()
	}
	mid := len(txs) / 2
	h1 := merkleRootProposerSlashings(txs[:mid])
	h2 := merkleRootProposerSlashings(txs[mid:])

	return chainhash.HashH(append(h1[:], h2[:]...))
}

// ProposerSlashingsRoot calculates the hash of the proposer slashings included in the block.
func (b *Block) ProposerSlashingsRoot() chainhash.Hash {
	return merkleRootProposerSlashings(b.ProposerSlashings)
}

func merkleRootRANDAOSlashings(txs []*RANDAOSlashing) chainhash.Hash {
	if len(txs) == 0 {
		return chainhash.Hash{}
	}
	if len(txs) == 1 {
		return txs[0].Hash()
	}
	mid := len(txs) / 2
	h1 := merkleRootRANDAOSlashings(txs[:mid])
	h2 := merkleRootRANDAOSlashings(txs[mid:])

	return chainhash.HashH(append(h1[:], h2[:]...))
}

// RANDAOSlashingsRoot calculates the merkle root of the RANDAO slashings included in the block.
func (b *Block) RANDAOSlashingsRoot() chainhash.Hash {
	return merkleRootRANDAOSlashings(b.RANDAOSlashings)
}

func merkleRootVoteSlashings(txs []*VoteSlashing) chainhash.Hash {
	if len(txs) == 0 {
		return chainhash.Hash{}
	}
	if len(txs) == 1 {
		return txs[0].Hash()
	}
	mid := len(txs) / 2
	h1 := merkleRootVoteSlashings(txs[:mid])
	h2 := merkleRootVoteSlashings(txs[mid:])

	return chainhash.HashH(append(h1[:], h2[:]...))
}

// VoteSlashingRoot calculates the merkle root of the vote slashings included in the block.
func (b *Block) VoteSlashingRoot() chainhash.Hash {
	return merkleRootVoteSlashings(b.VoteSlashings)
}

// Txs returns a slice of tx hashes.
func (b *Block) TxsHashes() []string {
	txs := make([]string, len(b.Txs))
	for i, tx := range b.Txs {
		txs[i] = tx.Hash().String()
	}
	return txs
}
