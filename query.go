package irc

import (
	"github.com/gissleh/irc/list"
)

// A Query is a target for direct messages to and from a specific nick.
type Query struct {
	user list.User
}

// Kind returns "channel"
func (query *Query) Kind() string {
	return "query"
}

// Name gets the query name
func (query *Query) Name() string {
	return query.user.Nick
}

func (query *Query) State() ClientStateTarget {
	return ClientStateTarget{
		Kind:  "query",
		Name:  query.user.Nick,
		Users: []list.User{query.user},
	}
}

// AddHandler handles messages routed to this channel by the client's event loop
func (query *Query) Handle(event *Event, client *Client) {
	switch event.Name() {
	case "packet.nick":
		{
			query.user.Nick = event.Arg(0)
		}
	case "packet.account":
		{
			account := ""
			if accountArg := event.Arg(0); accountArg != "" && accountArg != "*" {
				account = accountArg
			}

			query.user.Account = account
		}
	case "packet.chghost":
		{
			query.user.User = event.Arg(0)
			query.user.Host = event.Arg(1)
		}
	}
}
