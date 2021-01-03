// Code generated by MockGen. DO NOT EDIT.
// Source: ./scenarios.go

// Package mock_scenarios is a generated GoMock package.
package mock_scenarios

import (
	context "context"
	reflect "reflect"

	metadata "github.com/curious-kitten/scratch-post/pkg/metadata"
	gomock "github.com/golang/mock/gomock"
)

// MockAdder is a mock of Adder interface
type MockAdder struct {
	ctrl     *gomock.Controller
	recorder *MockAdderMockRecorder
}

// MockAdderMockRecorder is the mock recorder for MockAdder
type MockAdderMockRecorder struct {
	mock *MockAdder
}

// NewMockAdder creates a new mock instance
func NewMockAdder(ctrl *gomock.Controller) *MockAdder {
	mock := &MockAdder{ctrl: ctrl}
	mock.recorder = &MockAdderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockAdder) EXPECT() *MockAdderMockRecorder {
	return m.recorder
}

// AddOne mocks base method
func (m *MockAdder) AddOne(ctx context.Context, item interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddOne", ctx, item)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddOne indicates an expected call of AddOne
func (mr *MockAdderMockRecorder) AddOne(ctx, item interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddOne", reflect.TypeOf((*MockAdder)(nil).AddOne), ctx, item)
}

// MockGetter is a mock of Getter interface
type MockGetter struct {
	ctrl     *gomock.Controller
	recorder *MockGetterMockRecorder
}

// MockGetterMockRecorder is the mock recorder for MockGetter
type MockGetterMockRecorder struct {
	mock *MockGetter
}

// NewMockGetter creates a new mock instance
func NewMockGetter(ctrl *gomock.Controller) *MockGetter {
	mock := &MockGetter{ctrl: ctrl}
	mock.recorder = &MockGetterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockGetter) EXPECT() *MockGetterMockRecorder {
	return m.recorder
}

// Get mocks base method
func (m *MockGetter) Get(ctx context.Context, id string, item interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, id, item)
	ret0, _ := ret[0].(error)
	return ret0
}

// Get indicates an expected call of Get
func (mr *MockGetterMockRecorder) Get(ctx, id, item interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockGetter)(nil).Get), ctx, id, item)
}

// GetAll mocks base method
func (m *MockGetter) GetAll(ctx context.Context, items interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAll", ctx, items)
	ret0, _ := ret[0].(error)
	return ret0
}

// GetAll indicates an expected call of GetAll
func (mr *MockGetterMockRecorder) GetAll(ctx, items interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAll", reflect.TypeOf((*MockGetter)(nil).GetAll), ctx, items)
}

// MockIdentityGenerator is a mock of IdentityGenerator interface
type MockIdentityGenerator struct {
	ctrl     *gomock.Controller
	recorder *MockIdentityGeneratorMockRecorder
}

// MockIdentityGeneratorMockRecorder is the mock recorder for MockIdentityGenerator
type MockIdentityGeneratorMockRecorder struct {
	mock *MockIdentityGenerator
}

// NewMockIdentityGenerator creates a new mock instance
func NewMockIdentityGenerator(ctrl *gomock.Controller) *MockIdentityGenerator {
	mock := &MockIdentityGenerator{ctrl: ctrl}
	mock.recorder = &MockIdentityGeneratorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockIdentityGenerator) EXPECT() *MockIdentityGeneratorMockRecorder {
	return m.recorder
}

// AddMeta mocks base method
func (m *MockIdentityGenerator) AddMeta(author, objType string, identifiable metadata.Identifiable) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddMeta", author, objType, identifiable)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddMeta indicates an expected call of AddMeta
func (mr *MockIdentityGeneratorMockRecorder) AddMeta(author, objType, identifiable interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddMeta", reflect.TypeOf((*MockIdentityGenerator)(nil).AddMeta), author, objType, identifiable)
}