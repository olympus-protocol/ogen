package primitives

import (
	"github.com/golang/snappy"
	"github.com/olympus-protocol/ogen/pkg/burnproof"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
)

// MaxBlockSize defines the maximum bytes on a block object.
const MaxBlockSize = 1024 * 1024 * 2.5 // 2.5 MB

// Block is a block in the blockchain.
type Block struct {
	Header            *BlockHeader                        // 													= 372 bytes
	Votes             []*MultiValidatorVote               `ssz-max:"32"`   // MaxVotesPerBlock 				32 * 6474 		= 207168 bytes
	Txs               []*Tx                               `ssz-max:"5000"` // MaxTxsPerBlock					204 * 5000  	= 1020000 bytes
	TxsMulti          []*TxMulti                          `ssz-max:"128"`  // MaxTxsMultiPerBlock
	Deposits          []*Deposit                          `ssz-max:"128"`  // MaxDepositsPerBlock 				308 * 128 		= 39424 bytes
	Exits             []*Exit                             `ssz-max:"128"`  // MaxExitsPerBlock     			192 * 128 		= 24576 bytes
	VoteSlashings     []*VoteSlashing                     `ssz-max:"10"`   // MaxVoteSlashingPerBlock			666 * 10 		= 6660 bytes
	RANDAOSlashings   []*RANDAOSlashing                   `ssz-max:"20"`   // MaxRANDAOSlashingPerBlock   		152 * 20 		= 3040 bytes
	ProposerSlashings []*ProposerSlashing                 `ssz-max:"2"`    // MaxProposerSlashingPerBlock 		984 * 2 		= 1968 bytes
	GovernanceVotes   []*GovernanceVote                   `ssz-max:"128"`  // MaxGovernanceVotesPerBlock		260 * 128		= 33280 bytes
	CoinProofs        []*burnproof.CoinsProofSerializable `ssz-max:"128"`  // MaxCoinProofsPerBlock 			4321 * 128   	=
	Signature         [96]byte                            `ssz-size:"96"`  // 													= 96 bytes
	RandaoSignature   [96]byte                            `ssz-size:"96"`  // 													= 96 bytes
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

// TransactionMerkleRoot calculates the merkle root of the Txs in the block.
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

// TransactionMultiMerkleRoot calculates the merkle root of the TxsMulti in the block.
func (b *Block) TransactionMultiMerkleRoot() chainhash.Hash {
	return merkleRootTxsMulti(b.TxsMulti)
}

func merkleRootTxsMulti(txs []*TxMulti) chainhash.Hash {
	if len(txs) == 0 {
		return chainhash.Hash{}
	}
	if len(txs) == 1 {
		return txs[0].Hash()
	}
	mid := len(txs) / 2
	h1 := merkleRootTxsMulti(txs[:mid])
	h2 := merkleRootTxsMulti(txs[mid:])

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

// CoinProofsRoot calculates the merkle root of the CoinProofs in the block.
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

// GetTxs returns a slice with tx hashes
func (b *Block) GetTxs() []string {
	txs := make([]string, len(b.Txs))
	for i, tx := range b.Txs {
		txs[i] = tx.Hash().String()
	}
	return txs
}
