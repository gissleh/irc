package handlers

import (
	"github.com/gissleh/irc"
	"strconv"
	"strings"
	"time"
)

// CTCP implements the widely used CTCP commands (CLIENTINFO, VERSION, TIME, and PING), as well as the /ping command.
// It does not implement DCC.
//
// For every other CTCP command supported, you should expand the `ctcp.clientinfo.reply` client value like above.
func CTCP(event *irc.Event, client *irc.Client) {
	switch event.Name() {
	case "client.create":
		if r, ok := client.Value("ctcp.clientinfo.reply").(string); ok {
			if !strings.Contains(r, "ACTION PING TIME VERSION") {
				client.SetValue("ctcp.clientinfo.reply", r+" ACTION PING TIME VERSION")
			}
		} else {
			client.SetValue("ctcp.clientinfo.reply", "ACTION PING TIME VERSION")
		}
	case "ctcp.clientinfo":
		{
			response, ok := client.Value("ctcp.clientinfo.reply").(string)
			if !ok {
				response = "ACTION PING TIME VERSION"
			}

			client.SendCTCP("CLIENTINFO", event.Nick, true, response)
		}
	case "ctcp.version":
		{
			version := "github.com/gissleh/irc v1.0"
			if v, ok := client.Value("ctcp.version.reply").(string); ok {
				version = v
			}

			client.SendCTCP("VERSION", event.Nick, true, version)
		}
	case "ctcp.time":
		{
			client.SendCTCP("TIME", event.Nick, true, time.Now().Local().Format(time.RFC1123))
		}
	case "ctcp.ping":
		{
			client.SendCTCP("PING", event.Nick, true, event.Text)
		}
	case "input.ping":
		{
			args := strings.SplitN(event.Text, " ", 2)
			targetName := args[0]
			if targetName == "" {
				client.EmitNonBlocking(irc.NewErrorEvent("ctcp.pingarg", "/ping needs an argument"))
				break
			}

			client.SendCTCP("PING", targetName, false, strconv.FormatInt(time.Now().UnixNano()/1000000, 10))
		}
	}
}
