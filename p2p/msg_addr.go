package p2p

import (
	"errors"
	"fmt"
	"io"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/olympus-protocol/ogen/utils/serializer"
)

const MaxAddrPerMsg = 32
const MaxAddrPerPeer = 2

type MsgAddr struct {
	AddrList []peer.AddrInfo
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
		if len(na.Addrs) > MaxAddrPerPeer {
			return fmt.Errorf("too many addresses for message "+
				"[count %v, max %v]", len(na.Addrs), MaxAddrPerPeer)
		}
		if err := serializer.WriteVarInt(w, uint64(len(na.Addrs))); err != nil {
			return err
		}
		for _, a := range na.Addrs {
			b, err := a.MarshalBinary()
			if err != nil {
				return err
			}

			if err := serializer.WriteVarBytes(w, b); err != nil {
				return err
			}
		}

		b, err := na.ID.MarshalBinary()
		if err != nil {
			return err
		}

		if err := serializer.WriteVarBytes(w, b); err != nil {
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
	m.AddrList = make([]peer.AddrInfo, count)
	for i := range m.AddrList {
		countAddr, err := serializer.ReadVarInt(r)
		if err != nil {
			return err
		}

		if countAddr > MaxAddrPerPeer {
			return fmt.Errorf("too many addresses for message (count: %d, max: %d)", countAddr, MaxAddrPerPeer)
		}

		addrs := make([]multiaddr.Multiaddr, countAddr)
		for j := range addrs {
			addrBytes, err := serializer.ReadVarBytes(r)
			if err != nil {
				return err
			}

			if err := addrs[j].UnmarshalBinary(addrBytes); err != nil {
				return err
			}
		}

		peerIDBytes, err := serializer.ReadVarBytes(r)
		if err != nil {
			return err
		}

		if err := m.AddrList[i].ID.UnmarshalBinary(peerIDBytes); err != nil {
			return err
		}

		m.AddrList[i].Addrs = addrs
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
