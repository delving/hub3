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
