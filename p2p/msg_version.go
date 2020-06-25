package p2p

import (
	"time"

	"github.com/prysmaticlabs/go-ssz"
)

type MsgVersion struct {
	LastBlock uint64 // 8 bytes
	Nonce     uint64 // 8 bytes
	Timestamp uint64 // 8 bytes
}

func (m *MsgVersion) Marshal() ([]byte, error) {
	return ssz.Marshal(m)
}

func (m *MsgVersion) Unmarshal(b []byte) error {
	return ssz.Unmarshal(b, m)
}

func (m *MsgVersion) Command() string {
	return MsgVersionCmd
}

func (m *MsgVersion) MaxPayloadLength() uint32 {
	return 36
}

func NewMsgVersion(nonce uint64, lastBlock uint64) *MsgVersion {
	return &MsgVersion{
		Timestamp: uint64(time.Unix(time.Now().Unix(), 0).Unix()),
		Nonce:     nonce,
		LastBlock: lastBlock,
	}
}
