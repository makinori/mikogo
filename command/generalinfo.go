package command

import (
	"strings"

	"github.com/makinori/mikogo/env"
	"github.com/makinori/mikogo/irc"
)

func handleGeneralInfo(msg *irc.Message, args []string) {
	out := "hi im mikogo (commit=" + env.GIT_COMMIT +
		" go=" + env.GetGoVersion() + ")\n"
	out += "made by: https://maki.cafe\n"
	out += "named by: https://micae.la\n"
	out += "https://github.com/makinori/mikogo\n"
	msg.Client.Send(msg.Where, strings.TrimSpace(out))
}

var CommandGeneralInfo = Command{
	Name:        "info",
	Category:    "general",
	Description: "about me",
	Handle:      handleGeneralInfo,
}
