// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/nhost/hasura-storage/controller (interfaces: ContentStorage)

// Package mock_controller is a generated GoMock package.
package mock_controller

import (
	context "context"
	io "io"
	http "net/http"
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
	controller "github.com/nhost/hasura-storage/controller"
)

// MockContentStorage is a mock of ContentStorage interface.
type MockContentStorage struct {
	ctrl     *gomock.Controller
	recorder *MockContentStorageMockRecorder
}

// MockContentStorageMockRecorder is the mock recorder for MockContentStorage.
type MockContentStorageMockRecorder struct {
	mock *MockContentStorage
}

// NewMockContentStorage creates a new mock instance.
func NewMockContentStorage(ctrl *gomock.Controller) *MockContentStorage {
	mock := &MockContentStorage{ctrl: ctrl}
	mock.recorder = &MockContentStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockContentStorage) EXPECT() *MockContentStorageMockRecorder {
	return m.recorder
}

// CreatePresignedURL mocks base method.
func (m *MockContentStorage) CreatePresignedURL(arg0 string, arg1 time.Duration) (string, *controller.APIError) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreatePresignedURL", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(*controller.APIError)
	return ret0, ret1
}

// CreatePresignedURL indicates an expected call of CreatePresignedURL.
func (mr *MockContentStorageMockRecorder) CreatePresignedURL(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreatePresignedURL", reflect.TypeOf((*MockContentStorage)(nil).CreatePresignedURL), arg0, arg1)
}

// DeleteFile mocks base method.
func (m *MockContentStorage) DeleteFile(arg0 string) *controller.APIError {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteFile", arg0)
	ret0, _ := ret[0].(*controller.APIError)
	return ret0
}

// DeleteFile indicates an expected call of DeleteFile.
func (mr *MockContentStorageMockRecorder) DeleteFile(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteFile", reflect.TypeOf((*MockContentStorage)(nil).DeleteFile), arg0)
}

// GetFile mocks base method.
func (m *MockContentStorage) GetFile(arg0 string, arg1 http.Header) (*controller.File, *controller.APIError) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFile", arg0, arg1)
	ret0, _ := ret[0].(*controller.File)
	ret1, _ := ret[1].(*controller.APIError)
	return ret0, ret1
}

// GetFile indicates an expected call of GetFile.
func (mr *MockContentStorageMockRecorder) GetFile(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFile", reflect.TypeOf((*MockContentStorage)(nil).GetFile), arg0, arg1)
}

// GetFileWithPresignedURL mocks base method.
func (m *MockContentStorage) GetFileWithPresignedURL(arg0 context.Context, arg1, arg2 string, arg3 http.Header) (*controller.File, *controller.APIError) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFileWithPresignedURL", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(*controller.File)
	ret1, _ := ret[1].(*controller.APIError)
	return ret0, ret1
}

// GetFileWithPresignedURL indicates an expected call of GetFileWithPresignedURL.
func (mr *MockContentStorageMockRecorder) GetFileWithPresignedURL(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFileWithPresignedURL", reflect.TypeOf((*MockContentStorage)(nil).GetFileWithPresignedURL), arg0, arg1, arg2, arg3)
}

// ListFiles mocks base method.
func (m *MockContentStorage) ListFiles() ([]string, *controller.APIError) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListFiles")
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(*controller.APIError)
	return ret0, ret1
}

// ListFiles indicates an expected call of ListFiles.
func (mr *MockContentStorageMockRecorder) ListFiles() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListFiles", reflect.TypeOf((*MockContentStorage)(nil).ListFiles))
}

// PutFile mocks base method.
func (m *MockContentStorage) PutFile(arg0 io.ReadSeeker, arg1, arg2 string) (string, *controller.APIError) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PutFile", arg0, arg1, arg2)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(*controller.APIError)
	return ret0, ret1
}

// PutFile indicates an expected call of PutFile.
func (mr *MockContentStorageMockRecorder) PutFile(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PutFile", reflect.TypeOf((*MockContentStorage)(nil).PutFile), arg0, arg1, arg2)
}
