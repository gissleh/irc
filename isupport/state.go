package isupport

type State struct {
	Raw          map[string]string `json:"raw"`
	Prefixes     map[rune]rune     `json:"-"`
	ModeOrder    string            `json:"modeOrder"`
	PrefixOrder  string            `json:"prefixOrder"`
	ChannelModes []string          `json:"channelModes"`
}

func (state *State) Copy() *State {
	stateCopy := *state
	stateCopy.Raw = make(map[string]string, len(state.Raw))
	for key, value := range state.Raw {
		stateCopy.Raw[key] = value
	}

	return &stateCopy
}
