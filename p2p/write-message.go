package p2p

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
)

func WriteMessageWithEncodingN(w io.Writer, msg Message, net NetMagic) (int, error) {

	totalBytes := 0

	var command [serializer.CommandSize]byte
	cmd := msg.Command()
	if len(cmd) > serializer.CommandSize {
		str := fmt.Sprintf("command [%s] is too long [max %v]",
			cmd, serializer.CommandSize)
		err := errors.New(str)
		return totalBytes, err
	}
	copy(command[:], []byte(cmd))

	payload, err := msg.Marshal(&bw)
	if err != nil {
		return totalBytes, err
	}
	lenp := len(payload)

	if lenp > MaxMessagePayload {
		str := fmt.Sprintf("message payload is too large - encoded "+
			"%d bytes, but maximum message payload is %d bytes",
			lenp, MaxMessagePayload)
		err := errors.New(str)
		return totalBytes, err
	}

	mpl := msg.MaxPayloadLength()
	if uint32(lenp) > mpl {
		str := fmt.Sprintf("message payload is too large - encoded "+
			"%d bytes, but maximum message payload size for "+
			"messages of type [%s] is %d.", lenp, cmd, mpl)
		err := errors.New(str)
		return totalBytes, err
	}

	hdr := messageHeader{}
	hdr.magic = net
	hdr.command = cmd
	hdr.length = uint32(lenp)
	copy(hdr.checksum[:], chainhash.DoubleHashB(payload)[0:4])

	hw := bytes.NewBuffer(make([]byte, 0, MessageHeaderSize))
	_ = serializer.WriteElements(hw, hdr.magic, command, hdr.length, hdr.checksum)

	n, err := w.Write(hw.Bytes())
	totalBytes += n
	if err != nil {
		return totalBytes, err
	}

	n, err = w.Write(payload)
	totalBytes += n
	return totalBytes, err
}

func WriteMessageN(w io.Writer, msg Message, net NetMagic) (int, error) {
	return WriteMessageWithEncodingN(w, msg, net)
}

func WriteMessage(w io.Writer, msg Message, net NetMagic) error {
	_, err := WriteMessageN(w, msg, net)
	return err
}
