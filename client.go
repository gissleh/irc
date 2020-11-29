package irc

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	mathRand "math/rand"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gissleh/irc/ircutil"
	"github.com/gissleh/irc/isupport"
	"github.com/gissleh/irc/list"
)

var supportedCaps = []string{
	"server-time",
	"cap-notify",
	"multi-prefix",
	"userhost-in-names",
	"account-notify",
	"away-notify",
	"invite-notify",
	"extended-join",
	"chghost",
	"account-tag",
	"echo-message",
	"draft/languages",
	"sasl",
}

// ErrNoConnection is returned if you try to do something requiring a connection,
// but there is none.
var ErrNoConnection = errors.New("irc: no connection")

// ErrTargetAlreadyAdded is returned by Client.AddTarget if that target has already been
// added to the client.
var ErrTargetAlreadyAdded = errors.New("irc: target already added")

// ErrTargetConflict is returned by Client.AddTarget if there already exists a target
// matching the name and kind.
var ErrTargetConflict = errors.New("irc: target name and kind match existing target")

// ErrTargetNotFound is returned by Client.RemoveTarget if the target is not part of
// the client's target list
var ErrTargetNotFound = errors.New("irc: target not found")

// ErrTargetIsStatus is returned by Client.RemoveTarget if the target is the client's
// status target
var ErrTargetIsStatus = errors.New("irc: cannot remove status target")

// ErrDestroyed is returned by Client.Connect if you try to connect a destroyed client.
var ErrDestroyed = errors.New("irc: client destroyed")

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
	ready    bool
	isupport isupport.ISupport
	values   map[string]interface{}

	status  *Status
	targets []Target

	handlers []Handler
}

