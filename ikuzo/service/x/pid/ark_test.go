package pid

import (
	"testing"
)

func TestIsUUIDWithoutHyphens(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"12345678123456781234567812345678", true},
		{"12345678-1234-5678-1234-567812345678", false},
		{"1234567812345678123456781234567g", false},
		{"1234567812345678123456781234567", false},
	}

	for _, test := range tests {
		result := isUUIDWithoutHyphens(test.input)
		if result != test.expected {
			t.Errorf("isUUIDWithoutHyphens(%s) = %v; expected %v", test.input, result, test.expected)
		}
	}
}

func TestAddHyphensToUUID(t *testing.T) {
	input := "12345678123456781234567812345678"
	expected := "12345678-1234-5678-1234-567812345678"
	result := addHyphensToUUID(input)
	if result != expected {
		t.Errorf("addHyphensToUUID(%s) = %s; expected %s", input, result, expected)
	}
}

func TestValidateARK(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/ark:/12345/12345678123456781234567812345678", "ark:/12345/12345678-1234-5678-1234-567812345678"},
		{"/ark:/12345/12345678-1234-5678-1234-567812345678", "ark:/12345/12345678-1234-5678-1234-567812345678"},
		{"/ark:/12345/invaliduuid", "ark:/12345/invaliduuid"},
	}

	for _, test := range tests {
		result := validateARK(test.input)
		if result != test.expected {
			t.Errorf("validateARK(%s) = %s; expected %s", test.input, result, test.expected)
		}
	}
}

func TestNormalizeARK(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "N2T format with slash - remove slash",
			input:    "ark:/12345/67890",
			expected: "ark:12345/67890",
		},
		{
			name:     "Modern format without slash - add slash",
			input:    "ark:12345/67890",
			expected: "ark:/12345/67890",
		},
		{
			name:     "With UUID",
			input:    "ark:/63960/006b90dd-43a9-60bc-0e79-6b38c4d094ea",
			expected: "ark:63960/006b90dd-43a9-60bc-0e79-6b38c4d094ea",
		},
		{
			name:     "Modern format with UUID",
			input:    "ark:63960/006b90dd-43a9-60bc-0e79-6b38c4d094ea",
			expected: "ark:/63960/006b90dd-43a9-60bc-0e79-6b38c4d094ea",
		},
		{
			name:     "Not an ARK - no change",
			input:    "doi:10.1234/5678",
			expected: "doi:10.1234/5678",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := normalizeARK(test.input)
			if result != test.expected {
				t.Errorf("normalizeARK(%s) = %s; expected %s", test.input, result, test.expected)
			}
		})
	}
}

func TestNormalizeARKRoundTrip(t *testing.T) {
	// Test that normalizing twice returns the original
	tests := []string{
		"ark:/12345/67890",
		"ark:12345/67890",
		"ark:/63960/006b90dd-43a9-60bc-0e79-6b38c4d094ea",
		"ark:63960/006b90dd-43a9-60bc-0e79-6b38c4d094ea",
	}

	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			normalized := normalizeARK(test)
			backAgain := normalizeARK(normalized)
			if backAgain != test {
				t.Errorf("normalizeARK round trip failed: %s -> %s -> %s", test, normalized, backAgain)
			}
		})
	}
}
