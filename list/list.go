package list

import (
	"sort"
	"strings"
	"sync"

	"git.aiterp.net/gisle/irc/isupport"
)

// The List of users in a channel. It has all operations one would perform on
// users, like adding/removing modes and changing nicks.
type List struct {
	mutex    sync.RWMutex
	isupport *isupport.ISupport
	users    []*User
	index    map[string]*User
	autosort bool
}

// New creates a new list with the ISupport. The list can be reused between connections since the
// ISupport is simply cleared and repopulated, but it should be cleared.
func New(isupport *isupport.ISupport) *List {
	return &List{
		isupport: isupport,
		users:    make([]*User, 0, 64),
		index:    make(map[string]*User, 64),
		autosort: true,
	}
}

// InsertFromNamesToken inserts using a NAMES token to get the nick, user, host and prefixes.
// The format is `"@+Nick@user!hostmask.example.com"`
func (list *List) InsertFromNamesToken(namestoken string) (ok bool) {
	user := User{}

	// Parse prefixes and modes. @ and ! (It's IRCHighWay if you were wondering) are both
	// mode prefixes and that just makes a mess if leave them for last. It also supports
	// `multi-prefix`
	for i, ch := range namestoken {
		mode := list.isupport.Mode(ch)
		if mode == 0 {
			if i != 0 {
				namestoken = namestoken[i:]
			}
			break
		}

		user.Prefixes += string(ch)
		user.Modes += string(mode)
	}

	// Get the nick
	split := strings.Split(namestoken, "!")
	user.Nick = split[0]

	// Support `userhost-in-names`
	if len(split) == 2 {
		userhost := strings.Split(split[1], "@")
		if len(userhost) == 2 {
			user.User = userhost[0]
			user.Host = userhost[1]
		}
	}

	return list.Insert(user)
}

// Insert a user. Modes and prefixes will be cleaned up before insertion.
func (list *List) Insert(user User) (ok bool) {
	if len(user.Modes) > 0 {
		// IRCv3 promises they'll be ordered by rank in WHO and NAMES replies,
		// but one can never be too sure with IRC.
		user.Modes = list.isupport.SortModes(user.Modes)
		if len(user.Prefixes) < len(user.Modes) {
			user.Prefixes = list.isupport.Prefixes(user.Modes)
		} else {
			user.Prefixes = list.isupport.SortPrefixes(user.Prefixes)
		}
		user.updatePrefixedNick()
	} else {
		user.Prefixes = ""
		user.updatePrefixedNick()
	}

	list.mutex.Lock()
	defer list.mutex.Unlock()

	if list.index[strings.ToLower(user.Nick)] != nil {
		return false
	}

	list.users = append(list.users, &user)
	list.index[strings.ToLower(user.Nick)] = &user

	if list.autosort {
		list.sort()
	}

	return true
}

// AddMode adds a mode to a user. Redundant modes will be ignored. It returns true if
// the user can be found, even if the mode was redundant.
func (list *List) AddMode(nick string, mode rune) (ok bool) {
	if !list.isupport.IsPermissionMode(mode) {
		return false
	}

	list.mutex.RLock()
	defer list.mutex.RUnlock()

	user := list.index[strings.ToLower(nick)]
	if user == nil {
		return false
	}
	if strings.ContainsRune(user.Modes, mode) {
		return true
	}

	prevHighest := user.HighestMode()
	user.Modes = list.isupport.SortModes(user.Modes + string(mode))
	user.Prefixes = list.isupport.Prefixes(user.Modes)
	user.updatePrefixedNick()

	// Only sort if the new mode changed the highest mode.
	if list.autosort && prevHighest != user.HighestMode() {
		list.sort()
	}

	return true
}

