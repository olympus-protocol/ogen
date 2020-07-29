package p2p

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/olympus-protocol/ogen/utils/chainhash"
)

var (
	// ErrorChecksum returned when the message header checksum doesn't match.
	ErrorChecksum = errors.New("message checksum don't match")
	// ErrorAnnLength returned when the header length doesn't match the message length.
	ErrorAnnLength = errors.New("wrong announced length")
	// ErrorSizeExceed returned when the message exceed the maximum payload message.
	ErrorSizeExceed = errors.New("message exceed max payload length")
	// ErrorNetMismatch returned when the message doesn't match the specified network.
	ErrorNetMismatch = errors.New("wrong message network")
)

// MaxMsgSize is the maximum amount of bytes a message can have.
var MaxMsgSize = 1024 * 1024 * 64

const (
	// MsgVersionCmd is for version handshake
	MsgVersionCmd = "version"
	// MsgGetAddrCmd ask node for address
	MsgGetAddrCmd = "getaddr"
	// MsgAddrCmd an slice of address
	MsgAddrCmd = "addr"
	// MsgGetBlocksCmd ask a node for blocks
	MsgGetBlocksCmd = "getblocks"
	// MsgBlocksCmd slice with blocks
	MsgBlocksCmd = "blocks"
)

// Message interface for all the messages
type Message interface {
	Marshal() ([]byte, error)
	Unmarshal(b []byte) error
	Command() string
	MaxPayloadLength() uint64
}

// MessageHeader header of the message
type MessageHeader struct {
	Magic    uint64
	Command  [40]byte `ssz-size:"40"`
	Length   uint64
	Checksum [4]byte `ssz-size:"4"`
}

// Marshal serializes the data to bytes
func (h *MessageHeader) Marshal() ([]byte, error) {
	return h.MarshalSSZ()
}

// Unmarshal deserializes the data
func (h *MessageHeader) Unmarshal(b []byte) error {
	return h.UnmarshalSSZ(b)
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
	default:
		return nil, fmt.Errorf("unhandled command [%s]", command)
	}
	return msg, nil
}

// ReadMessage decodes the message from reader
func ReadMessage(r io.Reader, net uint32) (Message, error) {
	headerBuf := make([]byte, 60)
	_, err := io.ReadFull(r, headerBuf)
	if err != nil {
		return nil, err
	}
	header, err := readHeader(headerBuf, net)
	if err != nil {
		return nil, err
	}
	cmdB := bytes.Trim(header.Command[:], "\x00")
	cmd := string(cmdB)
	msg, err := makeEmptyMessage(cmd)
	if err != nil {
		return nil, err
	}
	msgB := make([]byte, header.Length)
	_, err = io.ReadFull(r, msgB)
	if err != nil {
		return nil, err
	}
	var checksum [4]byte
	copy(checksum[:], chainhash.DoubleHashB(msgB)[0:4])
	if header.Checksum != checksum {
		return nil, ErrorChecksum
	}

	if header.Length != uint64(len(msgB)) {
		return nil, ErrorAnnLength
	}
	err = msg.Unmarshal(msgB)
	if err != nil {
		return nil, err
	}
	if header.Length > msg.MaxPayloadLength() {
		return nil, ErrorSizeExceed
	}
	return msg, nil
}

func readHeader(h []byte, net uint32) (MessageHeader, error) {
	var header MessageHeader
	err := header.Unmarshal(h)
	if err != nil {
		return MessageHeader{}, err
	}
	if header.Magic != uint64(net) {
		return MessageHeader{}, ErrorNetMismatch
	}
	return header, nil
}

// WriteMessage writes the message to writer
func WriteMessage(w io.Writer, msg Message, net uint32) error {
	ser, err := msg.Marshal()
	if err != nil {
		return err
	}
	var checksum [4]byte
	copy(checksum[:], chainhash.DoubleHashB(ser)[0:4])
	hb, err := writeHeader(msg, net, uint64(len(ser)), checksum)
	if err != nil {
		return err
	}
	buf := []byte{}
	buf = append(buf, hb...)
	buf = append(buf, ser...)
	_, err = w.Write(buf)
	if err != nil {
		return err
	}
	return nil
}

func writeHeader(msg Message, net uint32, length uint64, checksum [4]byte) ([]byte, error) {
	cmd := [40]byte{}
	copy(cmd[:], []byte(msg.Command()))
	header := MessageHeader{
		Magic:    uint64(net),
		Command:  cmd,
		Length:   length,
		Checksum: checksum,
	}
	ser, err := header.Marshal()
	if err != nil {
		return nil, err
	}
	return ser, nil
}
