package main

import (
	"regexp"
	"strings"

	"github.com/makinori/mikogo/command"
	"github.com/makinori/mikogo/db"
	"github.com/makinori/mikogo/env"
	"github.com/makinori/mikogo/irc"
)

var (
	whiteSpaceRegexp = regexp.MustCompile(`\s+`)
)

func handleMessage(c *irc.Client, sender, where, msg string) {
	// fmt.Printf("%s in %s: %s\n", sender, where, msg)

	if strings.HasPrefix(where, "#") &&
		!strings.HasPrefix(msg, command.Prefix) {
		return
	}

	args := whiteSpaceRegexp.Split(strings.TrimSpace(msg), -1)
	args[0] = strings.TrimPrefix(args[0], command.Prefix)

	command.Run(c, sender, where, args)
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
