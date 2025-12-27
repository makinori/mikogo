package command

import (
	"fmt"

	"github.com/makinori/mikogo/irc"
)

func testPing(c *irc.Client, sender, where string) {
	c.Send(where, "pong!")
}

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

func handleTest(c *irc.Client, sender, where string, args []string) {
	if len(args) < 2 {
		c.Send(where, "either: ping, msgsize")
		return
	}

	switch args[1] {
	case "ping":
		testPing(c, sender, where)
	case "msgsize":
		testMsgsize(c, sender, where)
	}
}

var CommandAdminTest = Command{
	Name:        "test",
	Category:    "admin",
	Description: "various test functions",
	Handle:      handleTest,
}
