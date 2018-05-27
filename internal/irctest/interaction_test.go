package irctest_test

import (
	"net"
	"testing"

	"git.aiterp.net/gisle/irc/internal/irctest"
)

func TestInteraction(t *testing.T) {
	interaction := irctest.Interaction{
		Lines: []irctest.InteractionLine{
			{Kind: 'C', Data: "FIRST MESSAGE"},
			{Kind: 'S', Data: "SERVER MESSAGE"},
			{Kind: 'C', Data: "SECOND MESSAGE"},
		},
	}

	addr, err := interaction.Listen()
	if err != nil {
		t.Fatal("Listen:", err)
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal("Dial:", err)
	}

	_, err = conn.Write([]byte("FIRST MESSAGE\r\n"))
	if err != nil {
		t.Fatal("Write:", err)
	}

	buffer := make([]byte, 64)
	n, err := conn.Read(buffer)
	if err != nil {
		t.Fatal("Read:", err)
	}
	if string(buffer[:n]) != "SERVER MESSAGE\r\n" {
		t.Fatal("Read not correct:", string(buffer[:n]))
	}

	_, err = conn.Write([]byte("SECOND MESSAGE\r\n"))
	if err != nil {
		t.Fatal("Write 2:", err)
	}

	interaction.Wait()

	if interaction.Failure != nil {
		t.Error("Index:", interaction.Failure.Index)
		t.Error("Result:", interaction.Failure.Result)
		t.Error("NetErr:", interaction.Failure.NetErr)
		t.FailNow()
	}
}
