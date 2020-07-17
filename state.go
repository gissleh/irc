package irc

import (
	"github.com/gissleh/irc/isupport"
	"github.com/gissleh/irc/list"
)

type ClientState struct {
	ID        string          `json:"id"`
	Nick      string          `json:"nick"`
	User      string          `json:"user"`
	Host      string          `json:"host"`
	Connected bool            `json:"connected"`
	Ready     bool            `json:"quit"`
	ISupport  *isupport.State `json:"isupport"`
	Caps      []string        `json:"caps"`
	Targets   []TargetState   `json:"targets"`
}

type TargetState struct {
	ID    string      `json:"id"`
	Kind  string      `json:"kind"`
	Name  string      `json:"name"`
	Users []list.User `json:"users,omitempty"`
}
