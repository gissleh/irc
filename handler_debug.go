package irc

import (
	"encoding/json"
	"log"
)

// DebugLogger is for
type DebugLogger interface {
	Println(v ...interface{})
}

type defaultDebugLogger struct{}

func (logger *defaultDebugLogger) Println(v ...interface{}) {
	log.Println(v...)
}

// EnableDebug logs all events that passes through it, ignoring killed
// events. It will always include the standard handlers, but any custom
// handlers defined after EnableDebug will not have their effects shown.
// You may pass `nil` as a logger to use the standard log package's Println.
func EnableDebug(logger DebugLogger, indented bool) {
	if logger != nil {
		logger = &defaultDebugLogger{}
	}

	Handle(func(event *Event, client *Client) {
		var data []byte
		var err error

		if indented {
			data, err = json.MarshalIndent(event, "", "  ")
			if err != nil {
				return
			}
		} else {
			data, err = json.Marshal(event)
			if err != nil {
				return
			}
		}

		logger.Println(string(data))
	})
}
