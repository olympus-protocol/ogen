package p2p

import (
	"errors"
	"fmt"
	"github.com/grupokindynos/ogen/config"
	"github.com/grupokindynos/ogen/utils/serializer"
	"io"
	"time"
)

type MsgVersion struct {
	ProtocolVersion int32                 // 4 bytes
	UserAgent       string                // MaxUserAgentLen (bytes)
	LastBlock       int32                 // 4 bytes
	Nonce           uint64                // 8 bytes
	Services        ServiceFlag           // 8 bytes
	Timestamp       int64                 // 8 bytes
	AddrMe          serializer.NetAddress // 26 bytes
	AddrYou         serializer.NetAddress // 26 bytes
}

var (
	DefaultUserAgent = "/Ogen:" + config.OgenVersion() + "/"
)

const (
	MaxUserAgentLen = 40
)

func (m *MsgVersion) Encode(w io.Writer) error {
	err := validateUserAgent(m.UserAgent)
	if err != nil {
		return err
	}

	err = serializer.WriteElements(w, m.ProtocolVersion, m.Services,
		m.Timestamp)
	if err != nil {
		return err
	}

	err = serializer.WriteNetAddress(w, &m.AddrYou)
	if err != nil {
		return err
	}

	err = serializer.WriteNetAddress(w, &m.AddrMe)
	if err != nil {
		return err
	}

	err = serializer.WriteElement(w, m.Nonce)
	if err != nil {
		return err
	}

	err = serializer.WriteVarString(w, m.UserAgent)
	if err != nil {
		return err
	}

	err = serializer.WriteElement(w, m.LastBlock)
	if err != nil {
		return err
	}
	return nil
}

func (m *MsgVersion) Decode(r io.Reader) error {
	err := serializer.ReadElements(r, &m.ProtocolVersion, &m.Services,
		(*int64)(&m.Timestamp))
	if err != nil {
		return err
	}
	err = serializer.ReadNetAddress(r, &m.AddrYou)
	if err != nil {
		return err
	}
	err = serializer.ReadNetAddress(r, &m.AddrMe)
	if err != nil {
		return err
	}
	err = serializer.ReadElement(r, &m.Nonce)
	if err != nil {
		return err
	}
	userAgent, err := serializer.ReadVarString(r)
	if err != nil {
		return err
	}
	err = validateUserAgent(userAgent)
	if err != nil {
		return err
	}
	m.UserAgent = userAgent
	err = serializer.ReadElement(r, &m.LastBlock)
	if err != nil {
		return err
	}
	return nil
}

func (m *MsgVersion) Command() string {
	return MsgVersionCmd
}

func validateUserAgent(userAgent string) error {
	if len(userAgent) > MaxUserAgentLen {
		str := fmt.Sprintf("users agent too long [len %v, max %v]",
			len(userAgent), MaxUserAgentLen)
		err := errors.New(str)
		return err
	}
	return nil
}

func (m *MsgVersion) MaxPayloadLength() uint32 {
	return 124
}

func NewMsgVersion(me serializer.NetAddress, you serializer.NetAddress, nonce uint64,
	lastBlock int32) *MsgVersion {
	return &MsgVersion{
		ProtocolVersion: int32(ProtocolVersion),
		Services:        0,
		Timestamp:       time.Unix(time.Now().Unix(), 0).Unix(),
		AddrYou:         you,
		AddrMe:          me,
		Nonce:           nonce,
		UserAgent:       DefaultUserAgent,
		LastBlock:       lastBlock,
	}
}