// New creates a new client. The context can be context.Background if you want manually to
// tear down clients upon quitting.
func New(ctx context.Context, config Config) *Client {
	client := &Client{
		id:         generateClientID("C"),
		values:     make(map[string]interface{}),
		events:     make(chan *Event, 64),
		sends:      make(chan string, 64),
		capEnabled: make(map[string]bool),
		capData:    make(map[string]string),
		config:     config.WithDefaults(),
		status:     &Status{id: generateClientID("T")},
	}

	client.ctx, client.cancel = context.WithCancel(ctx)

	_ = client.AddTarget(client.status)

	go client.handleEventLoop()
	go client.handleSendLoop()

	client.EmitNonBlocking(NewEvent("client", "create"))

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

// CapData returns if there was any additional CAP data for the given capability.
func (client *Client) CapData(cap string) string {
	client.mutex.RLock()
	defer client.mutex.RUnlock()

	return client.capData[cap]
}

// CapEnabled returns whether an IRCv3 capability is enabled.
func (client *Client) CapEnabled(cap string) bool {
	client.mutex.RLock()
	defer client.mutex.RUnlock()

	return client.capEnabled[cap]
}

// Ready returns true if the client is marked as ready, which means that it has received the MOTD.
func (client *Client) Ready() bool {
	client.mutex.RLock()
	defer client.mutex.RUnlock()

	return client.ready
}

// HasQuit returns true if the client had manually quit. It should be checked before
// performing any reconnection logic.
func (client *Client) HasQuit() bool {
	client.mutex.RLock()
	defer client.mutex.RUnlock()

	return client.quit
}

func (client *Client) State() ClientState {
	client.mutex.RLock()

	state := ClientState{
		Nick:      client.nick,
		User:      client.user,
		Host:      client.host,
		Connected: client.conn != nil,
		Ready:     client.ready,
		Quit:      client.quit,
		ISupport:  client.isupport.State(),
		Caps:      make([]string, 0, len(client.capEnabled)),
		Targets:   make([]ClientStateTarget, 0, len(client.targets)),
	}

	for key, enabled := range client.capEnabled {
		if enabled {
			state.Caps = append(state.Caps, key)
		}
	}
	sort.Strings(state.Caps)

	for _, target := range client.targets {
		tstate := target.State()
		tstate.ID = target.ID()

		state.Targets = append(state.Targets, tstate)
	}

	client.mutex.RUnlock()

	return state
}

// Connect connects to the server by addr.
func (client *Client) Connect(addr string, ssl bool) (err error) {
	var conn net.Conn

	if client.Connected() {
		_ = client.Disconnect(false)
	}

	client.isupport.Reset()

	client.mutex.Lock()
	client.quit = false
	client.mutex.Unlock()

	client.EmitNonBlocking(NewEvent("client", "connecting"))

	if ssl {
		conn, err = tls.Dial("tcp", addr, &tls.Config{
			InsecureSkipVerify: client.config.SkipSSLVerification,
		})
		if err != nil {
			if !client.Destroyed() {
				client.EmitNonBlocking(NewErrorEvent("connect", "Connect failed: "+err.Error()))
			}
			return err
		}
	} else {
		conn, err = net.Dial("tcp", addr)
		if err != nil {
			if !client.Destroyed() {
				client.EmitNonBlocking(NewErrorEvent("connect", "Connect failed: "+err.Error()))
			}
			return err
		}
	}

	if client.Destroyed() {
		_ = conn.Close()
		return ErrDestroyed
	}

	client.EmitNonBlocking(NewEvent("client", "connect"))

	go func() {
		reader := bufio.NewReader(conn)
		replacer := strings.NewReplacer("\r", "", "\n", "")

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				client.EmitNonBlocking(NewErrorEvent("read", "Read failed: "+err.Error()))
				break
			}
			line = replacer.Replace(line)

			event, err := ParsePacket(line)
			if err != nil {
				client.mutex.RLock()
				hasQuit := client.quit
				client.mutex.RUnlock()

				if !hasQuit {
					client.EmitNonBlocking(NewErrorEvent("parse", "Read failed: "+err.Error()))
				}
				continue
			}

			client.EmitNonBlocking(event)
		}

		_ = client.conn.Close()

		client.mutex.Lock()
		client.conn = nil
		client.ready = false
		client.mutex.Unlock()

		client.EmitNonBlocking(NewEvent("client", "disconnect"))
	}()

	client.mutex.Lock()
	client.conn = conn
	client.mutex.Unlock()

	return nil
}

// Disconnect disconnects from the server. It will either return the
// close error, or ErrNoConnection if there is no connection. If
// markAsQuit is specified, HasQuit will return true until the next
// connections.
func (client *Client) Disconnect(markAsQuit bool) error {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	if markAsQuit {
		client.quit = true
	}

	if client.conn == nil {
		return ErrNoConnection
	}

	return client.conn.Close()
}

// Connected returns true if the client has a connection
func (client *Client) Connected() bool {
	client.mutex.RLock()
	defer client.mutex.RUnlock()

	return client.conn != nil
}

// Send sends a line to the server. A line-feed will be automatically added if one
// is not provided. If this isn't part of early registration, SendQueued might save
// you from a potential flood kick.
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

	_ = conn.SetWriteDeadline(time.Now().Add(time.Second * 30))
	_, err := conn.Write([]byte(line))
	if err != nil {
		client.EmitNonBlocking(NewErrorEvent("write", err.Error()))
		_ = client.Disconnect(false)
	}

	return err
}

// Sendf is Send with a fmt.Sprintf. If this isn't part of early registration,
// SendQueuedf might save you from a potential flood kick.
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

// SendCTCP sends a queued message with the following CTCP verb and text. If reply is true,
// it will use a NOTICE instead of PRIVMSG.
func (client *Client) SendCTCP(verb, targetName string, reply bool, text string) {
	ircVerb := "PRIVMSG"
	if reply {
		ircVerb = "NOTICE"
	}

	client.SendQueuedf("%s %s :\x01%s %s\x01", ircVerb, targetName, verb, text)
}

