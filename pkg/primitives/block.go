package primitives

import (
	"github.com/golang/snappy"
	"github.com/olympus-protocol/ogen/pkg/burnproof"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
)

// MaxBlockSize defines the maximum bytes on a block object.
const MaxBlockSize = BlockHeaderSize + 96 + 96 +
	(MaxMultiValidatorVoteSize * MaxVotesPerBlock) +
	(DepositSize * MaxDepositsPerBlock) +
	(ExitSize * MaxExitsPerBlock) +
	(PartialExitsSize * MaxPartialExitsPerBlock) +
	(MaxCoinProofSize * MaxCoinProofsPerBlock) +
	(MaxExecutionSize * MaxExecutionsPerBlock) +
	(TxSize * MaxTxsPerBlock) +
	(ProposerSlashingSize * MaxProposerSlashingsPerBlock) +
	(MaxVotesSlashingSize * MaxVoteSlashingsPerBlock) +
	(RANDAOSlashingSize * MaxRANDAOSlashingsPerBlock) +
	(MaxGovernanceVoteSize * MaxGovernanceVotesPerBlock) +
	(MaxMultiSignatureTxSize * MaxMultiSignatureTxsOnBlock)

// Block is a block in the blockchain.
type Block struct {
	Header            *BlockHeader                        // 																		= 500 bytes
	Signature         [96]byte                            // 																		= 96 bytes
	RandaoSignature   [96]byte                            // 																		= 96
	Votes             []*MultiValidatorVote               `ssz-max:"16"`    // MaxVotesPerBlock 					16 * 6479 		= 103664 bytes
	Deposits          []*Deposit                          `ssz-max:"32"`    // MaxDepositsPerBlock 					32 * 308 		= 9856 bytes
	Exits             []*Exit                             `ssz-max:"32"`    // MaxExitsPerBlock     				32 * 192 		= 6144 bytes
	PartialExit       []*PartialExit                      `ssz-max:"32"`    // MaxPartialExitsPerBlock            	32 * 200		= 6400 bytes
	CoinProofs        []*burnproof.CoinsProofSerializable `ssz-max:"64"`    // MaxCoinProofsPerBlock 				64 * 2317   	= 148288 bytes
	Executions        []*Execution                        `ssz-max:"128"`   // MaxExecutionsPerBlock				128 * 7168      = 917504 bytes
	Txs               []*Tx                               `ssz-max:"5000"`  // MaxTxsPerBlock						5000 * 188  	= 940000 bytes
	ProposerSlashings []*ProposerSlashing                 `ssz-max:"2"`     // MaxProposerSlashingsPerBlock 		2 * 1240 		= 2480 bytes
	VoteSlashings     []*VoteSlashing                     `ssz-max:"5"`     // MaxVoteSlashingsPerBlock				10 * 12948 		= 129480 bytes
	RANDAOSlashings   []*RANDAOSlashing                   `ssz-max:"20"`    // MaxRANDAOSlashingsPerBlock  			20 * 152 		= 3040 bytes
	GovernanceVotes   []*GovernanceVote                   `ssz-max:"128"`   // MaxGovernanceVotesPerBlock			128 * 264		= 33792 bytes
	MultiSignatureTxs []*MultiSignatureTx                 `ssz-max:"8"`     // MaxMultiSignatureTxsOnBlock			8 * 2231      	= 17848 bytes
}

// Marshal encodes the block.
func (b *Block) Marshal() ([]byte, error) {
	ser, err := b.MarshalSSZ()
	if err != nil {
		return nil, err
	}
	return snappy.Encode(nil, ser), nil
}

// Unmarshal decodes the block.
func (b *Block) Unmarshal(bb []byte) error {
	des, err := snappy.Decode(nil, bb)
	if err != nil {
		return err
	}
	return b.UnmarshalSSZ(des)
}

// Hash calculates the hash of the block.
func (b *Block) Hash() chainhash.Hash {
	return b.Header.Hash()
}

// GovernanceVoteMerkleRoot calculates the merkle root of the GovernanceVotes in the block.
func (b *Block) GovernanceVoteMerkleRoot() chainhash.Hash {
	return merkleRootGovernanceVotes(b.GovernanceVotes)
}

func merkleRootGovernanceVotes(governanceVote []*GovernanceVote) chainhash.Hash {
	if len(governanceVote) == 0 {
		return chainhash.Hash{}
	}
	if len(governanceVote) == 1 {
		return governanceVote[0].Hash()
	}
	mid := len(governanceVote) / 2
	h1 := merkleRootGovernanceVotes(governanceVote[:mid])
	h2 := merkleRootGovernanceVotes(governanceVote[mid:])

	return chainhash.HashH(append(h1[:], h2[:]...))
}

