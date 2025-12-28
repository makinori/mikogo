package command

import (
	"strings"

	"github.com/makinori/mikogo/cmdmenu"
	"github.com/makinori/mikogo/db"
	"github.com/makinori/mikogo/env"
	"github.com/makinori/mikogo/irc"
)

func handleAdminServer(c *irc.Client, sender, where string, args []string) {
	(&cmdmenu.Menu{
		Name: "server",
		PrintUsage: func(usage string) {
			if strings.HasPrefix(where, "#") {
				usage = Prefix + usage
			}
			c.Send(where, "usage: "+usage)
		},
		Commands: []cmdmenu.Runnable{
			&(cmdmenu.Command{
				Name: "list",
				Handle: func(args []string) {
					servers, err := db.Servers.GetAll()
					if err != nil {
						c.Send(where, "failed to get all: "+err.Error())
						return
					}

					out := "[home]: " + env.HOME_SERVER + "\n"
					for s := servers.Front(); s != nil; s = s.Next() {
						out += s.Key + ": " + s.Value.Address + "\n"
					}
					c.Send(where, strings.TrimSpace(out))
				},
			}),
			&(cmdmenu.Command{
				Name:  "add",
				Args:  2,
				Usage: "<name> <address>",
				Handle: func(args []string) {
					err := db.Servers.Add(
						args[0], db.Server{Address: args[1]},
					)
					if err != nil {
						c.Send(where, "failed to add: "+err.Error())
						return
					}

					c.Send(where, "server added! will connect")
					// TODO: irc.Sync()
				},
			}),
			&(cmdmenu.Command{
				Name:  "del",
				Args:  1,
				Usage: "<name>",
				Handle: func(args []string) {
					err := db.Servers.Delete(args[0])
					if err != nil {
						c.Send(where, "failed to delete: "+err.Error())
						return
					}

					c.Send(where, "server deleted! will disconnect")
					// TODO: irc.Sync()
				},
			}),
			&(cmdmenu.Menu{
				Name: "set",
				Commands: []cmdmenu.Runnable{
					&(cmdmenu.Command{
						Name:  "addr",
						Args:  2,
						Usage: "<name> <address>",
						Handle: func(args []string) {
							server, err := db.Servers.Get(args[0])
							if err != nil {
								c.Send(where, "failed to get: "+err.Error())
								return
							}

							server.Address = args[1]

							err = db.Servers.Put(args[0], server)
							if err != nil {
								c.Send(where, "failed to update: "+err.Error())
								return
							}

							c.Send(where, "server address updated! will reconnect")
							// TODO: irc.Sync()
						},
					}),
				},
			}),
		},
	}).Run(args[1:])
}

var CommandAdminServer = Command{
	Name:        "server",
	Category:    "admin",
	Description: "manage servers",
	Handle:      handleAdminServer,
}
