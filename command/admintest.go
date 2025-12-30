package command

import (
	"fmt"

	"github.com/makinori/mikogo/cmdmenu"
	"github.com/makinori/mikogo/irc"
)

func adminTestMsgsize(msg *irc.Message, args []string) {
	msg.Client.Send(msg.Where,
		"will send a few long messages and print byte length",
	)

	overhead := len(msg.Client.MakePrivmsg(msg.Where, ""))

	sendOfNBytes := func(size int) {
		paddingBytes := make([]byte, size-overhead)
		for i := range paddingBytes {
			paddingBytes[i] = '.'
		}
		info := fmt.Sprintf(
			"total:%d text:%d overhead:%d ", size, size-overhead, overhead,
		)
		text := info + string(paddingBytes[len(info):])
		out := msg.Client.MakePrivmsg(msg.Where, text)
		if len(out) == size {
			fmt.Fprint(msg.Client.Conn, out)
		} else {
			msg.Client.Send(msg.Where,
				"failed to make message. should not happen",
			)
		}
	}

	sendOfNBytes(200)
	sendOfNBytes(300)
	sendOfNBytes(400)
	sendOfNBytes(500)
	sendOfNBytes(512)

	msg.Client.Send(msg.Where, "hopefully the 512 one came through\n"+
		"will now send a few of 513 bytes and higher",
	)

	sendOfNBytes(513)
	sendOfNBytes(520)
	sendOfNBytes(530)
}

var adminTest = cmdmenu.Menu[irc.Message]{
	Name: "test",
	Commands: []cmdmenu.Runnable[irc.Message]{
		&cmdmenu.Command[irc.Message]{
			Name: "ping",
			Handle: func(msg *irc.Message, args []string) {
				msg.Client.Send(msg.Where, "pong!")
			},
		},
		&cmdmenu.Command[irc.Message]{
			Name:   "msgsize",
			Handle: adminTestMsgsize,
		},
		&cmdmenu.Command[irc.Message]{
			Name: "clientpanic",
			Handle: func(msg *irc.Message, args []string) {
				msg.Client.PanicOnNextPing = true
				msg.Client.Send(msg.Where, "will client panic on next ping")
			},
		},
		&cmdmenu.Command[irc.Message]{
			Name: "commandpanic",
			Handle: func(msg *irc.Message, args []string) {
				// should recover all the way back to command.go
				panic("test panic")
			},
		},
	},
}

func handleAdminTest(msg *irc.Message, args []string) {
	adminTest.Run(args[1:], msg, cmdmenuUsage(msg))
}

var CommandAdminTest = Command{
	Name:        "test",
	Category:    "admin",
	Description: "various test functions",
	Handle:      handleAdminTest,
}
