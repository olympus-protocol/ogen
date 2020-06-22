package primitives

import (
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
}

// Hash calculates the hash of the block.
func (b *Block) Hash() (chainhash.Hash, error) {
	return b.Header.Hash()
}

func merkleRootGovernanceVotes(votes []GovernanceVote) (chainhash.Hash, error) {
	if len(votes) == 0 {
		return chainhash.Hash{}, nil
	}
	if len(votes) == 1 {
		return votes[0].Hash()
	}
	mid := len(votes) / 2
	h1, err := merkleRootGovernanceVotes(votes[:mid])
	if err != nil {
		return chainhash.Hash{}, err
	}
	h2, err := merkleRootGovernanceVotes(votes[mid:])
	if err != nil {
		return chainhash.Hash{}, err
	}
	return chainhash.HashH(append(h1[:], h2[:]...)), nil
}

// GovernanceVoteMerkleRoot calculates the merkle root of the governance votes in the block.
func (b *Block) GovernanceVoteMerkleRoot() (chainhash.Hash, error) {
	return merkleRootGovernanceVotes(b.GovernanceVotes)
}

func merkleRootTxs(txs []Tx) (chainhash.Hash, error) {
	if len(txs) == 0 {
		return chainhash.Hash{}, nil
	}
	if len(txs) == 1 {
		return txs[0].Hash()
	}
	mid := len(txs) / 2
	h1, err := merkleRootTxs(txs[:mid])
	if err != nil {
		return chainhash.Hash{}, err
	}
	h2, err := merkleRootTxs(txs[mid:])
	if err != nil {
		return chainhash.Hash{}, err
	}

	return chainhash.HashH(append(h1[:], h2[:]...)), nil
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
func (b *Block) TransactionMerkleRoot() (chainhash.Hash, error) {
	return merkleRootTxs(b.Txs)
}

func merkleRootVotes(votes []MultiValidatorVote) (chainhash.Hash, error) {
	if len(votes) == 0 {
		return chainhash.Hash{}, nil
	}
	if len(votes) == 1 {
		return votes[0].Hash()
	}
	mid := len(votes) / 2
	h1, err := merkleRootVotes(votes[:mid])
	if err != nil {
		return chainhash.Hash{}, err
	}
	h2, err := merkleRootVotes(votes[mid:])
	if err != nil {
		return chainhash.Hash{}, err
	}
	return chainhash.HashH(append(h1[:], h2[:]...)), nil
}

// VotesMerkleRoot calculates the merkle root of the transactions in the block.
func (b *Block) VotesMerkleRoot() (chainhash.Hash, error) {
	return merkleRootVotes(b.Votes)
}

func merkleRootProposerSlashings(txs []ProposerSlashing) (chainhash.Hash, error) {
	if len(txs) == 0 {
		return chainhash.Hash{}, nil
	}
	if len(txs) == 1 {
		return txs[0].Hash()
	}
	mid := len(txs) / 2
	h1, err := merkleRootProposerSlashings(txs[:mid])
	if err != nil {
		return chainhash.Hash{}, err
	}
	h2, err := merkleRootProposerSlashings(txs[mid:])
	if err != nil {
		return chainhash.Hash{}, err
	}
	return chainhash.HashH(append(h1[:], h2[:]...)), nil
}

// ProposerSlashingsRoot calculates the hash of the proposer slashings included in the block.
func (b *Block) ProposerSlashingsRoot() (chainhash.Hash, error) {
	return merkleRootProposerSlashings(b.ProposerSlashings)
}

func merkleRootRANDAOSlashings(txs []RANDAOSlashing) (chainhash.Hash, error) {
	if len(txs) == 0 {
		return chainhash.Hash{}, nil
	}
	if len(txs) == 1 {
		return txs[0].Hash()
	}
	mid := len(txs) / 2
	h1, err := merkleRootRANDAOSlashings(txs[:mid])
	if err != nil {
		return chainhash.Hash{}, err
	}
	h2, err := merkleRootRANDAOSlashings(txs[mid:])
	if err != nil {
		return chainhash.Hash{}, err
	}

	return chainhash.HashH(append(h1[:], h2[:]...)), nil
}

// RANDAOSlashingsRoot calculates the merkle root of the RANDAO slashings included in the block.
func (b *Block) RANDAOSlashingsRoot() (chainhash.Hash, error) {
	return merkleRootRANDAOSlashings(b.RANDAOSlashings)
}

func merkleRootVoteSlashings(txs []VoteSlashing) (chainhash.Hash, error) {
	if len(txs) == 0 {
		return chainhash.Hash{}, nil
	}
	if len(txs) == 1 {
		return txs[0].Hash()
	}
	mid := len(txs) / 2
	h1, err := merkleRootVoteSlashings(txs[:mid])
	if err != nil {
		return chainhash.Hash{}, err
	}
	h2, err := merkleRootVoteSlashings(txs[mid:])
	if err != nil {
		return chainhash.Hash{}, err
	}

	return chainhash.HashH(append(h1[:], h2[:]...)), nil
}

// VoteSlashingRoot calculates the merkle root of the vote slashings included in the block.
func (b *Block) VoteSlashingRoot() (chainhash.Hash, error) {
	return merkleRootVoteSlashings(b.VoteSlashings)
}

// GetTxs returns
func (b *Block) GetTxs() ([]string, error) {
	txs := make([]string, len(b.Txs))
	for i, tx := range b.Txs {
		hash, err := tx.Hash()
		if err != nil {
			return nil, err
		}
		txs[i] = hash.String()
	}
	return txs, nil
}
