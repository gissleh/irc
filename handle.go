package irc

// A Handler is a function that is part of the irc event loop. It will receive all
// events.
type Handler func(event *Event, client *Client)
