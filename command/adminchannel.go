package command

import (
	"slices"
	"strings"

	"github.com/makinori/mikogo/cmdmenu"
	"github.com/makinori/mikogo/db"
	"github.com/makinori/mikogo/irc"
)

func adminChannelAdd(msg *irc.Message, args []string) {
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

func adminChannelRemove(msg *irc.Message, args []string) {
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

func adminChannelSync(msg *irc.Message, args []string) {
	msg.Client.SyncChannels()
	msg.Client.Send(msg.Where, "will resync channels")
}

var adminChannel = cmdmenu.Menu[irc.Message]{
	Name: "channel",
	Commands: []cmdmenu.Runnable[irc.Message]{
		&cmdmenu.Command[irc.Message]{
			Name:   "list",
			Handle: adminServerList,
		},
		&cmdmenu.Command[irc.Message]{
			Name:   "add",
			Args:   2,
			Usage:  "<server name> <channel name>",
			Handle: adminChannelAdd,
		},
		&cmdmenu.Command[irc.Message]{
			Name:   "remove",
			Args:   2,
			Usage:  "<server name> <channel name>",
			Handle: adminChannelRemove,
		},
		&cmdmenu.Command[irc.Message]{
			Name:   "sync",
			Handle: adminChannelSync,
		},
	},
}

func handleAdminChannel(msg *irc.Message, args []string) {
	adminChannel.Run(args[1:], msg, cmdmenuUsage(msg))
}

var CommandAdminChannel = Command{
	Name:        "channel",
	Category:    "admin",
	Description: "manage channels",
	Handle:      handleAdminChannel,
}
