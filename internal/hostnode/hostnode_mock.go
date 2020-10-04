// Code generated by MockGen. DO NOT EDIT.
// Source: internal/hostnode/hostnode.go

// Package hostnode is a generated GoMock package.
package hostnode

import (
	gomock "github.com/golang/mock/gomock"
	host "github.com/libp2p/go-libp2p-core/host"
	network "github.com/libp2p/go-libp2p-core/network"
	peer "github.com/libp2p/go-libp2p-core/peer"
	p2p "github.com/olympus-protocol/ogen/pkg/p2p"
	reflect "reflect"
)

// MockHostNode is a mock of HostNode interface
type MockHostNode struct {
	ctrl     *gomock.Controller
	recorder *MockHostNodeMockRecorder
}

// MockHostNodeMockRecorder is the mock recorder for MockHostNode
type MockHostNodeMockRecorder struct {
	mock *MockHostNode
}

// NewMockHostNode creates a new mock instance
func NewMockHostNode(ctrl *gomock.Controller) *MockHostNode {
	mock := &MockHostNode{ctrl: ctrl}
	mock.recorder = &MockHostNodeMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockHostNode) EXPECT() *MockHostNodeMockRecorder {
	return m.recorder
}

// Syncing mocks base method
func (m *MockHostNode) Syncing() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Syncing")
	ret0, _ := ret[0].(bool)
	return ret0
}

// Syncing indicates an expected call of Syncing
func (mr *MockHostNodeMockRecorder) Syncing() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Syncing", reflect.TypeOf((*MockHostNode)(nil).Syncing))
}

// GetHost mocks base method
func (m *MockHostNode) GetHost() host.Host {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetHost")
	ret0, _ := ret[0].(host.Host)
	return ret0
}

// GetHost indicates an expected call of GetHost
func (mr *MockHostNodeMockRecorder) GetHost() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetHost", reflect.TypeOf((*MockHostNode)(nil).GetHost))
}

// GetNetMagic mocks base method
func (m *MockHostNode) GetNetMagic() uint32 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNetMagic")
	ret0, _ := ret[0].(uint32)
	return ret0
}

// GetNetMagic indicates an expected call of GetNetMagic
func (mr *MockHostNodeMockRecorder) GetNetMagic() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNetMagic", reflect.TypeOf((*MockHostNode)(nil).GetNetMagic))
}

// DisconnectPeer mocks base method
func (m *MockHostNode) DisconnectPeer(p peer.ID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DisconnectPeer", p)
	ret0, _ := ret[0].(error)
	return ret0
}

// DisconnectPeer indicates an expected call of DisconnectPeer
func (mr *MockHostNodeMockRecorder) DisconnectPeer(p interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DisconnectPeer", reflect.TypeOf((*MockHostNode)(nil).DisconnectPeer), p)
}

// GetPeerInfos mocks base method
func (m *MockHostNode) GetPeerInfos() []peer.AddrInfo {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPeerInfos")
	ret0, _ := ret[0].([]peer.AddrInfo)
	return ret0
}

// GetPeerInfos indicates an expected call of GetPeerInfos
func (mr *MockHostNodeMockRecorder) GetPeerInfos() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPeerInfos", reflect.TypeOf((*MockHostNode)(nil).GetPeerInfos))
}

// GetPeerDirection mocks base method
func (m *MockHostNode) GetPeerDirection(id peer.ID) network.Direction {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPeerDirection", id)
	ret0, _ := ret[0].(network.Direction)
	return ret0
}

// GetPeerDirection indicates an expected call of GetPeerDirection
func (mr *MockHostNodeMockRecorder) GetPeerDirection(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPeerDirection", reflect.TypeOf((*MockHostNode)(nil).GetPeerDirection), id)
}

// GetPeerInfo mocks base method
func (m *MockHostNode) GetPeerInfo(id peer.ID) *peer.AddrInfo {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPeerInfo", id)
	ret0, _ := ret[0].(*peer.AddrInfo)
	return ret0
}

// GetPeerInfo indicates an expected call of GetPeerInfo
func (mr *MockHostNodeMockRecorder) GetPeerInfo(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPeerInfo", reflect.TypeOf((*MockHostNode)(nil).GetPeerInfo), id)
}

// RegisterHandler mocks base method
func (m *MockHostNode) RegisterHandler(message string, handler MessageHandler) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RegisterHandler", message, handler)
	ret0, _ := ret[0].(error)
	return ret0
}

// RegisterHandler indicates an expected call of RegisterHandler
func (mr *MockHostNodeMockRecorder) RegisterHandler(message, handler interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterHandler", reflect.TypeOf((*MockHostNode)(nil).RegisterHandler), message, handler)
}

// HandleStream mocks base method
func (m *MockHostNode) HandleStream(s network.Stream) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "HandleStream", s)
}

// HandleStream indicates an expected call of HandleStream
func (mr *MockHostNodeMockRecorder) HandleStream(s interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HandleStream", reflect.TypeOf((*MockHostNode)(nil).HandleStream), s)
}

// SendMessage mocks base method
func (m *MockHostNode) SendMessage(id peer.ID, msg p2p.Message) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendMessage", id, msg)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendMessage indicates an expected call of SendMessage
func (mr *MockHostNodeMockRecorder) SendMessage(id, msg interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMessage", reflect.TypeOf((*MockHostNode)(nil).SendMessage), id, msg)
}

// Broadcast mocks base method
func (m *MockHostNode) Broadcast(msg p2p.Message) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Broadcast", msg)
	ret0, _ := ret[0].(error)
	return ret0
}

// Broadcast indicates an expected call of Broadcast
func (mr *MockHostNodeMockRecorder) Broadcast(msg interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Broadcast", reflect.TypeOf((*MockHostNode)(nil).Broadcast), msg)
}
