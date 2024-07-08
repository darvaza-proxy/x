package reconnect

import "errors"

var (
	// ErrDoNotReconnect indicates the Waiter
	// instructed us to not reconnect
	ErrDoNotReconnect = errors.New("don't reconnect")
)
