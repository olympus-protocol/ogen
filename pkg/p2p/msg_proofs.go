package p2p

import "github.com/olympus-protocol/ogen/pkg/burnproof"

// MsgProofs is the struct that contains the node information during the version handshake.
type MsgProofs struct {
	Proofs []*burnproof.CoinsProofSerializable `ssz-max:"2048"`
}

// Marshal serializes the data to bytes
func (m *MsgProofs) Marshal() ([]byte, error) {
	return m.MarshalSSZ()
}

// Unmarshal deserializes the data
func (m *MsgProofs) Unmarshal(b []byte) error {
	return m.UnmarshalSSZ(b)
}

// Command returns the message topic
func (m *MsgProofs) Command() string {
	return MsgProofsCmd
}

// MaxPayloadLength returns the maximum size of the MsgVersion message.
func (m *MsgProofs) MaxPayloadLength() uint64 {
	return 2297 * 2048
}
