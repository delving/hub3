package namespace

// Namespace errors.
const (
	ErrNameSpaceNotFound       = Error("namespace not found")
	ErrNameSpaceDuplicateEntry = Error("prefix and base stored in different entries")
)

// Error represents a Namespace error.
type Error string

// Error returns the error message.
func (e Error) Error() string { return string(e) }
