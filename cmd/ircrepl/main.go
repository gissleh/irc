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
	"os/signal"
	"strings"
	"syscall"

	"github.com/gissleh/irc"
)

var flagNick = flag.String("nick", "Test", "The client nick")
var flagAlts = flag.String("alts", "Test2,Test3,Test4,Test5", "Alternative nicks to use")
var flagUser = flag.String("user", "test", "The client user/ident")
var flagPass = flag.String("pass", "", "The server password")
var flagServer = flag.String("server", "localhost:6667", "The server to connect to")
var flagSsl = flag.Bool("ssl", false, "Whether to connect securely")
var flagSkipVerify = flag.Bool("skip-verify", false, "Skip SSL verification")

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	flag.Parse()

	client := irc.New(ctx, irc.Config{
		Nick:                *flagNick,
		User:                *flagUser,
		Alternatives:        strings.Split(*flagAlts, ","),
		Password:            *flagPass,
		Languages:           []string{"no_NB", "no", "en_US", "en"},
		SkipSSLVerification: *flagSkipVerify,
	})

	client.AddHandler(handlers.Input)
	client.AddHandler(handlers.MRoleplay)

	err := client.Connect(*flagServer, *flagSsl)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to connect: %s", err)
		os.Exit(1)
	}

	var target irc.Target
	client.AddHandler(func(event *irc.Event, client *irc.Client) {
		if event.Name() == "input.target" {
			name := event.Arg(0)

			if client.ISupport().IsChannel(name) {
				log.Println("Set target channel", name)
				target = client.Target("target", name)
			} else if len(name) > 0 {
				log.Println("Set target query", name)
				target = client.Target("query", name)
			} else {
				log.Println("Set target status")
				target = client.Target("status", "status")
			}

			if target == nil {
				log.Println("Target does not exist, set to status")
				target = client.Target("status", "status")
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

		if event.Name() == "hook.remove_target" {
			if target != nil && target.Name() == event.Arg(2) && target.Kind() == event.Arg(1) {
				log.Println("Unset target ", event.Arg(1), event.Arg(2))
				target = nil
			}
		}

		if event.Name() == "hook.add_target" {
			log.Println("Set target ", event.Arg(1), event.Arg(2))
			target = client.Target(event.Arg(1), event.Arg(2))
		}

		if event.Name() == "client.disconnect" {
			os.Exit(0)
		}

		j, err := json.MarshalIndent(event, "", "    ")
		if err != nil {
			return
		}

		fmt.Println(string(j))
	})

	go func() {
		exitSignal := make(chan os.Signal)
		signal.Notify(exitSignal, os.Interrupt, os.Kill, syscall.SIGTERM)

		<-exitSignal

		client.Quit("Goodnight.")
	}()

	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		client.EmitInput(line[:len(line)-1], target)
	}
}
