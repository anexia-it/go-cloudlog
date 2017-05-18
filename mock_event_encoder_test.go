// Automatically generated by MockGen. DO NOT EDIT!
// Source: github.com/anexia-it/go-cloudlog (interfaces: EventEncoder)

package cloudlog

import (
	gomock "github.com/golang/mock/gomock"
)

// Mock of EventEncoder interface
type MockEventEncoder struct {
	ctrl     *gomock.Controller
	recorder *_MockEventEncoderRecorder
}

// Recorder for MockEventEncoder (not exported)
type _MockEventEncoderRecorder struct {
	mock *MockEventEncoder
}

func NewMockEventEncoder(ctrl *gomock.Controller) *MockEventEncoder {
	mock := &MockEventEncoder{ctrl: ctrl}
	mock.recorder = &_MockEventEncoderRecorder{mock}
	return mock
}

func (_m *MockEventEncoder) EXPECT() *_MockEventEncoderRecorder {
	return _m.recorder
}

func (_m *MockEventEncoder) EncodeEvent(_param0 interface{}) (map[string]interface{}, error) {
	ret := _m.ctrl.Call(_m, "EncodeEvent", _param0)
	ret0, _ := ret[0].(map[string]interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockEventEncoderRecorder) EncodeEvent(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "EncodeEvent", arg0)
}