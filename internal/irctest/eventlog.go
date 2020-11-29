package irctest

import "github.com/gissleh/irc"

type EventLog struct {
	events []*irc.Event
}

func (l *EventLog) First(kind, verb string) *irc.Event {
	for _, e := range l.events {
		if e.Verb() == verb && e.Kind() == kind {
			return e
		}
	}

	return nil
}

func (l *EventLog) Last(kind, verb string) *irc.Event {
	for i := len(l.events) - 1; i >= 0; i-- {
		e := l.events[i]
		if e.Verb() == verb && e.Kind() == kind {
			return e
		}
	}

	return nil
}

func (l *EventLog) Handler(event *irc.Event, _ *irc.Client) {
	l.events = append(l.events, event)
}
