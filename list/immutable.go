package list

// An Immutable is a wrapper around a userlist reference that provides a limited
// set of methods for reading a userlist's content
type Immutable struct {
	list *List
}

// User gets a user by nick
func (il Immutable) User(nick string) (u User, ok bool) {
	return il.list.User(nick)
}

// Users gets all the users in the list, in order
func (il Immutable) Users() []User {
	return il.list.Users()
}
