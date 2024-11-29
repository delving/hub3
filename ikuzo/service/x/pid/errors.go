package pid

import "errors"

var (
	ErrTombstone      = errors.New("pid is a tombstone")
	ErrPIDNotFound    = errors.New("pid is not found")
	ErrCannotStorePID = errors.New(`pid cannot be stored`)
)
