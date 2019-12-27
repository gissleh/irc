package isupport

import (
	"strconv"
	"strings"
	"sync"
)

// ISupport is a data structure containing server instructions about
// supported modes, encodings, lengths, prefixes, and so on. It is built
// from the 005 numeric's data, and has helper methods that makes sense
// of it. It's thread-safe through a reader/writer lock, so the locks will
// only block in the short duration post-registration when the 005s come in
type ISupport struct {
	lock  sync.RWMutex
	state State
}

// Get gets an isupport key. This is unprocessed data, and a helper should
// be used if available.
func (isupport *ISupport) Get(key string) (value string, ok bool) {
	isupport.lock.RLock()
	value, ok = isupport.state.Raw[key]
	isupport.lock.RUnlock()
	return
}

// Number gets a key and converts it to a number.
func (isupport *ISupport) Number(key string) (value int, ok bool) {
	isupport.lock.RLock()
	strValue, ok := isupport.state.Raw[key]
	isupport.lock.RUnlock()

	if !ok {
		return 0, ok
	}

	value, err := strconv.Atoi(strValue)
	if err != nil {
		return value, false
	}

	return value, ok
}

// ParsePrefixedNick parses a full nick into its components.
// Example: "@+HammerTime62" -> `"HammerTime62", "ov", "@+"`
func (isupport *ISupport) ParsePrefixedNick(fullnick string) (nick, modes, prefixes string) {
	isupport.lock.RLock()
	defer isupport.lock.RUnlock()

	if fullnick == "" || isupport.state.Prefixes == nil {
		return fullnick, "", ""
	}

	for i, ch := range fullnick {
		if mode, ok := isupport.state.Prefixes[ch]; ok {
			modes += string(mode)
			prefixes += string(ch)
		} else {
			nick = fullnick[i:]
			break
		}
	}

	return nick, modes, prefixes
}

// HighestPrefix gets the highest-level prefix declared by PREFIX
func (isupport *ISupport) HighestPrefix(prefixes string) rune {
	isupport.lock.RLock()
	defer isupport.lock.RUnlock()

	if len(prefixes) == 1 {
		return rune(prefixes[0])
	}

	for _, prefix := range isupport.state.PrefixOrder {
		if strings.ContainsRune(prefixes, prefix) {
			return prefix
		}
	}

	return rune(0)
}

// HighestMode gets the highest-level mode declared by PREFIX
func (isupport *ISupport) HighestMode(modes string) rune {
	isupport.lock.RLock()
	defer isupport.lock.RUnlock()

	if len(modes) == 1 {
		return rune(modes[0])
	}

	for _, mode := range isupport.state.ModeOrder {
		if strings.ContainsRune(modes, mode) {
			return mode
		}
	}

	return rune(0)
}

// IsModeHigher returns true if `current` is a higher mode than `other`.
func (isupport *ISupport) IsModeHigher(current rune, other rune) bool {
	isupport.lock.RLock()
	defer isupport.lock.RUnlock()

	if current == other {
		return false
	}
	if current == 0 {
		return false
	}
	if other == 0 {
		return true
	}

	for _, mode := range isupport.state.ModeOrder {
		if mode == current {
			return true
		} else if mode == other {
			return false
		}
	}

	return false
}

// SortModes returns the modes in order. Any unknown modes will be omitted.
func (isupport *ISupport) SortModes(modes string) string {
	result := ""

	for _, ch := range isupport.state.ModeOrder {
		for _, ch2 := range modes {
			if ch2 == ch {
				result += string(ch)
			}
		}
	}

	return result
}

// SortPrefixes returns the prefixes in order. Any unknown prefixes will be omitted.
func (isupport *ISupport) SortPrefixes(prefixes string) string {
	result := ""

	for _, ch := range isupport.state.PrefixOrder {
		for _, ch2 := range prefixes {
			if ch2 == ch {
				result += string(ch)
			}
		}
	}

	return result
}

// Mode gets the mode for the prefix.
func (isupport *ISupport) Mode(prefix rune) rune {
	isupport.lock.RLock()
	defer isupport.lock.RUnlock()

	return isupport.state.Prefixes[prefix]
}

