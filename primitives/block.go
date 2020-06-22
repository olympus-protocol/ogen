package primitives

import (
	"github.com/ferranbt/fastssz"
	"github.com/olympus-protocol/ogen/utils/chainhash"
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

	ssz.Marshaler
	ssz.Unmarshaler
}

func (b *Block) Marshal() ([]byte, error) {
	return b.MarshalSSZ()
}

func (b *Block) Unmarshal(bb []byte) error {
	return b.Unmarshal(bb)
}

// Hash calculates the hash of the block.
func (b *Block) Hash() (chainhash.Hash, error) {
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

// GovernanceVoteMerkleRoot calculates the merkle root of the governance votes in the block.
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
func (b *Block) ExitMerkleRoot() (chainhash.Hash, error) {
	return merkleRootDeposits(b.Deposits)
}

func merkleRootExits(exits []Exit) (chainhash.Hash, error) {
	if len(exits) == 0 {
		return chainhash.Hash{}, nil
	}
	if len(exits) == 1 {
		return exits[0].Hash()
	}
	mid := len(exits) / 2
	h1, err := merkleRootExits(exits[:mid])
	if err != nil {
		return chainhash.Hash{}, err
	}
	h2, err := merkleRootExits(exits[mid:])
	if err != nil {
		return chainhash.Hash{}, err
	}
	return chainhash.HashH(append(h1[:], h2[:]...)), nil
}

// DepositMerkleRoot calculates the merkle root of the deposits in the block.
func (b *Block) DepositMerkleRoot() (chainhash.Hash, error) {
	return merkleRootDeposits(b.Deposits)
}

func merkleRootDeposits(deposits []Deposit) (chainhash.Hash, error) {
	if len(deposits) == 0 {
		return chainhash.Hash{}, nil
	}
	if len(deposits) == 1 {
		return deposits[0].Hash()
	}
	mid := len(deposits) / 2
	h1, err := merkleRootDeposits(deposits[:mid])
	if err != nil {
		return chainhash.Hash{}, err
	}
	h2, err := merkleRootDeposits(deposits[mid:])
	if err != nil {
		return chainhash.Hash{}, err
	}
	return chainhash.HashH(append(h1[:], h2[:]...)), nil
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

// GetTxs returns
func (b *Block) GetTxs() []string {
	txs := make([]string, len(b.Txs))
	for i, tx := range b.Txs {
		txs[i] = tx.Hash().String()
	}
	return txs
}
