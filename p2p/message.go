package p2p

import (
	"fmt"
	"math"

	ssz "github.com/ferranbt/fastssz"
)

const (
	MsgVersionCmd   = "version"
	MsgGetAddrCmd   = "getaddr"
	MsgAddrCmd      = "addr"
	MsgGetBlocksCmd = "getblocks"
	MsgBlocksCmd    = "blocks"
	MsgTxCmd        = "tx"
)

const MessageHeaderSize = 24
const MaxMessagePayload = 1024 * 1024 * 32 // 32 MB

type Message interface {
	Marshal() ([]byte, error)
	Unmarshal(d []byte) error
	Command() string
}

type messageHeader struct {
	magic    NetMagic
	command  string
	length   uint32
	checksum [4]byte

	ssz.Marshaler
	ssz.Unmarshaler
}

// Marshal serializes the struct to bytes
func (mh *messageHeader) Marshal() ([]byte, error) {
	return mh.MarshalSSZ()
}

// Unmarshal deserializes the struct from bytes
func (mh *messageHeader) Unmarshal(b []byte) error {
	return mh.UnmarshalSSZ(b)
}

func makeEmptyMessage(command string) (Message, error) {
	var msg Message
	switch command {
	case MsgVersionCmd:
		msg = &MsgVersion{}
	case MsgAddrCmd:
		msg = &MsgAddr{}
	case MsgGetAddrCmd:
		msg = &MsgGetAddr{}
	case MsgBlocksCmd:
		msg = &MsgBlocks{}
	case MsgGetBlocksCmd:
		msg = &MsgGetBlocks{}
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
