package irc

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	mathRand "math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"git.aiterp.net/gisle/irc/ircutil"
	"git.aiterp.net/gisle/irc/isupport"
	"git.aiterp.net/gisle/irc/list"
)

var supportedCaps = []string{
	"server-time",
	"cap-notify",
	"multi-prefix",
	"userhost-in-names",
	"account-notify",
	"away-notify",
	"extended-join",
	"chghost",
	"account-tag",
}

// ErrNoConnection is returned if you try to do something requiring a connection,
// but there is none.
var ErrNoConnection = errors.New("irc: no connection")

// ErrTargetAlreadyAdded is returned by Client.AddTarget if that target has already been
// added to the client.
var ErrTargetAlreadyAdded = errors.New("irc: target already added")

// ErrTargetConflict is returned by Clinet.AddTarget if there already exists a target
// matching the name and kind.
var ErrTargetConflict = errors.New("irc: target name and kind match existing target")

// ErrTargetNotFound is returned by Clinet.RemoveTarget if the target is not part of
// the client's target list
var ErrTargetNotFound = errors.New("irc: target not found")

// ErrTargetIsStatus is returned by Clinet.RemoveTarget if the target is the client's
// status target
var ErrTargetIsStatus = errors.New("irc: cannot remove status target")

// A Client is an IRC client. You need to use New to construct it
type Client struct {
	id     string
	config Config

	mutex  sync.RWMutex
	conn   net.Conn
	ctx    context.Context
	cancel context.CancelFunc

	events chan *Event
	sends  chan string

	lastSend time.Time

	capEnabled    map[string]bool
	capData       map[string]string
	capsRequested []string

	nick     string
	user     string
	host     string
	quit     bool
	isupport isupport.ISupport
	values   map[string]interface{}

	status    *Status
	targets   []Target
	targteIds map[Target]string
}

// New creates a new client. The context can be context.Background if you want manually to
// tear down clients upon quitting.
func New(ctx context.Context, config Config) *Client {
	client := &Client{
		id:         generateClientID(),
		values:     make(map[string]interface{}),
		events:     make(chan *Event, 64),
		sends:      make(chan string, 64),
		capEnabled: make(map[string]bool),
		capData:    make(map[string]string),
		config:     config.WithDefaults(),
		targteIds:  make(map[Target]string, 16),
		status:     &Status{},
	}

	client.AddTarget(client.status)

	client.ctx, client.cancel = context.WithCancel(ctx)

	go client.handleEventLoop()
	go client.handleSendLoop()

	return client
}

// Context gets the client's context. It's cancelled if the parent context used
// in New is, or Destroy is called.
func (client *Client) Context() context.Context {
	return client.ctx
}

// ID gets the unique identifier for the client, which could be used in data structures
func (client *Client) ID() string {
	client.mutex.RLock()
	defer client.mutex.RUnlock()

	return client.id
}

// Nick gets the nick of the client
func (client *Client) Nick() string {
	client.mutex.RLock()
	defer client.mutex.RUnlock()

	return client.nick
}

// User gets the user/ident of the client
func (client *Client) User() string {
	client.mutex.RLock()
	defer client.mutex.RUnlock()

	return client.user
}

// Host gets the hostname of the client
func (client *Client) Host() string {
	client.mutex.RLock()
	defer client.mutex.RUnlock()

	return client.host
}

// ISupport gets the client's ISupport. This is mutable, and changes to it
// *will* affect the client.
func (client *Client) ISupport() *isupport.ISupport {
	return &client.isupport
}

