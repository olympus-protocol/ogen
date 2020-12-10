package models

type Vote struct {
	BlockHash             string
	Signature             string
	ParticipationBitfield string
	Hash                  string
	Data                  VoteData
}

type VoteData struct {
	Slot            int
	FromEpoch       int
	FromHash        string
	ToEpoch         int
	ToHash          string
	BeaconBlockHash string
	Nonce           int
}
