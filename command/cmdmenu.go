package command

import (
	"strings"

	"github.com/makinori/mikogo/irc"
)

func cmdmenuUsage(msg *irc.Message) func(usage string) {
	return func(usage string) {
		if strings.HasPrefix(msg.Where, "#") {
			usage = prefix + usage
		}
		msg.Client.Send(msg.Where, "usage: "+usage)
	}
}