// Connect connects to the server by addr.
func (client *Client) Connect(addr string, ssl bool) (err error) {
	var conn net.Conn

	if client.Connected() {
		client.Disconnect()
	}

	client.isupport.Reset()

	client.mutex.Lock()
	client.quit = false
	client.mutex.Unlock()

	client.EmitSync(context.Background(), NewEvent("client", "connecting"))

	if ssl {
		conn, err = tls.Dial("tcp", addr, &tls.Config{
			InsecureSkipVerify: client.config.SkipSSLVerification,
		})
		if err != nil {
			return err
		}
	} else {
		conn, err = net.Dial("tcp", addr)
		if err != nil {
			return err
		}
	}

	client.Emit(NewEvent("client", "connect"))

	go func() {
		reader := bufio.NewReader(conn)
		replacer := strings.NewReplacer("\r", "", "\n", "")

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			line = replacer.Replace(line)

			event, err := ParsePacket(line)
			if err != nil {
				continue
			}

			client.Emit(event)
		}

		client.mutex.Lock()
		client.conn = nil
		client.mutex.Unlock()

		client.Emit(NewEvent("client", "disconnect"))
	}()

	client.mutex.Lock()
	client.conn = conn
	client.mutex.Unlock()

	return nil
}

// Disconnect disconnects from the server. It will either return the
// close error, or ErrNoConnection if there is no connection
func (client *Client) Disconnect() error {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	if client.conn == nil {
		return ErrNoConnection
	}

	client.quit = true

	err := client.conn.Close()

	return err
}

// Connected returns true if the client has a connection
func (client *Client) Connected() bool {
	client.mutex.RLock()
	defer client.mutex.RUnlock()

	return client.conn != nil
}

// Send sends a line to the server. A line-feed will be automatically added if one
// is not provided.
func (client *Client) Send(line string) error {
	client.mutex.RLock()
	conn := client.conn
	client.mutex.RUnlock()

	if conn == nil {
		return ErrNoConnection
	}

	if !strings.HasSuffix(line, "\n") {
		line += "\r\n"
	}

	_, err := conn.Write([]byte(line))
	if err != nil {
		client.EmitNonBlocking(NewErrorEvent("network", err.Error()))
		client.Disconnect()
	}

	return err
}

// Sendf is Send with a fmt.Sprintf
func (client *Client) Sendf(format string, a ...interface{}) error {
	return client.Send(fmt.Sprintf(format, a...))
}

// SendQueued appends a message to a queue that will only send 2 messages
// per second to avoid flooding. If the queue is ull, a goroutine will be
// spawned to queue it, so this function will always return immediately.
// Order may not be guaranteed, however, but if you're sending 64 messages
// at once that may not be your greatest concern.
//
// Failed sends will be discarded quietly to avoid a backup from being
// thrown on a new connection.
func (client *Client) SendQueued(line string) {
	select {
	case client.sends <- line:
	default:
		go func() { client.sends <- line }()
	}
}

// SendQueuedf is SendQueued with a fmt.Sprintf
func (client *Client) SendQueuedf(format string, a ...interface{}) {
	client.SendQueued(fmt.Sprintf(format, a...))
}

// Emit sends an event through the client's event, and it will return immediately
// unless the internal channel is filled up. The returned context can be used to
// wait for the event, or the client's destruction.
func (client *Client) Emit(event Event) context.Context {
	event.ctx, event.cancel = context.WithCancel(client.ctx)
	client.events <- &event

	return event.ctx
}

// EmitNonBlocking is just like emit, but it will spin off a goroutine if the channel is full.
// This lets it be called from other handlers without ever blocking. See Emit for what the
// returned context is for.
func (client *Client) EmitNonBlocking(event Event) context.Context {
	event.ctx, event.cancel = context.WithCancel(client.ctx)

	select {
	case client.events <- &event:
	default:
		go func() { client.events <- &event }()
	}

	return event.ctx
}

// EmitSync emits an event and waits for either its context to complete or the one
// passed to it (e.g. a request's context). It's a shorthand for Emit with its
// return value used in a `select` along with a passed context.
func (client *Client) EmitSync(ctx context.Context, event Event) (err error) {
	eventCtx := client.Emit(event)

	select {
	case <-eventCtx.Done():
		{
			if err := eventCtx.Err(); err != context.Canceled {
				return err
			}

			return nil
		}
	case <-ctx.Done():
		{
			return ctx.Err()
		}
	}
}

