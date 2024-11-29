package pid

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestNewStoreResponse(t *testing.T) {
	tests := []struct {
		name     string
		pids     []*PID
		expected StoreResponse
	}{
		{
			name:     "empty pids",
			pids:     []*PID{},
			expected: StoreResponse{},
		},
		{
			name: "single pid",
			pids: []*PID{
				{ID: "urn:1", ExternalID: "ark:123/abc"},
			},
			expected: StoreResponse{
				CurrentPID: &PID{ID: "urn:1", ExternalID: "ark:123/abc"},
			},
		},
		{
			name: "multiple pids",
			pids: []*PID{
				{ID: "urn:1", ExternalID: "ark:123/abc", ReplacedBy: "urn:2"},
				{ID: "urn:2", ExternalID: "ark:123/abc"},
			},
			expected: StoreResponse{
				CurrentPID: &PID{ID: "urn:2", ExternalID: "ark:123/abc"},
				OtherIDs: []*PID{
					{ID: "urn:1", ExternalID: "ark:123/abc", ReplacedBy: "urn:2"},
				},
			},
		},
		{
			name: "multiple pids without replacedBy",
			pids: []*PID{
				{ID: "urn:1", ExternalID: "ark:123/abc", ModifiedAt: time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC)},
				{ID: "urn:2", ExternalID: "ark:123/abc", ModifiedAt: time.Date(2023, time.January, 2, 0, 0, 0, 0, time.UTC)},
			},
			expected: StoreResponse{
				CurrentPID: &PID{
					ID: "urn:2", ExternalID: "ark:123/abc", ModifiedAt: time.Date(2023, time.January, 2, 0, 0, 0, 0, time.UTC),
				},
				OtherIDs: []*PID{
					{ID: "urn:1", ExternalID: "ark:123/abc", ModifiedAt: time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC)},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewStoreResponse(tt.pids)
			if diff := cmp.Diff(tt.expected, got); diff != "" {
				t.Errorf("NewStoreResponse() mismatch (-expected +got):\n%s", diff)
			}
		})
	}
}
