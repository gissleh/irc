package irc

// NewErrorEvent makes an event of kind `error` and verb `code` with the text.
// It's absolutely trivial, but it's good to have standarized.
func NewErrorEvent(code, text, i18nKey string, raw error) Event {
	return NewErrorEventTarget(nil, code, text, i18nKey, raw)
}

func NewErrorEventTarget(target Target, code, text, i18nKey string, raw error) Event {
	event := NewEvent("error", code)
	event.Text = text

	if target != nil {
		event.targets = append(event.targets, target)
	}

	if i18nKey != "" {
		event.Tags["i18n_key"] = i18nKey
	}
	if raw != nil {
		event.Tags["raw"] = raw.Error()
	}

	return event
}
