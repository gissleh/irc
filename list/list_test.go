package list_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"git.aiterp.net/gisle/irc/isupport"
	"git.aiterp.net/gisle/irc/list"
)

var testISupport isupport.ISupport

func TestList(t *testing.T) {
	table := []struct {
		namestoken   string
		shouldInsert bool
		user         list.User
		order        []string
	}{
		{
			"@+Test!~test@example.com", true,
			list.User{
				Nick:         "Test",
				User:         "~test",
				Host:         "example.com",
				Modes:        "ov",
				Prefixes:     "@+",
				PrefixedNick: "@Test",
			},
			[]string{"@Test"},
		},
		{
			"+@Test2!~test2@example.com", true,
			list.User{
				Nick:         "Test2",
				User:         "~test2",
				Host:         "example.com",
				Modes:        "ov",
				Prefixes:     "@+",
				PrefixedNick: "@Test2",
			},
			[]string{"@Test", "@Test2"},
		},
		{
			"+Gissleh", true,
			list.User{
				Nick:         "Gissleh",
				User:         "",
				Host:         "",
				Modes:        "v",
				Prefixes:     "+",
				PrefixedNick: "+Gissleh",
			},
			[]string{"@Test", "@Test2", "+Gissleh"},
		},
		{
			"Guest!~guest@10.72.3.15", true,
			list.User{
				Nick:         "Guest",
				User:         "~guest",
				Host:         "10.72.3.15",
				Modes:        "",
				Prefixes:     "",
				PrefixedNick: "Guest",
			},
			[]string{"@Test", "@Test2", "+Gissleh", "Guest"},
		},
		{
			"@AOP!actualIdent@10.32.8.174", true,
			list.User{
				Nick:         "AOP",
				User:         "actualIdent",
				Host:         "10.32.8.174",
				Modes:        "o",
				Prefixes:     "@",
				PrefixedNick: "@AOP",
			},
			[]string{"@AOP", "@Test", "@Test2", "+Gissleh", "Guest"},
		},
		{
			"@ZOP!actualIdent@10.32.8.174", true,
			list.User{
				Nick:         "ZOP",
				User:         "actualIdent",
				Host:         "10.32.8.174",
				Modes:        "o",
				Prefixes:     "@",
				PrefixedNick: "@ZOP",
			},
			[]string{"@AOP", "@Test", "@Test2", "@ZOP", "+Gissleh", "Guest"},
		},
		{
			"+ZVoice!~zv@10.32.8.174", true,
			list.User{
				Nick:         "ZVoice",
				User:         "~zv",
				Host:         "10.32.8.174",
				Modes:        "v",
				Prefixes:     "+",
				PrefixedNick: "+ZVoice",
			},
			[]string{"@AOP", "@Test", "@Test2", "@ZOP", "+Gissleh", "+ZVoice", "Guest"},
		},
		{
			"+ZVoice!~zv@10.32.8.174", false,
			list.User{},
			[]string{"@AOP", "@Test", "@Test2", "@ZOP", "+Gissleh", "+ZVoice", "Guest"},
		},
	}

	list := list.New(&testISupport)

	for _, row := range table {
		t.Run("Insert_"+row.namestoken, func(t *testing.T) {
			ok := list.InsertFromNamesToken(row.namestoken)
			if ok && !row.shouldInsert {
				t.Error("Insert should have failed!")
				return
			}
			if !ok && row.shouldInsert {
				t.Error("Insert should NOT have failed!")
				return
			}

			if row.shouldInsert {
				user, ok := list.User(row.user.Nick)
				if !ok {
					t.Error("Could not find user.")
					return
				}

				jsonA, _ := json.MarshalIndent(user, "", "  ")
				jsonB, _ := json.MarshalIndent(row.user, "", "  ")

				t.Log("result =", string(jsonA))

				if string(jsonA) != string(jsonB) {
					t.Log("expectation =", string(jsonB))
					t.Error("Users did not match!")
				}
			}

			order := make([]string, 0, 16)
			for _, user := range list.Users() {
				order = append(order, user.PrefixedNick)
			}

			orderA := strings.Join(order, ", ")
			orderB := strings.Join(row.order, ", ")

			t.Log("order =", orderA)
			if orderA != orderB {
				t.Log("orderExpected =", orderB)
				t.Error("Order did not match!")
			}
		})
	}

	modeTable := []struct {
		add   bool
		mode  rune
		nick  string
		ok    bool
		order []string
	}{
		{
			true, 'o', "Gissleh", true,
			[]string{"@AOP", "@Gissleh", "@Test", "@Test2", "@ZOP", "+ZVoice", "Guest"},
		},
		{
			false, 'o', "Gissleh", true,
			[]string{"@AOP", "@Test", "@Test2", "@ZOP", "+Gissleh", "+ZVoice", "Guest"},
		},
		{
			true, 'o', "InvalidNick", false,
			[]string{"@AOP", "@Test", "@Test2", "@ZOP", "+Gissleh", "+ZVoice", "Guest"},
		},
		{
			true, 'v', "AOP", true,
			[]string{"@AOP", "@Test", "@Test2", "@ZOP", "+Gissleh", "+ZVoice", "Guest"},
		},
		{
			true, 'v', "ZOP", true,
			[]string{"@AOP", "@Test", "@Test2", "@ZOP", "+Gissleh", "+ZVoice", "Guest"},
		},
		{
			true, 'v', "Guest", true,
			[]string{"@AOP", "@Test", "@Test2", "@ZOP", "+Gissleh", "+Guest", "+ZVoice"},
		},
		{
			true, 'v', "Test", true,
			[]string{"@AOP", "@Test", "@Test2", "@ZOP", "+Gissleh", "+Guest", "+ZVoice"},
		},
		{
			false, 'v', "Test", true,
			[]string{"@AOP", "@Test", "@Test2", "@ZOP", "+Gissleh", "+Guest", "+ZVoice"},
		},
		{
			false, 'o', "Test", true,
			[]string{"@AOP", "@Test2", "@ZOP", "+Gissleh", "+Guest", "+ZVoice", "Test"},
		},
		{
			false, 'o', "AOP", true,
			[]string{"@Test2", "@ZOP", "+AOP", "+Gissleh", "+Guest", "+ZVoice", "Test"},
		},
		{
			true, 'x', "AOP", false,
			[]string{"@Test2", "@ZOP", "+AOP", "+Gissleh", "+Guest", "+ZVoice", "Test"},
		},
		{
			false, 'x', "ZOP", false,
			[]string{"@Test2", "@ZOP", "+AOP", "+Gissleh", "+Guest", "+ZVoice", "Test"},
		},
		{
			true, 'o', "UNKNOWN_USER", false,
			[]string{"@Test2", "@ZOP", "+AOP", "+Gissleh", "+Guest", "+ZVoice", "Test"},
		},
		{
			false, 'o', "UNKNOWN_USER", false,
			[]string{"@Test2", "@ZOP", "+AOP", "+Gissleh", "+Guest", "+ZVoice", "Test"},
		},
	}

	for i, row := range modeTable {
		t.Run(fmt.Sprintf("Mode_%d_%s", i, row.nick), func(t *testing.T) {
			var ok bool

			if row.add {
				ok = list.AddMode(row.nick, row.mode)
			} else {
				ok = list.RemoveMode(row.nick, row.mode)
			}

			if ok && !row.ok {
				t.Error("This should be not ok, but it is ok.")
			}
			if !ok && row.ok {
				t.Error("This is not ok.")
			}

			order := make([]string, 0, 16)
			for _, user := range list.Users() {
				order = append(order, user.PrefixedNick)
			}

			orderA := strings.Join(order, ", ")
			orderB := strings.Join(row.order, ", ")

			t.Log("order =", orderA)
			if orderA != orderB {
				t.Log("orderExpected =", orderB)
				t.Error("Order did not match!")
			}
		})
	}

	renameTable := []struct {
		from  string
		to    string
		ok    bool
		order []string
	}{
		{
			"ZOP", "AAOP", true,
			[]string{"@AAOP", "@Test2", "+AOP", "+Gissleh", "+Guest", "+ZVoice", "Test"},
		},
		{
			"Test", "ATest", true,
			[]string{"@AAOP", "@Test2", "+AOP", "+Gissleh", "+Guest", "+ZVoice", "ATest"},
		},
		{
			"AOP", "ZOP", true,
			[]string{"@AAOP", "@Test2", "+Gissleh", "+Guest", "+ZOP", "+ZVoice", "ATest"},
		},
		{
			"AOP", "ZOP", false,
			[]string{"@AAOP", "@Test2", "+Gissleh", "+Guest", "+ZOP", "+ZVoice", "ATest"},
		},
		{
			"ATest", "Test", true,
			[]string{"@AAOP", "@Test2", "+Gissleh", "+Guest", "+ZOP", "+ZVoice", "Test"},
		},
		{
			"Test2", "AAATest", true,
			[]string{"@AAATest", "@AAOP", "+Gissleh", "+Guest", "+ZOP", "+ZVoice", "Test"},
		},
		{
			"ZOP", "AAATest", false,
			[]string{"@AAATest", "@AAOP", "+Gissleh", "+Guest", "+ZOP", "+ZVoice", "Test"},
		},
		{
			"AAATest", "AAATest", true,
			[]string{"@AAATest", "@AAOP", "+Gissleh", "+Guest", "+ZOP", "+ZVoice", "Test"},
		},
	}

	for i, row := range renameTable {
		t.Run(fmt.Sprintf("Rename_%d_%s_%s", i, row.from, row.to), func(t *testing.T) {
			ok := list.Rename(row.from, row.to)
			if ok && !row.ok {
				t.Error("This should be not ok, but it is ok.")
			}
			if !ok && row.ok {
				t.Error("This is not ok.")
			}

			order := make([]string, 0, 16)
			for _, user := range list.Users() {
				order = append(order, user.PrefixedNick)
			}

			orderA := strings.Join(order, ", ")
			orderB := strings.Join(row.order, ", ")

			t.Log("order =", orderA)
			if orderA != orderB {
				t.Log("orderExpected =", orderB)
				t.Error("Order did not match!")
			}
		})
	}

	removeTable := []struct {
		nick  string
		ok    bool
		order []string
	}{
		{
			"AAOP", true,
			[]string{"@AAATest", "+Gissleh", "+Guest", "+ZOP", "+ZVoice", "Test"},
		},
		{
			"AAOP", false,
			[]string{"@AAATest", "+Gissleh", "+Guest", "+ZOP", "+ZVoice", "Test"},
		},
		{
			"Guest", true,
			[]string{"@AAATest", "+Gissleh", "+ZOP", "+ZVoice", "Test"},
		},
		{
			"ZOP", true,
			[]string{"@AAATest", "+Gissleh", "+ZVoice", "Test"},
		},
		{
			"ATest", false,
			[]string{"@AAATest", "+Gissleh", "+ZVoice", "Test"},
		},
		{
			"Test", true,
			[]string{"@AAATest", "+Gissleh", "+ZVoice"},
		},
	}

	for i, row := range removeTable {
		t.Run(fmt.Sprintf("Rename_%d_%s", i, row.nick), func(t *testing.T) {
			ok := list.Remove(row.nick)
			if ok && !row.ok {
				t.Error("This should be not ok, but it is ok.")
			}
			if !ok && row.ok {
				t.Error("This is not ok.")
			}

			order := make([]string, 0, 16)
			for _, user := range list.Users() {
				order = append(order, user.PrefixedNick)
			}

			if _, ok := list.User(row.nick); ok {
				t.Error("User is still there")
			}

			orderA := strings.Join(order, ", ")
			orderB := strings.Join(row.order, ", ")

			t.Log("order =", orderA)
			if orderA != orderB {
				t.Log("orderExpected =", orderB)
				t.Error("Order did not match!")
			}
		})
	}

	t.Run("AutoSort", func(t *testing.T) {
		list.SetAutoSort(false)

		if ok := list.InsertFromNamesToken("@+AAAAAAAAA"); !ok {
			t.Error("Failed to insert user @+AAAAAAAAA")
		}

		users := list.Users()
		last := users[len(users)-1]

		if last.PrefixedNick != "@AAAAAAAAA" {
			t.Error("@+AAAAAAAAA isn't last, "+last.PrefixedNick, "is.")
		}

		list.SetAutoSort(true)

		users = list.Users()
		last = users[len(users)-1]

		if last.PrefixedNick == "@AAAAAAAAA" {
			t.Error("@+AAAAAAAAA is still last after autosort was enabled. That's not right.")
		}
	})

	t.Run("Clear", func(t *testing.T) {
		list.Clear()

		if len(list.Users()) != 0 {
			t.Error("Clear failed!")
		}
	})
}

