// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	kbchat "github.com/keybase/go-keybase-chat-bot/kbchat"
	chat1 "github.com/keybase/go-keybase-chat-bot/kbchat/types/chat1"

	mock "github.com/stretchr/testify/mock"
)

// KeyBaseChat is an autogenerated mock type for the KeyBaseChat type
type KeyBaseChat struct {
	mock.Mock
}

// ListenForNewTextMessages provides a mock function with given fields:
func (_m *KeyBaseChat) ListenForNewTextMessages() (*kbchat.Subscription, error) {
	ret := _m.Called()

	var r0 *kbchat.Subscription
	if rf, ok := ret.Get(0).(func() *kbchat.Subscription); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*kbchat.Subscription)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SendReply provides a mock function with given fields: channel, replyTo, body, args
func (_m *KeyBaseChat) SendReply(channel chat1.ChatChannel, replyTo *chat1.MessageID, body string, args ...interface{}) (kbchat.SendResponse, error) {
	var _ca []interface{}
	_ca = append(_ca, channel, replyTo, body)
	_ca = append(_ca, args...)
	ret := _m.Called(_ca...)

	var r0 kbchat.SendResponse
	if rf, ok := ret.Get(0).(func(chat1.ChatChannel, *chat1.MessageID, string, ...interface{}) kbchat.SendResponse); ok {
		r0 = rf(channel, replyTo, body, args...)
	} else {
		r0 = ret.Get(0).(kbchat.SendResponse)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(chat1.ChatChannel, *chat1.MessageID, string, ...interface{}) error); ok {
		r1 = rf(channel, replyTo, body, args...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewKeyBaseChat interface {
	mock.TestingT
	Cleanup(func())
}

// NewKeyBaseChat creates a new instance of KeyBaseChat. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewKeyBaseChat(t mockConstructorTestingTNewKeyBaseChat) *KeyBaseChat {
	mock := &KeyBaseChat{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
