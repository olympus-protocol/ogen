package models

type Vote struct {
	BlockHash             [32]byte
	Signature             [96]byte
	ParticipationBitfield []byte
	Hash                  [32]byte `gorm:"primarykey"`
	Data                  VoteData `gorm:"foreignkey:Hash"`
}

type VoteData struct {
	Hash            [32]byte `gorm:"primarykey"`
	Slot            uint64
	FromEpoch       uint64
	FromHash        [32]byte
	ToEpoch         uint64
	ToHash          [32]byte
	BeaconBlockHash [32]byte
	Nonce           uint64
}
