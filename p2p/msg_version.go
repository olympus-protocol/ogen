package p2p

type MsgVersion struct {
	ProtocolVersion uint32
	LastBlock       uint64
	Nonce           uint64
	Timestamp       uint64
}
