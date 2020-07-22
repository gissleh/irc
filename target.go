package irc

// A Target is a handler for a message meant for a limited part of the client, like a channel or
// query
type Target interface {
	Kind() string
	Name() string
	Handle(event *Event, client *Client)
	State() ClientStateTarget
}
