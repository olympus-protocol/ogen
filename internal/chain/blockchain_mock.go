// Code generated by MockGen. DO NOT EDIT.
// Source: internal/chain/blockchain.go

// Package chain is a generated GoMock package.
package chain

import (
	gomock "github.com/golang/mock/gomock"
	blockdb "github.com/olympus-protocol/ogen/internal/blockdb"
	txindex "github.com/olympus-protocol/ogen/internal/txindex"
	chainhash "github.com/olympus-protocol/ogen/pkg/chainhash"
	primitives "github.com/olympus-protocol/ogen/pkg/primitives"
	reflect "reflect"
	time "time"
)

// MockBlockchain is a mock of Blockchain interface
type MockBlockchain struct {
	ctrl     *gomock.Controller
	recorder *MockBlockchainMockRecorder
}

// MockBlockchainMockRecorder is the mock recorder for MockBlockchain
type MockBlockchainMockRecorder struct {
	mock *MockBlockchain
}

// NewMockBlockchain creates a new mock instance
func NewMockBlockchain(ctrl *gomock.Controller) *MockBlockchain {
	mock := &MockBlockchain{ctrl: ctrl}
	mock.recorder = &MockBlockchainMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockBlockchain) EXPECT() *MockBlockchainMockRecorder {
	return m.recorder
}

// Start mocks base method
func (m *MockBlockchain) Start() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Start")
	ret0, _ := ret[0].(error)
	return ret0
}

// Start indicates an expected call of Start
func (mr *MockBlockchainMockRecorder) Start() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockBlockchain)(nil).Start))
}

// Stop mocks base method
func (m *MockBlockchain) Stop() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Stop")
}

// Stop indicates an expected call of Stop
func (mr *MockBlockchainMockRecorder) Stop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockBlockchain)(nil).Stop))
}

// State mocks base method
func (m *MockBlockchain) State() *StateService {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "State")
	ret0, _ := ret[0].(*StateService)
	return ret0
}

// State indicates an expected call of State
func (mr *MockBlockchainMockRecorder) State() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "State", reflect.TypeOf((*MockBlockchain)(nil).State))
}

// GenesisTime mocks base method
func (m *MockBlockchain) GenesisTime() time.Time {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenesisTime")
	ret0, _ := ret[0].(time.Time)
	return ret0
}

// GenesisTime indicates an expected call of GenesisTime
func (mr *MockBlockchainMockRecorder) GenesisTime() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenesisTime", reflect.TypeOf((*MockBlockchain)(nil).GenesisTime))
}

// GetBlock mocks base method
func (m *MockBlockchain) GetBlock(h chainhash.Hash) (*primitives.Block, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBlock", h)
	ret0, _ := ret[0].(*primitives.Block)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBlock indicates an expected call of GetBlock
func (mr *MockBlockchainMockRecorder) GetBlock(h interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBlock", reflect.TypeOf((*MockBlockchain)(nil).GetBlock), h)
}

// GetRawBlock mocks base method
func (m *MockBlockchain) GetRawBlock(h chainhash.Hash) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRawBlock", h)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRawBlock indicates an expected call of GetRawBlock
func (mr *MockBlockchainMockRecorder) GetRawBlock(h interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRawBlock", reflect.TypeOf((*MockBlockchain)(nil).GetRawBlock), h)
}

// GetAccountTxs mocks base method
func (m *MockBlockchain) GetAccountTxs(acc [20]byte) (txindex.AccountTxs, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAccountTxs", acc)
	ret0, _ := ret[0].(txindex.AccountTxs)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAccountTxs indicates an expected call of GetAccountTxs
func (mr *MockBlockchainMockRecorder) GetAccountTxs(acc interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAccountTxs", reflect.TypeOf((*MockBlockchain)(nil).GetAccountTxs), acc)
}

// GetTx mocks base method
func (m *MockBlockchain) GetTx(h chainhash.Hash) (*primitives.Tx, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTx", h)
	ret0, _ := ret[0].(*primitives.Tx)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetTx indicates an expected call of GetTx
func (mr *MockBlockchainMockRecorder) GetTx(h interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTx", reflect.TypeOf((*MockBlockchain)(nil).GetTx), h)
}

// GetLocatorHashes mocks base method
func (m *MockBlockchain) GetLocatorHashes() [][32]byte {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLocatorHashes")
	ret0, _ := ret[0].([][32]byte)
	return ret0
}

// GetLocatorHashes indicates an expected call of GetLocatorHashes
func (mr *MockBlockchainMockRecorder) GetLocatorHashes() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLocatorHashes", reflect.TypeOf((*MockBlockchain)(nil).GetLocatorHashes))
}

// Notify mocks base method
func (m *MockBlockchain) Notify(n BlockchainNotifee) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Notify", n)
}

// Notify indicates an expected call of Notify
func (mr *MockBlockchainMockRecorder) Notify(n interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Notify", reflect.TypeOf((*MockBlockchain)(nil).Notify), n)
}

// Unnotify mocks base method
func (m *MockBlockchain) Unnotify(n BlockchainNotifee) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Unnotify", n)
}

// Unnotify indicates an expected call of Unnotify
func (mr *MockBlockchainMockRecorder) Unnotify(n interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Unnotify", reflect.TypeOf((*MockBlockchain)(nil).Unnotify), n)
}

// UpdateChainHead mocks base method
func (m *MockBlockchain) UpdateChainHead(txn blockdb.DBUpdateTransaction, possible chainhash.Hash) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateChainHead", txn, possible)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateChainHead indicates an expected call of UpdateChainHead
func (mr *MockBlockchainMockRecorder) UpdateChainHead(txn, possible interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateChainHead", reflect.TypeOf((*MockBlockchain)(nil).UpdateChainHead), txn, possible)
}

// ProcessBlock mocks base method
func (m *MockBlockchain) ProcessBlock(block *primitives.Block) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProcessBlock", block)
	ret0, _ := ret[0].(error)
	return ret0
}

// ProcessBlock indicates an expected call of ProcessBlock
func (mr *MockBlockchainMockRecorder) ProcessBlock(block interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProcessBlock", reflect.TypeOf((*MockBlockchain)(nil).ProcessBlock), block)
}
