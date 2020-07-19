// Copyright 2020 Delving B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package domain

import (
	"errors"
	"fmt"
	"unicode"
)

// errors
var (
	ErrIDTooLong          = errors.New("identifier is too long")
	ErrIDNotLowercase     = errors.New("uppercase not allowed in identifier")
	ErrIDInvalidCharacter = errors.New("only letters are allowed in organization")
	ErrIDCannotBeEmpty    = errors.New("empty string is not a valid identifier")
	ErrIDExists           = errors.New("identifier already exists")
	ErrOrgNotFound        = errors.New("organization not found")
)

var (
	// MaxLengthID the maximum length of an identifier
	MaxLengthID = 10

	// protected organization names
	protected = []OrganizationID{
		OrganizationID("public"),
		OrganizationID("all"),
	}
)

type OrganizationFilter struct {
	// OffSet is the start of the results returned
	OffSet int
	// Limit is the number of items returned from the filter
	Limit int
	// Org can be used to filter the results based on the filled in value.
	// This is mostly usefull if you want to filter by attributes..
	Org Organization
}

// OrganizationID represents a short identifier for an Organization.
//
// The maximum length is MaxLengthID.
//
// In JSON the OrganizationID is represented as 'orgID'.
type OrganizationID string

// Organization is a basic building block for storing information.
// Everything that is stored by ikuzo must have an organization.ID as part of its metadata.
type Organization struct {
	ID          OrganizationID `json:"orgID"`
	Description string         `json:"description,omitempty"`
	// TODO(kiivihal): add organization config
	Config OrganizationConfig
	// create struct so that it can be initialized by config
}

type OrganizationConfig struct {
	// TODO(kiivihal): change to sub-struct later
	RDFBaseURL     string
	MintDatasetURL string
	MintOrgIDURL   string
}

// NewOrganizationID returns an OrganizationID and an error if the supplied input is invalid.
func NewOrganizationID(input string) (OrganizationID, error) {
	id := OrganizationID(input)
	if err := id.Valid(); err != nil {
		return OrganizationID(""), err
	}

	return id, nil
}

// Valid validates the identifier.
//
// - ErrIDTooLong is returned when ID is too long
//
// - ErrIDNotLowercase is returned when ID contains uppercase characters
//
// - ErrIDInvalidCharacter is returned when ID contains non-letters
//
func (id OrganizationID) Valid() error {
	if id == "" {
		return ErrIDCannotBeEmpty
	}

	if len(id) > MaxLengthID {
		return ErrIDTooLong
	}

	for _, p := range protected {
		if id == p {
			return ErrIDExists
		}
	}

	for _, r := range id {
		if !unicode.IsLower(r) && unicode.IsLetter(r) {
			return ErrIDNotLowercase
		}

		if !unicode.IsLetter(r) {
			return ErrIDInvalidCharacter
		}
	}

	return nil
}

// RawID returns the raw direct identifier string for an Organization
func (o *Organization) RawID() string {
	return string(o.ID)
}

func (o *Organization) NewDatasetURI(spec string) string {
	return fmt.Sprintf(o.Config.MintDatasetURL, o.Config.RDFBaseURL, spec)
}
