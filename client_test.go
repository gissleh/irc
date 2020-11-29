package irc_test

import (
	"context"
	"errors"
	"github.com/gissleh/irc/handlers"
	"testing"

	"github.com/gissleh/irc"
	"github.com/gissleh/irc/internal/irctest"
)

// Integration test below, brace yourself.
func TestClient(t *testing.T) {
	client := irc.New(context.Background(), irc.Config{
		Nick:            "Test",
		User:            "Tester",
		RealName:        "...",
		Alternatives:    []string{"Test2", "Test3", "Test4", "Test768"},
		SendRate:        1000,
		AutoJoinInvites: true,
	})

	logger := irctest.EventLog{}

	client.AddHandler(handlers.Input)
	client.AddHandler(handlers.MRoleplay)
	client.AddHandler(handlers.CTCP)
	client.AddHandler(logger.Handler)

	t.Logf("Client.ID = %#+v", client.ID())
	if client.ID() == "" {
		t.Fail()
	}

	interaction := irctest.Interaction{
		Strict: false,
		Lines: []irctest.InteractionLine{
			{Client: "CAP LS 302"},
			{Client: "NICK Test"},
			{Client: "USER Tester 8 * :..."},
			{Server: ":testserver.example.com NOTICE * :*** Checking your bits..."},
			{Server: ":testserver.example.com CAP * LS :multi-prefix chghost userhost-in-names vendorname/custom-stuff echo-message =malformed vendorname/advanced-custom-stuff=things,and,items"},
			{Server: ":testserver.example.com 433 * Test :Nick is not available"},
			{Client: "CAP REQ :multi-prefix chghost userhost-in-names echo-message"},
			{Server: ":testserver.example.com CAP * ACK :multi-prefix userhost-in-names"},
			{Client: "CAP END"},
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
			{Client: "NICK Test2"},
			{Server: ":testserver.example.com 433 * Test2 :Nick is not available"},
			{Client: "NICK Test3"},
			{Server: ":testserver.example.com 433 * Test3 :Nick is not available"},
			{Client: "NICK Test4"},
			{Server: ":testserver.example.com 433 * Test4 :Nick is not available"},
			{Client: "NICK Test768"},
			{Server: ":testserver.example.com 001 Test768 :Welcome to the TestServer Internet Relay Chat Network test"},
			{Client: "WHO Test768*"},
			{Server: ":testserver.example.com 002 Test768 :Your host is testserver.example.com[testserver.example.com/6667], running version charybdis-4-rc3"},
			{Server: ":testserver.example.com 003 Test768 :This server was created Fri Nov 25 2016 at 17:28:20 CET"},
			{Server: ":testserver.example.com 004 Test768 testserver.example.com charybdis-4-rc3 DQRSZagiloswxz CFILNPQbcefgijklmnopqrstvz bkloveqjfI"},
			{Server: ":testserver.example.com 005 Test768 FNC SAFELIST ELIST=CTU MONITOR=100 WHOX ETRACE KNOCK CHANTYPES=#& EXCEPTS INVEX CHANMODES=eIbq,k,flj,CFLNPQcgimnprstz CHANLIMIT=#&:15 :are supported by this server"},
			{Server: ":testserver.example.com 005 Test768 PREFIX=(ov)@+ MAXLIST=bqeI:100 MODES=4 NETWORK=TestServer STATUSMSG=@+ CALLERID=g CASEMAPPING=rfc1459 NICKLEN=30 MAXNICKLEN=31 CHANNELLEN=50 TOPICLEN=390 DEAF=D :are supported by this server"},
			{Server: ":testserver.example.com 005 Test768 TARGMAX=NAMES:1,LIST:1,KICK:1,WHOIS:1,PRIVMSG:4,NOTICE:4,ACCEPT:,MONITOR: EXTBAN=$,&acjmorsuxz| CLIENTVER=3.0 :are supported by this server"},
			{Server: ":testserver.example.com 251 Test768 :There are 0 users and 2 invisible on 1 servers"},
			{Server: ":testserver.example.com 254 Test768 1 :channels formed"},
			{Server: ":testserver.example.com 255 Test768 :I have 2 clients and 0 servers"},
			{Server: ":testserver.example.com 265 Test768 2 2 :Current local users 2, max 2"},
			{Server: ":testserver.example.com 266 Test768 2 2 :Current global users 2, max 2"},
			{Server: ":testserver.example.com 250 Test768 :Highest connection count: 2 (2 clients) (8 connections received)"},
			{Server: ":testserver.example.com 375 Test768 :- testserver.example.com Message of the Day - "},
			{Server: ":testserver.example.com 372 Test768 :- This server is only for testing irce, not chatting. If you happen"},
			{Server: ":testserver.example.com 372 Test768 :- to connect to it by accident, please disconnect immediately."},
			{Server: ":testserver.example.com 372 Test768 :-  "},
			{Server: ":testserver.example.com 372 Test768 :-  - #Test  :: Test Channel"},
			{Server: ":testserver.example.com 372 Test768 :-  - #Test2 :: Other Test Channel"},
			{Server: ":testserver.example.com 376 Test768 :End of /MOTD command."},
			{Server: ":testserver.example.com 352 Test768 * ~Tester testclient.example.com testserver.example.com Test768 H :0 ..."},
			{Server: ":Test768 MODE Test768 :+i"},
			{Server: "PING :testserver.example.com"}, // Ping/Pong to sync.
			{Client: "PONG :testserver.example.com"},
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
				client.Join("#Test")
				return nil
			}},
			{Client: "JOIN #Test"},
			{Server: ":Test768!~Tester@127.0.0.1 JOIN #Test *"},
			{Server: ":testserver.example.com 353 Test768 = #Test :Test768!~Tester@127.0.0.1 @+Gisle!irce@10.32.0.1"},
			{Server: ":testserver.example.com 366 Test768 #Test :End of /NAMES list."},
			{Server: "PING :testserver.example.com"}, // Ping/Pong to sync.
			{Client: "PONG :testserver.example.com"},
			{Callback: func() error {
				if client.Channel("#Test") == nil {
					return errors.New("Channel #Test not found")
				}

				return nil
			}},
			{Server: ":Gisle!~irce@10.32.0.1 MODE #Test +osv Test768 Test768"},
			{Server: ":Gisle!~irce@10.32.0.1 MODE #Test +N-s "},
			{Server: ":Test1234!~test2@172.17.37.1 JOIN #Test Test1234"},
			{Server: ":Test4321!~test2@172.17.37.1 JOIN #Test Test1234"},
			{Server: ":Gisle!~irce@10.32.0.1 MODE #Test +v Test1234"},
			{Server: "PING :testserver.example.com"}, // Ping/Pong to sync.
			{Client: "PONG :testserver.example.com"},
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
			{Server: ":Test1234!~test2@172.17.37.1 NICK Hunter2"},
			{Server: ":Hunter2!~test2@172.17.37.1 AWAY :Doing stuff"},
			{Server: ":Gisle!~irce@10.32.0.1 AWAY"},
			{Server: ":Gisle!~irce@10.32.0.1 PART #Test :Leaving the channel"},
			{Server: ":Hunter2!~test2@172.17.37.1 CHGHOST test2 some.awesome.virtual.host"},
			{Server: "@account=Hunter2 :Test4321!~test2@172.17.37.1 PRIVMSG #Test :Hello World."},
			{Server: "PING :testserver.example.com"}, // Ping/Pong to sync.
			{Client: "PONG :testserver.example.com"},
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

				event := logger.Last("packet", "PRIVMSG")
				if event == nil {
					return errors.New("did not find last channel message")
				}
				if event.ChannelTarget() == nil {
					return errors.New("event lacks channel target")
				}

				return nil
			}},
			{Server: ":Hunter2!~test2@172.17.37.1 PRIVMSG Test768 :Hello, World"},
			{Server: "PING :testserver.example.com"}, // Ping/Pong to sync.
			{Client: "PONG :testserver.example.com"},
			{Callback: func() error {
				query := client.Query("Hunter2")
				if query == nil {
					return errors.New("Did not find query")
				}

				event := logger.Last("packet", "PRIVMSG")
				if event == nil {
					return errors.New("did not find last query message")
				}
				if event.QueryTarget() == nil {
					return errors.New("event lacks query target")
				}

				return nil
			}},
			{Server: ":Hunter2!~test2@172.17.37.1 NICK SevenAsterisks"},
			{Server: "PING :testserver.example.com"}, // Ping/Pong to sync.
			{Client: "PONG :testserver.example.com"},
			{Callback: func() error {
				oldQuery := client.Query("Hunter2")
				if oldQuery != nil {
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
			{Client: "INVALIDCOMMAND stuff and things"},
			{Server: ":testserver.example.com 421 Test768 INVALIDCOMMAND :Unknown command"},
			{Callback: func() error {
				client.Say("SevenAsterisks", "hi!")
				return nil
			}},
			{Client: "PRIVMSG SevenAsterisks :hi!"},
			{Server: ":Test768!~Tester@127.0.0.1 PRIVMSG SevenAsterisks :hi!"},
			{Callback: func() error {
				event := logger.Last("packet", "PRIVMSG")
				if event == nil {
					return errors.New("did not find last query message")
				}
				if event.QueryTarget() == nil {
					return errors.New("event lacks query target")
				}
				if event.QueryTarget().Name() != "SevenAsterisks" {
					return errors.New("incorrect query target")
				}

				return nil
			}},
			{Callback: func() error {
				client.EmitInput("/invalidcommand stuff and things", nil)
				return nil
			}},
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
			{Client: "PRIVMSG #Test :\x01ACTION does stuff\x01"},
			{Client: "PRIVMSG #Test :\x01ACTION describes stuff\x01"},
			{Client: "PRIVMSG #Test :Hello, World"},
			{Client: "PRIVMSG #Test :Hello again"},
			{Server: ":Test768!~Tester@127.0.0.1 PRIVMSG #Test :\x01ACTION does stuff\x01"},
			{Server: ":Test768!~Tester@127.0.0.1 PRIVMSG #Test :\x01ACTION describes stuff\x01"},
			{Server: ":Test768!~Tester@127.0.0.1 PRIVMSG #Test :Hello, World"},
			{Server: ":Test768!~Tester@127.0.0.1 PRIVMSG #Test :Hello again"},
			{Callback: func() error {
				channel := client.Channel("#Test")
				if channel == nil {
					return errors.New("Channel #Test not found")
				}

				client.EmitInput("/m +N", channel)
				client.EmitInput("/npcac Test_NPC stuffs things", channel)
				return nil
			}},
			{Client: "MODE #Test +N"},
			{Client: "NPCA #Test Test_NPC :stuffs things"},
			{Server: ":Test768!~Tester@127.0.0.1 MODE #Test +N"},
			{Server: ":\x1FTest_NPC\x1F!Test768@npc.fakeuser.invalid PRIVMSG #Test :\x01ACTION stuffs things\x01"},
			{Callback: func() error {
				channel := client.Channel("#Test")
				if channel == nil {
					return errors.New("Channel #Test not found")
				}

				client.Describef(channel.Name(), "does stuff with %d things", 42)
				client.Sayf(channel.Name(), "Hello, %s", "World")
				return nil
			}},
			{Client: "PRIVMSG #Test :\x01ACTION does stuff with 42 things\x01"},
			{Client: "PRIVMSG #Test :Hello, World"},
			{Callback: func() error {
				channel := client.Channel("#Test")
				if channel == nil {
					return errors.New("#Test doesn't exist")
				}

				_, err := client.RemoveTarget(channel)
				return err
			}},
			{Client: "PART #Test"},
			{Server: ":Test768!~Tester@127.0.0.1 PART #Test"},
			{Server: "PING :testserver.example.com"}, // Ping/Pong to sync.
			{Client: "PONG :testserver.example.com"},
			{Callback: func() error {
				if client.Channel("#Test") != nil {
					return errors.New("#Test is still there.")
				}

				return nil
			}},
			{Callback: func() error {
				client.Join("#Test2")
				return nil
			}},
			{Client: "JOIN #Test2"},
			{Server: ":Test768!~Tester@127.0.0.1 JOIN #Test2 *"},
			{Server: ":testserver.example.com 353 Test768 = #Test2 :Test768!~Tester@127.0.0.1 +DoomedUser!doom@example.com @+ZealousMod!zeal@example.com"},
			{Server: ":testserver.example.com 366 Test768 #Test2 :End of /NAMES list."},
			{Server: "PING :testserver.example.com"}, // Ping/Pong to sync.
			{Client: "PONG :testserver.example.com"},
			{Callback: func() error {
				channel := client.Channel("#Test2")
				if channel == nil {
					return errors.New("Channel #Test2 not found")
				}

				return irctest.AssertUserlist(t, channel, "@ZealousMod", "+DoomedUser", "Test768")
			}},
			{Server: ":ZealousMod!zeal@example.com KICK #Test2 DoomedUser :Kickety kick"},
			{Server: "PING :testserver.example.com sync"}, // Ping/Pong to sync.
			{Client: "PONG :testserver.example.com sync"},
			{Callback: func() error {
				channel := client.Channel("#Test2")
				if channel == nil {
					return errors.New("Channel #Test2 not found")
				}

				return irctest.AssertUserlist(t, channel, "@ZealousMod", "Test768")
			}},
			{Server: ":ZealousMod!zeal@example.com KICK #Test2 Test768 :Kickety kick"},
			{Server: "PING :testserver.example.com sync"}, // Ping/Pong to sync.
			{Client: "PONG :testserver.example.com sync"},
			{Callback: func() error {
				if client.Channel("#Test2") != nil {
					return errors.New("#Test2 is still there.")
				}
				return nil
			}},
			{Server: ":testserver.example.com CAP Test768 NEW :invite-notify"},
			{Client: "CAP REQ :invite-notify"},
			{Server: ":testserver.example.com CAP Test768 ACK :invite-notify"},
			{Server: ":ZealousMod!zeal@example.com INVITE Test768 #Test2"},
			{Client: "JOIN #Test2"},
			{Server: ":Test768!~Tester@127.0.0.1 JOIN #Test2 *"},
			{Server: ":testserver.example.com 353 Test768 = #Test2 :Test768!~Tester@127.0.0.1 @+ZealousMod!zeal@example.com"},
			{Server: ":testserver.example.com 366 Test768 #Test2 :End of /NAMES list."},
			{Server: "PING :testserver.example.com"}, // Ping/Pong to sync.
			{Client: "PONG :testserver.example.com"},
			{Callback: func() error {
				if client.Channel("#Test2") == nil {
					return errors.New("#Test2 is not there.")
				}
				return nil
			}},
			{Server: ":ZealousMod!zeal@example.com INVITE Test768 #Test2"},
			{Server: ":ZealousMod!zeal@example.com INVITE DoomedUser #test768-do-not-join"},
			{Server: "PING :testserver.example.com"}, // Ping/Pong to sync.
			{Client: "PONG :testserver.example.com"},
		},
	}

	addr, err := interaction.Listen()
	if err != nil {
		t.Fatal("Listen:", err)
	}

	if err := client.Disconnect(false); err != irc.ErrNoConnection {
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
			if interaction.Lines[fail.Index].Server != "" {
				t.Error("Line.Server:", interaction.Lines[fail.Index].Server)
			}
			if interaction.Lines[fail.Index].Client != "" {
				t.Error("Line.Client:", interaction.Lines[fail.Index].Client)
			}
		}
	}

	for i, logLine := range interaction.Log {
		t.Logf("Log[%d] = %#+v", i, logLine)
	}
}

// TestParenthesesBug tests that the bugfix causing `((Remove :01 goofs!*))` to be parsed as an empty message. It was
// initially thought to be caused by the parentheses (like a hidden m_roleplay NPC attribution removal), hence the name
// for this function.
func TestParenthesesBug(t *testing.T) {
	gotMessage := true

	client := irc.New(context.Background(), irc.Config{
		Nick: "Stuff",
	})

	client.AddHandler(func(event *irc.Event, client *irc.Client) {
		if event.Name() == "packet.privmsg" || event.Nick == "Beans" {
			gotMessage = true

			if event.Text != "((Remove :01 goofs!*))" {
				t.Errorf("Expected: %#+v", "((Remove :01 goofs!*))")
				t.Errorf("Result:   %#+v", event.Text)
			}
		}
	})

	packet, err := irc.ParsePacket("@example/tag=32; :Beans!beans@beans.example.com PRIVMSG Stuff :((Remove :01 goofs!*))")
	if err != nil {
		t.Error("Parse", err)
	}

	client.EmitSync(context.Background(), packet)

	if !gotMessage {
		t.Error("Message was not received")
	}
}
