package command

import (
	"strings"

	"github.com/makinori/mikogo/env"
	"github.com/makinori/mikogo/irc"
)

func handleGeneralInfo(c *irc.Client, sender, where string, args []string) {
	out := "hi im mikogo (commit=" + env.GIT_COMMIT +
		" go=" + env.GetGoVersion() + ")\n"
	out += "made by: https://maki.cafe\n"
	out += "named by: https://micae.la\n"
	out += "https://github.com/makinori/mikogo\n"
	c.Send(sender, where, strings.TrimSpace(out))
}

var CommandGeneralInfo = Command{
	Name:        "info",
	Category:    "general",
	Description: "about me",
	Handle:      handleGeneralInfo,
}
