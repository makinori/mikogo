package irc

import (
	"bufio"
	"crypto/tls"
	"errors"
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

	RECONNECT_DURATION = time.Second * 10
)

var (
	PRIVMSG_REGEXP     = regexp.MustCompile(`^:(.+?)!.+? PRIVMSG (.+?) :(.+?)\r\n$`)
	WHOIS_REPLY_REGEXP = regexp.MustCompile(`^:.+? 311 .+ (.+?) (.+?) (.+?) \* (.+?)\r\n$`)
)

type Message struct {
	Client  *Client
	Sender  string
	Where   string
	Message string
}

var GlobalHandleMessage func(msg *Message)

type Client struct {
	Address string
	active  bool // for starting/stopping the client

	Conn  *tls.Conn
	state ConnState

	user string
	host string

	PanicOnNextPing bool
}

func (c *Client) MakePrivmsg(to string, msg string) string {
	out := fmt.Sprintf(":%s!%s@%s PRIVMSG %s :%s\r\n",
		env.NICK, c.user, c.host, to, msg,
	)
	if len(out) > 512 {
		slog.Warn("sent message too large", "bytes", len(out))
	}
	return out
}

func (c *Client) writeBatch(to string, lines []string) {
	id := fmt.Sprintf("%03d", rand.Intn(1000))
	fmt.Fprintf(c.Conn, "BATCH +%s draft/multiline %s\r\n", id, to)
	for i := range lines {
		fmt.Fprintf(c.Conn, "@batch=%s %s", id, c.MakePrivmsg(to, lines[i]))
	}
	fmt.Fprintf(c.Conn, "BATCH -%s\r\n", id)
}

func (c *Client) Send(to, msg string) {
	// TODO: handle messages over 512 bytes
	// SplitStringBySpace function available in old branch

	lines := strings.Split(msg, "\n")
	if len(lines) == 1 {
		fmt.Fprint(c.Conn, c.MakePrivmsg(to, msg))
		return
	}

	c.writeBatch(to, lines)
}

func (c *Client) handleMessage(msg string) {
	// debugMsg := msg
	// debugMsg = strings.ReplaceAll(debugMsg, "\r", "\\r")
	// debugMsg = strings.ReplaceAll(debugMsg, "\n", "\\n")
	// fmt.Println(debugMsg)

	// handle privmsg first cause we dont want anyone to attack the below
	matches := PRIVMSG_REGEXP.FindStringSubmatch(msg)
	if len(matches) > 0 {
		sender := matches[1]
		where := matches[2]
		if !strings.HasPrefix(where, "#") {
			// if direct message, "where" ends up being our nick
			where = sender
		}
		go GlobalHandleMessage(&Message{
			Client:  c,
			Sender:  sender,
			Where:   where,
			Message: matches[3],
		})
		return
	}

	if c.state == ConnStateConnecting &&
		strings.Contains(msg, " 001 "+env.NICK+" ") {
		slog.Info("connected!", "server", c.Address)
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
		slog.Info("got", "mask", c.user+"@"+c.host, "server", c.Address)

		return
	}
}

func (c *Client) connect() {
	if c.Conn != nil {
		c.Conn.Close()
		c.Conn = nil
	}

	c.state = ConnStateConnecting

	var err error
	c.Conn, err = tls.Dial("tcp", c.Address, &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		slog.Warn(
			"failed to connect. retrying...",
			"server", c.Address, "err", err,
		)
		return
	}

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

	reader := bufio.NewReader(c.Conn)
	for {
		msg, err := reader.ReadString('\n')
		if err == io.EOF || errors.Is(err, net.ErrClosed) {
			c.state = ConnStateDisconnected
			c.Conn = nil
			if c.active {
				slog.Warn("disconnected. retrying...", "server", c.Address)
			} else {
				slog.Info("disconnected by request", "server", c.Address)
			}
			break
		} else if err != nil {
			slog.Error(
				"failed to read message",
				"server", c.Address, "err", err,
			)
			// can just move on i suppose?
			continue
		}
		c.handleMessage(msg)
	}
}

// will set active to false
func (c *Client) delete() {
	c.active = false
	if c.Conn != nil {
		c.Conn.Close()
		c.Conn = nil
	}
}

func (c *Client) reconnect() {
	if !c.active {
		c.init()
		return
	}

	// will cause connection loop to reconnect
	if c.Conn != nil {
		c.Conn.Close()
		c.Conn = nil
	}
}

func (c *Client) recoverAndRestart() {
	r := recover()
	if r == nil {
		return
	}
	slog.Error("client panic", "err", r)
	c.reconnect()
}

func (c *Client) loop() {
	defer c.recoverAndRestart()
	for {
		if !c.active {
			return
		}
		c.connect() // will return if client disconnects
		if !c.active {
			return
		}
		time.Sleep(RECONNECT_DURATION)
	}
}

func (c *Client) ping() {
	if !c.active || c.state != ConnStateConnected {
		return
	}

	defer c.recoverAndRestart()
	if c.PanicOnNextPing {
		panic("test panic")
	}

	fmt.Fprintf(c.Conn, "PING hi\r\n")
}

func (c *Client) init() bool {
	if c.active {
		slog.Warn(
			"can't init client that's already active",
			"server", c.Address,
		)
		return false
	}

	c.active = true

	slog.Info("connecting...", "server", c.Address)

	go c.loop()

	return true
}

func newClient(address string) *Client {
	return &Client{
		Address: address,
	}
}