// SendCTCPf is SendCTCP with a fmt.Sprintf
func (client *Client) SendCTCPf(verb, targetName string, reply bool, format string, a ...interface{}) {
	client.SendCTCP(verb, targetName, reply, fmt.Sprintf(format, a...))
}

// Say sends a PRIVMSG with the target name and text, cutting the message if it gets too long.
func (client *Client) Say(targetName string, text string) {
	overhead := client.PrivmsgOverhead(targetName, false)
	cuts := ircutil.CutMessage(text, overhead)

	for _, cut := range cuts {
		client.SendQueuedf("PRIVMSG %s :%s", targetName, cut)
	}
}

// Sayf is Say with a fmt.Sprintf.
func (client *Client) Sayf(targetName string, format string, a ...interface{}) {
	client.Say(targetName, fmt.Sprintf(format, a...))
}

// Describe sends a CTCP ACTION with the target name and text, cutting the message if it gets too long.
func (client *Client) Describe(targetName string, text string) {
	overhead := client.PrivmsgOverhead(targetName, true)
	cuts := ircutil.CutMessage(text, overhead)

	for _, cut := range cuts {
		client.SendQueuedf("PRIVMSG %s :\x01ACTION %s\x01", targetName, cut)
	}
}

// Describef is Describe with a fmt.Sprintf.
func (client *Client) Describef(targetName string, format string, a ...interface{}) {
	client.Describe(targetName, fmt.Sprintf(format, a...))
}

// Emit sends an event through the client's event, and it will return immediately
// unless the internal channel is filled up. The returned context can be used to
// wait for the event, or the client's destruction.
func (client *Client) Emit(event Event) context.Context {
	event.ctx, event.cancel = context.WithCancel(client.ctx)
	if client.Destroyed() {
		event.cancel()
		return event.ctx
	}

	client.events <- &event

	return event.ctx
}

