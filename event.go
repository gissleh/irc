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

// MarshalJSON makes a JSON object from the event.
func (event *Event) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"kind":   event.kind,
		"verb":   event.verb,
		"text":   event.Text,
		"args":   event.Args,
		"tags":   event.Tags,
		"killed": event.killed,
		"hidden": event.hidden,
	})
}
