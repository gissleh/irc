package irc_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"git.aiterp.net/gisle/irc"
	"git.aiterp.net/gisle/irc/handlers"
	"git.aiterp.net/gisle/irc/internal/irctest"
)

// Integration test below, brace yourself.
func TestClient(t *testing.T) {
	irc.Handle(handlers.Input)
	irc.Handle(handlers.MRoleplay)

	client := irc.New(context.Background(), irc.Config{
		Nick:         "Test",
		User:         "Tester",
		RealName:     "...",
		Alternatives: []string{"Test2", "Test3", "Test4", "Test768"},
		SendRate:     1000,
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
			{Kind: 'S', Data: ":testserver.example.com CAP * LS :multi-prefix chghost userhost-in-names vendorname/custom-stuff echo-message =malformed vendorname/advanced-custom-stuff=things,and,items"},
			{Kind: 'C', Data: "CAP REQ :multi-prefix chghost userhost-in-names"},
			{Kind: 'S', Data: ":testserver.example.com CAP * ACK :multi-prefix userhost-in-names"},
			{Kind: 'C', Data: "CAP END"},
			{Callback: func() error {
				if !client.CapEnabled("multi-prefix") {
					return errors.New("multi-prefix cap should be enabled.")
				}
				if !client.CapEnabled("userhost-in-names") {
					return errors.New("userhost-in-names cap should be enabled.")
				}
				if client.CapEnabled("echo-message") {
					return errors.New("echo-message cap should not be enabled.")
				}
				if client.CapEnabled("") {
					return errors.New("(blank) cap should be enabled.")
				}

				return nil
			}},
			{Kind: 'S', Data: ":testserver.example.com 433 * Test :Nick is not available"},
			{Kind: 'C', Data: "NICK Test2"},
			{Kind: 'S', Data: ":testserver.example.com 433 * Test2 :Nick is not available"},
			{Kind: 'C', Data: "NICK Test3"},
			{Kind: 'S', Data: ":testserver.example.com 433 * Test3 :Nick is not available"},
			{Kind: 'C', Data: "NICK Test4"},
			{Kind: 'S', Data: ":testserver.example.com 433 * Test4 :Nick is not available"},
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
			{Kind: 'S', Data: ":testserver.example.com 375 Test768 :- testserver.example.com Message of the Day - "},
			{Kind: 'S', Data: ":testserver.example.com 372 Test768 :- This server is only for testing irce, not chatting. If you happen"},
			{Kind: 'S', Data: ":testserver.example.com 372 Test768 :- to connect to it by accident, please disconnect immediately."},
			{Kind: 'S', Data: ":testserver.example.com 372 Test768 :-  "},
			{Kind: 'S', Data: ":testserver.example.com 372 Test768 :-  - #Test  :: Test Channel"},
			{Kind: 'S', Data: ":testserver.example.com 372 Test768 :-  - #Test2 :: Other Test Channel"},
			{Kind: 'S', Data: ":testserver.example.com 376 Test768 :End of /MOTD command."},
			{Kind: 'S', Data: ":testserver.example.com 352 Test768 * ~Tester testclient.example.com testserver.example.com Test768 H :0 ..."},
			{Kind: 'S', Data: ":Test768 MODE Test768 :+i"},
			{Kind: 'S', Data: "PING :testserver.example.com"}, // Ping/Pong to sync.
			{Kind: 'C', Data: "PONG :testserver.example.com"},
			{Callback: func() error {
				if client.Nick() != "Test768" {
					return errors.New("client.Nick shouldn't be " + client.Nick())
				}
				if client.User() != "~Tester" {
					return errors.New("client.User shouldn't be " + client.User())
				}
				if client.Host() != "testclient.example.com" {
					return errors.New("client.Host shouldn't be " + client.Host())
				}

				return nil
			}},
			{Callback: func() error {
				err := client.Join("#Test")
				if err != nil {
					return fmt.Errorf("Failed to join #Test: %s", err)
				}

				return nil
			}},
			{Kind: 'C', Data: "JOIN #Test"},
			{Kind: 'S', Data: ":Test768!~test@127.0.0.1 JOIN #Test *"},
			{Kind: 'S', Data: ":testserver.example.com 353 Test768 = #Test :Test768!~test@127.0.0.1 @+Gisle!gisle@gisle.me"},
			{Kind: 'S', Data: ":testserver.example.com 366 Test768 #Test :End of /NAMES list."},
			{Kind: 'S', Data: ":Gisle!~irce@10.32.0.1 MODE #Test +osv Test768 Test768"},
			{Kind: 'S', Data: ":Gisle!~irce@10.32.0.1 MODE #Test +N-s "},
			{Kind: 'S', Data: ":Test1234!~test2@172.17.37.1 JOIN #Test Test1234"},
			{Kind: 'S', Data: ":Test4321!~test2@172.17.37.1 JOIN #Test Test1234"},
			{Kind: 'S', Data: ":Gisle!~irce@10.32.0.1 MODE #Test +v Test1234"},
			{Kind: 'S', Data: "PING :testserver.example.com"}, // Ping/Pong to sync.
			{Kind: 'C', Data: "PONG :testserver.example.com"},
			{Callback: func() error {
				channel := client.Channel("#Test")
				if channel == nil {
					return errors.New("Channel #Test not found")
				}

				err := irctest.AssertUserlist(t, channel, "@Gisle", "@Test768", "+Test1234", "Test4321")
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
			{Kind: 'S', Data: ":Hunter2!~test2@172.17.37.1 CHGHOST test2 some.awesome.virtual.host"},
			{Kind: 'S', Data: "@account=Hunter2 :Test4321!~test2@172.17.37.1 PRIVMSG #Test :Hello World."},
			{Kind: 'S', Data: "PING :testserver.example.com"}, // Ping/Pong to sync.
			{Kind: 'C', Data: "PONG :testserver.example.com"},
			{Callback: func() error {
				channel := client.Channel("#Test")
				if channel == nil {
					return errors.New("Channel #Test not found")
				}

				err := irctest.AssertUserlist(t, channel, "@Test768", "+Hunter2", "Test4321")
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
				if userHunter2.Host != "some.awesome.virtual.host" {
					return errors.New("Hunter2 should have changed the host: " + userHunter2.Host)
				}

				return nil
			}},
			{Kind: 'S', Data: ":Hunter2!~test2@172.17.37.1 PRIVMSG Test768 :Hello, World"},
			{Kind: 'S', Data: "PING :testserver.example.com"}, // Ping/Pong to sync.
			{Kind: 'C', Data: "PONG :testserver.example.com"},
			{Callback: func() error {
				query := client.Query("Hunter2")
				if query == nil {
					return errors.New("Did not find query")
				}

				return nil
			}},
			{Kind: 'S', Data: ":Hunter2!~test2@172.17.37.1 NICK SevenAsterisks"},
			{Kind: 'S', Data: "PING :testserver.example.com"}, // Ping/Pong to sync.
			{Kind: 'C', Data: "PONG :testserver.example.com"},
			{Callback: func() error {
				oldQuerry := client.Query("Hunter2")
				if oldQuerry != nil {
					return errors.New("Did find query by old name")
				}

				query := client.Query("SevenAsterisks")
				if query == nil {
					return errors.New("Did not find query by new name")
				}

				return nil
			}},
			{Callback: func() error {
				client.EmitInput("/invalidcommand stuff and things", nil)
				return nil
			}},
			{Kind: 'C', Data: "INVALIDCOMMAND stuff and things"},
			{Kind: 'S', Data: ":testserver.example.com 421 Test768 INVALIDCOMMAND :Unknown command"},
			{Callback: func() error {
				channel := client.Channel("#Test")
				if channel == nil {
					return errors.New("Channel #Test not found")
				}

				client.EmitInput("/me does stuff", channel)
				client.EmitInput("/describe #Test describes stuff", channel)
				client.EmitInput("/text Hello, World", channel)
				client.EmitInput("Hello again", channel)
				return nil
			}},
			{Kind: 'C', Data: "PRIVMSG #Test :\x01ACTION does stuff\x01"},
			{Kind: 'C', Data: "PRIVMSG #Test :\x01ACTION describes stuff\x01"},
			{Kind: 'C', Data: "PRIVMSG #Test :Hello, World"},
			{Kind: 'C', Data: "PRIVMSG #Test :Hello again"},
			{Kind: 'S', Data: ":Test768!~test@127.0.0.1 PRIVMSG #Test :\x01ACTION does stuff\x01"},
			{Kind: 'S', Data: ":Test768!~test@127.0.0.1 PRIVMSG #Test :\x01ACTION describes stuff\x01"},
			{Kind: 'S', Data: ":Test768!~test@127.0.0.1 PRIVMSG #Test :Hello, World"},
			{Kind: 'S', Data: ":Test768!~test@127.0.0.1 PRIVMSG #Test :Hello again"},
			{Callback: func() error {
				channel := client.Channel("#Test")
				if channel == nil {
					return errors.New("Channel #Test not found")
				}

				client.EmitInput("/m +N", channel)
				client.EmitInput("/npcac Test_NPC stuffs things", channel)
				return nil
			}},
			{Kind: 'C', Data: "MODE #Test +N"},
			{Kind: 'C', Data: "NPCA #Test Test_NPC :stuffs things"},
			{Kind: 'S', Data: ":Test768!~test@127.0.0.1 MODE #Test +N"},
			{Kind: 'S', Data: ":\x1FTest_NPC\x1F!Test768@npc.fakeuser.invalid PRIVMSG #Test :\x01ACTION stuffs things\x01"},
			{Callback: func() error {
				channel := client.Channel("#Test")
				if channel == nil {
					return errors.New("Channel #Test not found")
				}

				client.Describef(channel.Name(), "does stuff with %d things", 42)
				client.Sayf(channel.Name(), "Hello, %s", "World")
				return nil
			}},
			{Kind: 'C', Data: "PRIVMSG #Test :\x01ACTION does stuff with 42 things\x01"},
			{Kind: 'C', Data: "PRIVMSG #Test :Hello, World"},
			{Callback: func() error {
				err := client.Part("#Test")
				if err != nil {
					return fmt.Errorf("Failed to part #Test: %s", err)
				}

				return nil
			}},
			{Kind: 'C', Data: "PART #Test"},
		},
	}

	addr, err := interaction.Listen()
	if err != nil {
		t.Fatal("Listen:", err)
	}

	if err := client.Disconnect(); err != irc.ErrNoConnection {
		t.Errorf("It should fail to disconnect, got: %s", err)
	}

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

	for i, logLine := range interaction.Log {
		t.Logf("Log[%d] = %#+v", i, logLine)
	}
}