// EmitNonBlocking is just like emitInGlobalHandlers, but it will spin off a goroutine if the channel is full.
// This lets it be called from other handlers without ever blocking. See Emit for what the
// returned context is for.
func (client *Client) EmitNonBlocking(event Event) context.Context {
	event.ctx, event.cancel = context.WithCancel(client.ctx)
	if client.Destroyed() {
		event.cancel()
		return event.ctx
	}

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

// EmitInput emits an input event parsed from the line.
func (client *Client) EmitInput(line string, target Target) context.Context {
	event := ParseInput(line)

	client.mutex.RLock()
	if target != nil && client.TargetByID(target.ID()) == nil {
		client.EmitNonBlocking(NewErrorEvent("invalid_target", "Target does not exist."))

		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		return ctx
	}
	client.mutex.RUnlock()

	if target != nil {
		client.mutex.RLock()
		event.targets = append(event.targets, target)
		client.mutex.RUnlock()
	} else {
		client.mutex.RLock()
		event.targets = append(event.targets, client.status)
		client.mutex.RUnlock()
	}

	return client.Emit(event)
}

// Value gets a client value.
func (client *Client) Value(key string) interface{} {
	client.mutex.RLock()
	defer client.mutex.RUnlock()

	return client.values[key]
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
	_ = client.Disconnect(false)
	client.cancel()
	close(client.sends)

	client.Emit(NewEvent("client", "destroy"))

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
func (client *Client) Join(channels ...string) {
	client.SendQueuedf("JOIN %s", strings.Join(channels, ","))
}

// Part parts one or more channels.
func (client *Client) Part(channels ...string) {
	client.SendQueuedf("PART %s", strings.Join(channels, ","))
}

// Quit sends a quit message and marks the client as having quit, which
// means HasQuit() will return true.
func (client *Client) Quit(reason string) {
	client.mutex.Lock()
	client.quit = true
	client.mutex.Unlock()

	client.SendQueuedf("QUIT :%s", reason)
}

// Target gets a target by kind and name
func (client *Client) Target(kind string, name string) Target {
	client.mutex.RLock()
	defer client.mutex.RUnlock()

	for _, target := range client.targets {
		if target.Kind() == kind && strings.EqualFold(name, target.Name()) {
			return target
		}
	}

	return nil
}

// TargetByID gets a target by kind and name
func (client *Client) TargetByID(id string) Target {
	client.mutex.RLock()
	defer client.mutex.RUnlock()

	for _, target := range client.targets {
		if target.ID() == id {
			return target
		}
	}

	return nil
}

// Targets gets all targets of the given kinds.
func (client *Client) Targets(kinds ...string) []Target {
	if len(kinds) == 0 {
		client.mutex.Lock()
		targets := make([]Target, len(client.targets))
		copy(targets, client.targets)
		client.mutex.Unlock()

		return targets
	}

	client.mutex.Lock()
	targets := make([]Target, 0, len(client.targets))
	for _, target := range client.targets {
		for _, kind := range kinds {
			if target.Kind() == kind {
				targets = append(targets, target)
				break
			}
		}
	}
	client.mutex.Unlock()

	return targets
}

// Status gets the client's status target.
func (client *Client) Status() *Status {
	return client.status
}

// Channel is a shorthand for getting a channel target and type asserting it.
func (client *Client) Channel(name string) *Channel {
	target := client.Target("channel", name)
	if target == nil {
		return nil
	}

	return target.(*Channel)
}

// Channels gets all channel targets the client has.
func (client *Client) Channels() []*Channel {
	targets := client.Targets("channel")
	channels := make([]*Channel, len(targets))

	for i := range targets {
		channels[i] = targets[i].(*Channel)
	}

	return channels
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
func (client *Client) AddTarget(target Target) (err error) {
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

	client.targets = append(client.targets, target)

	event := NewEvent("hook", "add_target")
	event.Args = []string{target.ID(), target.Kind(), target.Name()}
	event.targets = []Target{target}
	client.EmitNonBlocking(event)

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
			id = target.ID()

			event := NewEvent("hook", "remove_target")
			event.Args = []string{target.ID(), target.Kind(), target.Name()}
			client.EmitNonBlocking(event)

			client.targets[i] = client.targets[len(client.targets)-1]
			client.targets = client.targets[:len(client.targets)-1]

			// Ensure the channel has been parted
			if channel, ok := target.(*Channel); ok && !channel.parted {
				client.SendQueuedf("PART %s", channel.Name())
			}

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

// AddHandler adds a handler. This is thread safe, unlike adding global handlers.
func (client *Client) AddHandler(handler Handler) {
	client.mutex.Lock()
	client.handlers = append(client.handlers[:0], client.handlers...)
	client.handlers = append(client.handlers, handler)
	client.mutex.Unlock()
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

				// Turn an unhandled input into a raw command.
				if event.kind == "input" && !event.preventedDefault {
					client.SendQueued(strings.ToUpper(event.verb) + " " + event.Text)
				}

				event.cancel()
			}
		case <-ticker.C:
			{
				event := NewEvent("client", "tick")
				event.ctx, event.cancel = context.WithCancel(client.ctx)

				client.handleEvent(&event)

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

	_ = client.Disconnect(false)
}

func (client *Client) handleSendLoop() {
	lastRefresh := time.Time{}
	queue := client.config.SendRate

	for line := range client.sends {
		now := time.Now()
		deltaTime := now.Sub(lastRefresh)

		if deltaTime < time.Second {
			queue--
			if queue <= 0 {
				time.Sleep(time.Second - deltaTime)
				lastRefresh = now

				queue = client.config.SendRate - 1
			}
		} else {
			lastRefresh = now
			queue = client.config.SendRate - 1
		}

		_ = client.Send(line)
	}
}

// handleEvent is always first and gets to break a few rules.
func (client *Client) handleEvent(event *Event) {
	sentCapEnd := false

	// Only use IRCv3 `server-time` to overwrite when requested. Frontends/dependents can still
	// get this information.
	if client.config.UseServerTime {
		if timeTag, ok := event.Tags["time"]; ok {
			serverTime, err := time.Parse(time.RFC3339Nano, timeTag)
			if err == nil && serverTime.Year() > 2000 {
				event.Time = serverTime
			}
		}
	}

	// For events that were created with targets, handle them now there now.
	for _, target := range event.targets {
		target.Handle(event, client)
	}

	switch event.name {

	// Ping Pong
	case "hook.tick":
		{
			client.mutex.RLock()
			lastSend := time.Since(client.lastSend)
			client.mutex.RUnlock()

			if lastSend > time.Second*120 {
				_ = client.Sendf("PING :%x%x%x", mathRand.Int63(), mathRand.Int63(), mathRand.Int63())
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

			_ = client.Send(message)
		}

	// Client Registration
	case "client.connect":
		{
			// Clear enabled caps and initiate negotiation.
			client.mutex.Lock()
			for key := range client.capEnabled {
				delete(client.capEnabled, key)
			}
			client.mutex.Unlock()
			_ = client.Send("CAP LS 302")

			// Send server password if configured.
			if client.config.Password != "" {
				_ = client.Sendf("PASS :%s", client.config.Password)
			}

			// Reuse nick or get from config
			nick := client.config.Nick
			client.mutex.RLock()
			if client.nick != "" {
				nick = client.nick
			}
			client.mutex.RUnlock()

			// Clear connection-specific data
			client.mutex.Lock()
			client.nick = ""
			client.user = ""
			client.host = ""
			client.capsRequested = client.capsRequested[:0]
			for key := range client.capData {
				delete(client.capData, key)
			}
			for key := range client.capEnabled {
				delete(client.capEnabled, key)
			}
			client.mutex.Unlock()

			// Start registration.
			_ = client.Sendf("NICK %s", nick)
			_ = client.Sendf("USER %s 8 * :%s", client.config.User, client.config.RealName)
		}

	// Welcome message
	case "packet.001":
		{
			client.mutex.Lock()
			client.nick = event.Args[0]
			client.mutex.Unlock()

			// Send a WHO right away to gather enough client information for precise message cutting.
			_ = client.Sendf("WHO %s", event.Args[0])
		}

	// Nick rotation
	case "packet.431", "packet.432", "packet.433", "packet.436":
		{
			lockNickChange, _ := client.Value("internal.lockNickChange").(bool)

			// Ignore if client is registered
			if client.Nick() != "" {
				break
			}

			nick := event.Args[1]
			newNick := ""

			// "AltN" -> "AltN+1", ...
			prev := client.config.Nick
			for _, alt := range client.config.Alternatives {
				if nick == prev {
					newNick = alt
					break
				}

				prev = alt
			}

			if newNick == "" {
				// "LastAlt" -> "Nick23962"
				newNick = fmt.Sprintf("%s%05d", client.config.Nick, mathRand.Int31n(99999))
			}

			if lockNickChange {
				client.SetValue("internal.primedNickChange", newNick)
			} else {
				_ = client.Sendf("NICK %s", newNick)
			}
		}

	case "packet.nick":
		{
			client.handleInTargets(event.Nick, event)

			if event.Nick == client.nick {
				client.SetValue("nick", event.Arg(0))
			}
		}

	// ISupport
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
					client.SetValue("internal.lockNickChange", true)

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
							if supportedCaps[i] == key {
								client.mutex.Lock()
								client.capsRequested = append(client.capsRequested, key)
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

							_ = client.Send("CAP REQ :" + requestedCaps)
						} else {
							sentCapEnd = true
							_ = client.Send("CAP END")
						}
					}
				}
			case "ACK":
				{
					for _, token := range capTokens {
						client.mutex.Lock()
						if !client.capEnabled[token] {
							client.capEnabled[token] = true
						}
						client.mutex.Unlock()

						// Special cases for supported tokens
						switch token {
						case "sasl":
							{
								if client.config.SASL == nil {
									break
								}

								mechanisms := strings.Split(client.capData[token], ",")
								selectedMechanism := ""
								if len(mechanisms) == 0 || mechanisms[0] == "" {
									selectedMechanism = "PLAIN"
								}
								for _, mechanism := range mechanisms {
									if mechanism == "PLAIN" && selectedMechanism == "" {
										selectedMechanism = "PLAIN"
									}
								}

								// TODO: Add better mechanisms
								if selectedMechanism != "" {
									_ = client.Sendf("AUTHENTICATE %s", selectedMechanism)
									client.SetValue("sasl.usingMethod", "PLAIN")
								}
							}

						case "draft/languages":
							{
								if len(client.config.Languages) == 0 {
									break
								}

								// draft/languages=15,en,~bs,~de,~el,~en-AU,~es,~fi,~fr-FR,~it,~no,~pl,~pt-BR,~ro,~tr-TR,~zh-CN
								langData := strings.Split(client.capData[token], ",")
								if len(langData) < 0 {
									break
								}
								maxCount, err := strconv.Atoi(langData[0])
								if err != nil {
									break
								}

								languages := make([]string, 0, maxCount)

							LanguageLoop:
								for _, lang := range client.config.Languages {
									for _, lang2 := range langData[1:] {
										if strings.HasPrefix(lang2, "~") {
											lang2 = lang2[1:]
										}
										if strings.EqualFold(lang, lang2) {
											languages = append(languages, lang)
											if len(languages) >= maxCount {
												break LanguageLoop
											}
										}
									}
								}

								if len(languages) > 0 {
									_ = client.Send("LANGUAGE " + strings.Join(languages, " "))
								}
							}
						}
					}

					if !client.Ready() {
						sentCapEnd = true
						_ = client.Send("CAP END")
					}
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

					_ = client.Send("CAP REQ :" + requestedCaps)
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
						_ = client.Send("CAP REQ :" + strings.Join(requests, " "))
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

	// SASL
	case "packet.authenticate":
		{
			if event.Arg(0) != "+" {
				break
			}

			method, ok := client.Value("sasl.usingMethod").(string)
			if !ok {
				break
			}

			switch method {
			case "PLAIN":
				{
					parts := [][]byte{
						[]byte(client.config.SASL.AuthenticationIdentity),
						[]byte(client.config.SASL.AuthorizationIdentity),
						[]byte(client.config.SASL.Password),
					}
					plainString := base64.StdEncoding.EncodeToString(bytes.Join(parts, []byte{0x00}))

					_ = client.Sendf("AUTHENTICATE %s", plainString)
				}
			}
		}
	case "packet.904": // Auth failed
		{
			// Cancel authentication.
			_ = client.Sendf("AUTHENTICATE *")
			client.SetValue("sasl.usingMethod", (interface{})(nil))
		}
	case "packet.903", "packet.906": // Auth ended
		{
			// A bit dirty, but it'll get the nick rotation started again.
			if client.Nick() == "" {
				_ = client.Sendf("NICK %s", client.config.Nick)
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
				channel = &Channel{
					id:       generateClientID("T"),
					name:     event.Arg(0),
					userlist: list.New(&client.isupport),
				}
				_ = client.AddTarget(channel)
			} else {
				channel = client.Channel(event.Arg(0))
			}

			client.handleInTarget(channel, event)
		}

	case "packet.part":
		{
			channel := client.Channel(event.Arg(0))
			if channel == nil {
				break
			}

			if event.Nick == client.nick {
				channel.parted = true
				_, _ = client.RemoveTarget(channel)
			} else {
				client.handleInTarget(channel, event)
			}
		}

	case "packet.kick":
		{
			channel := client.Channel(event.Arg(0))
			if channel == nil {
				break
			}

			if event.Arg(1) == client.nick {
				channel.parted = true
				_, _ = client.RemoveTarget(channel)
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

	case "packet.invite":
		{
			inviteeNick := event.Arg(0)
			channelName := event.Arg(1)
			channel := client.Channel(channelName)

			if client.config.AutoJoinInvites && inviteeNick == client.Nick() {
				if channel == nil {
					client.Join(channelName)
				}
			}

			// Add channel target for rendering invite-notify invitations.
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
			// Target the message
			target := Target(client.status)
			targetName := event.Arg(0)

			if targetName == client.nick {
				queryTarget := client.Target("query", event.Nick)
				if queryTarget == nil {
					query := &Query{
						id: client.id,
						user: list.User{
							Nick: event.Nick,
							User: event.User,
							Host: event.Host,
						},
					}
					if accountTag, ok := event.Tags["account"]; ok {
						query.user.Account = accountTag
					}

					_ = client.AddTarget(query)
					event.RenderTags["spawned"] = query.id

					queryTarget = query
				}

				target = queryTarget
			} else {
				channel := client.Channel(targetName)
				if channel != nil {
					if user, ok := channel.UserList().User(event.Nick); ok {
						event.RenderTags["prefixedNick"] = user.PrefixedNick
					}

					target = channel
				}
			}

			client.handleInTarget(target, event)
		}

	case "packet.notice":
		{
			// Find channel target
			targetName := event.Arg(0)
			if client.isupport.IsChannel(targetName) {
				channel := client.Channel(targetName)
				if channel != nil {
					if user, ok := channel.UserList().User(event.Nick); ok {
						event.RenderTags["prefixedNick"] = user.PrefixedNick
					}

					client.handleInTarget(channel, event)
				}
			} else {
				// Try to target by mentioned channel name.
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

	// Auto-rejoin
	case "packet.376", "packet.422":
		{
			client.mutex.RLock()
			channels := make([]string, 0, len(client.targets))
			rejoinEvent := NewEvent("info", "rejoin")
			for _, target := range client.targets {
				if channel, ok := target.(*Channel); ok {
					channels = append(channels, channel.Name())

					rejoinEvent.targets = append(rejoinEvent.targets, target)
				}
			}
			client.mutex.RUnlock()

			if len(channels) > 0 {
				_ = client.Sendf("JOIN %s", strings.Join(channels, ","))
				client.EmitNonBlocking(rejoinEvent)
			}

			client.mutex.Lock()
			client.ready = true
			client.mutex.Unlock()

			client.EmitNonBlocking(NewEvent("hook", "ready"))
		}
	}

	if sentCapEnd {
		client.SetValue("internal.lockNickChange", false)

		if primedNick, _ := client.Value("internal.primedNickChange").(string); primedNick != "" {
			_ = client.Sendf("NICK %s", primedNick)
		}
	}

	if len(event.targets) == 0 {
		client.handleInTarget(client.status, event)
	}

	client.mutex.RLock()
	clientHandlers := client.handlers
	client.mutex.RUnlock()

	for _, handler := range clientHandlers {
		handler(event, client)
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
			}
		case *Query:
			{
				if target.user.Nick == nick {
					target.Handle(event, client)

					event.targets = append(event.targets, target)
				}
			}
		case *Status:
			{
				if client.nick == event.Nick {
					target.Handle(event, client)

					event.targets = append(event.targets, target)
				}
			}
		}
	}

	client.mutex.RUnlock()
}

func (client *Client) handleInTarget(target Target, event *Event) {
	if target == nil {
		return
	}

	event.targets = append(event.targets, target)
	target.Handle(event, client)
}

func generateClientID(prefix string) string {
	buffer := [12]byte{}
	_, err := rand.Read(buffer[:])

	// Ugly fallback if crypto rand doesn't work.
	if err != nil {
		mathRand.Read(buffer[:])
	}

	binary.BigEndian.PutUint32(buffer[4:], uint32(time.Now().Unix()))

	return prefix + hex.EncodeToString(buffer[:])[1:]
}
