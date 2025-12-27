package command

import (
	"fmt"

	"github.com/makinori/mikogo/irc"
)

func handleMsgsize(c *irc.Client, sender, where string, args []string) {
	c.Send(sender, where,
		"will send a few long messages and print byte length",
	)

	to := irc.GetRecipient(sender, where)
	overhead := len(c.MakePrivmsg(to, ""))

	sendOfNBytes := func(size int) {
		paddingBytes := make([]byte, size-overhead)
		for i := range paddingBytes {
			paddingBytes[i] = '.'
		}
		info := fmt.Sprintf(
			"text:%d overhead:%d total:%d ", size-overhead, overhead, size,
		)
		text := info + string(paddingBytes[len(info):])
		msg := c.MakePrivmsg(to, text)
		if len(msg) == size {
			fmt.Fprint(c.Conn, msg)
		} else {
			c.Send(sender, where, "failed to make message. should not happen")
		}
	}

	sendOfNBytes(200)
	sendOfNBytes(300)
	sendOfNBytes(400)
	sendOfNBytes(500)
	sendOfNBytes(512)
	c.Send(sender, where, "hopefully the 512 one came through\n"+
		"will now send a few of 513 bytes and higher",
	)
	sendOfNBytes(513)
	sendOfNBytes(520)
	sendOfNBytes(530)
}

var CommandTestingMsgsize = Command{
	Name:        "msgsize",
	Category:    "testing",
	Description: "test max message size",
	OwnerOnly:   false,
	Handle:      handleMsgsize,
}