// ExitMerkleRoot calculates the merkle root of the Exits in the block.
func (b *Block) ExitMerkleRoot() chainhash.Hash {
	return merkleRootExits(b.Exits)
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

// DepositMerkleRoot calculates the merkle root of the Deposits in the block.
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

// TxsMerkleRoot calculates the merkle root of the Txs in the block.
func (b *Block) TxsMerkleRoot() chainhash.Hash {
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

// MultiSignatureTxsMerkleRoot calculates the merkle root of the TxsMulti in the block.
func (b *Block) MultiSignatureTxsMerkleRoot() chainhash.Hash {
	return merkleRootMultiSignaturesTxs(b.MultiSignatureTxs)
}

func merkleRootMultiSignaturesTxs(txs []*MultiSignatureTx) chainhash.Hash {
	if len(txs) == 0 {
		return chainhash.Hash{}
	}
	if len(txs) == 1 {
		return txs[0].Hash()
	}
	mid := len(txs) / 2
	h1 := merkleRootMultiSignaturesTxs(txs[:mid])
	h2 := merkleRootMultiSignaturesTxs(txs[mid:])

	return chainhash.HashH(append(h1[:], h2[:]...))
}

// VotesMerkleRoot calculates the merkle root of the Votes in the block.
func (b *Block) VotesMerkleRoot() chainhash.Hash {
	return merkleRootVotes(b.Votes)
}

func merkleRootVotes(votes []*MultiValidatorVote) chainhash.Hash {
	if len(votes) == 0 {
		return chainhash.Hash{}
	}
	if len(votes) == 1 {
		return votes[0].Data.Hash()
	}
	mid := len(votes) / 2
	h1 := merkleRootVotes(votes[:mid])
	h2 := merkleRootVotes(votes[mid:])

	return chainhash.HashH(append(h1[:], h2[:]...))
}

// ProposerSlashingsRoot calculates the merkle root of the ProposerSlashings in the block.
func (b *Block) ProposerSlashingsRoot() chainhash.Hash {
	return merkleRootProposerSlashing(b.ProposerSlashings)
}

func merkleRootProposerSlashing(slashings []*ProposerSlashing) chainhash.Hash {
	if len(slashings) == 0 {
		return chainhash.Hash{}
	}
	if len(slashings) == 1 {
		return slashings[0].Hash()
	}
	mid := len(slashings) / 2
	h1 := merkleRootProposerSlashing(slashings[:mid])
	h2 := merkleRootProposerSlashing(slashings[mid:])

	return chainhash.HashH(append(h1[:], h2[:]...))
}

// RANDAOSlashingsRoot calculates the merkle root of the RANDAOSlashings in the block.
func (b *Block) RANDAOSlashingsRoot() chainhash.Hash {
	return merkleRootRandaoSlashing(b.RANDAOSlashings)
}

func merkleRootRandaoSlashing(slashings []*RANDAOSlashing) chainhash.Hash {
	if len(slashings) == 0 {
		return chainhash.Hash{}
	}
	if len(slashings) == 1 {
		return slashings[0].Hash()
	}
	mid := len(slashings) / 2
	h1 := merkleRootRandaoSlashing(slashings[:mid])
	h2 := merkleRootRandaoSlashing(slashings[mid:])

	return chainhash.HashH(append(h1[:], h2[:]...))
}

// VoteSlashingRoot calculates the merkle root of the VoteSlashings in the block.
func (b *Block) VoteSlashingRoot() chainhash.Hash {
	return merkleRootVoteSlashing(b.VoteSlashings)
}

func merkleRootVoteSlashing(slashings []*VoteSlashing) chainhash.Hash {
	if len(slashings) == 0 {
		return chainhash.Hash{}
	}
	if len(slashings) == 1 {
		return slashings[0].Hash()
	}
	mid := len(slashings) / 2
	h1 := merkleRootVoteSlashing(slashings[:mid])
	h2 := merkleRootVoteSlashing(slashings[mid:])

	return chainhash.HashH(append(h1[:], h2[:]...))
}

// CoinProofsMerkleRoot calculates the merkle root of the CoinProofs in the block.
func (b *Block) CoinProofsMerkleRoot() chainhash.Hash {
	return merkleRootCoinProofs(b.CoinProofs)
}

func merkleRootCoinProofs(proofs []*burnproof.CoinsProofSerializable) chainhash.Hash {
	if len(proofs) == 0 {
		return chainhash.Hash{}
	}
	if len(proofs) == 1 {
		return proofs[0].Hash()
	}
	mid := len(proofs) / 2

	h1 := merkleRootCoinProofs(proofs[:mid])
	h2 := merkleRootCoinProofs(proofs[mid:])

	return chainhash.HashH(append(h1[:], h2[:]...))
}

// PartialExitsMerkleRoot calculates the merkle root of the PartialExit in the block.
func (b *Block) PartialExitsMerkleRoot() chainhash.Hash {
	return merkleRootPartialExit(b.PartialExit)
}

func merkleRootPartialExit(e []*PartialExit) chainhash.Hash {
	if len(e) == 0 {
		return chainhash.Hash{}
	}
	if len(e) == 1 {
		return e[0].Hash()
	}
	mid := len(e) / 2

	h1 := merkleRootPartialExit(e[:mid])
	h2 := merkleRootPartialExit(e[mid:])

	return chainhash.HashH(append(h1[:], h2[:]...))
}

// ExecutionsMarkleRoot calculates the merkle root of the Executions in the block.
func (b *Block) ExecutionsMerkleRoot() chainhash.Hash {
	return merkleRootExecutions(b.Executions)
}

func merkleRootExecutions(e []*Execution) chainhash.Hash {
	if len(e) == 0 {
		return chainhash.Hash{}
	}
	if len(e) == 1 {
		return e[0].Hash()
	}
	mid := len(e) / 2

	h1 := merkleRootExecutions(e[:mid])
	h2 := merkleRootExecutions(e[mid:])

	return chainhash.HashH(append(h1[:], h2[:]...))
}

// GetTxs returns a slice with tx hashes
func (b *Block) GetTxs() []string {
	txs := make([]string, len(b.Txs))
	for i, tx := range b.Txs {
		txs[i] = tx.Hash().String()
	}
	return txs
}
