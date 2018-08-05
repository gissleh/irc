package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"git.aiterp.net/gisle/irc"
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

	irc.Handle(func(event *irc.Event, client *irc.Client) {
		json, err := json.MarshalIndent(event, "", "    ")
		if err != nil {
			return
		}

		fmt.Println(string(json))
	})

	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		client.EmitInput(string(line[:len(line)-1]), client.Status())
	}
}
