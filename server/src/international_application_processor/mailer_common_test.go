package international_application_processor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeAndFormatName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"John Doe", "John Doe"},
		{"Jóhn Doe", "Jóhn Doe"},
		{" John Doe ", "John Doe"},
		{"john <doe>\t", "John Doe"},
		{"!@#John123", "John123"},
		{"jane doe the 3rd", "Jane Doe The 3Rd"},
		{"", ""},
	}

	for _, test := range tests {
		result := sanitizeAndFormatName(test.input)
		assert.Equal(t, test.expected, result, "Expected sanitized and formatted name to match")
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		input    string
		hasError bool
	}{
		{"test@example.com", false},
		{"user@domain.co.uk", false},
		{"user@مثال.com", false},
		{"user@tld", false},
		{"user@", true},
		{"invalid-email", true},
		{"", true},
	}

	for _, test := range tests {
		err := validateEmail(test.input)
		if test.hasError {
			assert.Error(t, err, "Expected an error for invalid email")
		} else {
			assert.NoError(t, err, "Expected no error for valid email")
		}
	}
}

func TestSanitizeState(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"CA", "CA"},
		{"California123", "California"},
		{"!@#CA", "CA"},
		{"", ""},
	}

	for _, test := range tests {
		result := sanitizeState(test.input)
		assert.Equal(t, test.expected, result, "Expected sanitized state to match")
	}
}
