package primitives

import (
	"errors"

	"github.com/golang/snappy"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/prysmaticlabs/go-ssz"
)

// ErrorBlockSize returns when the decompresed size of the block exceed MaxBlockSize
var ErrorBlockSize = errors.New("the block size is too big")

// MaxBlockSize defines the maximum bytes on a block object.
const MaxBlockSize = 1024 * 1024 * 2 // 2 MB

// Block is a block in the blockchain.
type Block struct {
	Header            *BlockHeader
	Votes             []*MultiValidatorVote `ssz-max:"32"`   // MaxVotesPerBlock
	Txs               []*Tx                 `ssz-max:"1000"` // MaxTxsPerBlock
	Deposits          []*Deposit            `ssz-max:"32"`   // MaxDepositsPerBlock
	Exits             []*Exit               `ssz-max:"32"`   // MaxExitsPerBlock
	VoteSlashings     []*VoteSlashing       `ssz-max:"10"`   // MaxVoteSlashingsPerBlock
	RANDAOSlashings   []*RANDAOSlashing     `ssz-max:"20"`   // MaxRANDAOSlashingsPerBlock
	ProposerSlashings []*ProposerSlashing   `ssz-max:"2"`    // MaxProposerSlashingsPerBlock
	GovernanceVotes   []*GovernanceVote     `ssz-max:"128"`  // MaxGovernanceVotesPerBlock
	Signature         [96]byte              `ssz-size:"96"`
	RandaoSignature   [96]byte              `ssz-size:"96"`
}

// Marshal encodes the block.
func (b *Block) Marshal() ([]byte, error) {
	bb, err := b.MarshalSSZ()
	if err != nil {
		return nil, err
	}
	if len(bb) > MaxBlockHeaderBytes {
		return nil, ErrorBlockHeaderSize
	}
	return snappy.Encode(nil, bb), nil
}

// Unmarshal decodes the block.
func (b *Block) Unmarshal(bb []byte) error {
	d, err := snappy.Decode(nil, bb)
	if err != nil {
		return err
	}
	if len(d) > MaxBlockHeaderBytes {
		return ErrorBlockHeaderSize
	}
	return b.UnmarshalSSZ(d)
}

// Hash calculates the hash of the block.
func (b *Block) Hash() chainhash.Hash {
	return b.Header.Hash()
}

// GovernanceVoteMerkleRoot calculates the merkle root of the governance votes in the block.
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
	hash, _ := ssz.HashTreeRoot(b.Deposits)
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
