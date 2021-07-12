package primitives

// BlockHeader
const BlockHeaderSize = 12 + (14 * 32) // 468 bytes

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

// Tx
const MaxTxsPerBlock = 5000
const TxSize = 20 + 48 + (8 * 3) + 96 // 188 bytes

// ProposerSlashing
const MaxProposerSlashingsPerBlock = 2
const ProposerSlashingSize = (96 * 2) + 48 + (BlockHeaderSize * 2) // 1160 bytes

// VoteSlashing
const MaxVoteSlashingsPerBlock = 5
const MaxVotesSlashingSize = MaxMultiValidatorVoteSize*2 + 8 // 12966 bytes

// RANDAOSlashing
const MaxRANDAOSlashingsPerBlock = 20
const RANDAOSlashingSize = 96 + 48 + 8 // 152 bytes