// Prefix gets the prefix for the mode. It's a bit slower
// than the other way around, but is a far less frequently
// used.
func (isupport *ISupport) Prefix(mode rune) rune {
	isupport.lock.RLock()
	defer isupport.lock.RUnlock()

	for prefix, mappedMode := range isupport.state.Prefixes {
		if mappedMode == mode {
			return prefix
		}
	}

	return rune(0)
}

// Prefixes gets the prefixes in the order of the modes, skipping any
// invalid modes.
func (isupport *ISupport) Prefixes(modes string) string {
	result := ""

	for _, mode := range modes {
		prefix := isupport.Prefix(mode)
		if prefix != mode {
			result += string(prefix)
		}
	}

	return result
}

// IsChannel returns whether the target name is a channel.
func (isupport *ISupport) IsChannel(targetName string) bool {
	isupport.lock.RLock()
	defer isupport.lock.RUnlock()

	return strings.Contains(isupport.state.Raw["CHANTYPES"], string(targetName[0]))
}

// IsPermissionMode returns whether the flag is a permission mode
func (isupport *ISupport) IsPermissionMode(flag rune) bool {
	isupport.lock.RLock()
	defer isupport.lock.RUnlock()

	return strings.ContainsRune(isupport.state.ModeOrder, flag)
}

// ModeTakesArgument returns true if the mode takes an argument
func (isupport *ISupport) ModeTakesArgument(flag rune, plus bool) bool {
	isupport.lock.RLock()
	defer isupport.lock.RUnlock()

	// Permission modes always take an argument.
	if strings.ContainsRune(isupport.state.ModeOrder, flag) {
		return true
	}

	// Modes in category A and B always takes an argument
	if strings.ContainsRune(isupport.state.ChannelModes[0], flag) || strings.ContainsRune(isupport.state.ChannelModes[1], flag) {
		return true
	}

	// Modes in category C only takes one when added
	if plus && strings.ContainsRune(isupport.state.ChannelModes[1], flag) {
		return true
	}

	// Modes in category D and outside never does
	return false
}

// ChannelModeType returns a number from 0 to 3 based on what block of mode
// in the CHANMODES variable it fits into. If it's not found at all, it will
// return -1
func (isupport *ISupport) ChannelModeType(mode rune) int {
	isupport.lock.RLock()
	defer isupport.lock.RUnlock()

	// User permission modes function exactly like the first block
	// when it comes to add/remove
	if strings.ContainsRune(isupport.state.ModeOrder, mode) {
		return 0
	}

	for i, block := range isupport.state.ChannelModes {
		if strings.ContainsRune(block, mode) {
			return i
		}
	}

	return -1
}

// Set sets an isupport key, and related structs. This should only be used
// if a 005 packet contains the Key-Value pair or if it can be "polyfilled"
// in some other way.
func (isupport *ISupport) Set(key, value string) {
	key = strings.ToUpper(key)

	isupport.lock.Lock()

	if isupport.state.Raw == nil {
		isupport.state.Raw = make(map[string]string, 32)
	}

	isupport.state.Raw[key] = value

	switch key {
	case "PREFIX": // PREFIX=(ov)@+
		{
			split := strings.SplitN(value[1:], ")", 2)

			isupport.state.PrefixOrder = split[1]
			isupport.state.ModeOrder = split[0]
			isupport.state.Prefixes = make(map[rune]rune, len(split[0]))
			for i, ch := range split[0] {
				isupport.state.Prefixes[rune(split[1][i])] = ch
			}
		}
	case "CHANMODES": // CHANMODES=eIbq,k,flj,CFLNPQcgimnprstz
		{
			isupport.state.ChannelModes = strings.Split(value, ",")
		}
	}

	isupport.lock.Unlock()
}

// State gets a copy of the isupport state.
func (isupport *ISupport) State() *State {
	return isupport.state.Copy()
}

// Reset clears everything.
func (isupport *ISupport) Reset() {
	isupport.lock.Lock()
	isupport.state.PrefixOrder = ""
	isupport.state.ModeOrder = ""
	isupport.state.Prefixes = nil
	isupport.state.ChannelModes = nil

	for key := range isupport.state.Raw {
		delete(isupport.state.Raw, key)
	}
	isupport.lock.Unlock()
}
