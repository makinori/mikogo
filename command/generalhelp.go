package command

import (
	"strings"

	"github.com/elliotchance/orderedmap/v3"
	"github.com/makinori/mikogo/env"
	"github.com/makinori/mikogo/irc"
)

func handleGeneralHelp(msg *irc.Message, args []string) {
	categories := orderedmap.NewOrderedMap[string, []*Command]()

	for i := range commands {
		command := commands[i]
		if !canSenderRunCommand(msg, command) {
			continue
		}
		category, _ := categories.Get(command.Category)
		category = append(category, command)
		categories.Set(command.Category, category)
	}

	out := ""
	if msg.Sender == env.OWNER {
		out = "hi " + msg.Sender + " <3\n"
	}

	for name, commands := range categories.AllFromBack() {
		out += name + ":\n"
		for i := range commands {
			command := commands[i]
			out += "  " + command.Name + ": " + command.Description + "\n"
		}
	}

	msg.Client.Send(msg.Where, strings.TrimSpace(out))
}

var CommandGeneralHelp = Command{
	Name:        "help",
	Category:    "general",
	Description: "show all commands",
	Handle:      handleGeneralHelp,
}
