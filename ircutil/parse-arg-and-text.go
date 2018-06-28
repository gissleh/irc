package ircutil

import (
	"strings"
)

// ParseArgAndText parses a text like "#Channel stuff and things" into "#Channel"
// and "stuff and things". This is commonly used for input commands which has
// no standard
func ParseArgAndText(s string) (arg, text string) {
	spaceIndex := strings.Index(s, " ")
	if spaceIndex == -1 {
		return s, ""
	}

	return s[:spaceIndex], s[spaceIndex+1:]
}
