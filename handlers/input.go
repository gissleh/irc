package handlers

import (
	"git.aiterp.net/gisle/irc"
	"git.aiterp.net/gisle/irc/ircutil"
)

// Input handles the default input.
func Input(event *irc.Event, client *irc.Client) {
	switch event.Name() {

	// /msg sends an action to a target specified before the message.
	case "input.msg":
		{
			event.Kill()

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
			event.Kill()

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
			event.Kill()

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
			}
		}

	// /describe sends an action to a target specified before the message, like /msg.
	case "input.describe":
		{
			event.Kill()

			targetName, text := ircutil.ParseArgAndText(event.Text)
			if targetName == "" || text == "" {
				client.EmitNonBlocking(irc.NewErrorEvent("input", "Usage: /describe <target> <text...>"))
				break
			}

			overhead := client.PrivmsgOverhead(targetName, true)
			cuts := ircutil.CutMessage(text, overhead)
			for _, cut := range cuts {
				client.SendCTCP("ACTION", targetName, false, cut)
			}
		}

	// /m is a shorthand for /mode that targets the current channel
	case "input.m":
		{
			event.Kill()

			if event.Text == "" {
				client.EmitNonBlocking(irc.NewErrorEvent("input", "Usage: /m <text...>"))
				break
			}

			channel := event.ChannelTarget()
			if channel == nil {
				client.EmitNonBlocking(irc.NewErrorEvent("input", "Target is not a channel"))
				break
			}

			client.SendQueuedf("MODE %s %s", channel.Name(), event.Text)
		}
	}
}
