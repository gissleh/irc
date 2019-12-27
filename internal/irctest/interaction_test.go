package irctest_test

import (
	"net"
	"testing"

	"github.com/gissleh/irc/internal/irctest"
)

func TestInteraction(t *testing.T) {
	interaction := irctest.Interaction{
		Lines: []irctest.InteractionLine{
			{Client: "FIRST MESSAGE"},
			{Server: "SERVER MESSAGE"},
			{Client: "SECOND MESSAGE"},
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
