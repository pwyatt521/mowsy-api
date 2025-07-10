package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidEmail(t *testing.T) {
	testCases := []struct {
		name     string
		email    string
		expected bool
	}{
		{"Valid simple email", "test@example.com", true},
		{"Valid email with subdomain", "user@mail.example.com", true},
		{"Valid email with numbers", "user123@example.com", true},
		{"Valid email with plus", "user+tag@example.com", true},
		{"Valid email with dots", "first.last@example.com", true},
		{"Valid email with dashes", "user-name@example.com", true},
		{"Valid email with underscores", "user_name@example.com", true},
		{"Invalid email missing @", "userexample.com", false},
		{"Invalid email missing domain", "user@", false},
		{"Invalid email missing user", "@example.com", false},
		{"Invalid email missing TLD", "user@example", false},
		{"Invalid email with spaces", "user @example.com", false},
		{"Invalid email with double @", "user@@example.com", false},
		{"Empty email", "", false},
		{"Invalid email with invalid chars", "user<>@example.com", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsValidEmail(tc.email)
			assert.Equal(t, tc.expected, result, "Email: %s", tc.email)
		})
	}
}

func TestIsValidPhone(t *testing.T) {
	testCases := []struct {
		name     string
		phone    string
		expected bool
	}{
		{"Valid US phone basic", "555-123-4567", true},
		{"Valid US phone with parentheses", "(555) 123-4567", true},
		{"Valid US phone with dots", "555.123.4567", true},
		{"Valid US phone no separators", "5551234567", true},
		{"Valid US phone with +1", "+1-555-123-4567", true},
		{"Valid US phone with 1", "1-555-123-4567", true},
		{"Valid US phone with spaces", "555 123 4567", true},
		{"Invalid phone too short", "555-123", false},
		{"Invalid phone too long", "555-123-45678", false},
		{"Invalid phone with letters", "555-ABC-4567", false},
		{"Empty phone", "", false},
		{"Invalid phone format", "555-123-456", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsValidPhone(tc.phone)
			assert.Equal(t, tc.expected, result, "Phone: %s", tc.phone)
		})
	}
}

func TestIsValidPassword(t *testing.T) {
	testCases := []struct {
		name     string
		password string
		expected bool
	}{
		{"Valid 8 character password", "password", true},
		{"Valid long password", "verylongpasswordwithmanycharacters", true},
		{"Valid password with special chars", "Pass@123", true},
		{"Invalid short password", "pass", false},
		{"Invalid 7 character password", "passwor", false},
		{"Empty password", "", false},
		{"Exactly 8 characters", "12345678", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsValidPassword(tc.password)
			assert.Equal(t, tc.expected, result, "Password length: %d", len(tc.password))
		})
	}
}

func TestIsValidZipCode(t *testing.T) {
	testCases := []struct {
		name     string
		zipCode  string
		expected bool
	}{
		{"Valid 5 digit zip", "12345", true},
		{"Valid zip with zeros", "00123", true},
		{"Invalid 4 digit zip", "1234", false},
		{"Invalid 6 digit zip", "123456", false},
		{"Invalid zip with letters", "1234A", false},
		{"Invalid zip with spaces", "12 345", false},
		{"Invalid zip with dashes", "12-345", false},
		{"Empty zip code", "", false},
		{"Invalid zip with special chars", "123@5", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsValidZipCode(tc.zipCode)
			assert.Equal(t, tc.expected, result, "Zip code: %s", tc.zipCode)
		})
	}
}

func TestSanitizeString(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"String with leading spaces", "  hello", "hello"},
		{"String with trailing spaces", "hello  ", "hello"},
		{"String with both leading and trailing spaces", "  hello  ", "hello"},
		{"String with tabs", "\thello\t", "hello"},
		{"String with newlines", "\nhello\n", "hello"},
		{"String with mixed whitespace", " \t\nhello \t\n", "hello"},
		{"Empty string", "", ""},
		{"String with only spaces", "   ", ""},
		{"Normal string", "hello", "hello"},
		{"String with internal spaces", "hello world", "hello world"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := SanitizeString(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsValidRating(t *testing.T) {
	testCases := []struct {
		name     string
		rating   int
		expected bool
	}{
		{"Valid rating 1", 1, true},
		{"Valid rating 2", 2, true},
		{"Valid rating 3", 3, true},
		{"Valid rating 4", 4, true},
		{"Valid rating 5", 5, true},
		{"Invalid rating 0", 0, false},
		{"Invalid rating 6", 6, false},
		{"Invalid negative rating", -1, false},
		{"Invalid large rating", 10, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsValidRating(tc.rating)
			assert.Equal(t, tc.expected, result, "Rating: %d", tc.rating)
		})
	}
}

func TestIsNumeric(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Valid numeric string", "12345", true},
		{"Valid single digit", "5", true},
		{"Valid with leading zero", "01234", true},
		{"Invalid with letters", "123a5", false},
		{"Invalid with spaces", "123 45", false},
		{"Invalid with special chars", "123@5", false},
		{"Empty string", "", true}, // Empty string is considered numeric in this implementation
		{"Invalid with decimal", "123.45", false},
		{"Invalid with negative", "-123", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isNumeric(tc.input)
			assert.Equal(t, tc.expected, result, "Input: %s", tc.input)
		})
	}
}