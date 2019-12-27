package irc_test

import (
	"context"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/gissleh/irc"
)

func TestHandle(t *testing.T) {
	rng := rand.NewSource(time.Now().UnixNano())
	eventName := strconv.FormatInt(rng.Int63(), 36) + strconv.FormatInt(rng.Int63(), 36) + strconv.FormatInt(rng.Int63(), 36)

	client := irc.New(context.Background(), irc.Config{})
	event := irc.NewEvent("test", eventName)
	handled := false

	irc.AddHandler(func(event *irc.Event, client *irc.Client) {
		t.Log("Got:", event.Kind(), event.Verb())

		if event.Kind() == "test" && event.Verb() == eventName {
			handled = true
		}
	})

	client.EmitSync(context.Background(), event)
	if !handled {
		t.Error("Event wasn't handled")
	}
}

func BenchmarkHandle(b *testing.B) {
	rng := rand.NewSource(time.Now().UnixNano())
	eventName := strconv.FormatInt(rng.Int63(), 36) + strconv.FormatInt(rng.Int63(), 36) + strconv.FormatInt(rng.Int63(), 36)

	client := irc.New(context.Background(), irc.Config{})
	event := irc.NewEvent("test", eventName)

	b.Run("Emit", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			client.Emit(event)
		}
	})

	b.Run("EmitSync", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			client.EmitSync(context.Background(), event)
		}
	})
}
