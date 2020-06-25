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

// GovernanceVoteMerkleRoot calculates the merkle root of the governance votes in the
// block.
func (b *Block) GovernanceVoteMerkleRoot() chainhash.Hash {
	hash, _ := ssz.HashTreeRoot(b.GovernanceVotes)
	return chainhash.Hash(hash)
}

// ExitMerkleRoot calculates the merkle root of the exits in the block.
func (b *Block) ExitMerkleRoot() chainhash.Hash {
	hash, _ := ssz.HashTreeRoot(b.Exits)
	return chainhash.Hash(hash)
}

// DepositMerkleRoot calculates the merkle root of the deposits in the block.
func (b *Block) DepositMerkleRoot() chainhash.Hash {
	hash, _ := ssz.HashTreeRoot(b.DepositMerkleRoot)
	return chainhash.Hash(hash)
}

// TransactionMerkleRoot calculates the merkle root of the transactions in the block.
func (b *Block) TransactionMerkleRoot() chainhash.Hash {
	hash, _ := ssz.HashTreeRoot(b.Txs)
	return chainhash.Hash(hash)
}

// VotesMerkleRoot calculates the merkle root of the transactions in the block.
func (b *Block) VotesMerkleRoot() chainhash.Hash {
	hash, _ := ssz.HashTreeRoot(b.Votes)
	return chainhash.Hash(hash)
}

// ProposerSlashingsRoot calculates the hash of the proposer slashings included in the block.
func (b *Block) ProposerSlashingsRoot() chainhash.Hash {
	hash, _ := ssz.HashTreeRoot(b.ProposerSlashings)
	return chainhash.Hash(hash)
}

// RANDAOSlashingsRoot calculates the merkle root of the RANDAO slashings included in the block.
func (b *Block) RANDAOSlashingsRoot() chainhash.Hash {
	hash, _ := ssz.HashTreeRoot(b.RANDAOSlashings)
	return chainhash.Hash(hash)
}

// VoteSlashingRoot calculates the merkle root of the vote slashings included in the block.
func (b *Block) VoteSlashingRoot() chainhash.Hash {
	hash, _ := ssz.HashTreeRoot(b.VoteSlashings)
	return chainhash.Hash(hash)
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
