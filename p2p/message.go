package p2p

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/golang/snappy"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/prysmaticlabs/go-ssz"
)

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
	MaxPayloadLength() uint32
}

// MessageHeader header of the message
type MessageHeader struct {
	Magic    uint32
	Command  [40]byte
	Length   uint32
	Checksum [4]byte
}

func (h *MessageHeader) Marshal() ([]byte, error) {
	return ssz.Marshal(h)
}

func (h *MessageHeader) Unmarshal(b []byte) error {
	return ssz.Unmarshal(b, h)
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
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	data, err := snappy.Decode(nil, b)
	if err != nil {
		return nil, err
	}
	header, err := readHeader(data[0:52], net)
	if err != nil {
		return nil, err
	}
	cmdB := bytes.Trim(header.Command[:], "\x00")
	cmd := string(cmdB)
	msg, err := makeEmptyMessage(cmd)
	if err != nil {
		return nil, err
	}
	var checksum [4]byte
	copy(checksum[:], chainhash.DoubleHashB(data[52:])[0:4])
	if header.Checksum != checksum {
		return nil, errors.New("checksum don't match")
	}
	if header.Length != uint32(len(data[52:])) {
		return nil, errors.New("wrong announced length")
	}
	err = msg.Unmarshal(data[52:])
	if err != nil {
		return nil, err
	}
	if header.Length > msg.MaxPayloadLength() {
		return nil, errors.New("message exceed max payload length")
	}
	return msg, nil
}

func readHeader(h []byte, net uint32) (MessageHeader, error) {
	var header MessageHeader
	err := header.Unmarshal(h)
	if err != nil {
		return MessageHeader{}, err
	}
	if header.Magic != net {
		return MessageHeader{}, errors.New("wrong message network")
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
	hb, err := writeHeader(msg, net, uint32(len(ser)), checksum)
	if err != nil {
		return err
	}
	buf := []byte{}
	buf = append(buf, hb...)
	buf = append(buf, ser...)
	compressed := snappy.Encode(nil, buf)
	_, err = w.Write(compressed)
	if err != nil {
		return err
	}
	return nil
}

func writeHeader(msg Message, net uint32, length uint32, checksum [4]byte) ([]byte, error) {
	cmd := [40]byte{}
	copy(cmd[:], []byte(msg.Command()))
	header := MessageHeader{
		Magic:    net,
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
