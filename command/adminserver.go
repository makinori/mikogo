package command

import (
	"strings"

	"github.com/makinori/mikogo/db"
	"github.com/makinori/mikogo/env"
	"github.com/makinori/mikogo/irc"
)

func handleAdminServer(c *irc.Client, sender, where string, args []string) {
	if len(args) < 2 {
		c.Send(where, "either: list, add, del, setaddr")
		return
	}

	switch args[1] {
	case "list":
		servers, err := db.Servers.GetAll()
		if err != nil {
			c.Send(where, "failed to get all: "+err.Error())
			return
		}

		out := "[home]: " + env.HOME_SERVER + "\n"
		for server := servers.Front(); server != nil; server = server.Next() {
			out += server.Key + ": " + server.Value.Address + "\n"
		}
		c.Send(where, strings.TrimSpace(out))

	case "add":
		if len(args) < 4 {
			c.Send(where, "usage: <name> <address>")
			return
		}

		err := db.Servers.Add(args[2], db.Server{Address: args[3]})
		if err != nil {
			c.Send(where, "failed to add: "+err.Error())
			return
		}

		c.Send(where, "server added! will connect")
		// TODO: irc.Sync()

	case "del":
		if len(args) < 3 {
			c.Send(where, "usage: <name>")
			return
		}

		err := db.Servers.Delete(args[2])
		if err != nil {
			c.Send(where, "failed to delete: "+err.Error())
			return
		}

		c.Send(where, "server deleted! will disconnect")
		// TODO: irc.Sync()

	case "setaddr":
		if len(args) < 4 {
			c.Send(where, "usage: <name> <address>")
			return
		}

		server, err := db.Servers.Get(args[2])
		if err != nil {
			c.Send(where, "failed to get: "+err.Error())
			return
		}

		server.Address = args[3]

		err = db.Servers.Put(args[2], server)
		if err != nil {
			c.Send(where, "failed to update: "+err.Error())
			return
		}

		c.Send(where, "server address updated! will reconnect")
		// TODO: irc.Sync()
	}
}

var CommandAdminServer = Command{
	Name:        "server",
	Category:    "admin",
	Description: "manage servers",
	Handle:      handleAdminServer,
}
