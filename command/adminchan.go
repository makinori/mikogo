package command

import (
	"github.com/makinori/mikogo/irc"
)

func handleAdminChan(c *irc.Client, sender, where string, args []string) {
	c.Send(where, "not implemented lol")
}

var CommandAdminChan = Command{
	Name:        "chan",
	Category:    "admin",
	Description: "manage channels and servers",
	Handle:      handleAdminChan,
}
