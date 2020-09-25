// Code generated by MockGen. DO NOT EDIT.
// Source: internal/hostnode/database.go

// Package hostnode is a generated GoMock package.
package hostnode

import (
	gomock "github.com/golang/mock/gomock"
	crypto "github.com/libp2p/go-libp2p-core/crypto"
	peer "github.com/libp2p/go-libp2p-core/peer"
	reflect "reflect"
)

// MockDatabase is a mock of Database interface
type MockDatabase struct {
	ctrl     *gomock.Controller
	recorder *MockDatabaseMockRecorder
}

// MockDatabaseMockRecorder is the mock recorder for MockDatabase
type MockDatabaseMockRecorder struct {
	mock *MockDatabase
}

// NewMockDatabase creates a new mock instance
func NewMockDatabase(ctrl *gomock.Controller) *MockDatabase {
	mock := &MockDatabase{ctrl: ctrl}
	mock.recorder = &MockDatabaseMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockDatabase) EXPECT() *MockDatabaseMockRecorder {
	return m.recorder
}

// SavePeer mocks base method
func (m *MockDatabase) SavePeer(pinfo peer.AddrInfo) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SavePeer", pinfo)
	ret0, _ := ret[0].(error)
	return ret0
}

// SavePeer indicates an expected call of SavePeer
func (mr *MockDatabaseMockRecorder) SavePeer(pinfo interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SavePeer", reflect.TypeOf((*MockDatabase)(nil).SavePeer), pinfo)
}

// BanscorePeer mocks base method
func (m *MockDatabase) BanscorePeer(pinfo peer.AddrInfo, weight uint16) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BanscorePeer", pinfo, weight)
	ret0, _ := ret[0].(error)
	return ret0
}

// BanscorePeer indicates an expected call of BanscorePeer
func (mr *MockDatabaseMockRecorder) BanscorePeer(pinfo, weight interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BanscorePeer", reflect.TypeOf((*MockDatabase)(nil).BanscorePeer), pinfo, weight)
}

// GetSavedPeers mocks base method
func (m *MockDatabase) GetSavedPeers() ([]peer.AddrInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSavedPeers")
	ret0, _ := ret[0].([]peer.AddrInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSavedPeers indicates an expected call of GetSavedPeers
func (mr *MockDatabaseMockRecorder) GetSavedPeers() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSavedPeers", reflect.TypeOf((*MockDatabase)(nil).GetSavedPeers))
}

// GetPrivKey mocks base method
func (m *MockDatabase) GetPrivKey() (crypto.PrivKey, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPrivKey")
	ret0, _ := ret[0].(crypto.PrivKey)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPrivKey indicates an expected call of GetPrivKey
func (mr *MockDatabaseMockRecorder) GetPrivKey() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPrivKey", reflect.TypeOf((*MockDatabase)(nil).GetPrivKey))
}
