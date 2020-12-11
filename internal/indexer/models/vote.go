package models

type Vote struct {
	BlockHash             string
	ParticipationBitfield string
	Hash                  string   `gorm:"primaryKey"`
	Data                  VoteData `gorm:"foreignKey:Hash"`
}

type VoteData struct {
	Hash            string `gorm:"primaryKey"`
	Slot            uint64
	FromEpoch       uint64
	FromHash        string
	ToEpoch         uint64
	ToHash          string
	BeaconBlockHash string
	Nonce           string
}
