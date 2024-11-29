package pid

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/matryer/is"
)

func TestInferType(t *testing.T) {
	is := is.New(t)
	tests := []struct {
		name       string
		externalID string
		expected   Type
	}{
		{"ARK", "ark:/12345/x", Ark},
		{"DOI", "10.1234/abcd", DOI},
		{"Handle", "hdl:12345/abc", Handle},
		{"Undefined", "unknown", Undefined},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PID{ExternalID: tt.externalID}
			p.inferType()
			is.Equal(tt.expected, p.Type)
		})
	}
}

func TestIsDOI(t *testing.T) {
	is := is.New(t)
	is.True(isDOI("10.1234/abcd"))
	is.True(isDOI("doi:10.1234/abcd"))
	is.True(!isDOI("not-a-doi"))
}

func TestIsHandle(t *testing.T) {
	is := is.New(t)
	is.True(isHandle("hdl:12345/abc"))
	is.True(!isHandle("not-a-handle"))
}

func TestIsDifferent(t *testing.T) {
	tests := []struct {
		name     string
		pid1     PID
		pid2     PID
		expected bool
	}{
		{
			name:     "same PIDs",
			pid1:     PID{ID: "1", ExternalID: "subject1"},
			pid2:     PID{ID: "1", ExternalID: "subject1"},
			expected: false,
		},
		{
			name:     "different IDs",
			pid1:     PID{ID: "1", ExternalID: "subject1"},
			pid2:     PID{ID: "2", ExternalID: "subject1"},
			expected: true,
		},
		{
			name:     "different ExternalIDs",
			pid1:     PID{ID: "1", ExternalID: "subject1"},
			pid2:     PID{ID: "1", ExternalID: "subject2"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &tt.pid1
			got := p.IsDifferent(&tt.pid2)
			if diff := cmp.Diff(tt.expected, got); diff != "" {
				t.Errorf("IsDifferent() mismatch (-expected +got):\n%s", diff)
			}
		})
	}
}
