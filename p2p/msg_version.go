package p2p

import (
	"time"

	"github.com/prysmaticlabs/go-ssz"
)

type MsgVersion struct {
	ProtocolVersion int32  // 4 bytes
	LastBlock       uint64 // 8 bytes
	Nonce           uint64 // 8 bytes
	Timestamp       int64  // 8 bytes
}

// Marshal serializes the struct to bytes
func (m *MsgVersion) Marshal() ([]byte, error) {
	return ssz.Marshal(m)
}

// Unmarshal deserializes the struct from bytes
func (m *MsgVersion) Unmarshal(b []byte) error {
	return ssz.Unmarshal(b, m)
}

func (m *MsgVersion) Command() string {
	return MsgVersionCmd
}

func (m *MsgVersion) MaxPayloadLength() uint32 {
	return 36
}

func NewMsgVersion(lastBlock uint64) *MsgVersion {
	return &MsgVersion{
		ProtocolVersion: int32(ProtocolVersion),
		Timestamp:       time.Unix(time.Now().Unix(), 0).Unix(),
		LastBlock:       lastBlock,
	}
}
