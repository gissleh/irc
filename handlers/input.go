package handlers

import (
	"github.com/gissleh/irc"
	"github.com/gissleh/irc/ircutil"
	"time"
)

// Input handles the default input.
func Input(event *irc.Event, client *irc.Client) {
	switch event.Name() {

	// /msg sends an action to a target specified before the message.
	case "input.msg":
		{
			event.PreventDefault()

			targetName, text := ircutil.ParseArgAndText(event.Text)
			if targetName == "" || text == "" {
				client.EmitNonBlocking(irc.NewErrorEvent("input", "Usage: /msg <target> <text...>"))
				break
			}

			overhead := client.PrivmsgOverhead(targetName, true)
			cuts := ircutil.CutMessage(text, overhead)
			for _, cut := range cuts {
				client.Sendf("PRIVMSG %s :%s", targetName, cut)
			}
		}

	// /text (or text without a command) sends a message to the target.
	case "input.text":
		{
			event.PreventDefault()

			if event.Text == "" {
				client.EmitNonBlocking(irc.NewErrorEvent("input", "Usage: /text <text...>"))
				break
			}

			target := event.Target("query", "channel")
			if target == nil {
				client.EmitNonBlocking(irc.NewErrorEvent("input", "Target is not a channel or query"))
				break
			}

			overhead := client.PrivmsgOverhead(target.Name(), false)
			cuts := ircutil.CutMessage(event.Text, overhead)
			for _, cut := range cuts {
				client.SendQueuedf("PRIVMSG %s :%s", target.Name(), cut)
			}
		}

	// /me and /action sends a CTCP ACTION.
	case "input.me", "input.action":
		{
			event.PreventDefault()

			if event.Text == "" {
				client.EmitNonBlocking(irc.NewErrorEvent("input", "Usage: /me <text...>"))
				break
			}

			target := event.Target("query", "channel")
			if target == nil {
				client.EmitNonBlocking(irc.NewErrorEvent("input", "Target is not a channel or query"))
				break
			}

			overhead := client.PrivmsgOverhead(target.Name(), true)
			cuts := ircutil.CutMessage(event.Text, overhead)
			for _, cut := range cuts {
				client.SendCTCP("ACTION", target.Name(), false, cut)

				if !client.CapEnabled("echo-message") {
					event := irc.NewEvent("echo", "action")
					event.Time = time.Now()
					event.Nick = client.Nick()
					event.User = client.User()
					event.Host = client.Host()
					event.Args = []string{target.Name()}
					event.Text = cut

					client.EmitNonBlocking(event)
				}
			}
		}

	// /describe sends an action to a target specified before the message, like /msg.
	case "input.describe":
		{
			event.PreventDefault()

			targetName, text := ircutil.ParseArgAndText(event.Text)
			if targetName == "" || text == "" {
				client.EmitNonBlocking(irc.NewErrorEvent("input", "Usage: /describe <target> <text...>"))
				break
			}

			overhead := client.PrivmsgOverhead(targetName, true)
			cuts := ircutil.CutMessage(text, overhead)
			for _, cut := range cuts {
				client.SendCTCP("ACTION", targetName, false, cut)

				if !client.CapEnabled("echo-message") {
					event := irc.NewEvent("echo", "action")
					event.Time = time.Now()
					event.Nick = client.Nick()
					event.User = client.User()
					event.Host = client.Host()
					event.Args = []string{targetName}
					event.Text = cut

					client.EmitNonBlocking(event)
				}
			}
		}

	// /m is a shorthand for /mode that targets the current channel
	case "input.m":
		{
			event.PreventDefault()

			if event.Text == "" {
				client.EmitNonBlocking(irc.NewErrorEvent("input", "Usage: /m <modes and args...>"))
				break
			}

			if channel := event.ChannelTarget(); channel != nil {
				client.SendQueuedf("MODE %s %s", channel.Name(), event.Text)
			} else if status := event.StatusTarget(); status != nil {
				client.SendQueuedf("MODE %s %s", client.Nick(), event.Text)
			} else {
				client.EmitNonBlocking(irc.NewErrorEvent("input", "Target is not a channel or status"))
			}
		}

	case "input.quit", "input.disconnect":
		{
			event.PreventDefault()

			reason := event.Text
			if reason == "" {
				reason = "Client Quit"
			}

			client.Quit(reason)
		}
	}
}
