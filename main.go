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

func main() {
	err := db.Init()
	if err != nil {
		panic(err)
	}

	var client *irc.Client

	handleMessage := func(sender, where, msg string) {
		// fmt.Printf("%s in %s: %s\n", sender, where, msg)

		if strings.HasPrefix(where, "#") &&
			!strings.HasPrefix(msg, command.Prefix) {
			return
		}

		args := whiteSpaceRegexp.Split(strings.TrimSpace(msg), -1)
		args[0] = strings.TrimPrefix(args[0], command.Prefix)

		command.Run(client, sender, where, args)
	}

	client, err = irc.Init(env.HOME_SERVER, handleMessage)
	if err != nil {
		panic(err)
	}

	keepAlive := make(chan struct{})
	<-keepAlive
}
