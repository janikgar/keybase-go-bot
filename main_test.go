package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/janikgar/keybase-go-bot/mocks"
	"github.com/keybase/go-keybase-chat-bot/kbchat"
	"github.com/keybase/go-keybase-chat-bot/kbchat/types/chat1"
	"github.com/stretchr/testify/require"
)

func createTextMessage(msg string) kbchat.SubscriptionMessage {
	return kbchat.SubscriptionMessage{
		Message: chat1.MsgSummary{
			Id: 1,
			Channel: chat1.ChatChannel{
				Name:        "test",
				Public:      true,
				MembersType: "a",
				TopicType:   "b",
				TopicName:   "c",
			},
			Content: chat1.MsgContent{
				TypeName: "text",
				Text: &chat1.MsgTextContent{
					Body: msg,
				},
			},
		},
		Conversation: chat1.ConvSummary{},
	}
}

func createNonTextMessage(msg string) kbchat.SubscriptionMessage {
	return kbchat.SubscriptionMessage{
		Message: chat1.MsgSummary{
			Id: 1,
			Channel: chat1.ChatChannel{
				Name:        "test",
				Public:      true,
				MembersType: "a",
				TopicType:   "b",
				TopicName:   "c",
			},
			Content: chat1.MsgContent{
				TypeName: "image",
				Text: &chat1.MsgTextContent{
					Body: msg,
				},
			},
		},
		Conversation: chat1.ConvSummary{},
	}
}

func TestReadSub(t *testing.T) {
	testMessage := createTextMessage("test")

	nonTextMessage := kbchat.SubscriptionMessage{
		Message: chat1.MsgSummary{
			Content: chat1.MsgContent{
				TypeName: "image",
			},
		},
		Conversation: chat1.ConvSummary{},
	}

	emptyMessage := kbchat.SubscriptionMessage{}

	cases := []struct {
		msg         kbchat.SubscriptionMessage
		expectedMsg kbchat.SubscriptionMessage
		err         error
	}{
		{testMessage, testMessage, nil},
		{testMessage, emptyMessage, errors.New("test")},
		{nonTextMessage, emptyMessage, errors.New("not text")},
	}

	for _, c := range cases {
		sub := mocks.NewSubReader(t)
		sub.On("Read").Return(c.msg, c.err)

		msg, err := readSub(sub)

		if err != nil {
			require.Equal(t, emptyMessage, msg)
			require.Contains(t, err.Error(), c.err.Error())
		} else {
			require.Equal(t, c.msg.Message.Content.Text, msg.Message.Content.Text)
		}
	}
}

func captureOutput(t *testing.T, f func()) string {
	var buf bytes.Buffer
	log.SetOutput(&buf)

	fail = func(format string, v ...any) {
		log.Printf(format, v...)
	}

	f()

	fakeStdout, err := ioutil.ReadAll(&buf)
	log.SetOutput(os.Stdout)

	if err != nil {
		t.Error(err.Error())
	}

	return string(fakeStdout)
}

func TestInitDotenv(t *testing.T) {
	dotenv = "itdoesnotexist"

	fakeStdout := captureOutput(t, func() { setupEnv() })

	require.Contains(t, fakeStdout, "could not load")
}

func TestParseMessages(t *testing.T) {
	kbc := mocks.NewKeyBaseChat(t)

	cases := []struct {
		message          kbchat.SubscriptionMessage
		expectedOutput   any
		expectedError    error
		expectedResponse string
	}{
		{createTextMessage("test"), "test", nil, ""},
		{createTextMessage("fail"), "fail", errors.New("fail"), ""},
		{createTextMessage("ip"), "looking up", nil, "1.1.1.1"},
		{createTextMessage("ip"), "message read failed", errors.New("ip"), "could not get ip address"},
		{createNonTextMessage("nontext"), "nontext", errors.New("nontext"), "nontext"},
	}

	for _, c := range cases {
		sub := mocks.NewSubReader(t)
		sub.On("Read").Return(c.message, c.expectedError).Maybe()

		kbc.On("SendReply", c.message.Message.Channel, &c.message.Message.Id, c.expectedResponse).Return(
			kbchat.SendResponse{},
			nil,
		).Maybe()

		body, bodyWrite := io.Pipe()
		go func() {
			fmt.Fprint(bodyWrite, c.expectedResponse)
			bodyWrite.Close()
		}()

		httpReq := mocks.NewRequests(t)
		httpReq.On("Get", "https://api.ipify.org").Return(&http.Response{
			StatusCode: 200,
			Body:       body,
		}, nil).Maybe()

		if c.expectedError != nil && c.expectedError.Error() == "ip" {
			fmt.Println(c.expectedError.Error())
			fakeStdout := captureOutput(t, func() { parseMessages(kbc, sub, httpReq) })

			if c.expectedOutput != "" {
				require.Contains(t, fakeStdout, c.expectedOutput)
			}
			require.Contains(t, fakeStdout, c.expectedError.Error())
		} else if c.expectedError != nil {
			fakeStdout := captureOutput(t, func() { parseMessages(kbc, sub, httpReq) })

			if c.expectedOutput != "" {
				require.Contains(t, fakeStdout, c.expectedOutput)
			}
		} else {
			fakeStdout := captureOutput(t, func() { parseMessages(kbc, sub, httpReq) })

			if c.expectedOutput != "" {
				require.Contains(t, fakeStdout, c.expectedOutput)
			}
		}

	}
}

func TestReply(t *testing.T) {
	kbc := mocks.NewKeyBaseChat(t)
	msg := createTextMessage("reply to this")

	kbc.On("SendReply", msg.Message.Channel, &msg.Message.Id, "this is a reply").Return(
		kbchat.SendResponse{},
		nil,
	)

	err := reply(kbc, msg, "this is a reply")
	require.Nil(t, err)
}

func TestMainLoop(t *testing.T) {
	httpReq := mocks.NewRequests(t)

	cases := []struct {
		sub *kbchat.Subscription
		err error
	}{
		{kbchat.NewSubscription(), nil},
		{kbchat.NewSubscription(), errors.New("fail")},
	}

	for _, c := range cases {
		kbc := mocks.NewKeyBaseChat(t)
		kbc.On("ListenForNewTextMessages").Return(c.sub, c.err)

		fakeStdout := captureOutput(t, func() {
			go mainLoop(kbc, httpReq)
			func() {
				time.Sleep(time.Second * 2)
				reply(kbc, createTextMessage("bye"), "")
			}()
		})

		require.Contains(t, fakeStdout, "bot started")
		if c.err != nil {
			fmt.Printf("expected error: %+v\n", c.err)
			require.Contains(t, fakeStdout, "could not start subscription")
		}
	}

}

func TestMain(t *testing.T) {
	kbLoc = "itdoesnotexist"

	fakeStdout := captureOutput(t, func() { main() })

	require.Contains(t, fakeStdout, "could not start")
	require.NotPanics(t, func() { main() })
}
