package irctest

import (
	"bufio"
	"net"
	"strings"
	"sync"
	"time"
)

// An Interaction is a "simulated" server that will trigger the
// client.
type Interaction struct {
	wg sync.WaitGroup

	Strict  bool
	Lines   []InteractionLine
	Log     []string
	Failure *InteractionFailure
}

// Listen listens for a client in a separate goroutine.
func (interaction *Interaction) Listen() (addr string, err error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}

	lines := make([]InteractionLine, len(interaction.Lines))
	copy(lines, interaction.Lines)

	go func() {
		interaction.wg.Add(1)
		defer interaction.wg.Done()

		conn, err := listener.Accept()
		if err != nil {
			interaction.Failure = &InteractionFailure{
				Index: -1, NetErr: err,
			}

			return
		}

		defer conn.Close()

		reader := bufio.NewReader(conn)

		for i := 0; i < len(lines); i++ {
			line := lines[i]

			if line.Server != "" {
				_ = conn.SetWriteDeadline(time.Now().Add(time.Second * 2))
				_, err := conn.Write(append([]byte(line.Server), '\r', '\n'))
				if err != nil {
					interaction.Failure = &InteractionFailure{
						Index: i, NetErr: err,
					}
					return
				}
			} else if line.Client != "" {
				_ = conn.SetReadDeadline(time.Now().Add(time.Second * 2))
				input, err := reader.ReadString('\n')
				if err != nil {
					interaction.Failure = &InteractionFailure{
						Index: i, NetErr: err,
					}
					return
				}
				input = strings.Replace(input, "\r", "", -1)
				input = strings.Replace(input, "\n", "", 1)

				match := line.Client
				success := false

				if strings.HasSuffix(match, "*") {
					success = strings.HasPrefix(input, match[:len(match)-1])
				} else {
					success = match == input
				}

				interaction.Log = append(interaction.Log, input)

				if !success {
					if !interaction.Strict {
						i--
						continue
					}

					interaction.Failure = &InteractionFailure{
						Index: i, Result: input,
					}
					return
				}
			} else if line.Callback != nil {
				err := line.Callback()
				if err != nil {
					interaction.Failure = &InteractionFailure{
						Index: i, CBErr: err,
					}
					return
				}
			}
		}
	}()

	return listener.Addr().String(), nil
}

// Wait waits for the setup to be done. It's safe to check
// Failure after that.
func (interaction *Interaction) Wait() {
	interaction.wg.Wait()
}

// InteractionFailure signifies a test failure.
type InteractionFailure struct {
	Index  int
	Result string
	NetErr error
	CBErr  error
}

// InteractionLine is part of an interaction, whether it is a line
// that is sent to a client or a line expected from a client.
type InteractionLine struct {
	Client   string
	Server   string
	Callback func() error
}
