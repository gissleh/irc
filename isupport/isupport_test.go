package isupport_test

import (
	"github.com/gissleh/irc/isupport"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

var isupportMessages = "FNC SAFELIST ELIST=CTU MONITOR=100 WHOX ETRACE KNOCK CHANTYPES=#& EXCEPTS INVEX CHANMODES=eIbq,k,flj,CFLNPQcgimnprstz CHANLIMIT=#&:15 PREFIX=(aovh)~@+% MAXLIST=bqeI:100 MODES=4 NETWORK=TestServer STATUSMSG=@+% CALLERID=g CASEMAPPING=rfc1459 NICKLEN=30 MAXNICKLEN=31 CHANNELLEN=50 TOPICLEN=390 DEAF=D TARGMAX=NAMES:1,LIST:1,KICK:1,WHOIS:1,PRIVMSG:4,NOTICE:4,ACCEPT:,MONITOR: EXTBAN=$,&acjmorsuxz| CLIENTVER=3.0"

var is isupport.ISupport

func init() {
	for _, token := range strings.Split(isupportMessages, " ") {
		pair := strings.SplitN(token, "=", 2)
		if len(pair) == 2 {
			is.Set(pair[0], pair[1])
		} else {
			is.Set(pair[0], "")
		}
	}
}

func TestISupport_ParsePrefixedNick(t *testing.T) {
	table := []struct {
		Full     string
		Prefixes string
		Modes    string
		Nick     string
	}{
		{"User", "", "", "User"},
		{"+User", "+", "v", "User"},
		{"@%+User", "@%+", "ohv", "User"},
		{"~User", "~", "a", "User"},
	}

	for _, row := range table {
		t.Run(row.Full, func(t *testing.T) {
			nick, modes, prefixes := is.ParsePrefixedNick(row.Full)

			assert.Equal(t, row.Nick, nick)
			assert.Equal(t, row.Modes, modes)
			assert.Equal(t, row.Prefixes, prefixes)
		})
	}
}

func TestISupport_IsChannel(t *testing.T) {
	table := map[string]bool{
		"#Test":        true,
		"&Test":        true,
		"User":         false,
		"+Stuff":       false,
		"#TestAndSuch": true,
		"@astrwef":     false,
	}

	for channelName, isChannel := range table {
		t.Run(channelName, func(t *testing.T) {
			assert.Equal(t, isChannel, is.IsChannel(channelName))
		})
	}
}

func TestISupport_IsPermissionMode(t *testing.T) {
	table := map[rune]bool{
		'#': false,
		'+': false,
		'o': true,
		'v': true,
		'h': true,
		'a': true,
		'g': false,
		'p': false,
	}

	for flag, expected := range table {
		t.Run(string(flag), func(t *testing.T) {
			assert.Equal(t, expected, is.IsPermissionMode(flag))
		})
	}
}
