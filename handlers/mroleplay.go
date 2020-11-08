package handlers

import (
	"strings"

	"github.com/gissleh/irc"
	"github.com/gissleh/irc/ircutil"
)

// MRoleplay is a handler that adds commands for cutting NPC commands, as well as cleaning up
// the input from the server. It's named after Charybdis IRCd's m_roleplay module.
func MRoleplay(event *irc.Event, client *irc.Client) {
	switch event.Name() {
	case "input.enablerp", "input.disablerp":
		{
			sign := "+"
			if event.Verb() == "disablerp" {
				sign = "-"
			}

			// If the target is a channel, use RPCHAN or, if not stated, N.
			chanMode, chanModeOk := client.ISupport().Get("RPCHAN")
			channel := event.ChannelTarget()
			if channel != nil {
				if chanModeOk {
					client.SendQueuedf("MODE %s %s%s", channel.Name(), sign, chanMode)
				} else {
					client.SendQueuedf("MODE %s %sN", channel.Name(), sign)
				}
			}

			// Otherwise enable it on yourself, but only if RPUSER is set as that is not supported.
			// by servers without this ISupport tag.
			userMode, userModeOk := client.ISupport().Get("RPUSER")
			query := event.QueryTarget()
			status := event.StatusTarget()
			if (query != nil || status != nil) && userModeOk {
				client.SendQueuedf("MODE %s%s", sign, userMode)
			}
		}

	// Parse roleplaying messages, and replace underscored-nick with a render tag.
	case "packet.privmsg", "ctcp.action":
		{
			// Detect m_roleplay
			if strings.HasPrefix(event.Nick, "\x1F") {
				event.Nick = event.Nick[1 : len(event.Nick)-2]
				if event.Verb() == "PRIVMSG" {
					event.RenderTags["mRoleplay"] = "npc"
				} else {
					event.RenderTags["mRoleplay"] = "npca"
				}
			} else if strings.HasPrefix(event.Nick, "=") {
				event.RenderTags["mRoleplay"] = "scene"
			} else {
				break
			}

			// Some servers use this.
			lastSpace := strings.LastIndex(event.Text, " ")
			lastParentheses := strings.LastIndex(event.Text, "(")
			if lastParentheses != -1 && lastSpace != -1 && lastParentheses == lastSpace+1 {
				event.Text = event.Text[:lastSpace]
			}
		}

	// NPC commands
	case "input.npcc", "input.npcac":
		{
			isAction := event.Verb() == "npcac"
			nick, text := ircutil.ParseArgAndText(event.Text)
			if nick == "" || text == "" {
				client.EmitNonBlocking(irc.NewErrorEvent("input", "Usage: /"+event.Verb()+" <nick> <text...>"))
				break
			}

			channel := event.ChannelTarget()
			if channel == nil {
				client.EmitNonBlocking(irc.NewErrorEvent("input", "Target is not a channel"))
				break
			}

			overhead := ircutil.MessageOverhead("\x1f"+nick+"\x1f", client.Nick(), "npc.fakeuser.invalid", channel.Name(), isAction)
			cuts := ircutil.CutMessage(text, overhead)

			for _, cut := range cuts {
				npcCommand := "NPC"
				if isAction {
					npcCommand = "NPCA"
				}

				client.SendQueuedf("%s %s %s :%s", npcCommand, channel.Name(), nick, cut)
			}

			event.PreventDefault()
		}

	// Scene/narrator command
	case "input.scenec", "input.narratorc":
		{
			if event.Text == "" {
				client.EmitNonBlocking(irc.NewErrorEvent("input", "Usage: /"+event.Verb()+" <text...>"))
				break
			}

			channel := event.ChannelTarget()
			if channel == nil {
				client.EmitNonBlocking(irc.NewErrorEvent("input", "Target is not a channel"))
				break
			}

			overhead := ircutil.MessageOverhead("=Scene=", client.Nick(), "npc.fakeuser.invalid", channel.Name(), false)
			cuts := ircutil.CutMessage(event.Text, overhead)
			for _, cut := range cuts {
				client.SendQueuedf("SCENE %s :%s", channel.Name(), cut)
			}

			event.PreventDefault()
		}
	}
}
