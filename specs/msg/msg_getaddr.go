package msg

type MsgGetAddr struct{}

func (m *MsgGetAddr) Command() string {
	return MsgGetAddrCmd
}