func init() {
	isupportData := map[string]string{
		"FNC":         "",
		"SAFELIST":    "",
		"ELIST":       "CTU",
		"MONITOR":     "100",
		"WHOX":        "",
		"ETRACE":      "",
		"KNOCK":       "",
		"CHANTYPES":   "#&",
		"EXCEPTS":     "",
		"INVEX":       "",
		"CHANMODES":   "eIbq,k,flj,CFLNPQcgimnprstz",
		"CHANLIMIT":   "#&:15",
		"PREFIX":      "(ov)@+",
		"MAXLIST":     "bqeI:100",
		"MODES":       "4",
		"NETWORK":     "TestServer",
		"STATUSMSG":   "@+",
		"CALLERID":    "g",
		"CASEMAPPING": "rfc1459",
		"NICKLEN":     "30",
		"MAXNICKLEN":  "31",
		"CHANNELLEN":  "50",
		"TOPICLEN":    "390",
		"DEAF":        "D",
		"TARGMAX":     "NAMES:1,LIST:1,KICK:1,WHOIS:1,PRIVMSG:4,NOTICE:4,ACCEPT:,MONITOR:",
		"EXTBAN":      "$,&acjmorsuxz|",
		"CLIENTVER":   "3.0",
	}

	for key, value := range isupportData {
		testISupport.Set(key, value)
	}
}