// Value gets a client value.
func (client *Client) Value(key string) (v interface{}, ok bool) {
	client.mutex.RLock()
	v, ok = client.values[key]
	client.mutex.RUnlock()

	return
}

// SetValue sets a client value.
func (client *Client) SetValue(key string, value interface{}) {
	client.mutex.Lock()
	client.values[key] = value
	client.mutex.Unlock()
}

// Destroy destroys the client, which will lead to a disconnect. Cancelling the
// parent context will do the same.
func (client *Client) Destroy() {
	client.Disconnect()
	client.cancel()
	close(client.sends)
	close(client.events)
}

// Destroyed returns true if the client has been destroyed, either by
// Destroy or the parent context.
func (client *Client) Destroyed() bool {
	select {
	case <-client.ctx.Done():
		return true
	default:
		return false
	}
}

// PrivmsgOverhead returns the overhead on a privmsg to the target. If `action` is true,
// it will also count the extra overhead of a CTCP ACTION.
func (client *Client) PrivmsgOverhead(targetName string, action bool) int {
	client.mutex.RLock()
	defer client.mutex.RUnlock()

	// Return a really safe estimate if user or host is missing.
	if client.user == "" || client.host == "" {
		return 200
	}

	return ircutil.MessageOverhead(client.nick, client.user, client.host, targetName, action)
}

// Join joins one or more channels without a key.
func (client *Client) Join(channels ...string) error {
	return client.Sendf("JOIN %s", strings.Join(channels, ","))
}

// Target gets a target by kind and name
func (client *Client) Target(kind string, name string) Target {
	client.mutex.RLock()
	defer client.mutex.RUnlock()

	for _, target := range client.targets {
		if target.Kind() == kind && target.Name() == name {
			return target
		}
	}

	return nil
}

// Channel is a shorthand for getting a channel target and type asserting it.
func (client *Client) Channel(name string) *Channel {
	target := client.Target("channel", name)
	if target == nil {
		return nil
	}

	return target.(*Channel)
}

// Query is a shorthand for getting a query target and type asserting it.
func (client *Client) Query(name string) *Query {
	target := client.Target("query", name)
	if target == nil {
		return nil
	}

	return target.(*Query)
}

// AddTarget adds a target to the client, generating a unique ID for it.
func (client *Client) AddTarget(target Target) (id string, err error) {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	for i := range client.targets {
		if target == client.targets[i] {
			err = ErrTargetAlreadyAdded
			return
		} else if target.Kind() == client.targets[i].Kind() && target.Name() == client.targets[i].Name() {
			err = ErrTargetConflict
			return
		}
	}

	id = generateClientID()
	client.targets = append(client.targets, target)
	client.targteIds[target] = id

	return
}

// RemoveTarget removes a target to the client
func (client *Client) RemoveTarget(target Target) (id string, err error) {
	if target == client.status {
		return "", ErrTargetIsStatus
	}

	client.mutex.Lock()
	defer client.mutex.Unlock()

	for i := range client.targets {
		if target == client.targets[i] {
			id = client.targteIds[target]

			client.targets[i] = client.targets[len(client.targets)-1]
			client.targets = client.targets[:len(client.targets)-1]
			delete(client.targteIds, target)

			return
		}
	}

	err = ErrTargetNotFound
	return
}

// FindUser checks each channel to find user info about a user.
func (client *Client) FindUser(nick string) (u list.User, ok bool) {
	client.mutex.RLock()
	defer client.mutex.RUnlock()

	for _, target := range client.targets {
		channel, ok := target.(*Channel)
		if !ok {
			continue
		}

		user, ok := channel.UserList().User(nick)
		if !ok {
			continue
		}

		return user, true
	}

	return list.User{}, false
}

