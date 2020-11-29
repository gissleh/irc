package irc_test

import (
	"github.com/gissleh/irc"
	"github.com/stretchr/testify/assert"
	"testing"
)

type packetTestRow struct {
	Data string
	Kind string
	Verb string
	Args []string
	Text string
	Tags map[string]string
}

var packetTestTable = []packetTestRow{
	{":test.server PING Test", "packet", "PING", []string{"Test"}, "", map[string]string{}},
	{":test.server PING :Test", "packet", "PING", []string{}, "Test", map[string]string{}},
	{":Test2!test@test.example.com PRIVMSG Tester :\x01ACTION hello to you.\x01", "ctcp", "ACTION", []string{"Tester"}, "hello to you.", map[string]string{}},
}

func TestParsePacket(t *testing.T) {
	for _, row := range packetTestTable {
		t.Run(row.Data, func(t *testing.T) {
			event, err := irc.ParsePacket(row.Data)
			if err != nil {
				t.Error("Parse Failed", err)
				return
			}

			assert.Equal(t, row.Kind, event.Kind(), "kind")
			assert.Equal(t, row.Verb, event.Verb(), "kind")
			assert.Equal(t, row.Args, event.Args, "kind")
			assert.Equal(t, row.Text, event.Text, "kind")
		})
	}
}
