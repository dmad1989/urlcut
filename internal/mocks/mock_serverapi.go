// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/dmad1989/urlcut/internal/serverapi (interfaces: App,Conf)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"

	jsonobject "github.com/dmad1989/urlcut/internal/jsonobject"
)

// MockApp is a mock of App interface.
type MockApp struct {
	ctrl     *gomock.Controller
	recorder *MockAppMockRecorder
}

// MockAppMockRecorder is the mock recorder for MockApp.
type MockAppMockRecorder struct {
	mock *MockApp
}

// NewMockApp creates a new mock instance.
func NewMockApp(ctrl *gomock.Controller) *MockApp {
	mock := &MockApp{ctrl: ctrl}
	mock.recorder = &MockAppMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockApp) EXPECT() *MockAppMockRecorder {
	return m.recorder
}

// Cut mocks base method.
func (m *MockApp) Cut(arg0 context.Context, arg1 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Cut", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Cut indicates an expected call of Cut.
func (mr *MockAppMockRecorder) Cut(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Cut", reflect.TypeOf((*MockApp)(nil).Cut), arg0, arg1)
}

// DeleteUrls mocks base method.
func (m *MockApp) DeleteUrls(arg0 string, arg1 jsonobject.ShortIds) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "DeleteUrls", arg0, arg1)
}

// DeleteUrls indicates an expected call of DeleteUrls.
func (mr *MockAppMockRecorder) DeleteUrls(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteUrls", reflect.TypeOf((*MockApp)(nil).DeleteUrls), arg0, arg1)
}

// GetKeyByValue mocks base method.
func (m *MockApp) GetKeyByValue(arg0 context.Context, arg1 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetKeyByValue", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetKeyByValue indicates an expected call of GetKeyByValue.
func (mr *MockAppMockRecorder) GetKeyByValue(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetKeyByValue", reflect.TypeOf((*MockApp)(nil).GetKeyByValue), arg0, arg1)
}

// GetUserURLs mocks base method.
func (m *MockApp) GetUserURLs(arg0 context.Context) (jsonobject.Batch, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserURLs", arg0)
	ret0, _ := ret[0].(jsonobject.Batch)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserURLs indicates an expected call of GetUserURLs.
func (mr *MockAppMockRecorder) GetUserURLs(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserURLs", reflect.TypeOf((*MockApp)(nil).GetUserURLs), arg0)
}

// PingDB mocks base method.
func (m *MockApp) PingDB(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PingDB", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// PingDB indicates an expected call of PingDB.
func (mr *MockAppMockRecorder) PingDB(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PingDB", reflect.TypeOf((*MockApp)(nil).PingDB), arg0)
}

// UploadBatch mocks base method.
func (m *MockApp) UploadBatch(arg0 context.Context, arg1 jsonobject.Batch) (jsonobject.Batch, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UploadBatch", arg0, arg1)
	ret0, _ := ret[0].(jsonobject.Batch)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UploadBatch indicates an expected call of UploadBatch.
func (mr *MockAppMockRecorder) UploadBatch(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UploadBatch", reflect.TypeOf((*MockApp)(nil).UploadBatch), arg0, arg1)
}

// MockConf is a mock of Conf interface.
type MockConf struct {
	ctrl     *gomock.Controller
	recorder *MockConfMockRecorder
}

// MockConfMockRecorder is the mock recorder for MockConf.
type MockConfMockRecorder struct {
	mock *MockConf
}

// NewMockConf creates a new mock instance.
func NewMockConf(ctrl *gomock.Controller) *MockConf {
	mock := &MockConf{ctrl: ctrl}
	mock.recorder = &MockConfMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockConf) EXPECT() *MockConfMockRecorder {
	return m.recorder
}

// GetShortAddress mocks base method.
func (m *MockConf) GetShortAddress() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetShortAddress")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetShortAddress indicates an expected call of GetShortAddress.
func (mr *MockConfMockRecorder) GetShortAddress() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetShortAddress", reflect.TypeOf((*MockConf)(nil).GetShortAddress))
}

// GetURL mocks base method.
func (m *MockConf) GetURL() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetURL")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetURL indicates an expected call of GetURL.
func (mr *MockConfMockRecorder) GetURL() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetURL", reflect.TypeOf((*MockConf)(nil).GetURL))
}
