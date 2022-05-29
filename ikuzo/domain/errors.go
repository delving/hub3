package domain

import "errors"

// errors
var (
	ErrIDTooLong          = errors.New("identifier is too long")
	ErrIDNotLowercase     = errors.New("uppercase not allowed in identifier")
	ErrIDInvalidCharacter = errors.New("only letters and numbers are allowed in organization")
	ErrIDCannotBeEmpty    = errors.New("empty string is not a valid identifier")
	ErrIDExists           = errors.New("identifier already exists")
	ErrOrgNotFound        = errors.New("organization not found")
	ErrHubIDInvalid       = errors.New("HubID strings is invalid")
)
