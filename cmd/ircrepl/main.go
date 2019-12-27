package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gissleh/irc/handlers"
	"log"
	"os"
	"strings"

	"github.com/gissleh/irc"
)

var flagNick = flag.String("nick", "Test", "The client nick")
var flagAlts = flag.String("alts", "Test2,Test3,Test4,Test5", "Alternative nicks to use")
var flagUser = flag.String("user", "test", "The client user/ident")
var flagPass = flag.String("pass", "", "The server password")
var flagServer = flag.String("server", "localhost:6667", "The server to connect to")
var flagSsl = flag.Bool("ssl", false, "Wether to connect securely")

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	flag.Parse()

	irc.AddHandler(handlers.Input)
	irc.AddHandler(handlers.MRoleplay)

	client := irc.New(ctx, irc.Config{
		Nick:         *flagNick,
		User:         *flagUser,
		Alternatives: strings.Split(*flagAlts, ","),
		Password:     *flagPass,
	})

	err := client.Connect(*flagServer, *flagSsl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %s", err)
	}

	var target irc.Target
	irc.AddHandler(func(event *irc.Event, client *irc.Client) {
		if event.Name() == "input.target" {
			name := event.Arg(0)

			if client.ISupport().IsChannel(name) {
				log.Println("Set target channel", name)
				target = client.Channel(name)
			} else if len(name) > 0 {
				log.Println("Set target query", name)
				target = client.Query(name)
			} else {
				log.Println("Set target status")
				target = client.Status()
			}

			if target == nil {
				log.Println("Target does not exist, set to status")
				target = client.Status()
			}

			event.PreventDefault()
			return
		}

		if event.Name() == "input.clientstatus" {
			j, err := json.MarshalIndent(client.State(), "", "    ")
			if err != nil {
				return
			}

			fmt.Println(string(j))

			event.PreventDefault()
			return
		}

		j, err := json.MarshalIndent(event, "", "    ")
		if err != nil {
			return
		}

		fmt.Println(string(j))
	})

	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		client.EmitInput(string(line[:len(line)-1]), target)
	}
}