func (client *Client) handleEventLoop() {
	ticker := time.NewTicker(time.Second * 30)

	for {
		select {
		case event, ok := <-client.events:
			{
				if !ok {
					goto end
				}

				client.handleEvent(event)
				emit(event, client)

				event.cancel()
			}
		case <-ticker.C:
			{
				event := NewEvent("client", "tick")
				event.ctx, event.cancel = context.WithCancel(client.ctx)

				client.handleEvent(&event)
				emit(&event, client)

				event.cancel()
			}
		case <-client.ctx.Done():
			{
				goto end
			}
		}
	}

end:

	ticker.Stop()

	client.Disconnect()

	event := NewEvent("client", "destroy")
	event.ctx, event.cancel = context.WithCancel(client.ctx)

	client.handleEvent(&event)
	emit(&event, client)

	event.cancel()
}

func (client *Client) handleSendLoop() {
	lastRefresh := time.Time{}
	queue := 2

	for line := range client.sends {
		now := time.Now()
		deltaTime := now.Sub(lastRefresh)

		if deltaTime < time.Second {
			queue--
			if queue <= 0 {
				time.Sleep(time.Second - deltaTime)
				lastRefresh = now

				queue = 0
			}
		} else {
			lastRefresh = now
		}

		client.Send(line)
	}
}

