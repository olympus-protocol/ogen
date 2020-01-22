package p2p

import (
	"fmt"
	"io"
	"math"
)

const (
	MsgVersionCmd   = "version"
	MsgVerackCmd    = "verack"
	MsgPingCmd      = "ping"
	MsgPongCmd      = "pong"
	MsgGetAddrCmd   = "getaddr"
	MsgAddrCmd      = "addr"
	MsgGetBlocksCmd = "getblocks"
	MsgBlockCmd     = "block"
	MsgBlocksInvCmd = "blocksinv"
	MsgTxCmd        = "tx"
)

const MessageHeaderSize = 24
const MaxMessagePayload = 1024 * 1024 * 32 // 32 MB

type Message interface {
	Decode(io.Reader) error
	Encode(io.Writer) error
	Command() string
	MaxPayloadLength() uint32
}

type messageHeader struct {
	magic    NetMagic
	command  string
	length   uint32
	checksum [4]byte
}

func makeEmptyMessage(command string) (Message, error) {
	var msg Message
	switch command {
	case MsgVersionCmd:
		msg = &MsgVersion{}
	case MsgVerackCmd:
		msg = &MsgVerack{}
	case MsgPingCmd:
		msg = &MsgPing{}
	case MsgPongCmd:
		msg = &MsgPong{}
	case MsgAddrCmd:
		msg = &MsgGetAddr{}
	case MsgGetAddrCmd:
		msg = &MsgGetAddr{}
	case MsgBlockCmd:
		msg = &MsgBlock{}
	case MsgGetBlocksCmd:
		msg = &MsgGetBlocks{}
	case MsgBlocksInvCmd:
		msg = &MsgBlockInv{}
	case MsgTxCmd:
		msg = &MsgTx{}
	default:
		return nil, fmt.Errorf("unhandled command [%s]", command)
	}
	return msg, nil
}

// VarIntSerializeSize returns the number of bytes it would take to serialize
// val as a variable length integer.
func VarIntSerializeSize(val uint64) int {
	// The value is small enough to be represented by itself, so it's
	// just 1 byte.
	if val < 0xfd {
		return 1
	}

	// Discriminant 1 byte plus 2 bytes for the uint16.
	if val <= math.MaxUint16 {
		return 3
	}

	// Discriminant 1 byte plus 4 bytes for the uint32.
	if val <= math.MaxUint32 {
		return 5
	}

	// Discriminant 1 byte plus 8 bytes for the uint64.
	return 9
}
