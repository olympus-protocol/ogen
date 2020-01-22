package serializer

import (
	"encoding/binary"
	"io"
	"net"
	"time"
)

type NetAddress struct {
	Timestamp int64  // 8 bytes
	IP        net.IP // 16 bytes
	Port      uint16 // 2 bytes
}

func NewNetAddress(
	timestamp time.Time, ip net.IP, port uint16) *NetAddress {
	na := NetAddress{
		Timestamp: time.Unix(timestamp.Unix(), 0).Unix(),
		IP:        ip,
		Port:      port,
	}
	return &na
}

func WriteNetAddress(w io.Writer, na *NetAddress) error {
	err := WriteElement(w, uint64(na.Timestamp))
	if err != nil {
		return err
	}
	var ip [16]byte
	if na.IP != nil {
		copy(ip[:], na.IP.To16())
	}
	err = WriteElements(w, ip)
	if err != nil {
		return err
	}

	return binary.Write(w, bigEndian, na.Port)
}

func ReadNetAddress(r io.Reader, na *NetAddress) error {
	var ip [16]byte
	err := ReadElement(r, &na.Timestamp)
	if err != nil {
		return err
	}
	err = ReadElements(r, &ip)
	if err != nil {
		return err
	}
	port, err := binarySerializer.Uint16(r, bigEndian)
	if err != nil {
		return err
	}
	*na = NetAddress{
		Timestamp: na.Timestamp,
		IP:        net.IP(ip[:]),
		Port:      port,
	}
	return nil
}
