package irc

import (
	"context"
	"encoding/json"
	"time"
)

// An Event is any thing that passes through the irc client's event loop. It's not thread safe, because it's processed
// in sequence and should not be used off the goroutine that processed it.
type Event struct {
	kind string
	verb string
	name string

	Time time.Time
	Nick string
	User string
	Host string
	Args []string
	Text string
	Tags map[string]string

	ctx    context.Context
	cancel context.CancelFunc
	killed bool
	hidden bool

	targets   []Target
	targetIds map[Target]string

	RenderTags map[string]string
}

// NewEvent makes a new event with Kind, Verb, Time set and Args and Tags initialized.
func NewEvent(kind, verb string) Event {
	return Event{
		kind: kind,
		verb: verb,
		name: kind + "." + verb,

		Time: time.Now(),
		Args: make([]string, 0, 4),
		Tags: make(map[string]string),

		targetIds: make(map[Target]string),

		RenderTags: make(map[string]string),
	}
}

// Kind gets the event's kind
func (event *Event) Kind() string {
	return event.kind
}

// Verb gets the event's verb
func (event *Event) Verb() string {
	return event.verb
}

// Name gets the event name, which is Kind and Verb separated by a dot.
func (event *Event) Name() string {
	return event.name
}

// IsEither returns true if the event has the kind and one of the verbs.
func (event *Event) IsEither(kind string, verbs ...string) bool {
	if event.kind != kind {
		return false
	}

	for i := range verbs {
		if event.verb == verbs[i] {
			return true
		}
	}

	return false
}

// Context gets the event's context if it's part of the loop, or `context.Background` otherwise. client.Emit
// will set this context on its copy and return it.
func (event *Event) Context() context.Context {
	if event.ctx == nil {
		return context.Background()
	}

	return event.ctx
}

// Kill stops propagation of the event. The context will be killed once
// the current event handler returns.
func (event *Event) Kill() {
	event.killed = true
}

// Killed returns true if Kill has been called.
func (event *Event) Killed() bool {
	return event.killed
}

// Hide will not stop propagation, but it will allow output handlers to know not to
// render it.
func (event *Event) Hide() {
	event.hidden = true
}

// Hidden returns true if Hide has been called.
func (event *Event) Hidden() bool {
	return event.hidden
}

// Arg gets the argument by index. The rationale behind it is that some
// servers may use it for the last argument in JOINs and such.
func (event *Event) Arg(index int) string {
	if index < 0 || index > len(event.Args) {
		return ""
	}

	if index == len(event.Args) {
		return event.Text
	}

	return event.Args[index]
}

// MarshalJSON makes a JSON object from the event.
func (event *Event) MarshalJSON() ([]byte, error) {
	data := eventJSONData{
		Name:       event.Name(),
		Kind:       event.kind,
		Verb:       event.verb,
		Time:       event.Time,
		Nick:       event.Nick,
		User:       event.User,
		Host:       event.Host,
		Args:       event.Args,
		Text:       event.Text,
		Tags:       event.Tags,
		RenderTags: event.RenderTags,
	}

	data.Targets = make([]string, 0, len(event.targets))
	for _, target := range event.targets {
		data.Targets = append(data.Targets, event.targetIds[target])
	}

	return json.Marshal(data)
}

type eventJSONData struct {
	Name       string            `json:"name"`
	Kind       string            `json:"kind"`
	Verb       string            `json:"verb"`
	Time       time.Time         `json:"time"`
	Nick       string            `json:"nick"`
	User       string            `json:"user"`
	Host       string            `json:"host"`
	Args       []string          `json:"args"`
	Text       string            `json:"text"`
	Tags       map[string]string `json:"tags"`
	Targets    []string          `json:"targets"`
	RenderTags map[string]string `json:"renderTags"`
}
