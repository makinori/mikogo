package command

import (
	"strings"

	"github.com/elliotchance/orderedmap/v3"
	"github.com/makinori/mikogo/env"
	"github.com/makinori/mikogo/irc"
)

func handleGeneralHelp(c *irc.Client, sender, where string, args []string) {
	categories := orderedmap.NewOrderedMap[string, []*Command]()

	for i := range commands {
		command := commands[i]
		if !canSenderRunCommand(c, sender, command) {
			continue
		}
		category, _ := categories.Get(command.Category)
		category = append(category, command)
		categories.Set(command.Category, category)
	}

	out := ""
	if sender == env.OWNER {
		out = "hi " + sender + " <3\n"
	}

	for cat := categories.Front(); cat != nil; cat = cat.Next() {
		out += cat.Key + ":\n"
		for i := range cat.Value {
			command := cat.Value[i]
			out += "  " + command.Name + ": " + command.Description + "\n"
		}
	}

	c.Send(where, strings.TrimSpace(out))
}

var CommandGeneralHelp = Command{
	Name:        "help",
	Category:    "general",
	Description: "show all commands",
	Handle:      handleGeneralHelp,
}
