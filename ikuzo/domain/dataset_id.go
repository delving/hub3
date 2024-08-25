package domain

import (
	"fmt"
	"unicode"
)

// DatasetID represents a short identifier for an Dataset.
//
// The maximum length is MaxLengthID. The minimum length is MinLengthID
//
// In JSON the DatasetID is represented as 'datasetID'.
type DatasetID struct {
	OrgID OrganizationID `json:"orgID,omitempty"`
	ID    string         `json:"datasetID,omitempty"`
}

// NewDatasetID returns an DatasetID and an error if the supplied input is invalid.
func NewDatasetID(orgID, datasetID string) (DatasetID, error) {
	oi, err := NewOrganizationID(orgID)
	if err != nil {
		return DatasetID{}, err
	}

	ds := DatasetID{oi, datasetID}
	if err := ds.Valid(); err != nil {
		return DatasetID{}, err
	}

	return ds, nil
}

// Valid validates the identifier.
//
// - ErrIDCannotBeEmpty is returned when the id is empty
//
// - ErrIDTooLong is returned when ID is too long
//
// - ErrIDTooShort is returned when ID is too short
//
// - ErrIDInvalidCharacter is returned when ID contains non-letters, excluding '-'
func (id *DatasetID) Valid() error {
	if id.ID == "" {
		return ErrIDCannotBeEmpty
	}

	if len(id.ID) > MaxLengthDatasetID {
		return fmt.Errorf("maximum length = %d; %w", MaxLengthDatasetID, ErrIDTooLong)
	}

	if len(id.ID) < MinLengthID {
		return fmt.Errorf("minimu length = %d; %w", MinLengthID, ErrIDTooShort)
	}

	for _, p := range protectedDatasetIDs {
		if id.ID == p {
			return ErrIDExists
		}
	}

	for _, r := range id.ID {
		// allow letters and numbers
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '-' {
			continue
		}

		return ErrIDInvalidCharacter
	}

	return nil
}

func (id *DatasetID) String() string {
	return fmt.Sprintf("%s_%s", id.OrgID.String(), id.ID)
}
