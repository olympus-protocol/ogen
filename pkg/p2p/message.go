package p2p

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/olympus-protocol/ogen/pkg/chainhash"
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

const (
	// MsgBlockCmd is a single block element
	MsgBlockCmd = "block"
	// MsgTxCmd is a single tx element
	MsgTxCmd = "tx"
	// MsgDepositCmd is a single deposit element
	MsgDepositCmd = "deposit"
	// MsgDepositCmd is a deposit slice element
	MsgDepositsCmd = "deposits"
	// MsgVoteCmd is a single vote element
	MsgVoteCmd = "vote"
	// MsgValidatorStart is a validator hello element
	MsgValidatorStartCmd = "validator_hello"
	// MsgExitCmd is a exit element
	MsgExitCmd = "exit"
	// MsgExitsCmd is a exit slice element
	MsgExitsCmd = "exits"
	// MsgGovernanceCmd is a exit element
	MsgGovernanceCmd = "governance_vote"
	// MsgMultiSignatureTx is a exit element
	MsgMultiSignatureTxCmd = "multi_sig_tx"
	// MsgVersionCmd is for version handshake
	MsgVersionCmd = "version"
	// MsgGetBlocksCmd ask a node for blocks
	MsgGetBlocksCmd = "getblocks"
	// MsgFinalizationCmd announce a peer to reached state finalization
	MsgFinalizationCmd = "finalized"
	// MsgProofsCmd is coin redeem
	MsgProofsCmd = "proofs"
	// MsgPartialExitsCmd subtract coins from a contract
	MsgPartialExitsCmd = "partialexit"
	// MsgExecution executes a bytecode that modifies the state
	MsgExecutionCmd = "execute"
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
func (m *MessageHeader) Marshal() ([]byte, error) {
	return m.MarshalSSZ()
}

// Unmarshal deserializes the data
func (m *MessageHeader) Unmarshal(b []byte) error {
	return m.UnmarshalSSZ(b)
}

func makeEmptyMessage(command string) (Message, error) {
	var msg Message
	switch command {
	case MsgVersionCmd:
		msg = &MsgVersion{}
	case MsgGetBlocksCmd:
		msg = &MsgGetBlocks{}
	case MsgBlockCmd:
		msg = &MsgBlock{}
	case MsgTxCmd:
		msg = &MsgTx{}
	case MsgMultiSignatureTxCmd:
		msg = &MsgMultiSignatureTx{}
	case MsgDepositCmd:
		msg = &MsgDeposit{}
	case MsgDepositsCmd:
		msg = &MsgDeposits{}
	case MsgVoteCmd:
		msg = &MsgVote{}
	case MsgExitCmd:
		msg = &MsgExit{}
	case MsgExitsCmd:
		msg = &MsgExits{}
	case MsgValidatorStartCmd:
		msg = &MsgValidatorStart{}
	case MsgGovernanceCmd:
		msg = &MsgGovernance{}
	case MsgFinalizationCmd:
		msg = &MsgFinalization{}
	case MsgProofsCmd:
		msg = &MsgProofs{}
	case MsgPartialExitsCmd:
		msg = &MsgPartialExits{}
	case MsgExecutionCmd:
		msg = &MsgExecution{}

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
	fmt.Println(header.Length, msg.MaxPayloadLength())
	if header.Length > msg.MaxPayloadLength() {
		return nil, ErrorSizeExceed
	}

	return msg, nil
}

func readHeader(h []byte, net uint32) (*MessageHeader, error) {
	header := new(MessageHeader)

	err := header.Unmarshal(h)
	if err != nil {
		return nil, err
	}

	if header.Magic != uint64(net) {
		return nil, ErrorNetMismatch
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

	var buf []byte
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
	copy(cmd[:], msg.Command())

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
