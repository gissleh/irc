package handlers

import (
	"strings"

	"git.aiterp.net/gisle/irc"
	"git.aiterp.net/gisle/irc/ircutil"
)

// MRoleplay is a handler that adds commands for cutting NPC commands, as well as cleaning up
// the input from the server. It's named after Charybdis IRCd's m_roleplay module.
func MRoleplay(event *irc.Event, client *irc.Client) {
	switch event.Name() {
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

			lastSpace := strings.LastIndex(event.Text, " ")
			lastParanthesis := strings.LastIndex(event.Text, "(")
			if lastParanthesis != -1 && lastSpace != -1 && lastParanthesis == lastSpace+1 {
				event.Text = event.Text[:lastSpace]
			}
		}
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
			cuts := ircutil.CutMessage(event.Text, overhead)

			for _, cut := range cuts {
				npcCommand := "NPCA"
				if event.Verb() == "npcc" {
					npcCommand = "NPC"
				}

				client.SendQueuedf("%s %s :%s", npcCommand, channel.Name(), cut)
			}

			event.Kill()
		}
	case "input.scenec":
		{
			if event.Text == "" {
				client.EmitNonBlocking(irc.NewErrorEvent("input", "Usage: /scenec <text...>"))
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

			event.Kill()
		}
	}
}
