package command

import (
	"fmt"
	"log/slog"

	"github.com/makinori/mikogo/irc"
)

func testMsgsize(c *irc.Client, sender, where string) {
	c.Send(where,
		"will send a few long messages and print byte length",
	)

	overhead := len(c.MakePrivmsg(where, ""))

	sendOfNBytes := func(size int) {
		paddingBytes := make([]byte, size-overhead)
		for i := range paddingBytes {
			paddingBytes[i] = '.'
		}
		info := fmt.Sprintf(
			"total:%d text:%d overhead:%d ", size, size-overhead, overhead,
		)
		text := info + string(paddingBytes[len(info):])
		msg := c.MakePrivmsg(where, text)
		if len(msg) == size {
			fmt.Fprint(c.Conn, msg)
		} else {
			c.Send(where, "failed to make message. should not happen")
		}
	}

	sendOfNBytes(200)
	sendOfNBytes(300)
	sendOfNBytes(400)
	sendOfNBytes(500)
	sendOfNBytes(512)

	c.Send(where, "hopefully the 512 one came through\n"+
		"will now send a few of 513 bytes and higher",
	)

	sendOfNBytes(513)
	sendOfNBytes(520)
	sendOfNBytes(530)
}

func handleTest(msg *irc.Message, args []string) {
	// use TODO: cmdmenu

	if len(args) < 2 {
		msg.Client.Send(msg.Where, "usage: test <subcommand>\n"+
			"  ping, msgsize, clientpanic, commandpanic",
		)
		return
	}

	switch args[1] {
	case "ping":
		msg.Client.Send(msg.Where, "pong!")

	case "msgsize":
		testMsgsize(msg.Client, msg.Sender, msg.Where)

	case "clientpanic":
		msg.Client.PanicOnNextPing = true
		msg.Client.Send(msg.Where, "will client panic on next ping")
		slog.Info("admin requested test panic")

	case "commandpanic":
		panic("test panic")

	default:
		msg.Client.Send(msg.Where, "unknown subcommand")
	}
}

var CommandAdminTest = Command{
	Name:        "test",
	Category:    "admin",
	Description: "various test functions",
	Handle:      handleTest,
}
