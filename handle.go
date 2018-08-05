package irc

// A Handler is a function that is part of the irc event loop. It will receive all
// events.
type Handler func(event *Event, client *Client)

var eventHandler struct {
	handlers []Handler
}

func emit(event *Event, client *Client) {
	for _, handler := range eventHandler.handlers {
		handler(event, client)
	}
}

// Handle adds a new handler to the irc handling. The handler may be called from multiple threads at the same
// time, so external resources should be locked if there are multiple clients. Adding handlers is not thread
// safe and should be done prior to clients being created.Handle. Also, this handler will block the individual
// client's event loop, so long operations that include network requests and the like should be done in a
// goroutine with the needed data **copied** from the handler function.
func Handle(handler Handler) {
	eventHandler.handlers = append(eventHandler.handlers, handler)
}

func init() {
	eventHandler.handlers = make([]Handler, 0, 8)
}
