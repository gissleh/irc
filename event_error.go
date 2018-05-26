package irc

// NewErrorEvent makes an event of kind `error` and verb `code` with the text.
// It's absolutely trivial, but it's good to have standarized.
func NewErrorEvent(code, text string) Event {
	event := NewEvent("error", code)
	event.Text = text

	return event
}
