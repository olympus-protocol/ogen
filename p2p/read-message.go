package p2p

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
)

func readMessageHeader(r io.Reader) (int, *messageHeader, error) {
	var headerBytes [MessageHeaderSize]byte
	n, err := io.ReadFull(r, headerBytes[:])
	if err != nil {
		return n, nil, err
	}
	hr := bytes.NewReader(headerBytes[:])

	hdr := messageHeader{}
	var command [serializer.CommandSize]byte
	serializer.ReadElements(hr, &hdr.magic, &command, &hdr.length, &hdr.checksum)

	hdr.command = string(bytes.TrimRight(command[:], string(0)))
	return n, &hdr, nil
}

func discardInput(r io.Reader, n uint32) {
	maxSize := uint32(10 * 1024) // 10k at a time
	numReads := n / maxSize
	bytesRemaining := n % maxSize
	if n > 0 {
		buf := make([]byte, maxSize)
		for i := uint32(0); i < numReads; i++ {
			io.ReadFull(r, buf)
		}
	}
	if bytesRemaining > 0 {
		buf := make([]byte, bytesRemaining)
		io.ReadFull(r, buf)
	}
}

func ReadMessageWithEncodingN(r io.Reader, net NetMagic) (Message, error) {
	_, hdr, err := readMessageHeader(r)
	if err != nil {
		return nil, err
	}

	if hdr.length > MaxMessagePayload {
		str := fmt.Sprintf("message payload is too large - header "+
			"indicates %d bytes, but max message payload is %d "+
			"bytes.", hdr.length, MaxMessagePayload)
		err = errors.New(str)
		return nil, err

	}

	if hdr.magic != net {
		discardInput(r, hdr.length)
		str := fmt.Sprintf("message from other network [%v]", hdr.magic)
		err = errors.New(str)
		return nil, err
	}

	command := hdr.command
	if !utf8.ValidString(command) {
		discardInput(r, hdr.length)
		str := fmt.Sprintf("invalid command %v", []byte(command))
		err = errors.New(str)
		return nil, err
	}

	msg, err := makeEmptyMessage(command)
	if err != nil {
		discardInput(r, hdr.length)
		return nil, fmt.Errorf("error creating new command %s: %s", command, err)
	}

	mpl := msg.MaxPayloadLength()
	if hdr.length > mpl {
		discardInput(r, hdr.length)
		str := fmt.Sprintf("payload exceeds max length - header "+
			"indicates %v bytes, but max payload size for "+
			"messages of type [%v] is %v.", hdr.length, command, mpl)
		err = errors.New(str)
		return nil, err
	}

	payload := make([]byte, hdr.length)
	_, err = io.ReadFull(r, payload)
	if err != nil {
		return nil, fmt.Errorf("error reading payload of command %s: %s", command, err)
	}

	checksum := chainhash.DoubleHashB(payload)[0:4]
	if !bytes.Equal(checksum[:], hdr.checksum[:]) {
		str := fmt.Sprintf("payload checksum failed - header "+
			"indicates %v, but actual checksum is %v.",
			hdr.checksum, checksum)
		err = errors.New(str)
		return nil, err
	}

	err = msg.Unmarshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error decoding payload of command %s: %s", command, err)
	}

	return msg, nil
}
