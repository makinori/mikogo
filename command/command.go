package command

import (
	"slices"
	"strings"

	"github.com/makinori/mikogo/env"
	"github.com/makinori/mikogo/irc"
)

const (
	Prefix = "m/"
)

type Command struct {
	Name        string
	Category    string
	Description string
	Handle      func(c *irc.Client, sender, where string, args []string)
}

var (
	commands = []*Command{}

	ownerOnlyCategories = []string{
		"admin",
		"testing",
	}
)

func init() {
	commands = append(commands,
		&CommandGeneralHelp,
		&CommandGeneralInfo,

		&CommandAdminChan,

		&CommandTestingPing,
		&CommandTestingMsgsize,
	)
}

func getUnknownCommandMsg(where string) string {
	if strings.HasPrefix(where, "#") {
		return "unknown command. type " + Prefix + "help"
	} else {
		return "unknown command. type help"
	}
}

func canSenderRunCommand(sender string, command *Command) bool {
	if sender == env.OWNER {
		return true
	}
	// if command.OwnerOnly ||
	return !slices.Contains(ownerOnlyCategories, command.Category)
}

// args[0] should be prefix stripped
func Run(c *irc.Client, sender, where string, args []string) {
	if len(args) == 0 {
		c.Send(sender, where, getUnknownCommandMsg(where))
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
		c.Send(sender, where, getUnknownCommandMsg(where))
		return
	}

	if canSenderRunCommand(sender, commands[foundCommand]) {
		commands[foundCommand].Handle(c, sender, where, args)
	} else {
		c.Send(sender, where, "sorry you can't run that command :(")
		c.Send(env.OWNER, env.OWNER,
			sender+" tried to run "+commands[foundCommand].Name,
		)
	}
}
