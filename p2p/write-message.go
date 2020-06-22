package p2p

import (
	"errors"
	"fmt"
	"io"

	"github.com/olympus-protocol/ogen/utils/chainhash"
)

func WriteMessageWithEncodingN(w io.Writer, msg Message, net NetMagic) (int, error) {

	totalBytes := 0

	var command []byte
	cmd := msg.Command()

	copy(command[:], []byte(cmd))

	payload, err := msg.Marshal()
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

	//mpl := msg.MaxPayloadLength()
	//if uint32(lenp) > mpl {
	//	str := fmt.Sprintf("message payload is too large - encoded "+
	//		"%d bytes, but maximum message payload size for "+
	//		"messages of type [%s] is %d.", lenp, cmd, mpl)
	//	err := errors.New(str)
	//	return totalBytes, err
	//}

	hdr := messageHeader{}
	hdr.magic = net
	hdr.command = cmd
	hdr.length = uint32(lenp)
	copy(hdr.checksum[:], chainhash.DoubleHashB(payload)[0:4])

	hwBytes, err := hdr.Marshal()
	if err != nil {
		return totalBytes, err
	}
	n, err := w.Write(hwBytes)
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
