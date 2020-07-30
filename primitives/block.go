package primitives

import (
	"errors"

	ssz "github.com/ferranbt/fastssz"
	"github.com/golang/snappy"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

// ErrorBlockSize returns when the decompresed size of the block exceed MaxBlockSize
var ErrorBlockSize = errors.New("the block size is too big")

// MaxBlockSize defines the maximum bytes on a block object.
const MaxBlockSize = 1024 * 1024 * 2 // 2 MB

// Votes is the struct on the block that contains block votes.
type Votes struct {
	Votes []*MultiValidatorVote `ssz-max:"32"` // MaxVotesPerBlock
}

// Txs is the struct on the block that contains block txs.
type Txs struct {
	Txs []*Tx `ssz-max:"1000"` // MaxTxsPerBlock
}

// Deposits is the struct on the block that contains block deposits.
type Deposits struct {
	Deposits []*Deposit `ssz-max:"32"` // MaxDepositsPerBlock
}

// Exits is the struct on the block that contains block exits.
type Exits struct {
	Exits []*Exit `ssz-max:"32"` // MaxExitsPerBlock
}

// VoteSlashings is the struct on the block that contains block vote slashings.
type VoteSlashings struct {
	VoteSlashings []*VoteSlashing `ssz-max:"10"` // MaxVoteSlashingsPerBlock
}

// RANDAOSlashings is the struct on the block that contains block randao slashings.
type RANDAOSlashings struct {
	RANDAOSlashings []*RANDAOSlashing `ssz-max:"20"` // MaxRANDAOSlashingsPerBlock
}

// ProposerSlashings is the struct on the block that contains block proposer slashings.
type ProposerSlashings struct {
	ProposerSlashings []*ProposerSlashing `ssz-max:"2"` // MaxProposerSlashingsPerBlock
}

// GovernanceVotes is the struct on the block that contains block governance votes.
type GovernanceVotes struct {
	GovernanceVotes []*GovernanceVote `ssz-max:"128"` // MaxGovernanceVotesPerBlock
}

// Block is a block in the blockchain.
type Block struct {
	Header            *BlockHeader
	Votes             *Votes
	Txs               *Txs
	Deposits          *Deposits
	Exits             *Exits
	VoteSlashings     *VoteSlashings
	RANDAOSlashings   *RANDAOSlashings
	ProposerSlashings *ProposerSlashings
	GovernanceVotes   *GovernanceVotes
	Signature         [96]byte `ssz-size:"96"`
	RandaoSignature   [96]byte `ssz-size:"96"`
}

// Marshal encodes the block.
func (b *Block) Marshal() ([]byte, error) {
	bb, err := b.MarshalSSZ()
	if err != nil {
		return nil, err
	}
	if len(bb) > MaxBlockSize {
		return nil, ErrorBlockSize
	}
	return snappy.Encode(nil, bb), nil
}

// Unmarshal decodes the block.
func (b *Block) Unmarshal(bb []byte) error {
	d, err := snappy.Decode(nil, bb)
	if err != nil {
		return err
	}
	if len(d) > MaxBlockSize {
		return ErrorBlockSize
	}
	return b.UnmarshalSSZ(d)
}

// Hash calculates the hash of the block.
func (b *Block) Hash() chainhash.Hash {
	return b.Header.Hash()
}

// GovernanceVoteMerkleRoot calculates the merkle root of the governance votes in the block.
func (b *Block) GovernanceVoteMerkleRoot() chainhash.Hash {
	h, _ := ssz.HashWithDefaultHasher(b.GovernanceVotes)
	return h
}

// ExitMerkleRoot calculates the merkle root of the exits in the block.
func (b *Block) ExitMerkleRoot() chainhash.Hash {
	h, _ := ssz.HashWithDefaultHasher(b.Exits)
	return h
}

// DepositMerkleRoot calculates the merkle root of the deposits in the block.
func (b *Block) DepositMerkleRoot() chainhash.Hash {
	h, _ := ssz.HashWithDefaultHasher(b.Deposits)
	return h
}

// TransactionMerkleRoot calculates the merkle root of the transactions in the block.
func (b *Block) TransactionMerkleRoot() chainhash.Hash {
	h, _ := ssz.HashWithDefaultHasher(b.Txs)
	return h
}

// VotesMerkleRoot calculates the merkle root of the transactions in the block.
func (b *Block) VotesMerkleRoot() chainhash.Hash {
	h, _ := ssz.HashWithDefaultHasher(b.Votes)
	return h
}

// ProposerSlashingsRoot calculates the hash of the proposer slashings included in the block.
func (b *Block) ProposerSlashingsRoot() chainhash.Hash {
	h, _ := ssz.HashWithDefaultHasher(b.ProposerSlashings)
	return h
}

// RANDAOSlashingsRoot calculates the merkle root of the RANDAO slashings included in the block.
func (b *Block) RANDAOSlashingsRoot() chainhash.Hash {
	h, _ := ssz.HashWithDefaultHasher(b.RANDAOSlashings)
	return h
}

// VoteSlashingRoot calculates the merkle root of the vote slashings included in the block.
func (b *Block) VoteSlashingRoot() chainhash.Hash {
	h, _ := ssz.HashWithDefaultHasher(b.VoteSlashings)
	return h
}

// GetTxs returns a slice with tx hashes
func (b *Block) GetTxs() []string {
	txs := make([]string, len(b.Txs.Txs))
	for i, tx := range b.Txs.Txs {
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
