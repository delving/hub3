package domain

import (
	"fmt"
	"strings"
)

type HubID struct {
	DatasetID DatasetID
	LocalID   string
}

func NewHubID(input string) (HubID, error) {
	parts := strings.SplitN(input, "_", 3)
	if len(parts) != 3 {
		return HubID{}, ErrHubIDInvalid
	}

	for _, v := range parts {
		if v == "" {
			return HubID{}, ErrHubIDInvalid
		}
	}

	datasetID, err := NewDatasetID(parts[0], parts[1])
	if err != nil {
		return HubID{}, err
	}

	return HubID{
		DatasetID: datasetID,
		LocalID:   parts[2],
	}, nil
}

func (h HubID) OrgID() OrganizationID {
	return h.DatasetID.OrgID
}

func (h HubID) String() string {
	return fmt.Sprintf("%s_%s", h.DatasetID.String(), h.LocalID)
}
