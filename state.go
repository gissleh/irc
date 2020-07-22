package irc

import (
	"github.com/gissleh/irc/isupport"
	"github.com/gissleh/irc/list"
)

// ClientState is a serializable snapshot of the client's state.
type ClientState struct {
	ID        string              `json:"id"`
	Nick      string              `json:"nick"`
	User      string              `json:"user"`
	Host      string              `json:"host"`
	Connected bool                `json:"connected"`
	Ready     bool                `json:"ready"`
	Quit      bool                `json:"quit"`
	ISupport  *isupport.State     `json:"isupport"`
	Caps      []string            `json:"caps"`
	Targets   []ClientStateTarget `json:"targets"`
}

// ClientStateTarget is a part of the ClientState representing a target's state at the time of snapshot.
type ClientStateTarget struct {
	ID    string      `json:"id"`
	Kind  string      `json:"kind"`
	Name  string      `json:"name"`
	Users []list.User `json:"users,omitempty"`
}
