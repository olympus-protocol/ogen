package p2p

import (
	"errors"
	"fmt"
	"github.com/olympus-protocol/ogen/utils/serializer"
	"io"
)

const MaxAddrPerMsg = 1000

type MsgAddr struct {
	AddrList []*serializer.NetAddress
}

func (m *MsgAddr) AddAddress(na *serializer.NetAddress) error {
	if len(m.AddrList)+1 > MaxAddrPerMsg {
		str := fmt.Sprintf("too many addresses in message [max %v]",
			MaxAddrPerMsg)
		return errors.New(str)
	}

	m.AddrList = append(m.AddrList, na)
	return nil
}

func (m *MsgAddr) AddAddresses(netAddrs ...*serializer.NetAddress) error {
	for _, na := range netAddrs {
		err := m.AddAddress(na)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *MsgAddr) Encode(w io.Writer) error {
	count := len(m.AddrList)
	if count > MaxAddrPerMsg {
		str := fmt.Sprintf("too many addresses for message "+
			"[count %v, max %v]", count, MaxAddrPerMsg)
		return errors.New(str)
	}
	err := serializer.WriteVarInt(w, uint64(count))
	if err != nil {
		return err
	}
	for _, na := range m.AddrList {
		err = serializer.WriteNetAddress(w, na)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *MsgAddr) Decode(r io.Reader) error {
	count, err := serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	if count > MaxAddrPerMsg {
		str := fmt.Sprintf("too many addresses for message "+
			"[count %v, max %v]", count, MaxAddrPerMsg)
		return errors.New(str)
	}
	addrList := make([]serializer.NetAddress, count)
	m.AddrList = make([]*serializer.NetAddress, 0, count)
	for i := uint64(0); i < count; i++ {
		na := &addrList[i]
		err := serializer.ReadNetAddress(r, na)
		if err != nil {
			return err
		}
		err = m.AddAddress(na)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *MsgAddr) Command() string {
	return MsgAddrCmd
}

func (m *MsgAddr) MaxPayloadLength() uint32 {
	netAddressSize := 26 // Max NetAddress size
	return uint32(MaxAddrPerMsg * netAddressSize)
}

func NewMsgAddr() *MsgAddr {
	return &MsgAddr{}
}
