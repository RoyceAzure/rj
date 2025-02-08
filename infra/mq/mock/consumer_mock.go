// Code generated by MockGen. DO NOT EDIT.
// Source: ./consumer.go

// Package mockmq is a generated GoMock package.
package mockmq

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockIConsumer is a mock of IConsumer interface.
type MockIConsumer struct {
	ctrl     *gomock.Controller
	recorder *MockIConsumerMockRecorder
}

// MockIConsumerMockRecorder is the mock recorder for MockIConsumer.
type MockIConsumerMockRecorder struct {
	mock *MockIConsumer
}

// NewMockIConsumer creates a new mock instance.
func NewMockIConsumer(ctrl *gomock.Controller) *MockIConsumer {
	mock := &MockIConsumer{ctrl: ctrl}
	mock.recorder = &MockIConsumerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIConsumer) EXPECT() *MockIConsumerMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockIConsumer) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockIConsumerMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockIConsumer)(nil).Close))
}

// Consume mocks base method.
func (m *MockIConsumer) Consume(queueName string, handler func([]byte) error) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Consume", queueName, handler)
	ret0, _ := ret[0].(error)
	return ret0
}

// Consume indicates an expected call of Consume.
func (mr *MockIConsumerMockRecorder) Consume(queueName, handler interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Consume", reflect.TypeOf((*MockIConsumer)(nil).Consume), queueName, handler)
}
