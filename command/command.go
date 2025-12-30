package command

import (
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"strings"

	"github.com/makinori/mikogo/env"
	"github.com/makinori/mikogo/irc"
)

const (
	prefix = "m/"
)

type Command struct {
	Name        string
	Category    string
	Description string
	Handle      func(msg *irc.Message, args []string)
}

var (
	commands = []*Command{}

	ownerOnlyCategories = []string{
		"admin",
		"testing",
	}

	whiteSpaceRegexp = regexp.MustCompile(`\s+`)
)

func init() {
	commands = append(commands,
		&CommandGeneralHelp,
		&CommandGeneralInfo,

		&CommandAdminServer,
		&CommandAdminTest,
	)
}

func sendUnknownCommand(msg *irc.Message) {
	if strings.HasPrefix(msg.Where, "#") {
		msg.Client.Send(msg.Where, "unknown command. type "+prefix+"help")
	} else {
		msg.Client.Send(msg.Where, "unknown command. type help")
	}
}

func canSenderRunCommand(msg *irc.Message, command *Command) bool {
	// only allow on home server incase there's a malicious server
	if msg.Client.Address == env.HOME_SERVER && msg.Sender == env.OWNER {
		return true
	}

	return !slices.Contains(ownerOnlyCategories, command.Category)
}

func Run(msg *irc.Message) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		msg.Client.Send(msg.Where, fmt.Sprintf("command panicked: %v", r))
		slog.Warn("command panicked", "err", r)
	}()

	if strings.HasPrefix(msg.Where, "#") &&
		!strings.HasPrefix(msg.Message, prefix) {
		return
	}

	args := whiteSpaceRegexp.Split(strings.TrimSpace(msg.Message), -1)
	args[0] = strings.TrimPrefix(args[0], prefix)

	if len(args) == 0 {
		sendUnknownCommand(msg)
		return
	}

	name := strings.ToLower(args[0])

	foundCommand := -1
	for i := range commands {
		if commands[i].Name == name {
			foundCommand = i
			break
		}
	}
	if foundCommand == -1 {
		sendUnknownCommand(msg)
		return
	}

	if canSenderRunCommand(msg, commands[foundCommand]) {
		commands[foundCommand].Handle(msg, args)
	} else {
		msg.Client.Send(msg.Where, "sorry you can't run that command :(")
		irc.ReportIncident(fmt.Sprintf(
			`"%s" tried to run "%s" on "%s"`,
			msg.Sender, msg.Message, msg.Client.Address,
		))
	}
}