// handleEvent is always first and gets to break a few rules.
func (client *Client) handleEvent(event *Event) {
	// IRCv3 `server-time`
	if timeTag, ok := event.Tags["time"]; ok {
		serverTime, err := time.Parse(time.RFC3339Nano, timeTag)
		if err == nil && serverTime.Year() > 2000 {
			event.Time = serverTime
		}
	}

	switch event.name {

	// Ping Pong
	case "hook.tick":
		{
			client.mutex.RLock()
			lastSend := time.Since(client.lastSend)
			client.mutex.RUnlock()

			if lastSend > time.Second*120 {
				client.Sendf("PING :%x%x%x", mathRand.Int63(), mathRand.Int63(), mathRand.Int63())
			}
		}
	case "packet.ping":
		{
			message := "PONG"
			for _, arg := range event.Args {
				message += " " + arg
			}
			if event.Text != "" {
				message += " :" + event.Text
			}

			client.Send(message + "")
		}

	// Client Registration
	case "client.connect":
		{
			client.Send("CAP LS 302")

			if client.config.Password != "" {
				client.Sendf("PASS :%s", client.config.Password)
			}

			nick := client.config.Nick
			client.mutex.RLock()
			if client.nick != "" {
				nick = client.nick
			}
			client.mutex.RUnlock()
			client.Sendf("NICK %s", nick)

			client.Sendf("USER %s 8 * :%s", client.config.User, client.config.RealName)
		}

	case "packet.001":
		{
			client.mutex.Lock()
			client.nick = event.Args[0]
			client.mutex.Unlock()

			client.Sendf("WHO %s", event.Args[0])
		}

	case "packet.443":
		{
			client.mutex.RLock()
			hasRegistered := client.nick != ""
			client.mutex.RUnlock()

			if !hasRegistered {
				nick := event.Args[1]

				// "AltN" -> "AltN+1", ...
				prev := client.config.Nick
				sent := false
				for _, alt := range client.config.Alternatives {
					if nick == prev {
						client.Sendf("NICK %s", alt)
						sent = true
						break
					}

					prev = alt
				}

				if !sent {
					// "LastAlt" -> "Nick23962"
					client.Sendf("NICK %s%05d", client.config.Nick, mathRand.Int31n(99999))
				}
			}
		}

	case "packet.nick":
		{
			client.handleInTargets(event.Nick, event)

			if event.Nick == client.nick {
				client.SetValue("nick", event.Arg(0))
			}
		}

	// Handle ISupport
	case "packet.005":
		{
			for _, token := range event.Args[1:] {
				kvpair := strings.Split(token, "=")

				if len(kvpair) == 2 {
					client.isupport.Set(kvpair[0], kvpair[1])
				} else {
					client.isupport.Set(kvpair[0], "")
				}
			}
		}

	// Capability negotiation
	case "packet.cap":
		{
			capCommand := event.Args[1]
			capTokens := strings.Split(event.Text, " ")

			switch capCommand {
			case "LS":
				{
					for _, token := range capTokens {
						split := strings.SplitN(token, "=", 2)
						key := split[0]
						if len(key) == 0 {
							continue
						}

						if len(split) == 2 {
							client.capData[key] = split[1]
						}

						for i := range supportedCaps {
							if supportedCaps[i] == token {
								client.mutex.Lock()
								client.capsRequested = append(client.capsRequested, token)
								client.mutex.Unlock()

								break
							}
						}
					}

					if len(event.Args) < 3 || event.Args[2] != "*" {
						client.mutex.RLock()
						requestedCount := len(client.capsRequested)
						client.mutex.RUnlock()

						if requestedCount > 0 {
							client.mutex.RLock()
							requestedCaps := strings.Join(client.capsRequested, " ")
							client.mutex.RUnlock()

							client.Send("CAP REQ :" + requestedCaps)
						} else {
							client.Send("CAP END")
						}
					}
				}
			case "ACK":
				{
					for _, token := range capTokens {
						client.mutex.Lock()
						if client.capEnabled[token] {
							client.capEnabled[token] = true
						}
						client.mutex.Unlock()
					}

					client.Send("CAP END")
				}
			case "NAK":
				{
					// Remove offenders
					for _, token := range capTokens {
						client.mutex.Lock()
						for i := range client.capsRequested {
							if token == client.capsRequested[i] {
								client.capsRequested = append(client.capsRequested[:i], client.capsRequested[i+1:]...)
								break
							}
						}
						client.mutex.Unlock()
					}

					client.mutex.RLock()
					requestedCaps := strings.Join(client.capsRequested, " ")
					client.mutex.RUnlock()

					client.Send("CAP REQ :" + requestedCaps)
				}
			case "NEW":
				{
					requests := make([]string, 0, len(capTokens))

					for _, token := range capTokens {
						for i := range supportedCaps {
							if supportedCaps[i] == token {
								requests = append(requests, token)
							}
						}
					}

					if len(requests) > 0 {
						client.Send("CAP REQ :" + strings.Join(requests, " "))
					}
				}
			case "DEL":
				{
					for _, token := range capTokens {
						client.mutex.Lock()
						if client.capEnabled[token] {
							client.capEnabled[token] = false
						}
						client.mutex.Unlock()
					}
				}
			}
		}

	// User/host detection
	case "packet.352": // WHO reply
		{
			// Example args: test * ~irce 127.0.0.1 localhost.localnetwork Gissleh H :0 ...
			nick := event.Args[5]
			user := event.Args[2]
			host := event.Args[3]

			if nick == client.nick {
				client.mutex.Lock()
				client.user = user
				client.host = host
				client.mutex.Unlock()
			}
		}

	case "packet.chghost":
		{
			if event.Nick == client.nick {
				client.mutex.Lock()
				client.user = event.Args[1]
				client.host = event.Args[2]
				client.mutex.Unlock()
			}

			// This may be relevant in channels where the client resides.
			client.handleInTargets(event.Nick, event)
		}

	// Channel join/leave/mode handling
	case "packet.join":
		{
			var channel *Channel

			if event.Nick == client.nick {
				channel = &Channel{name: event.Arg(0), userlist: list.New(&client.isupport)}
				client.AddTarget(channel)
			} else {
				channel = client.Channel(event.Arg(0))
			}

			client.handleInTarget(channel, event)

			if channel != nil {
				channel.Handle(event, client)
			}
		}

	case "packet.part":
		{
			channel := client.Channel(event.Arg(0))
			if channel == nil {
				break
			}

			channel.Handle(event, client)

			if event.Nick == client.nick {
				client.RemoveTarget(channel)
			} else {
				client.handleInTarget(channel, event)
			}
		}

	case "packet.quit":
		{
			client.handleInTargets(event.Nick, event)
		}

	case "packet.353": // NAMES
		{
			channel := client.Channel(event.Arg(2))
			if channel != nil {
				client.handleInTarget(channel, event)
			}
		}

	case "packet.366": // End of NAMES
		{
			channel := client.Channel(event.Arg(1))
			if channel != nil {
				client.handleInTarget(channel, event)
			}
		}

	case "packet.mode":
		{
			targetName := event.Arg(0)

			if client.isupport.IsChannel(targetName) {
				channel := client.Channel(targetName)
				if channel != nil {
					client.handleInTarget(channel, event)
				}
			}
		}

	// Message parsing
	case "packet.privmsg", "ctcp.action":
		{
			// Target the mssage
			target := Target(client.status)
			spawned := false
			targetName := event.Arg(0)
			if targetName == client.nick {
				target := client.Target("query", targetName)
				if target == nil {
					query := &Query{user: list.User{
						Nick: event.Nick,
						User: event.User,
						Host: event.Host,
					}}

					client.AddTarget(query)

					spawned = true

					target = query
				}
			} else {
				channel := client.Channel(targetName)
				if channel != nil {
					if user, ok := channel.UserList().User(event.Nick); ok {
						event.RenderTags["prefixedNick"] = user.PrefixedNick
					}

					target = channel
				} else {
					target = client.status
				}
			}

			client.handleInTarget(target, event)

			if spawned {
				// TODO: Message has higher importance // 0:Normal, 1:Important, 2:Highlight
			}
		}

	case "packet.notice":
		{
			// Try to target by mentioned channel name
			for _, token := range strings.Fields(event.Text) {
				if client.isupport.IsChannel(token) {
					channel := client.Channel(token)
					if channel == nil {
						continue
					}

					if user, ok := channel.UserList().User(event.Nick); ok {
						event.RenderTags["prefixedNick"] = user.PrefixedNick
					}

					client.handleInTarget(channel, event)
					break
				}
			}

			// Otherwise, it belongs in the status target
			if len(event.targets) == 0 {
				client.status.Handle(event, client)
				client.handleInTarget(client.status, event)
			}
		}

	// account-notify
	case "packet.account":
		{
			client.handleInTargets(event.Nick, event)
		}

	// away-notify
	case "packet.away":
		{
			client.handleInTargets(event.Nick, event)
		}
	}

	if len(event.targets) == 0 {
		client.handleInTarget(client.status, event)
	}
}

