// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	kbchat "github.com/keybase/go-keybase-chat-bot/kbchat"
	mock "github.com/stretchr/testify/mock"
)

// SubReader is an autogenerated mock type for the SubReader type
type SubReader struct {
	mock.Mock
}

// Read provides a mock function with given fields:
func (_m *SubReader) Read() (kbchat.SubscriptionMessage, error) {
	ret := _m.Called()

	var r0 kbchat.SubscriptionMessage
	if rf, ok := ret.Get(0).(func() kbchat.SubscriptionMessage); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(kbchat.SubscriptionMessage)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewSubReader interface {
	mock.TestingT
	Cleanup(func())
}

// NewSubReader creates a new instance of SubReader. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewSubReader(t mockConstructorTestingTNewSubReader) *SubReader {
	mock := &SubReader{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
