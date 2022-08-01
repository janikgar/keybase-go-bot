package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/keybase/go-keybase-chat-bot/kbchat"
	"github.com/keybase/go-keybase-chat-bot/kbchat/types/chat1"
)

type KeyBaseChat interface {
	ListenForNewTextMessages() (*kbchat.Subscription, error)
	SendReply(channel chat1.ChatChannel, replyTo *chat1.MessageID, body string, args ...interface{}) (kbchat.SendResponse, error)
}

type SubReader interface {
	Read() (kbchat.SubscriptionMessage, error)
}

type Logger interface {
	Fatalf(format string, v ...any)
}

type NativeLogger struct{}

func (n NativeLogger) Fatalf(format string, v ...any) {
	log.Fatalf(format, v...)
}

var (
	kbLoc  string
	kbc    KeyBaseChat
	err    error
	logger Logger
	fail   func(string, ...any)
)

var dotenv = ".env"

func setupEnv() {
	if err := godotenv.Load(dotenv); err != nil {
		fail("could not load .env file: %s", err.Error())
		return
	}

	kbLoc = os.Getenv("KB_LOCATION")
	// hassApiKey = os.Getenv("HASS_API_KEY")
}

func init() {
	setupEnv()
	logger = new(NativeLogger)
	fail = logger.Fatalf
}

func readSub(sub SubReader) (kbchat.SubscriptionMessage, error) {
	msg, err := sub.Read()
	if err != nil {
		return kbchat.SubscriptionMessage{}, fmt.Errorf("message read failed: %s", err.Error())
	}

	if msg.Message.Content.TypeName != "text" {
		return kbchat.SubscriptionMessage{}, errors.New("message read failed: not text")
	}

	return msg, nil
}

func reply(kbc KeyBaseChat, msg kbchat.SubscriptionMessage, reply string) error {
	if reply == "" {
		return nil
	}

	_, err := kbc.SendReply(msg.Message.Channel, &msg.Message.Id, reply)
	if err != nil {
		return fmt.Errorf("error sending reply: %s", err.Error())
	}
	return nil
}

func parseMessages(kbc KeyBaseChat, sub SubReader, httpReq Requests) {
	msg, err := readSub(sub)

	if err != nil {
		fail(err.Error())
		return
	}

	if msg.Message.Content.Text == nil {
		fail("no content")
		return
	}

	body := msg.Message.Content.Text.Body
	input := strings.ToLower(strings.TrimSpace(body))

	fmt.Printf("Input: %s\n", input)

	switch input {
	case "ip":
		ipAddr, err := getIp(httpReq)
		if err != nil {
			fail("could not get ip address: %s", err.Error())
		}
		reply(kbc, msg, ipAddr)
	case "bye":
		os.Exit(0)
	default:
		log.Println(input)
	}
}

func mainLoop(kbc KeyBaseChat, httpReq Requests) {
	log.Println("bot started")

	sub, err := kbc.ListenForNewTextMessages()
	if err != nil {
		fail("could not start subscription: %s", err.Error())
		return
	}

	for {
		parseMessages(kbc, sub, httpReq)
	}
}

func main() {
	if kbc, err = kbchat.Start(kbchat.RunOptions{KeybaseLocation: kbLoc}); err != nil {
		fail("could not start: %s", err.Error())
		return
	}

	httpReq = new(httpRequests)
	mainLoop(kbc, httpReq)
}
