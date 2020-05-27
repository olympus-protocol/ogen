package p2p

import (
	"fmt"
)

const (
	ProtocolVersion uint32 = 70000
)

type NetMagic uint32

const (
	MainNet NetMagic = 0xd9b4bef9
)

var bnStrings = map[NetMagic]string{
	MainNet: "MainNet",
}

func (n NetMagic) String() string {
	if s, ok := bnStrings[n]; ok {
		return s
	}

	return fmt.Sprintf("Unknown Net (%d)", uint32(n))
}
