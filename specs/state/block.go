package state

// Block is a block on the blockchain
type Block struct {
	Header            *BlockHeader
	Votes             []*MultiValidatorVote `ssz-max:"1099511627776"`
	Txs               []*Tx                 `ssz-max:"1099511627776"`
	Deposits          []*Deposit            `ssz-max:"1099511627776"`
	Exits             []*Exit               `ssz-max:"1099511627776"`
	VoteSlashings     []*VoteSlashing       `ssz-max:"1099511627776"`
	RANDAOSlashings   []*RANDAOSlashing     `ssz-max:"1099511627776"`
	ProposerSlashings []*ProposerSlashing   `ssz-max:"1099511627776"`
	GovernanceVotes   []*GovernanceVote     `ssz-max:"1099511627776"`
	Signature         []byte                `ssz-size:"96"`
	RandaoSignature   []byte                `ssz-size:"96"`
}
