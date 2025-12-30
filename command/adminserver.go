package command

import (
	"slices"
	"strings"

	"github.com/makinori/mikogo/cmdmenu"
	"github.com/makinori/mikogo/db"
	"github.com/makinori/mikogo/irc"
)

func adminServerList(msg *irc.Message, args []string) {
	servers, err := db.Servers.GetAll()
	if err != nil {
		msg.Client.Send(msg.Where, "failed to get all: "+err.Error())
		return
	}

	out := ""
	for name, server := range servers.AllFromBack() {
		out += name + " (" + server.Address + "): "
		if len(server.Channels) == 0 {
			out += "none\n"
		} else {
			out += strings.Join(server.Channels, ", ") + "\n"
		}
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

func adminServerChannelsAdd(msg *irc.Message, args []string) {
	channel := args[1]
	if !strings.HasPrefix(channel, "#") {
		channel = "#" + channel
	}

	server, err, _ := db.Servers.Get(args[0])
	if err != nil {
		msg.Client.Send(msg.Where, "failed to get: "+err.Error())
		return
	}

	if slices.Contains(server.Channels, channel) {
		msg.Client.Send(msg.Where, "already in channel")
		return
	}

	server.Channels = append(server.Channels, channel)

	err = db.Servers.Put(args[0], server)
	if err != nil {
		msg.Client.Send(msg.Where, "failed to put: "+err.Error())
		return
	}

	msg.Client.Send(msg.Where, "added channel! will join")

	irc.Sync()
}

func adminServerChannelsRemove(msg *irc.Message, args []string) {
	channel := args[1]
	if !strings.HasPrefix(channel, "#") {
		channel = "#" + channel
	}

	server, err, _ := db.Servers.Get(args[0])
	if err != nil {
		msg.Client.Send(msg.Where, "failed to get: "+err.Error())
		return
	}

	i := slices.Index(server.Channels, channel)
	if i == -1 {
		msg.Client.Send(msg.Where, "not in channel")
		return
	}

	server.Channels = slices.Delete(server.Channels, i, i+1)

	err = db.Servers.Put(args[0], server)
	if err != nil {
		msg.Client.Send(msg.Where, "failed to put: "+err.Error())
		return
	}

	msg.Client.Send(msg.Where, "removed channel! will leave")

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
		&cmdmenu.Menu[irc.Message]{
			Name: "channels",
			Commands: []cmdmenu.Runnable[irc.Message]{
				&cmdmenu.Command[irc.Message]{
					Name:   "add",
					Args:   2,
					Usage:  "<server name> <channel name>",
					Handle: adminServerChannelsAdd,
				},
				&cmdmenu.Command[irc.Message]{
					Name:   "remove",
					Args:   2,
					Usage:  "<server name> <channel name>",
					Handle: adminServerChannelsRemove,
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
