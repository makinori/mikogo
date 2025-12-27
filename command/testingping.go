package command

import (
	"github.com/makinori/mikogo/irc"
)

func handleTestingPing(c *irc.Client, sender, where string, args []string) {
	c.Send(sender, where, "pong!")
}

var CommandTestingPing = Command{
	Name:        "ping",
	Category:    "testing",
	Description: "should pong",
	Handle:      handleTestingPing,
}
