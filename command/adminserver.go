package command

import (
	"fmt"
	"slices"
	"strings"

	"github.com/makinori/mikogo/cmdmenu"
	"github.com/makinori/mikogo/db"
	"github.com/makinori/mikogo/irc"
	"github.com/makinori/mikogo/ircf"
)

func adminServerList(msg *irc.Message, args []string) {
	servers, err := db.Servers.GetAll()
	if err != nil {
		msg.Client.Send(msg.Where, "failed to get all: "+err.Error())
		return
	}

	out := ""
	for name, server := range servers.AllFromBack() {
		client := irc.GetClient(name)
		currentChannels := client.CurrentChannels()

		formattedChannels := make([]string, len(server.Channels))
		for i, channel := range server.Channels {
			if slices.Contains(currentChannels, channel) {
				formattedChannels[i] = ircf.Color(98, 43).Format(channel)
			} else {
				formattedChannels[i] = ircf.Color(98, 40).Format(channel)
			}
		}

		out += fmt.Sprintf(
			"%s addr=%s state=%s\n  %s\n",
			ircf.BoldWhite.Format(name),
			ircf.BoldWhite.Format(server.Address),
			client.FormattedState(),
			ircf.Bold().Format(strings.Join(formattedChannels, ", ")),
		)
	}

	msg.Client.Send(msg.Where, strings.TrimSpace(out))
}

func adminServerAdd(msg *irc.Message, args []string) {
	if args[0] == "home" {
		msg.Client.Send(msg.Where, "cannot add home server")
		return
	}

	serverName, _, err, ok := db.GetServerByAddress(args[1])
	if err != nil {
		msg.Client.Send(msg.Where, "failed to get server by address: "+err.Error())
		return
	}
	if ok {
		msg.Client.Send(msg.Where, "server with same address already exists: "+serverName)
		return
	}

	err = db.Servers.Add(
		args[0], db.Server{Address: args[1]},
	)
	if err != nil {
		msg.Client.Send(msg.Where, "failed to add: "+err.Error())
		return
	}

	msg.Client.Send(msg.Where, "server added! will connect")

	irc.Sync()
}

func adminServerRemove(msg *irc.Message, args []string) {
	if args[0] == "home" {
		msg.Client.Send(msg.Where, "cannot remove home server")
		return
	}

	err := db.Servers.Delete(args[0])
	if err != nil {
		msg.Client.Send(msg.Where, "failed to remove: "+err.Error())
		return
	}

	msg.Client.Send(msg.Where, "server removed! will disconnect")

	irc.Sync()
}

func adminServerSetAddr(msg *irc.Message, args []string) {
	if args[0] == "home" {
		msg.Client.Send(msg.Where, "cannot update home server address")
		return
	}

	serverName, _, err, ok := db.GetServerByAddress(args[1])
	if err != nil {
		msg.Client.Send(msg.Where, "failed to get server by address: "+err.Error())
		return
	}
	if ok {
		msg.Client.Send(msg.Where, "server with same address already exists: "+serverName)
		return
	}

	server, err, _ := db.Servers.Get(args[0])
	if err != nil {
		msg.Client.Send(msg.Where, "failed to get: "+err.Error())
		return
	}

	server.Address = args[1]

	err = db.Servers.Put(args[0], server)
	if err != nil {
		msg.Client.Send(msg.Where, "failed to update: "+err.Error())
		return
	}

	msg.Client.Send(msg.Where, "server address updated! will reconnect")

	irc.Sync()
}

var adminServer = cmdmenu.Menu[irc.Message]{
	Name: "server",
	Commands: []cmdmenu.Runnable[irc.Message]{
		&cmdmenu.Command[irc.Message]{
			Name:   "list",
			Handle: adminServerList,
		},
		&cmdmenu.Command[irc.Message]{
			Name:   "add",
			Args:   2,
			Usage:  "<name> <address>",
			Handle: adminServerAdd,
		},
		&cmdmenu.Command[irc.Message]{
			Name:   "remove",
			Args:   1,
			Usage:  "<name>",
			Handle: adminServerRemove,
		},
		&cmdmenu.Menu[irc.Message]{
			Name: "set",
			Commands: []cmdmenu.Runnable[irc.Message]{
				&cmdmenu.Command[irc.Message]{
					Name:   "addr",
					Args:   2,
					Usage:  "<name> <address>",
					Handle: adminServerSetAddr,
				},
			},
		},
	},
}

func handleAdminServer(msg *irc.Message, args []string) {
	adminServer.Run(args[1:], msg, cmdmenuUsage(msg))
}

var CommandAdminServer = Command{
	Name:        "server",
	Category:    "admin",
	Description: "manage servers",
	Handle:      handleAdminServer,
}
