package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
	"github.com/keybase/go-keybase-chat-bot/kbchat"
	"github.com/keybase/go-keybase-chat-bot/kbchat/types/chat1"
)

type KeyBaseChat interface {
	ListenForNewTextMessages() (*kbchat.Subscription, error)
	SendReply(channel chat1.ChatChannel, replyTo *chat1.MessageID, body string, args ...interface{}) (kbchat.SendResponse, error)
	AdvertiseCommands(ad kbchat.Advertisement) (kbchat.SendResponse, error)
}

type SubReader interface {
	Read() (kbchat.SubscriptionMessage, error)
}

type Logger interface {
	Printf(format string, v ...any)
}

type NativeLogger struct{}

func (n NativeLogger) Printf(format string, v ...any) {
	log.Printf(format, v...)
}

var (
	kbLoc      string
	kbc        KeyBaseChat
	err        error
	logger     Logger
	fail       func(string, ...any)
	hassApiKey string
	exitFunc   func(int)
)

var dotenv = ".env"

func setupEnv() {
	if err := godotenv.Load(dotenv); err != nil {
		fail("could not load .env file: %s", err.Error())
		return
	}

	kbLoc = os.Getenv("KB_LOCATION")
	hassApiKey = os.Getenv("HASS_API_KEY")
}

func init() {
	setupEnv()
	logger = new(NativeLogger)
	fail = logger.Printf
	exitFunc = os.Exit
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

	body := msg.Message.Content.Text.Body
	input := strings.ToLower(strings.TrimSpace(body))

	ip := regexp.MustCompile(`^!ip`)
	bye := regexp.MustCompile(`^!bye`)
	home := regexp.MustCompile(`^!home`)

	cmds := []chat1.UserBotCommandInput{
		{
			Name:        "ip",
			Description: "Get current IP address",
		},
		{
			Name:        "bye",
			Description: "Kill the bot",
		},
		{
			Name:        "home",
			Description: "Interact with home automation",
		},
	}

	adv := kbchat.Advertisement{
		Alias: "j2bot",
		Advertisements: []chat1.AdvertiseCommandAPIParam{
			{
				Typ:      "public",
				Commands: cmds,
			},
		},
	}

	_, err = kbc.AdvertiseCommands(adv)
	if err != nil {
		fail("could not advertise commands: %s", err.Error())
	}

	if ip.MatchString(input) {
		ipAddr, err := getIp(httpReq)
		if err != nil {
			fail("could not get ip address: %s", err.Error())
		}
		reply(kbc, msg, ipAddr)
	} else if bye.MatchString(input) {
		exitFunc(0)
	} else if home.MatchString(input) {
		hassStrings := strings.Split(input, " ")
		hassString := strings.Join(hassStrings[1:], "/")
		hassUrl := fmt.Sprintf("http://home-assistant.home.lan:8123/api/%s", hassString)
		hassOutput, err := getFromHass(httpReq, hassUrl)
		if err != nil {
			fail("error communicating with Home Assistant: %s", err.Error())
		}
		log.Println(hassOutput)
		reply(kbc, msg, hassOutput)
	} else {
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
	options := kbchat.RunOptions{
		KeybaseLocation: kbLoc,
		StartService:    true,
	}
	if kbc, err = kbchat.Start(options); err != nil {
		fail("could not start: %s", err.Error())
		return
	}

	httpReq = new(httpRequests)
	mainLoop(kbc, httpReq)
}
