package irc

// A Status contains
type Status struct {
}

// Kind returns "status"
func (status *Status) Kind() string {
	return "status"
}

// Name returns "status"
func (status *Status) Name() string {
	return "Status"
}

// Handle handles messages routed to this status by the client's event loop
func (status *Status) Handle(event *Event, client *Client) {

}
