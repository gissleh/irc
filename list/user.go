package list

// A User represents a member of a userlist.
type User struct {
	Nick         string `json:"nick"`
	User         string `json:"user,omitempty"`
	Host         string `json:"host,omitempty"`
	Account      string `json:"account,omitempty"`
	Modes        string `json:"modes"`
	Prefixes     string `json:"prefixes"`
	PrefixedNick string `json:"prefixedNick"`
}

// UserPatch is used in List.Patch to apply changes to a user
type UserPatch struct {
	User         string
	Host         string
	Account      string
	ClearAccount bool
}

// HighestMode returns the highest mode.
func (user *User) HighestMode() rune {
	if len(user.Modes) == 0 {
		return 0
	}

	return rune(user.Modes[0])
}

// PrefixedNick gets the full nick.
func (user *User) updatePrefixedNick() {
	if len(user.Prefixes) == 0 {
		user.PrefixedNick = user.Nick
		return
	}

	user.PrefixedNick = string(user.Prefixes[0]) + user.Nick
}
