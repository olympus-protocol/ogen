package p2p

import (
	"io"
	"time"

	"github.com/olympus-protocol/ogen/utils/serializer"
)

type MsgVersion struct {
	ProtocolVersion int32       // 4 bytes
	LastBlock       uint64      // 8 bytes
	Nonce           uint64      // 8 bytes
	Timestamp       int64       // 8 bytes
}

func (m *MsgVersion) Encode(w io.Writer) error {
	err := serializer.WriteElements(w, m.ProtocolVersion, m.Timestamp, m.Nonce, m.LastBlock)
	if err != nil {
		return err
	}
	return nil
}

func (m *MsgVersion) Decode(r io.Reader) error {
	err := serializer.ReadElements(r, &m.ProtocolVersion,
		(*int64)(&m.Timestamp), &m.Nonce, &m.LastBlock)
	if err != nil {
		return err
	}
	return nil
}

func (m *MsgVersion) Command() string {
	return MsgVersionCmd
}

func (m *MsgVersion) MaxPayloadLength() uint32 {
	return 36
}

func NewMsgVersion(nonce uint64, lastBlock uint64) *MsgVersion {
	return &MsgVersion{
		ProtocolVersion: int32(ProtocolVersion),
		Timestamp:       time.Unix(time.Now().Unix(), 0).Unix(),
		Nonce:           nonce,
		LastBlock:       lastBlock,
	}
}
