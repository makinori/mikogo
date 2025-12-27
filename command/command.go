package command

import (
	"strings"

	"github.com/makinori/mikogo/env"
	"github.com/makinori/mikogo/irc"
)

const (
	Prefix            = "m/"
	UnknownCommandMsg = "unknown command. use " + Prefix + "help"
)

type Command struct {
	Name        string
	Category    string
	Description string
	OwnerOnly   bool
	Handle      func(c *irc.Client, sender, where string, args []string)
}

var commands = []*Command{}

func init() {
	commands = append(commands,
		&CommandGeneralHelp,
		&CommandGeneralInfo,
	)

	if env.DEV {
		commands = append(commands,
			&CommandTestingPing,
			&CommandTestingMsgsize,
		)
	}
}

// args[0] should be prefix stripped
func Run(c *irc.Client, sender, where string, args []string) {
	if len(args) == 0 {
		c.Send(sender, where, UnknownCommandMsg)
		return
	}

	// TODO: handle recover

	name := strings.ToLower(args[0])

	foundCommand := -1
	for i := range commands {
		if commands[i].Name == name {
			foundCommand = i
			break
		}
	}
	if foundCommand == -1 {
		c.Send(sender, where, UnknownCommandMsg)
		return
	}

	commands[foundCommand].Handle(c, sender, where, args)
}
