package ircutil

import (
	"bytes"
	"unicode/utf8"
)

// MessageOverhead calculates the overhead in a `PRIVMSG` sent by a client
// with the given nick, user, host and target name. A `NOTICE` is shorter, so
// it is safe to use the same function for it.
func MessageOverhead(nick, user, host, target string, action bool) int {
	template := ":!@ PRIVMSG  :"
	if action {
		template += "\x01ACTION \x01"
	}

	return len(template) + len(nick) + len(user) + len(host) + len(target)
}

// CutMessage returns cuts of the message with the given overhead. If there
// there are tokens longer than the cutLength, it will call CutMessageNoSpace
// instead.
func CutMessage(text string, overhead int) []string {
	tokens := bytes.Split([]byte(text), []byte{' '})
	cutLength := 510 - overhead
	for _, token := range tokens {
		if len(token) >= cutLength {
			return CutMessageNoSpace(text, overhead)
		}
	}

	result := make([]string, 0, (len(text)/(cutLength))+1)
	current := make([]byte, 0, cutLength)
	for _, token := range tokens {
		if (len(current) + 1 + len(token)) > cutLength {
			result = append(result, string(current))
			current = current[:0]
		}

		if len(current) > 0 {
			current = append(current, ' ')
		}
		current = append(current, token...)
	}

	return append(result, string(current))
}

// CutMessageNoSpace cuts the messages per utf-8 rune.
func CutMessageNoSpace(text string, overhead int) []string {
	cutLength := 510 - overhead
	result := make([]string, 0, (len(text)/(cutLength))+1)
	current := ""

	for _, r := range text {
		if len(current)+utf8.RuneLen(r) > cutLength {
			result = append(result, current)
			current = ""
		}

		current += string(r)
	}

	return append(result, current)
}
