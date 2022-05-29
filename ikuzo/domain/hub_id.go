package domain

import (
	"fmt"
	"strings"
)

type HubID struct {
	OrgID     string
	DatasetID string
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

	return HubID{
		OrgID:     parts[0],
		DatasetID: parts[1],
		LocalID:   parts[2],
	}, nil
}

func (h HubID) String() string {
	return fmt.Sprintf("%s_%s_%s", h.OrgID, h.DatasetID, h.LocalID)
}
