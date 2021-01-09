package primitives

// BlockHeader
const BlockHeaderSize = 20 + (14 * 32) // 468 bytes

// Vote
const MaxVotesPerBlock = 16
const VoteDataSize = 32 * 4                                // 128 bytes
const MaxMultiValidatorVoteSize = VoteDataSize + 96 + 6255 // 6479 bytes

// Deposits
const MaxDepositsPerBlock = 32
const DepositDataSize = 48 + 96 + 20          // 164 bytes
const DepositSize = DepositDataSize + 48 + 96 // 308 bytes

// Exits
const MaxExitsPerBlock = 32
const ExitSize = (48 * 2) + 96 // 192 bytes

// PartialExits
const MaxPartialExitsPerBlock = 32
const PartialExitsSize = (48 * 2) + 96 + 8 // 200 bytes

// CoinProofs
const MaxCoinProofsPerBlock = 64
const MaxCoinProofSize = 8 + 25 + 192 + 44 + (32 * 64) // 2317 bytes

// Execution
const MaxExecutionsPerBlock = 128
const MaxExecutionSize = 48 + 20 + (7168 + 4) + 96 + 8 + 8 // 32952 bytes

// Tx
const MaxTxsPerBlock = 30000
const TxSize = 20 + 48 + (8 * 3) + 96 // 188 bytes

// ProposerSlashing
const MaxProposerSlashingsPerBlock = 2
const ProposerSlashingSize = (96 * 2) + 48 + (BlockHeaderSize * 2) // 1240 bytes

// VoteSlashing
const MaxVoteSlashingsPerBlock = 5
const MaxVotesSlashingSize = MaxMultiValidatorVoteSize*2 + 8 // 12948 bytes

// RANDAOSlashing
const MaxRANDAOSlashingsPerBlock = 20
const RANDAOSlashingSize = 96 + 48 + 8 // 152 bytes

// GovernanceVote
const MaxGovernanceVotesPerBlock = 128
const MaxGovernanceVoteSize = 8 + 8 + 96 + 48 + 100 + 4 // 264 bytes

// Multipub
const MaxPublicKeysOnMultipub = 15
const MaxMultipubSize = 8 + (48 * 15) + 4 // 732

// Multisig
const MaxMultisigSize = MaxMultipubSize + (96 * 15) + 15 // 2187 bytes

// MultiSignatureTx
const MaxMultiSignatureTxsOnBlock = 8
const MaxMultiSignatureTxSize = MaxMultisigSize + 20 + (8 * 3) // 2231

// ValidatorHelloMessage
const MaxValidatorHelloMessageSize = 128 + 16 + 96 + 5
