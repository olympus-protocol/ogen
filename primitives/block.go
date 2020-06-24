package primitives

import (
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/prysmaticlabs/go-ssz"
)

const (
	maxBlockSize = 1024 * 512 // 512 kilobytes
)

// Block is a block in the blockchain.
type Block struct {
	Header            BlockHeader
	Votes             []MultiValidatorVote
	Txs               []Tx
	Deposits          []Deposit
	Exits             []Exit
	VoteSlashings     []VoteSlashing
	RANDAOSlashings   []RANDAOSlashing
	ProposerSlashings []ProposerSlashing
	GovernanceVotes   []GovernanceVote
	Signature         []byte
	RandaoSignature   []byte
}

// Marshal encodes the block.
func (b *Block) Marshal() ([]byte, error) {
	return ssz.Marshal(b)
}

// Unmarshal decodes the block.
func (b *Block) Unmarshal(by []byte) error {
	return ssz.Unmarshal(by, b)
}

// Hash calculates the hash of the block.
func (b *Block) Hash() chainhash.Hash {
	return b.Header.Hash()
}

func merkleRootGovernanceVotes(votes []GovernanceVote) chainhash.Hash {
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

// GovernanceVoteMerkleRoot calculates the merkle root of the governance votes in the
// block.
func (b *Block) GovernanceVoteMerkleRoot() chainhash.Hash {
	return merkleRootGovernanceVotes(b.GovernanceVotes)
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

func merkleRootProposerSlashings(txs []ProposerSlashing) chainhash.Hash {
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

func merkleRootRANDAOSlashings(txs []RANDAOSlashing) chainhash.Hash {
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

func merkleRootVoteSlashings(txs []VoteSlashing) chainhash.Hash {
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

// GetTxs returns a slice with tx hashes
func (b *Block) GetTxs() []string {
	txs := make([]string, len(b.Txs))
	for i, tx := range b.Txs {
		txs[i] = tx.Hash().String()
	}
	return txs
}

// SerializedTx return a slice serialized transactions that include one of the passed accounts.
func (b *Block) SerializedTx(accounts map[[20]byte]struct{}) []byte {
	return []byte{}
}

// SerializedEpochs return a slice serialized epochs that include one of the passed public keys.
func (b *Block) SerializedEpochs(accounts map[[48]byte]struct{}) []byte {
	return []byte{}
}
