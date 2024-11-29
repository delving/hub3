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

func TestExtractRelativePath(t *testing.T) {
	is := is.New(t)

	tests := []struct {
		name     string
		uri      string
		expected string
		wantErr  bool
	}{
		{
			name:     "ARK URI",
			uri:      "https://n2t.net/ark:/63960/006b90dd-43a9-60bc-0e79-6b38c4d094ea",
			expected: "ark:/63960/006b90dd-43a9-60bc-0e79-6b38c4d094ea",
			wantErr:  false,
		},
		{
			name:     "Simple URI",
			uri:      "https://example.com/path/to/resource",
			expected: "path/to/resource",
			wantErr:  false,
		},
		{
			name:     "URI with no path",
			uri:      "https://example.com",
			expected: "",
			wantErr:  false,
		},
		{
			name:     "URI with query parameters",
			uri:      "https://example.com/path?query=value",
			expected: "path",
			wantErr:  false,
		},
		{
			name:     "URI with fragment",
			uri:      "https://example.com/path#fragment",
			expected: "path",
			wantErr:  false,
		},
		{
			name:     "Invalid URI",
			uri:      "://invalid",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := is.New(t)

			result, err := extractRelativePath(tt.uri)

			if tt.wantErr {
				is.True(err != nil) // Should have an error
			} else {
				is.NoErr(err)                 // Should not have an error
				is.Equal(tt.expected, result) // Result should match expected
			}
		})
	}
}
