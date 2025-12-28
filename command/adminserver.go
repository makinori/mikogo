package command

import (
	"strings"

	"github.com/makinori/mikogo/cmdmenu"
	"github.com/makinori/mikogo/db"
	"github.com/makinori/mikogo/env"
	"github.com/makinori/mikogo/irc"
)

func adminServerList(msg *irc.Message, args []string) {
	servers, err := db.Servers.GetAll()
	if err != nil {
		msg.Client.Send(msg.Where, "failed to get all: "+err.Error())
		return
	}

	out := "[home]: " + env.HOME_SERVER + "\n"
	for s := servers.Front(); s != nil; s = s.Next() {
		out += s.Key + ": " + s.Value.Address + "\n"
	}
	msg.Client.Send(msg.Where, strings.TrimSpace(out))
}

func adminServerAdd(msg *irc.Message, args []string) {
	err := db.Servers.Add(
		args[0], db.Server{Address: args[1]},
	)
	if err != nil {
		msg.Client.Send(msg.Where, "failed to add: "+err.Error())
		return
	}

	msg.Client.Send(msg.Where, "server added! will connect")
	// TODO: irc.Sync()
}

func adminServerDel(msg *irc.Message, args []string) {
	err := db.Servers.Delete(args[0])
	if err != nil {
		msg.Client.Send(msg.Where, "failed to delete: "+err.Error())
		return
	}

	msg.Client.Send(msg.Where, "server deleted! will disconnect")
	// TODO: irc.Sync()
}

func adminServerSetAddr(c *irc.Message, args []string) {
	server, err := db.Servers.Get(args[0])
	if err != nil {
		c.Client.Send(c.Where, "failed to get: "+err.Error())
		return
	}

	server.Address = args[1]

	err = db.Servers.Put(args[0], server)
	if err != nil {
		c.Client.Send(c.Where, "failed to update: "+err.Error())
		return
	}

	c.Client.Send(c.Where, "server address updated! will reconnect")
	// TODO: irc.Sync()
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
			Name:   "del",
			Args:   1,
			Usage:  "<name>",
			Handle: adminServerDel,
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
	adminServer.Run(args[1:], msg, func(usage string) {
		if strings.HasPrefix(msg.Where, "#") {
			usage = prefix + usage
		}
		msg.Client.Send(msg.Where, "usage: "+usage)
	})
}

var CommandAdminServer = Command{
	Name:        "server",
	Category:    "admin",
	Description: "manage servers",
	Handle:      handleAdminServer,
}
