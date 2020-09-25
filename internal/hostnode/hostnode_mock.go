// Code generated by MockGen. DO NOT EDIT.
// Source: internal/hostnode/hostnode.go

// Package hostnode is a generated GoMock package.
package hostnode

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	host "github.com/libp2p/go-libp2p-core/host"
	network "github.com/libp2p/go-libp2p-core/network"
	peer "github.com/libp2p/go-libp2p-core/peer"
	protocol "github.com/libp2p/go-libp2p-core/protocol"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
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

// Topic mocks base method
func (m *MockHostNode) Topic(topic string) (*pubsub.Topic, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Topic", topic)
	ret0, _ := ret[0].(*pubsub.Topic)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Topic indicates an expected call of Topic
func (mr *MockHostNodeMockRecorder) Topic(topic interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Topic", reflect.TypeOf((*MockHostNode)(nil).Topic), topic)
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

// GetContext mocks base method
func (m *MockHostNode) GetContext() context.Context {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetContext")
	ret0, _ := ret[0].(context.Context)
	return ret0
}

// GetContext indicates an expected call of GetContext
func (mr *MockHostNodeMockRecorder) GetContext() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetContext", reflect.TypeOf((*MockHostNode)(nil).GetContext))
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

// IsConnected mocks base method
func (m *MockHostNode) IsConnected() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsConnected")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsConnected indicates an expected call of IsConnected
func (mr *MockHostNodeMockRecorder) IsConnected() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsConnected", reflect.TypeOf((*MockHostNode)(nil).IsConnected))
}

// PeersConnected mocks base method
func (m *MockHostNode) PeersConnected() int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PeersConnected")
	ret0, _ := ret[0].(int)
	return ret0
}

// PeersConnected indicates an expected call of PeersConnected
func (mr *MockHostNodeMockRecorder) PeersConnected() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PeersConnected", reflect.TypeOf((*MockHostNode)(nil).PeersConnected))
}

// GetPeerList mocks base method
func (m *MockHostNode) GetPeerList() []peer.ID {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPeerList")
	ret0, _ := ret[0].([]peer.ID)
	return ret0
}

// GetPeerList indicates an expected call of GetPeerList
func (mr *MockHostNodeMockRecorder) GetPeerList() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPeerList", reflect.TypeOf((*MockHostNode)(nil).GetPeerList))
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

// ConnectedToPeer mocks base method
func (m *MockHostNode) ConnectedToPeer(id peer.ID) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ConnectedToPeer", id)
	ret0, _ := ret[0].(bool)
	return ret0
}

// ConnectedToPeer indicates an expected call of ConnectedToPeer
func (mr *MockHostNodeMockRecorder) ConnectedToPeer(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ConnectedToPeer", reflect.TypeOf((*MockHostNode)(nil).ConnectedToPeer), id)
}

// Notify mocks base method
func (m *MockHostNode) Notify(notifee network.Notifiee) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Notify", notifee)
}

// Notify indicates an expected call of Notify
func (mr *MockHostNodeMockRecorder) Notify(notifee interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Notify", reflect.TypeOf((*MockHostNode)(nil).Notify), notifee)
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

// Stop mocks base method
func (m *MockHostNode) Stop() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Stop")
}

// Stop indicates an expected call of Stop
func (mr *MockHostNodeMockRecorder) Stop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockHostNode)(nil).Stop))
}

// Start mocks base method
func (m *MockHostNode) Start() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Start")
	ret0, _ := ret[0].(error)
	return ret0
}

// Start indicates an expected call of Start
func (mr *MockHostNodeMockRecorder) Start() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockHostNode)(nil).Start))
}

// SetStreamHandler mocks base method
func (m *MockHostNode) SetStreamHandler(id protocol.ID, handleStream func(network.Stream)) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetStreamHandler", id, handleStream)
}

// SetStreamHandler indicates an expected call of SetStreamHandler
func (mr *MockHostNodeMockRecorder) SetStreamHandler(id, handleStream interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetStreamHandler", reflect.TypeOf((*MockHostNode)(nil).SetStreamHandler), id, handleStream)
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
