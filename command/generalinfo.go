package command

import (
	"strings"

	"github.com/makinori/mikogo/irc"
)

func handleInfo(c *irc.Client, sender, where string, args []string) {
	// TODO: add git hash
	out := "hi im mikogo\n"
	out += "made by: https://maki.cafe\n"
	out += "named by: https://micae.la\n"
	out += "https://github.com/makinori/mikogo\n"
	c.Send(sender, where, strings.TrimSpace(out))
}

var CommandGeneralInfo = Command{
	Name:        "info",
	Category:    "general",
	Description: "about me",
	OwnerOnly:   false,
	Handle:      handleInfo,
}
