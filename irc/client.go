package irc

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/makinori/mikogo/env"
)

type ConnState = uint8

const (
	ConnStateConnecting ConnState = iota
	ConnStateConnected
	ConnStateDisconnected
)

var (
	PRIVMSG_REGEXP     = regexp.MustCompile(`^:(.+?)!.+? PRIVMSG (.+?) :(.+?)\r\n$`)
	WHOIS_REPLY_REGEXP = regexp.MustCompile(`^:.+? 311 .+ (.+?) (.+?) (.+?) \* (.+?)\r\n$`)
)

type Client struct {
	address string
	active  bool

	Conn  *tls.Conn
	state ConnState

	user string
	host string
}

func (c *Client) Close() {
	c.active = false
	c.Conn.Close()
}

func (c *Client) MakePrivmsg(to string, msg string) string {
	return fmt.Sprintf(":%s!%s@%s PRIVMSG %s :%s\r\n",
		env.NICK, c.user, c.host, to, msg,
	)
}

func (c *Client) writeBatch(to string, lines []string) {
	id := fmt.Sprintf("%03d", rand.Intn(1000))
	fmt.Fprintf(c.Conn, "BATCH +%s draft/multiline %s\r\n", id, to)
	for i := range lines {
		fmt.Fprintf(c.Conn, "@batch=%s %s", id, c.MakePrivmsg(to, lines[i]))
	}
	fmt.Fprintf(c.Conn, "BATCH -%s\r\n", id)
}

// bool true if channel
func GetRecipient(sender, where string) string {
	if strings.HasPrefix(where, "#") {
		return where
	} else {
		return sender
	}
}

func (c *Client) Send(sender, where, msg string) {
	to := GetRecipient(sender, where)

	// TODO: handle messages over 512 bytes
	// SplitStringBySpace function available in old branch

	lines := strings.Split(msg, "\n")
	if len(lines) == 1 {
		fmt.Fprint(c.Conn, c.MakePrivmsg(to, msg))
		return
	}

	c.writeBatch(to, lines)
}

type HandleMessageFunc func(sender, where, msg string)

func (c *Client) handleMessage(msg string, botHandleMessage HandleMessageFunc) {
	// debugMsg := msg
	// debugMsg = strings.ReplaceAll(debugMsg, "\r", "\\r")
	// debugMsg = strings.ReplaceAll(debugMsg, "\n", "\\n")
	// fmt.Println(debugMsg)

	// handle privmsg first cause we dont want anyone to attack the below
	matches := PRIVMSG_REGEXP.FindStringSubmatch(msg)
	if len(matches) > 0 {
		go botHandleMessage(matches[1], matches[2], matches[3])
		return
	}

	if c.state == ConnStateConnecting &&
		strings.Contains(msg, " 001 "+env.NICK+" ") {
		slog.Info("connected to", "server", c.address)
		c.state = ConnStateConnected
		return
	}

	// response to self whois
	// TODO: what if server changes our mask?

	if c.user == "" && c.host == "" && strings.Contains(
		msg, " 311 "+env.NICK+" "+env.NICK,
	) {
		matches := WHOIS_REPLY_REGEXP.FindStringSubmatch(msg)
		if len(matches) == 0 {
			return
		}

		if matches[1] != env.NICK {
			return
		}

		c.user = matches[2]
		c.host = matches[3]
		slog.Info("got", "mask", c.user+"@"+c.host, "server", c.address)

		return
	}
}

func (c *Client) loop(botHandleMessage HandleMessageFunc) {
	reader := bufio.NewReader(c.Conn)
	for {
		msg, err := reader.ReadString('\n')
		if err == io.EOF {
			c.state = ConnStateDisconnected
			slog.Error("server closed connection")
			break
		} else if err != nil {
			slog.Error("failed to read message", "err", err)
			continue
		}
		c.handleMessage(msg, botHandleMessage)
	}
}

func Init(
	address string, handleMessage HandleMessageFunc,
) (*Client, error) {
	c := Client{
		address: address,
		active:  true,
	}

	var err error
	c.Conn, err = tls.Dial("tcp", address, &tls.Config{
		InsecureSkipVerify: true,
	})

	// TODO: should keep trying as clients should auto reconnect forever

	if err != nil {
		return nil, err
	}

	tcpConn, _ := c.Conn.NetConn().(*net.TCPConn)
	tcpConn.SetKeepAlive(true)

	// TODO: what if nick not available

	fmt.Fprintf(c.Conn, "NICK %s\r\n", env.NICK)
	fmt.Fprintf(c.Conn, "USER %s 0 * :%s\r\n",
		env.NICK, env.NICK,
	)

	// bot mode b or B
	fmt.Fprintf(c.Conn, "MODE %s +b\r\n", env.NICK)
	fmt.Fprintf(c.Conn, "MODE %s +B\r\n", env.NICK)

	// self whois for privmsg prefix
	fmt.Fprintf(c.Conn, "WHOIS %s\r\n", env.NICK)

	go func() {
		for {
			if !c.active {
				// user closed
				return
			}
			if c.state == ConnStateDisconnected {
				time.Sleep(time.Second * 5)
				continue
			}
			// TODO: is this the correct way to do it?
			fmt.Fprintf(c.Conn, "PING a\r\n")
			time.Sleep(time.Second * 60)
		}
	}()

	go c.loop(handleMessage)

	return &c, nil
}
