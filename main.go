package main

import (
	"github.com/makinori/mikogo/command"
	"github.com/makinori/mikogo/db"
	"github.com/makinori/mikogo/env"
	"github.com/makinori/mikogo/irc"
)

func handleMessage(msg *irc.Message) {
	// fmt.Printf("%s in %s: %s\n", sender, where, msg)

	command.Run(msg)
}

func main() {
	err := db.Init()
	if err != nil {
		panic(err)
	}

	irc.GlobalHandleMessage = handleMessage
	irc.Init(env.HOME_SERVER)

	keepAlive := make(chan struct{})
	<-keepAlive
}
