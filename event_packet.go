package irc

import (
	"errors"
	"strings"
	"time"
)

var unescapeTags = strings.NewReplacer("\\\\", "\\", "\\:", ";", "\\s", " ", "\\r", "\r", "\\n", "\n")

// ParsePacket parses an irc line and returns an event that's either of kind `packet`, `ctcp` or `ctcpreply`
func ParsePacket(line string) (Event, error) {
	event := NewEvent("packet", "")
	event.Time = time.Now()

	if len(line) == 0 {
		return event, errors.New("irc: empty line")
	}

	// Parse tags
	if line[0] == '@' {
		split := strings.SplitN(line, " ", 2)
		if len(split) < 2 {
			return event, errors.New("irc: incomplete packet")
		}

		tagTokens := strings.Split(split[0][1:], ";")
		for _, token := range tagTokens {
			kv := strings.SplitN(token, "=", 2)

			if len(kv) == 2 {
				event.Tags[kv[0]] = unescapeTags.Replace(kv[1])
			} else {
				event.Tags[kv[0]] = ""
			}
		}

		line = split[1]
	}

	// Parse prefix
	if line[0] == ':' {
		split := strings.SplitN(line, " ", 2)
		if len(split) < 2 {
			return event, errors.New("ParsePacket: incomplete packet")
		}

		prefixTokens := strings.Split(split[0][1:], "!")

		event.Nick = prefixTokens[0]
		if len(prefixTokens) > 1 {
			userhost := strings.Split(prefixTokens[1], "@")

			if len(userhost) < 2 {
				return event, errors.New("ParsePacket: invalid user@host format")
			}

			event.User = userhost[0]
			event.Host = userhost[1]
		}

		line = split[1]
	}

	// Parse body
	split := strings.SplitN(line, " :", 2)
	tokens := strings.Split(split[0], " ")

	if len(split) == 2 {
		event.Text = split[1]
	}

	event.verb = tokens[0]
	event.Args = tokens[1:]

	// Parse CTCP
	if (event.verb == "PRIVMSG" || event.verb == "NOTICE") && strings.HasPrefix(event.Text, "\x01") {
		verbtext := strings.SplitN(strings.Replace(event.Text, "\x01", "", 2), " ", 2)

		if event.verb == "PRIVMSG" {
			event.kind = "ctcp"
		} else {
			event.kind = "ctcp-reply"
		}

		event.verb = verbtext[0]
		if len(verbtext) == 2 {
			event.Text = verbtext[1]
		} else {
			event.Text = ""
		}
	}

	event.name = event.kind + "." + strings.ToLower(event.verb)
	return event, nil
}
