// Package validator provides utilities to provide a uniform valditation experience.
//
// This validator is adapted from Alex Edwards' "let's go further" book, see
// https://lets-go-further.alexedwards.net
package validator

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/go-multierror"
)

type Validator struct {
	Errors           map[string]error
	PrefixKeyInError bool
}

// New returns a new empty Validator
func New() *Validator {
	return &Validator{Errors: make(map[string]error)}
}

// Valid return a boolean if errors have been encountered
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// AddError adds an error to the validator
//
// Per key only one error message can be added.
// Additional errors will silently be ignored.
//
// When both err and message are given they are wrapped in a new Error.
//
// Either err or message must given a non-empty value, otherwise the error is
// silently ignored.
func (v *Validator) AddError(key string, err error, message string) {
	if _, exists := v.Errors[key]; !exists {
		if err == nil && message == "" {
			return
		}

		errMsg := message
		if v.PrefixKeyInError {
			errMsg = fmt.Sprintf("%s; %s", key, message)
		}

		if err == nil && message != "" {
			err = fmt.Errorf(errMsg)
		} else if err != nil && message != "" {
			err = fmt.Errorf("%w: %s", err, errMsg)
		}

		v.Errors[key] = err
	}
}

// Check evaluates ok and adds messages if not true.
//
// This can be used to build easy to read list of checks.
//
//		v.Check(f.Page > 0, "page", errors.New("invalid param"), "must be greater than zero")
//		v.Check(f.Page <= 10_000_000, "page", nil, "must be a maximum of 10 million")
func (v *Validator) Check(ok bool, key string, err error, message string) {
	if !ok {
		v.AddError(key, err, message)
	}
}

// ErrorOrNil wraps all the errors in a single error or returns nil
func (v *Validator) ErrorOrNil() error {
	var result error
	for _, err := range v.Errors {
		result = multierror.Append(result, err)
	}

	return result
}

// In checks if value is part of list.
func In(value string, list ...string) bool {
	for i := range list {
		if value == list[i] {
			return true
		}
	}

	return false
}

// Matches checks value against the supplied rx
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// Unique checks if all values are unique
//
// When values is an empty list it returns true.
func Unique(values []string) bool {
	uniqueValues := make(map[string]bool)

	for _, value := range values {
		uniqueValues[value] = true
	}

	return len(values) == len(uniqueValues)
}
