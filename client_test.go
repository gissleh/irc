package irc_test

import (
	"context"
	"errors"
	"testing"

	"git.aiterp.net/gisle/irc"
	"git.aiterp.net/gisle/irc/internal/irctest"
)

func TestClient(t *testing.T) {
	client := irc.New(context.Background(), irc.Config{
		Nick:         "Test",
		User:         "Tester",
		RealName:     "...",
		Alternatives: []string{"Test2", "Test3", "Test4", "Test768"},
	})

	t.Logf("Client.ID = %#+v", client.ID())
	if client.ID() == "" {
		t.Fail()
	}

	interaction := irctest.Interaction{
		Strict: false,
		Lines: []irctest.InteractionLine{
			{Kind: 'C', Data: "CAP LS 302"},
			{Kind: 'C', Data: "NICK Test"},
			{Kind: 'C', Data: "USER Tester 8 * :..."},
			{Kind: 'S', Data: ":testserver.example.com CAP * LS :multi-prefix userhost-in-names"},
			{Kind: 'C', Data: "CAP REQ :multi-prefix userhost-in-names"},
			{Kind: 'S', Data: ":testserver.example.com CAP * ACK :multi-prefix userhost-in-names"},
			{Kind: 'C', Data: "CAP END"},
			{Kind: 'S', Data: ":testserver.example.com 443 * Test :Nick is not available"},
			{Kind: 'C', Data: "NICK Test2"},
			{Kind: 'S', Data: ":testserver.example.com 443 * Test2 :Nick is not available"},
			{Kind: 'C', Data: "NICK Test3"},
			{Kind: 'S', Data: ":testserver.example.com 443 * Test3 :Nick is not available"},
			{Kind: 'C', Data: "NICK Test4"},
			{Kind: 'S', Data: ":testserver.example.com 443 * Test4 :Nick is not available"},
			{Kind: 'C', Data: "NICK Test768"},
			{Kind: 'S', Data: ":testserver.example.com 001 Test768 :Welcome to the TestServer Internet Relay Chat Network test"},
			{Kind: 'C', Data: "WHO Test768*"},
			{Kind: 'S', Data: ":testserver.example.com 002 Test768 :Your host is testserver.example.com[testserver.example.com/6667], running version charybdis-4-rc3"},
			{Kind: 'S', Data: ":testserver.example.com 003 Test768 :This server was created Fri Nov 25 2016 at 17:28:20 CET"},
			{Kind: 'S', Data: ":testserver.example.com 004 Test768 testserver.example.com charybdis-4-rc3 DQRSZagiloswxz CFILNPQbcefgijklmnopqrstvz bkloveqjfI"},
			{Kind: 'S', Data: ":testserver.example.com 005 Test768 FNC SAFELIST ELIST=CTU MONITOR=100 WHOX ETRACE KNOCK CHANTYPES=#& EXCEPTS INVEX CHANMODES=eIbq,k,flj,CFLNPQcgimnprstz CHANLIMIT=#&:15 :are supported by this server"},
			{Kind: 'S', Data: ":testserver.example.com 005 Test768 PREFIX=(ov)@+ MAXLIST=bqeI:100 MODES=4 NETWORK=TestServer STATUSMSG=@+ CALLERID=g CASEMAPPING=rfc1459 NICKLEN=30 MAXNICKLEN=31 CHANNELLEN=50 TOPICLEN=390 DEAF=D :are supported by this server"},
			{Kind: 'S', Data: ":testserver.example.com 005 Test768 TARGMAX=NAMES:1,LIST:1,KICK:1,WHOIS:1,PRIVMSG:4,NOTICE:4,ACCEPT:,MONITOR: EXTBAN=$,&acjmorsuxz| CLIENTVER=3.0 :are supported by this server"},
			{Kind: 'S', Data: ":testserver.example.com 251 Test768 :There are 0 users and 2 invisible on 1 servers"},
			{Kind: 'S', Data: ":testserver.example.com 254 Test768 1 :channels formed"},
			{Kind: 'S', Data: ":testserver.example.com 255 Test768 :I have 2 clients and 0 servers"},
			{Kind: 'S', Data: ":testserver.example.com 265 Test768 2 2 :Current local users 2, max 2"},
			{Kind: 'S', Data: ":testserver.example.com 266 Test768 2 2 :Current global users 2, max 2"},
			{Kind: 'S', Data: ":testserver.example.com 250 Test768 :Highest connection count: 2 (2 clients) (8 connections received)"},
			{Kind: 'S', Data: ":testserver.example.com 352 Test768 * ~Tester testclient.example.com testserver.example.com Test768 H :0 ..."},
			{Kind: 'S', Data: ":testserver.example.com 375 Test768 :- testserver.example.com Message of the Day - "},
			{Kind: 'S', Data: ":testserver.example.com 372 Test768 :- This server is only for testing irce, not chatting. If you happen"},
			{Kind: 'S', Data: ":testserver.example.com 372 Test768 :- to connect to it by accident, please disconnect immediately."},
			{Kind: 'S', Data: ":testserver.example.com 372 Test768 :-  "},
			{Kind: 'S', Data: ":testserver.example.com 372 Test768 :-  - #Test  :: Test Channel"},
			{Kind: 'S', Data: ":testserver.example.com 372 Test768 :-  - #Test2 :: Other Test Channel"},
			{Kind: 'S', Data: ":testserver.example.com 376 Test768 :End of /MOTD command."},
			{Kind: 'S', Data: ":Test768 MODE Test768 :+i"},
			{Kind: 'C', Data: "JOIN #Test"},
			{Kind: 'S', Data: ":Test768!~test@127.0.0.1 JOIN #Test *"},
			{Kind: 'S', Data: ":testserver.example.com 353 Test768 = #Test :Test768!~test@127.0.0.1 @+Gisle!gisle@gisle.me"},
			{Kind: 'S', Data: ":testserver.example.com 366 Test768 #Test :End of /NAMES list."},
			{Kind: 'S', Data: ":Gisle!~irce@10.32.0.1 MODE #Test +osv Test768 Test768"},
			{Kind: 'S', Data: ":Gisle!~irce@10.32.0.1 MODE #Test +N-s "},
			{Kind: 'S', Data: ":Test1234!~test2@172.17.37.1 JOIN #Test Test1234"},
			{Kind: 'S', Data: ":Gisle!~irce@10.32.0.1 MODE #Test +v Test1234"},
			{Kind: 'S', Data: "PING :archgisle.lan"}, // Ping/Pong to sync.
			{Kind: 'C', Data: "PONG :archgisle.lan"},
			{Callback: func() error {
				channel := client.Channel("#Test")
				if channel == nil {
					return errors.New("Channel #Test not found")
				}

				err := irctest.AssertUserlist(t, channel, "@Gisle", "@Test768", "+Test1234")
				if err != nil {
					return err
				}

				userTest1234, ok := channel.UserList().User("Test1234")
				if !ok {
					return errors.New("Test1234 not found")
				}
				if userTest1234.Account != "Test1234" {
					return errors.New("Test1234 did not get account from extended-join")
				}

				return nil
			}},
			{Kind: 'S', Data: ":Test1234!~test2@172.17.37.1 NICK Hunter2"},
			{Kind: 'S', Data: ":Hunter2!~test2@172.17.37.1 AWAY :Doing stuff"},
			{Kind: 'S', Data: ":Gisle!~irce@10.32.0.1 AWAY"},
			{Kind: 'S', Data: ":Gisle!~irce@10.32.0.1 PART #Test :Leaving the channel"},
			{Kind: 'S', Data: "PING :archgisle.lan"}, // Ping/Pong to sync.
			{Kind: 'C', Data: "PONG :archgisle.lan"},
			{Callback: func() error {
				channel := client.Channel("#Test")
				if channel == nil {
					return errors.New("Channel #Test not found")
				}

				err := irctest.AssertUserlist(t, channel, "@Test768", "+Hunter2")
				if err != nil {
					return err
				}

				_, ok := channel.UserList().User("Test1234")
				if ok {
					return errors.New("Test1234 is still there")
				}

				userHunter2, ok := channel.UserList().User("Hunter2")
				if !ok {
					return errors.New("Test1234 not found")
				}
				if userHunter2.Account != "Test1234" {
					return errors.New("Hunter2 did not persist account post nick change")
				}
				if !userHunter2.IsAway() {
					return errors.New("Hunter2 should be away")
				}
				if userHunter2.Away != "Doing stuff" {
					return errors.New("Hunter2 has the wrong away message: " + userHunter2.Away)
				}

				return nil
			}},
		},
	}

	addr, err := interaction.Listen()
	if err != nil {
		t.Fatal("Listen:", err)
	}

	irc.Handle(func(event *irc.Event, client *irc.Client) {
		if event.Name() == "packet.376" {
			client.SendQueued("JOIN #Test")
		}
	})

	err = client.Connect(addr, false)
	if err != nil {
		t.Fatal("Connect:", err)
		return
	}

	interaction.Wait()

	fail := interaction.Failure
	if fail != nil {
		t.Error("Index:", fail.Index)
		t.Error("NetErr:", fail.NetErr)
		t.Error("CBErr:", fail.CBErr)
		t.Error("Result:", fail.Result)
		if fail.Index >= 0 {
			t.Error("Line.Kind:", interaction.Lines[fail.Index].Kind)
			t.Error("Line.Data:", interaction.Lines[fail.Index].Data)
		}
	}

	if client.Nick() != "Test768" {
		t.Errorf("Nick: %#+v != %#+v (Expectation)", client.Nick(), "Test768")
	}
	if client.User() != "~Tester" {
		t.Errorf("User: %#+v != %#+v (Expectation)", client.User(), "~Tester")
	}
	if client.Host() != "testclient.example.com" {
		t.Errorf("Host: %#+v != %#+v (Expectation)", client.Host(), "testclient.example.com")
	}

	for i, logLine := range interaction.Log {
		t.Logf("Log[%d] = %#+v", i, logLine)
	}
}
