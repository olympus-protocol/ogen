package state

type BlockHeader struct {
	Version                    uint32
	TxMerkleRoot               []byte `ssz-size:"32"`
	VoteMerkleRoot             []byte `ssz-size:"32"`
	DepositMerkleRoot          []byte `ssz-size:"32"`
	ExitMerkleRoot             []byte `ssz-size:"32"`
	VoteSlashingMerkleRoot     []byte `ssz-size:"32"`
	RANDAOSlashingMerkleRoot   []byte `ssz-size:"32"`
	ProposerSlashingMerkleRoot []byte `ssz-size:"32"`
	GovernanceVotesMerkleRoot  []byte `ssz-size:"32"`
	PrevBlockHash              []byte `ssz-size:"32"`
	Timestamp                  uint64
	Slot                       uint64
	StateRoot                  []byte `ssz-size:"32"`
	FeeAddress                 []byte `ssz-size:"20"`
}
