// Package validator provides utilities to provide a uniform valditation experience.
//
// This validator is adapted from Alex Edwards' "let's go further" book, see
// https://lets-go-further.alexedwards.net
package validator

import "regexp"

type Validator struct {
	Errors map[string]string
}

// New returns a new empty Validator
func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// Valid return a boolean if errors have been encountered
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// AddError adds an error to the validator
//
// Per key only one error message can be added.
// Additional errors will silently be ignored.
func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

// Check evaluates ok and adds messages if not true.
//
// This can be used to build easy to read list of checks.
//
//		v.Check(f.Page > 0, "page", "must be greater than zero")
//		v.Check(f.Page <= 10_000_000, "page", "must be a maximum of 10 million")
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
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
