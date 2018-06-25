package irctest

import (
	"errors"
	"strings"
	"testing"

	"git.aiterp.net/gisle/irc"
)

// AssertUserlist compares the userlist to a list of prefixed nicks
func AssertUserlist(t *testing.T, channel *irc.Channel, assertedOrder ...string) error {
	users := channel.UserList().Users()
	order := make([]string, 0, len(users))
	for _, user := range users {
		order = append(order, user.PrefixedNick)
	}

	orderA := strings.Join(order, ", ")
	orderB := strings.Join(assertedOrder, ", ")

	if orderA != orderB {
		t.Logf("Userlist: %s", orderA)
		t.Logf("Asserted: %s", orderB)

		t.Fail()

		return errors.New("Userlists does not match")
	}

	return nil
}