// RemoveMode adds a mode to a user. It returns true if
// the user can be found, even if the mode was not there.
func (list *List) RemoveMode(nick string, mode rune) (ok bool) {
	if !list.isupport.IsPermissionMode(mode) {
		return false
	}

	list.mutex.RLock()
	defer list.mutex.RUnlock()

	user := list.index[strings.ToLower(nick)]
	if user == nil {
		return false
	}
	if !strings.ContainsRune(user.Modes, mode) {
		return true
	}

	prevHighest := user.HighestMode()
	user.Modes = strings.Replace(user.Modes, string(mode), "", 1)
	user.Prefixes = strings.Replace(user.Prefixes, string(list.isupport.Prefix(mode)), "", 1)
	user.updatePrefixedNick()

	// Only sort if the new mode changed the highest mode.
	if list.autosort && prevHighest != user.HighestMode() {
		list.sort()
	}

	return true
}

// Rename renames a user. It will return true if user by `from` exists, or if user by `to` does not exist.
func (list *List) Rename(from, to string) (ok bool) {
	fromKey := strings.ToLower(from)
	toKey := strings.ToLower(to)

	list.mutex.Lock()
	defer list.mutex.Unlock()

	// Sanitiy check
	user := list.index[fromKey]
	if user == nil {
		return false
	}
	if from == to {
		return true
	}
	existing := list.index[toKey]
	if existing != nil {
		return false
	}

	user.Nick = to
	user.updatePrefixedNick()

	delete(list.index, fromKey)
	list.index[toKey] = user

	if list.autosort {
		list.sort()
	}

	return true
}

// Remove a user from the userlist.
func (list *List) Remove(nick string) (ok bool) {
	list.mutex.Lock()
	defer list.mutex.Unlock()

	user := list.index[strings.ToLower(nick)]
	if user == nil {
		return false
	}

	for i := range list.users {
		if list.users[i] == user {
			list.users = append(list.users[:i], list.users[i+1:]...)
			break
		}
	}
	delete(list.index, strings.ToLower(nick))

	return true
}

// User gets a copy of the user by nick, or an empty user if there is none.
func (list *List) User(nick string) (u User, ok bool) {
	list.mutex.RLock()
	defer list.mutex.RUnlock()

	user := list.index[strings.ToLower(nick)]
	if user == nil {
		return User{}, false
	}

	return *user, true
}

// Users gets a copy of the users in the list's current state.
func (list *List) Users() []User {
	result := make([]User, len(list.users))
	list.mutex.RLock()
	for i := range list.users {
		result[i] = *list.users[i]
	}
	list.mutex.RUnlock()

	return result
}

// Patch allows editing a limited subset of the user's properties.
func (list *List) Patch(nick string, patch UserPatch) (ok bool) {
	list.mutex.Lock()
	defer list.mutex.Unlock()

	for _, user := range list.users {
		if strings.EqualFold(nick, user.Nick) {
			if patch.Account != "" || patch.ClearAccount {
				user.Account = patch.Account
			}

			if patch.User != "" {
				user.User = patch.User
			}

			if patch.Host != "" {
				user.Host = patch.Host
			}

			return true
		}
	}

	return false
}

// SetAutoSort enables or disables automatic sorting, which by default is enabled.
// Dislabing it makes sense when doing a massive operation. Enabling it will trigger
// a sort.
func (list *List) SetAutoSort(autosort bool) {
	list.mutex.Lock()
	list.autosort = autosort
	list.sort()
	list.mutex.Unlock()
}

// Clear removes all users in a list.
func (list *List) Clear() {
	list.mutex.Lock()

	list.users = list.users[:0]
	for key := range list.index {
		delete(list.index, key)
	}

	list.mutex.Unlock()
}

// Immutable gets an immutable version of the list.
func (list *List) Immutable() Immutable {
	return Immutable{list: list}
}

func (list *List) sort() {
	sort.Slice(list.users, func(i, j int) bool {
		a := list.users[i]
		b := list.users[j]

		aMode := a.HighestMode()
		bMode := b.HighestMode()

		if aMode != bMode {
			return list.isupport.IsModeHigher(aMode, bMode)
		}

		return strings.ToLower(a.Nick) < strings.ToLower(b.Nick)
	})
}
