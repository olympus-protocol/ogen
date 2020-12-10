package models

type Vote struct {
	BlockHash             string
	Signature             string
	ParticipationBitfield string
	Hash                  string `gorm:"primarykey"`
	Data                  VoteData
}

type VoteData struct {
	Hash            string `gorm:"primarykey"`
	Slot            int
	FromEpoch       int
	FromHash        string
	ToEpoch         int
	ToHash          string
	BeaconBlockHash string
	Nonce           int
}
