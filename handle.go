package irc

import (
	"sync"
)

// A Handler is a function that is part of the irc event loop. It will receive all
// events that haven't been killed up to that point.
type Handler func(event *Event, client *Client)

var eventHandler struct {
	mutex    sync.RWMutex
	handlers []Handler
}

func emit(event *Event, client *Client) {
	eventHandler.mutex.RLock()
	for _, handler := range eventHandler.handlers {
		handler(event, client)
	}
	eventHandler.mutex.RUnlock()
}

// Handle adds a new handler to the irc handling. It returns a pointer that can be passed to RemoveHandler
// later on to unsubscribe.
func Handle(handler Handler) *Handler {
	eventHandler.mutex.Lock()
	defer eventHandler.mutex.Unlock()

	eventHandler.handlers = append(eventHandler.handlers, handler)
	return &eventHandler.handlers[len(eventHandler.handlers)-1]
}

// RemoveHandler unregisters a handler.
func RemoveHandler(handlerPtr *Handler) (ok bool) {
	eventHandler.mutex.Lock()
	defer eventHandler.mutex.Unlock()

	for i := range eventHandler.handlers {
		if &eventHandler.handlers[i] == handlerPtr {
			eventHandler.handlers = append(eventHandler.handlers[:i], eventHandler.handlers[i+1:]...)
			return true
		}
	}

	return false
}

func init() {
	eventHandler.handlers = make([]Handler, 0, 8)
}
