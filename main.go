package main

import (
	"flag"
	"fmt"

	"github.com/keybase/go-keybase-chat-bot/kbchat"
)

var (
	kbLoc string
	kbc   *kbchat.API
	err   error
)

func main() {
	flag.StringVar(&kbLoc, "keybase", "keybase", "location of the Keybase app")
	flag.Parse()

	fmt.Printf("%+v\n", kbLoc)
}