func (client *Client) handleInTargets(nick string, event *Event) {
	client.mutex.RLock()
	for i := range client.targets {
		switch target := client.targets[i].(type) {
		case *Channel:
			{
				if nick != "" {
					if _, ok := target.UserList().User(event.Nick); !ok {
						continue
					}
				}

				target.Handle(event, client)

				event.targets = append(event.targets, target)
				event.targetIds[target] = client.targteIds[target]
			}
		case *Query:
			{
				if target.user.Nick == nick {
					target.Handle(event, client)

					event.targets = append(event.targets, target)
					event.targetIds[target] = client.targteIds[target]
				}
			}
		case *Status:
			{
				if client.nick == event.Nick {
					target.Handle(event, client)

					event.targets = append(event.targets, target)
					event.targetIds[target] = client.targteIds[target]
				}
			}
		}
	}

	client.mutex.RUnlock()
}

func (client *Client) handleInTarget(target Target, event *Event) {
	client.mutex.RLock()

	target.Handle(event, client)

	event.targets = append(event.targets, target)
	event.targetIds[target] = client.targteIds[target]

	client.mutex.RUnlock()
}

func generateClientID() string {
	bytes := make([]byte, 12)
	_, err := rand.Read(bytes)

	// Ugly fallback if crypto rand doesn't work.
	if err != nil {
		rng := mathRand.NewSource(time.Now().UnixNano())
		result := strconv.FormatInt(rng.Int63(), 16)
		for len(result) < 24 {
			result += strconv.FormatInt(rng.Int63(), 16)
		}

		return result[:24]
	}

	binary.BigEndian.PutUint32(bytes[4:], uint32(time.Now().Unix()))

	return hex.EncodeToString(bytes)
}
