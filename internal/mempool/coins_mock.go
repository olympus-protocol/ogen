// Code generated by MockGen. DO NOT EDIT.
// Source: internal/mempool/coins.go

// Package mempool is a generated GoMock package.
package mempool

import (
	gomock "github.com/golang/mock/gomock"
	state "github.com/olympus-protocol/ogen/internal/state"
	primitives "github.com/olympus-protocol/ogen/pkg/primitives"
	reflect "reflect"
)

// MockCoinsMempool is a mock of CoinsMempool interface
type MockCoinsMempool struct {
	ctrl     *gomock.Controller
	recorder *MockCoinsMempoolMockRecorder
}

// MockCoinsMempoolMockRecorder is the mock recorder for MockCoinsMempool
type MockCoinsMempoolMockRecorder struct {
	mock *MockCoinsMempool
}

// NewMockCoinsMempool creates a new mock instance
func NewMockCoinsMempool(ctrl *gomock.Controller) *MockCoinsMempool {
	mock := &MockCoinsMempool{ctrl: ctrl}
	mock.recorder = &MockCoinsMempoolMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockCoinsMempool) EXPECT() *MockCoinsMempoolMockRecorder {
	return m.recorder
}

// Add mocks base method
func (m *MockCoinsMempool) Add(item primitives.Tx, state *primitives.CoinsState) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Add", item, state)
	ret0, _ := ret[0].(error)
	return ret0
}

// Add indicates an expected call of Add
func (mr *MockCoinsMempoolMockRecorder) Add(item, state interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockCoinsMempool)(nil).Add), item, state)
}

// RemoveByBlock mocks base method
func (m *MockCoinsMempool) RemoveByBlock(b *primitives.Block) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RemoveByBlock", b)
}

// RemoveByBlock indicates an expected call of RemoveByBlock
func (mr *MockCoinsMempoolMockRecorder) RemoveByBlock(b interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveByBlock", reflect.TypeOf((*MockCoinsMempool)(nil).RemoveByBlock), b)
}

// Get mocks base method
func (m *MockCoinsMempool) Get(maxTransactions uint64, s state.State) ([]*primitives.Tx, state.State) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", maxTransactions, s)
	ret0, _ := ret[0].([]*primitives.Tx)
	ret1, _ := ret[1].(state.State)
	return ret0, ret1
}

// Get indicates an expected call of Get
func (mr *MockCoinsMempoolMockRecorder) Get(maxTransactions, s interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockCoinsMempool)(nil).Get), maxTransactions, s)
}
