package irc

import "git.aiterp.net/gisle/irc/list"

// A Channel is a target that manages the userlist
type Channel struct {
	name     string
	userlist list.List
}

// Kind returns "channel"
func (channel *Channel) Kind() string {
	return "channel"
}

// Name gets the channel name
func (channel *Channel) Name() string {
	return channel.name
}

// UserList gets the channel userlist
func (channel *Channel) UserList() list.Immutable {
	return channel.userlist.Immutable()
}

// Handle handles messages routed to this channel by the client's event loop
func (channel *Channel) Handle(event *Event, client *Client) {
	switch event.Name() {
	case "packet.join":
		{
			// Support extended-join
			account := ""
			if accountArg := event.Arg(1); accountArg != "" && accountArg != "*" {
				account = accountArg
			}

			channel.userlist.Insert(list.User{
				Nick:    event.Nick,
				User:    event.User,
				Host:    event.Host,
				Account: account,
			})
		}
	case "packet.part", "packet.quit":
		{
			channel.userlist.Remove(event.Nick)
		}
	case "packet.account":
		{
			newAccount := event.Arg(0)

			if newAccount != "*" && newAccount != "" {
				channel.userlist.Patch(event.Nick, list.UserPatch{Account: newAccount})
			} else {
				channel.userlist.Patch(event.Nick, list.UserPatch{ClearAccount: true})
			}
		}
	case "packet.chghost":
		{
			newUser := event.Arg(0)
			newHost := event.Arg(1)

			channel.userlist.Patch(event.Nick, list.UserPatch{User: newUser, Host: newHost})
		}
	}
}
