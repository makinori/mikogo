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
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/makinori/mikogo/env"
	"github.com/makinori/mikogo/ircf"
)

// TODO: better logging system so we dont keep writing "server", c.Address

type ConnState = uint8

const (
	ConnStateConnecting ConnState = iota
	ConnStateConnected
	ConnStateDisconnected

	RECONNECT_DURATION = time.Second * 10
)

var (
	RE_PRIVMSG     = regexp.MustCompile(`^:(.+?)!.+? PRIVMSG (.+?) :(.+?)\r\n$`)
	RE_KICK        = regexp.MustCompile(`^:(.+?)!.+? KICK (#.+?) .+? :?(.+?)\r\n$`)
	RE_WHOIS_REPLY = regexp.MustCompile(`^:.+? 311 .+ (.+?) (.+?) (.+?) \* (.+?)\r\n$`)
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

	channelsCurrent []string
	channelsTarget  []string
	channelsMutex   *sync.RWMutex
}

func (c *Client) slog() *slog.Logger {
	return slog.Default().With("server", c.Address)
}

func (c *Client) FormattedState() string {
	switch c.state {
	case ConnStateConnecting:
		return ircf.Bold().Color(98, 41).Format("connecting")
	case ConnStateConnected:
		return ircf.Bold().Color(98, 43).Format("connected")
	case ConnStateDisconnected:
		return ircf.Bold().Color(98, 40).Format("disconnected")
	}
	return ircf.Bold().Color(98, 40).Format(
		fmt.Sprintf("unknown: %v", c.state),
	)
}

func (c *Client) CurrentChannels() []string {
	c.channelsMutex.RLock()
	defer c.channelsMutex.RUnlock()
	return slices.Concat(c.channelsCurrent) // make copy
}

func (c *Client) setTargetChannels(target []string) {
	c.channelsMutex.Lock()
	defer c.channelsMutex.Unlock()
	c.channelsTarget = slices.Concat(target) // make copy
}

func (c *Client) SyncChannels() {
	if !c.active || c.state != ConnStateConnected {
		c.slog().Warn("can't sync channels if inactive or disconnected")
		return
	}

	c.channelsMutex.Lock()
	defer c.channelsMutex.Unlock()

	for _, target := range c.channelsTarget {
		if slices.Contains(c.channelsCurrent, target) {
			continue
		}

		if !strings.HasPrefix(target, "#") {
			c.slog().Warn("can't join invalid channel", "name", target)
			continue
		}

		fmt.Fprintf(c.Conn, "JOIN %s\r\n", target)
		// TODO: implementation doesnt handle JOIN fails
		c.channelsCurrent = append(c.channelsCurrent, target)
		c.slog().Info("channel joined", "name", target)
	}

	i := 0
	for _, current := range c.channelsCurrent {
		if slices.Contains(c.channelsTarget, current) {
			c.channelsCurrent[i] = current
			i++
			continue
		}

		fmt.Fprintf(c.Conn, "PART %s\r\n", current)
		c.slog().Info("channel left", "name", current)
	}
	c.channelsCurrent = c.channelsCurrent[:i]
}

func (c *Client) MakePrivmsg(to string, msg string) string {
	out := fmt.Sprintf(":%s!%s@%s PRIVMSG %s :%s\r\n",
		env.NICK, c.user, c.host, to, msg,
	)
	if len(out) > 512 {
		c.slog().Warn("sent message too large", "bytes", len(out))
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

func (c *Client) handleKick(sender string, where string, reason string) {
	c.channelsMutex.Lock()
	defer c.channelsMutex.Unlock()

	c.slog().Info("kicked", "sender", sender, "where", where, "reason", reason)

	ReportIncident(fmt.Sprintf(
		"kicked from %s by %s on %s for %s",
		ircf.BoldWhite.Format(where),
		ircf.BoldWhite.Format(sender),
		ircf.BoldWhite.Format(c.Address),
		ircf.BoldWhite.Format(reason),
	))

	i := slices.Index(c.channelsCurrent, where)
	if i == -1 {
		c.slog().Warn("was never in channel?", "where", where)
		return
	}

	c.channelsCurrent = slices.Delete(c.channelsCurrent, i, i+1)
}

func (c *Client) handleMessage(msg string) {
	// debugMsg := msg
	// debugMsg = strings.ReplaceAll(debugMsg, "\r", "\\r")
	// debugMsg = strings.ReplaceAll(debugMsg, "\n", "\\n")
	// fmt.Println(debugMsg)

	// handle privmsg first cause we dont want anyone to attack the below
	matches := RE_PRIVMSG.FindStringSubmatch(msg)
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
		c.slog().Info("connected!")
		c.state = ConnStateConnected
		c.SyncChannels()
		return
	}

	matches = RE_KICK.FindStringSubmatch(msg)
	if len(matches) > 0 {
		c.handleKick(matches[1], matches[2], matches[3])
		return
	}

	// response to self whois
	// TODO: what if server changes our mask?

	if c.user == "" && c.host == "" && strings.Contains(
		msg, " 311 "+env.NICK+" "+env.NICK,
	) {
		matches := RE_WHOIS_REPLY.FindStringSubmatch(msg)
		if len(matches) == 0 {
			return
		}

		if matches[1] != env.NICK {
			return
		}

		c.user = matches[2]
		c.host = matches[3]
		c.slog().Info("got mask", "mask", c.user+"@"+c.host)

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
		c.slog().Warn("failed to connect. retrying...", "err", err)
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
				c.slog().Warn("disconnected. retrying...")
			} else {
				c.slog().Info("disconnected by request")
			}
			break
		} else if err != nil {
			c.slog().Error("failed to read message", "err", err)
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
	c.slog().Error("client panic", "err", r)
	c.PanicOnNextPing = false
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
		c.slog().Warn("can't init client that's already active")
		return false
	}

	c.active = true

	c.slog().Info("connecting...")

	go c.loop()

	return true
}

func newClient(address string) *Client {
	return &Client{
		Address:       address,
		channelsMutex: &sync.RWMutex{},
	}
}
