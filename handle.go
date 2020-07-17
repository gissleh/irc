package irc

// A Handler is a function that is part of the irc event loop. It will receive all
// events.
type Handler func(event *Event, client *Client)

var globalHandlers = make([]Handler, 0, 8)

// AddHandler adds a new handler to the irc handling. The handler may be called from multiple threads at the same
// time, so external resources should be locked if there are multiple clients. Adding handlers is not thread
// safe and should be done prior to clients being created. Also, this handler will block the individual
// client's event loop, so long operations that include network requests and the like should be done in a
// goroutine with the needed data **copied** from the handler function.
func AddHandler(handler Handler) {
	globalHandlers = append(globalHandlers, handler)
}

func init() {
	globalHandlers = make([]Handler, 0, 8)
}
