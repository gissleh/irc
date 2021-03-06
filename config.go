package irc

import (
	"strconv"
)

// The Config for an IRC client.
type Config struct {
	// The nick that you go by. By default it's "IrcUser"
	Nick string `json:"nick"`

	// Alternatives are a list of nicks to try if Nick is occupied, in order of preference. By default
	// it's your nick with numbers 1 through 9.
	Alternatives []string `json:"alternatives"`

	// User is sent along with all messages and commonly shown before the @ on join, quit, etc....
	// Some servers tack on a ~ in front of it if you do not have an ident server.
	User string `json:"user"`

	// RealName is shown in WHOIS as your real name. By default "..."
	RealName string `json:"realName"`

	// SkipSSLVerification disables SSL certificate verification. Do not do this
	// in production.
	SkipSSLVerification bool `json:"skipSslVerification"`

	// The Password used upon connection. This is not your NickServ/SASL password!
	Password string `json:"password"`

	// The rate (lines per second) to send with Client.SendQueued. Default is 2, which is how
	// clients that don't excess flood does it.
	SendRate int `json:"sendRate"`

	// Languages to request.
	Languages []string `json:"languages"`

	// Auto-join on invite (bad idea).
	AutoJoinInvites bool `json:"autoJoinInvites"`

	// Whether to use the server time tag to overwrite event time.
	UseServerTime bool `json:"useServerTime"`

	// Use SASL authorization if supported.
	SASL *SASLConfig `json:"sasl"`
}

type SASLConfig struct {
	AuthenticationIdentity string `json:"authenticationIdentity"`
	AuthorizationIdentity  string `json:"authorizationIdentity"`
	Password               string `json:"password"`
}

// WithDefaults returns the config with the default values
func (config Config) WithDefaults() Config {
	if config.Nick == "" {
		config.Nick = "IrcUser"
	}
	if config.User == "" {
		config.User = "IrcUser"
	}
	if config.RealName == "" {
		config.RealName = "..."
	}

	if len(config.Alternatives) == 0 {
		config.Alternatives = make([]string, 9)
		for i := 0; i < 9; i++ {
			config.Alternatives[i] = config.Nick + strconv.FormatInt(int64(i+1), 10)
		}
	}

	if config.SendRate <= 0 {
		config.SendRate = 2
	}

	return config
}
