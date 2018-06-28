package irc

import (
	"strings"
	"time"
)

// ParseInput parses an input command into an event.
func ParseInput(line string) Event {
	event := NewEvent("input", "")
	event.Time = time.Now()

	if strings.HasPrefix(line, "/") {
		split := strings.SplitN(line[1:], " ", 2)
		event.verb = strings.ToLower(split[0])
		if len(split) == 2 {
			event.Text = split[1]
		}
	} else {
		event.Text = line
		event.verb = "text"
	}

	event.name = event.kind + "." + event.verb

	return event
}
