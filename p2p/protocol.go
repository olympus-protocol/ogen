package p2p

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	ProtocolVersion uint32 = 70000
)

type NetMagic uint32

type ServiceFlag uint64

func (f ServiceFlag) String() string {
	if f == 0 {
		return "0x0"
	}
	s := ""
	for _, flag := range orderedSFStrings {
		if f&flag == flag {
			s += sfStrings[flag] + "|"
			f -= flag
		}
	}
	s = strings.TrimRight(s, "|")
	if f != 0 {
		s += "|0x" + strconv.FormatUint(uint64(f), 16)
	}
	s = strings.TrimLeft(s, "|")
	return s
}

const (
	Node ServiceFlag = 1 << iota
	MasterNode
)

var sfStrings = map[ServiceFlag]string{
	Node:       "Node",
	MasterNode: "MasterNode",
}

var orderedSFStrings = []ServiceFlag{
	Node,
	MasterNode,
}

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
